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

	"gorm.io/gorm"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	sobservability "github.com/HiIamJeff67/notegic-backend/shared/platform/observability"
	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"

	durablejobconfig "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/configs"
	durablejobexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/exceptions"
	routineexecution "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask"
	routinetaskmatchers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/matchers"
	routinetaskresolvers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/resolvers"
	routinetaskrecoverers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/recovery/recoverers"
	realtimegatewayproducers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/transports/realtimegateway/producers"
	status "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/transports/status"
	yjsworkertransport "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/transports/yjsworker"
	validation "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/validations"
	routinetaskworker "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/workers/routinetask"
)

type Application struct {
	healthy           atomic.Bool
	ready             atomic.Bool
	routineTaskEngine *routinetaskworker.Engine
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
	routineTaskClaimer := routinetaskworker.NewClaimer(db, validation.New())
	routineTaskLifecycleProducer := realtimegatewayproducers.NewRoutineTaskLifecycleProducer(kafkaProducer)
	routineTaskCompletionProducer := realtimegatewayproducers.NewRoutineTaskCompletionProducer(kafkaProducer)
	rootShelfRepository := srepositories.NewRootShelfRepository(db, scopes.NewRootShelfScope())
	stationRepository := srepositories.NewStationRepository(db, scopes.NewStationScope())
	subShelfRepository := srepositories.NewSubShelfRepository(db, scopes.NewSubShelfScope(), rootShelfRepository)
	bulkBlockPackRepository := srepositories.NewBulkBlockPackRepository(db, scopes.NewBlockPackScope())
	blockPackRepository := srepositories.NewBlockPackRepository(db, scopes.NewBlockPackScope(), bulkBlockPackRepository, subShelfRepository)
	blockRepository := srepositories.NewBlockRepository(db, scopes.NewBlockScope(), bulkBlockPackRepository)
	routineRepository := srepositories.NewRoutineRepository(db, scopes.NewRoutineScope(), stationRepository)
	materialRepository := srepositories.NewMaterialRepository(db, scopes.NewMaterialScope(), subShelfRepository)
	routineTaskRecordRepository := srepositories.NewRoutineTaskRecordRepository(db, scopes.NewRoutineTaskRecordScope())
	patternResolver := routinetaskresolvers.NewRoutineTaskPatternResolver(
		routinetaskresolvers.NewBlockPatternResolver(blockRepository),
		routinetaskresolvers.NewBlockPackPatternResolver(blockPackRepository),
	)
	templateBlockMatcher := routinetaskmatchers.NewRoutineTaskTemplateMatcher()
	routineTaskExecutionService := routineexecution.NewRoutineTaskExecutionService(
		validation.New(),
		db,
		routineTaskRecordRepository,
		patternResolver,
		templateBlockMatcher,
		subShelfRepository,
		blockPackRepository,
		blockRepository,
		routineRepository,
		materialRepository,
		yjsworkertransport.NewDocumentInitializationClient(config.YjsDocumentInitialization),
		yjsworkertransport.NewBlockPackUpdateClient(config.YjsDocumentInitialization),
	)
	routineTaskPlanService := routineexecution.NewPlanService(
		db,
		validation.New(),
		durablejobexceptions.NewRoutineTaskException(),
	)
	routineTaskEngine := routinetaskworker.NewEngine(
		config,
		routineTaskClaimer,
		routineTaskPlanService,
		routineTaskExecutionService,
		routineTaskLifecycleProducer,
		routineTaskCompletionProducer,
	)
	routineTaskEngine.SetRoutineTaskRecoverer(
		routinetaskrecoverers.NewStaleRecordRecoverer(db),
	)
	a.routineTaskEngine = routineTaskEngine
	shutdownRoutineTaskEngine := routineTaskEngine.Start(context.Background())

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
