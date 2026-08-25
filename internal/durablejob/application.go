package durablejob

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	sobservability "github.com/HiIamJeff67/notegic-backend/shared/platform/observability"
	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	durablejobconfig "github.com/HiIamJeff67/notegic-backend/internal/durablejob/configs"
	data "github.com/HiIamJeff67/notegic-backend/internal/durablejob/data/postgres"
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
	loadConfig() durablejobconfig.Config
	loadKafkaConnectionConfig() skafka.ConnectionConfig
	initializeObservability() func()
	initializeKafka(skafka.ConnectionConfig, func()) *skafka.Producer
	initializeWorkers(durablejobconfig.Config, skafka.ConnectionConfig, *skafka.Producer) func()
	buildRouter() *http.ServeMux
	startHTTP(durablejobconfig.Config, *http.ServeMux, func(), *skafka.Producer, func()) func()
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

func (a *Application) loadConfig() durablejobconfig.Config {
	config, err := durablejobconfig.LoadConfig()
	if err != nil {
		panic(err)
	}
	return config
}

func (a *Application) loadKafkaConnectionConfig() skafka.ConnectionConfig {
	kafkaConnectionConfig, err := skafka.LoadConnectionConfig()
	if err != nil {
		panic(err)
	}
	return kafkaConnectionConfig
}

func (a *Application) initializeObservability() func() {
	return sobservability.Initialize(
		context.Background(),
		sobservability.LoadConfig("notegic-durable-job"),
	)
}

func (a *Application) initializeKafka(
	config skafka.ConnectionConfig,
	shutdownObservability func(),
) *skafka.Producer {
	kafkaProducer, err := skafka.NewProducer(skafka.ClientConfig{
		ConnectionConfig: config,
		ClientId:         "notegic-durable-job",
	})
	if err != nil {
		shutdownObservability()
		panic(err)
	}
	if err := kafkaProducer.Ping(context.Background()); err != nil {
		kafkaProducer.Close()
		shutdownObservability()
		panic(err)
	}
	return kafkaProducer
}

func (a *Application) initializeWorkers(
	config durablejobconfig.Config,
	kafkaConnection skafka.ConnectionConfig,
	kafkaProducer *skafka.Producer,
) func() {
	db, err := data.Connect(config.Postgres)
	if err != nil {
		panic(err)
	}
	if err := spostgres.Migrate(
		db,
		ctypes.Runtime_DurableJob,
		data.DatabaseMigrationManifest,
	); err != nil {
		_ = data.Disconnect(db)
		panic(fmt.Errorf("failed to initialize DurableJob database schema: %w", err))
	}
	data.SetDefaultDB(db)

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
		_ = data.Disconnect(db)
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
	shutdownWorkers func(),
	kafkaProducer *skafka.Producer,
	shutdownObservability func(),
) func() {
	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		shutdownWorkers()
		kafkaProducer.Close()
		shutdownObservability()
		panic(err)
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
		// Stop HTTP traffic before stopping workers and Kafka.
		a.ready.Store(false)
		a.healthy.Store(false)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			fmt.Println("Failed to shutdown DurableJob server: ", err)
		}
		shutdownWorkers()
		kafkaProducer.Close()
		shutdownObservability()
	}
}

func (a *Application) Start() func() {
	shutdownObservability := a.initializeObservability()
	config := a.loadConfig()
	kafkaConnection := a.loadKafkaConnectionConfig()
	kafkaProducer := a.initializeKafka(kafkaConnection, shutdownObservability)
	shutdownWorkers := a.initializeWorkers(config, kafkaConnection, kafkaProducer)
	router := a.buildRouter()
	return a.startHTTP(config, router, shutdownWorkers, kafkaProducer, shutdownObservability)
}

// make sure Application struct followed the ApplicationInterface implementations
var _ ApplicationInterface = (*Application)(nil)
