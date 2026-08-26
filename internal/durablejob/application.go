package durablejob

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	sobservability "github.com/HiIamJeff67/notegic-backend/shared/platform/observability"
	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	"gorm.io/gorm"

	durablejobconfig "github.com/HiIamJeff67/notegic-backend/internal/durablejob/configs"
	routinetask "github.com/HiIamJeff67/notegic-backend/internal/durablejob/routinetask"
	routineexecution "github.com/HiIamJeff67/notegic-backend/internal/durablejob/routinetask/execution"
	realtimegatewayproducers "github.com/HiIamJeff67/notegic-backend/internal/durablejob/transports/realtimegateway/producers"
	status "github.com/HiIamJeff67/notegic-backend/internal/durablejob/transports/status"
	validation "github.com/HiIamJeff67/notegic-backend/internal/durablejob/validations"
)

type Application struct {
	healthy           atomic.Bool
	ready             atomic.Bool
	routineTaskEngine *routinetask.Engine
}

type ApplicationInterface interface {
	initializeObservability() func()
	initializeDatabase(durablejobconfig.Config) (*gorm.DB, error)
	initializeKafka(skafka.ConnectionConfig) (*skafka.Producer, error)
	initializeWorkers(durablejobconfig.Config, *gorm.DB, skafka.ConnectionConfig, *skafka.Producer) func()
	buildRouter() *http.ServeMux
	startHTTP(durablejobconfig.Config, *http.ServeMux) (func(), error)
	Start() func()
	IsHealthy() bool
	IsReady() bool
}

func NewApplication() *Application {
	return &Application{}
}

func (a *Application) IsHealthy() bool {
	return a.healthy.Load()
}

func (a *Application) IsReady() bool {
	return a.ready.Load() && a.routineTaskEngine != nil && a.routineTaskEngine.IsReady()
}

func (a *Application) initializeObservability() func() {
	return sobservability.Initialize(
		context.Background(),
		sobservability.LoadConfig("notegic-durable-job"),
	)
}

func (a *Application) initializeKafka(
	config skafka.ConnectionConfig,
) (*skafka.Producer, error) {
	kafkaProducer, err := skafka.NewProducer(skafka.ClientConfig{
		ConnectionConfig: config,
		ClientId:         "notegic-durable-job",
	})
	if err != nil {
		return nil, err
	}
	if err := kafkaProducer.Ping(context.Background()); err != nil {
		kafkaProducer.Close()
		return nil, err
	}
	return kafkaProducer, nil
}

func (a *Application) initializeDatabase(config durablejobconfig.Config) (*gorm.DB, error) {
	if config.Postgres.User != ctypes.Runtime_DurableJob.RoleName() {
		return nil, fmt.Errorf("DurableJob PostgreSQL user must be %q", ctypes.Runtime_DurableJob.RoleName())
	}
	db, err := spostgres.Connect(config.Postgres)
	if err != nil {
		return nil, fmt.Errorf("failed to connect DurableJob database: %w", err)
	}
	return db, nil
}

func (a *Application) initializeWorkers(
	config durablejobconfig.Config,
	db *gorm.DB,
	kafkaConnection skafka.ConnectionConfig,
	kafkaProducer *skafka.Producer,
) func() {

	// Construct and start the durable-job workers that claim and execute tasks.
	routineTaskEngine := routinetask.NewEngine(config)
	a.routineTaskEngine = routineTaskEngine
	routineTaskLifecycleProducer := realtimegatewayproducers.NewRoutineTaskLifecycleProducer(kafkaProducer)
	routineTaskClaimer := routinetask.NewClaimer(db, validation.New())
	routineTaskExecutionService := routineexecution.NewRoutineTaskExecutionService(
		validation.New(),
		db,
		routinetask.NewDocumentInitializationClient(config.YjsDocumentInitialization),
	)
	routineTaskResultWriter := routinetask.NewRoutineTaskResultWriter(db, routineTaskExecutionService)
	routineTaskEngine.SetResultPublisher(routineTaskResultWriter.Write)
	routineTaskEngine.SetRoutineTaskRunningPublisher(
		routineTaskLifecycleProducer.ProduceRoutineTaskRunning,
	)
	shutdownRoutineTaskEngine := routineTaskEngine.Start(
		context.Background(),
		func(ctx context.Context, request cdurablejob.ClaimRoutineTasksRequestDto) error {
			response, exception := routineTaskClaimer.ClaimRoutineTasks(ctx, request)
			if exception != nil {
				return exception
			}
			return routineTaskEngine.HandleRoutineTaskAssignments(ctx, response.Assignments)
		},
	)

	return func() {
		shutdownRoutineTaskEngine()
	}
}

func (a *Application) buildRouter() *http.ServeMux {
	mux := http.NewServeMux()
	status.ConfigureStartedRouter(mux, a.IsHealthy)
	status.ConfigureHealthRouter(mux, a.IsReady)
	return mux
}

func (a *Application) startHTTP(
	config durablejobconfig.Config,
	mux *http.ServeMux,
) (func(), error) {
	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		return nil, err
	}
	a.healthy.Store(true)
	a.ready.Store(a.routineTaskEngine.IsReady())
	server := &http.Server{Handler: mux}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	return func() {
		a.ready.Store(false)
		a.healthy.Store(false)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			fmt.Println("Failed to shutdown DurableJob server: ", err)
		}
	}, nil
}

func (a *Application) Start() func() {
	shutdownObservability := a.initializeObservability()
	var (
		db              *gorm.DB
		kafkaProducer   *skafka.Producer
		shutdownHTTP    func()
		shutdownWorkers func()
	)
	var shutdownOnce sync.Once
	shutdown := func() {
		shutdownOnce.Do(func() {
			if shutdownHTTP != nil {
				shutdownHTTP()
			}
			if shutdownWorkers != nil {
				shutdownWorkers()
			}
			if kafkaProducer != nil {
				kafkaProducer.Close()
			}
			if db != nil {
				_ = spostgres.Disconnect(db)
			}
			shutdownObservability()
		})
	}
	fail := func(err error) {
		shutdown()
		panic(err)
	}

	config, err := durablejobconfig.LoadConfig()
	if err != nil {
		fail(err)
	}
	kafkaConnection, err := skafka.LoadConnectionConfig()
	if err != nil {
		fail(err)
	}
	db, err = a.initializeDatabase(config)
	if err != nil {
		fail(err)
	}
	kafkaProducer, err = a.initializeKafka(kafkaConnection)
	if err != nil {
		fail(err)
	}
	shutdownWorkers = a.initializeWorkers(config, db, kafkaConnection, kafkaProducer)
	router := a.buildRouter()
	shutdownHTTP, err = a.startHTTP(config, router)
	if err != nil {
		fail(err)
	}
	return shutdown
}

// make sure Application struct followed the ApplicationInterface implementations
var _ ApplicationInterface = (*Application)(nil)
