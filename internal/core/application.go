package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"

	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sauthcode "github.com/HiIamJeff67/notegic-backend/shared/lib/authcode"
	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	sobservability "github.com/HiIamJeff67/notegic-backend/shared/platform/observability"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	sredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	coreconfig "github.com/HiIamJeff67/notegic-backend/internal/core/configs"
	apikeycache "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/apikey"
	userdata "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata"
	storage "github.com/HiIamJeff67/notegic-backend/internal/core/data/storage"
	apikeyservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/apikey"
	authservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/auth"
	blockservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/blocks"
	materialservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/material"
	otherservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/other"
	realtimeservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/realtime"
	routineservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines"
	shelfservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/shelves"
	userservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/user"
	coretransports "github.com/HiIamJeff67/notegic-backend/internal/core/transports"
	durablejobrouters "github.com/HiIamJeff67/notegic-backend/internal/core/transports/durablejob/routers"
	emailtransport "github.com/HiIamJeff67/notegic-backend/internal/core/transports/email"
	coremiddlewares "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/middlewares"
	gatewayrouters "github.com/HiIamJeff67/notegic-backend/internal/core/transports/gateway/routers"
	status "github.com/HiIamJeff67/notegic-backend/internal/core/transports/status"
	yjsworkertransport "github.com/HiIamJeff67/notegic-backend/internal/core/transports/yjsworker"
	yjsworkerconsumers "github.com/HiIamJeff67/notegic-backend/internal/core/transports/yjsworker/consumers"
	yjsworkerproducers "github.com/HiIamJeff67/notegic-backend/internal/core/transports/yjsworker/producers"
	validation "github.com/HiIamJeff67/notegic-backend/internal/core/validations"
	coreworkers "github.com/HiIamJeff67/notegic-backend/internal/core/workers"
)

type Application struct {
	healthy atomic.Bool
	ready   atomic.Bool
}

type ApplicationInterface interface {
	initializeObservability() func()
	initializeDatabase(spostgres.Config) (*gorm.DB, error)
	initializeRedis(sredis.Config) (*sredis.ClientSet, error)
	initializeCacheClients(coreconfig.Config, *sredis.ClientSet) (*userdata.UserDataCacheClient, *apikeycache.APIKeyCacheClient, error)
	initializeYjsClient(coreconfig.Config) *yjsworkertransport.DocumentInitializationClient
	initializeKafka(skafka.ConnectionConfig) (*skafka.Producer, bool)
	initializeWorkers(coreconfig.Config, *gorm.DB, skafka.ConnectionConfig, *skafka.Producer) func()
	buildRouter(coreconfig.Config, *gorm.DB, *skafka.Producer, *userdata.UserDataCacheClient, *yjsworkertransport.DocumentInitializationClient, *apikeycache.APIKeyCacheClient) *gin.Engine
	startHTTP(coreconfig.Config, *gin.Engine, bool) (func(), error)
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
	return a.ready.Load()
}

func (a *Application) buildRouter(
	config coreconfig.Config,
	db *gorm.DB,
	kafkaProducer *skafka.Producer,
	userDataCacheClient *userdata.UserDataCacheClient,
	yjsDocumentInitializationClient *yjsworkertransport.DocumentInitializationClient,
	apiKeyCacheClient *apikeycache.APIKeyCacheClient,
) *gin.Engine {
	validator := validation.New()

	rootShelfScope := sscopes.NewRootShelfScope()
	stationScope := sscopes.NewStationScope()
	blockScope := sscopes.NewBlockScope()
	blockPackScope := sscopes.NewBlockPackScope()
	subShelfScope := sscopes.NewSubShelfScope()
	materialScope := sscopes.NewMaterialScope()
	routineScope := sscopes.NewRoutineScope()
	routineTagScope := sscopes.NewRoutineTagScope()
	routineTaskScope := sscopes.NewRoutineTaskScope()
	routineTaskRecordScope := sscopes.NewRoutineTaskRecordScope()
	itemScope := sscopes.NewItemScope()

	userRepository := srepositories.NewUserRepository(db)
	userInfoRepository := srepositories.NewUserInfoRepository(db)
	userAccountRepository := srepositories.NewUserAccountRepository(db)
	userSettingRepository := srepositories.NewUserSettingRepository(db)
	rootShelfRepository := srepositories.NewRootShelfRepository(db, rootShelfScope)
	stationRepository := srepositories.NewStationRepository(db, stationScope)
	usersToShelvesRepository := srepositories.NewUsersToShelvesRepository(db)
	usersToStationsRepository := srepositories.NewUsersToStationsRepository(db)
	blockRepository := srepositories.NewBlockRepository(db, blockScope)
	blockPackRepository := srepositories.NewBlockPackRepository(db, blockPackScope)
	subShelfRepository := srepositories.NewSubShelfRepository(db, subShelfScope)
	materialRepository := srepositories.NewMaterialRepository(db, materialScope)
	routineRepository := srepositories.NewRoutineRepository(db, routineScope)
	routineTagRepository := srepositories.NewRoutineTagRepository(db, routineTagScope)
	routineTaskRepository := srepositories.NewRoutineTaskRepository(db, routineTaskScope)
	routineTaskRecordRepository := srepositories.NewRoutineTaskRecordRepository(db, routineTaskRecordScope)
	itemRepository := srepositories.NewItemRepository(db, itemScope)
	outboxEventRepository := srepositories.NewOutboxEventRepository(db)
	inMemoryStorage := storage.NewInMemoryStorage()

	oauthService := authservices.NewOAuthService(config.OAuthGoogle.OAuthConfig())
	emailClient := emailtransport.NewClient(
		db,
	)

	authService := authservices.NewAuthService(
		validator,
		db,
		userRepository,
		userInfoRepository,
		userAccountRepository,
		userSettingRepository,
		rootShelfRepository,
		outboxEventRepository,
		oauthService,
		emailClient,
		userDataCacheClient,
		sauthcode.New(),
	)
	rootShelfService := shelfservices.NewRootShelfService(
		validator,
		db,
		rootShelfScope,
		rootShelfRepository,
		usersToShelvesRepository,
		blockPackRepository,
	)
	stationService := routineservices.NewStationService(
		validator,
		db,
		stationScope,
		stationRepository,
		usersToStationsRepository,
	)
	userSettingService := userservices.NewUserSettingService(
		validator,
		db,
		userSettingRepository,
	)
	userInfoService := userservices.NewUserInfoService(
		validator,
		db,
		userInfoRepository,
		userDataCacheClient,
	)
	userAccountService := userservices.NewUserAccountService(
		validator,
		db,
		userRepository,
		userAccountRepository,
		srepositories.NewUserQuotaRepository(db),
		oauthService,
	)
	userService := userservices.NewUserService(
		validator,
		db,
		userRepository,
		userDataCacheClient,
	)
	blockService := blockservices.NewBlockService(
		validator,
		db,
		blockScope,
		blockPackScope,
		subShelfScope,
		blockPackRepository,
		blockRepository,
	)
	realtimeService := realtimeservices.NewRealtimeService(
		validator,
		db,
		blockPackRepository,
	)
	routineTagService := routineservices.NewRoutineTagService(
		validator,
		db,
		routineTagRepository,
	)
	routineTaskRecordService := routineservices.NewRoutineTaskRecordService(
		validator,
		db,
		routineTaskRecordRepository,
	)
	subShelfService := shelfservices.NewSubShelfService(
		validator,
		db,
		inMemoryStorage,
		subShelfScope,
		subShelfRepository,
		rootShelfRepository,
		materialRepository,
		blockPackRepository,
	)
	blockPackService := blockservices.NewBlockPackService(
		validator,
		db,
		blockPackScope,
		subShelfRepository,
		blockPackRepository,
	)
	materialService := materialservices.NewMaterialService(
		validator,
		db,
		inMemoryStorage,
		materialScope,
		subShelfRepository,
		materialRepository,
		config.StorageKeySalt,
	)
	routineService := routineservices.NewRoutineService(
		validator,
		db,
		routineScope,
		stationRepository,
		routineRepository,
		routineTagRepository,
		routineTaskRepository,
		itemRepository,
	)
	routineTaskExecutionService := routineservices.NewRoutineTaskExecutionService(
		validator,
		db,
		yjsDocumentInitializationClient,
	)
	routineTaskService := routineservices.NewRoutineTaskService(
		validator,
		db,
		routineTaskScope,
		routineTaskRepository,
		routineTaskRecordRepository,
		srepositories.NewUserQuotaRepository(db),
		routineTaskExecutionService,
	)
	themeService := otherservices.NewThemeService(db)
	itemService := shelfservices.NewItemService(db, itemScope)
	badgeService := otherservices.NewBadgeService(db)
	apiKeyRepository := srepositories.NewAPIKeyRepository(db)
	apiKeyService := apikeyservices.NewAPIKeyService(validator, db, apiKeyRepository, apiKeyCacheClient)
	authMiddleware := coremiddlewares.AuthMiddleware(userRepository, userDataCacheClient, db)
	apiKeyMiddleware := coremiddlewares.APIKeyMiddleware(
		apiKeyRepository,
		userRepository,
		apiKeyCacheClient,
	)

	router := gatewayrouters.NewRouter(gatewayrouters.RouterDependencies{
		Auth: gatewayrouters.AuthRouterDependencies{
			Service:             authService,
			AuthMiddleware:      authMiddleware,
			UserDataCacheClient: userDataCacheClient,
		},
		APIKey: gatewayrouters.APIKeyRouterDependencies{
			Service:        apiKeyService,
			AuthMiddleware: authMiddleware,
		},
		RootShelf: gatewayrouters.RootShelfRouterDependencies{
			Service: rootShelfService, AuthMiddleware: authMiddleware, APIKeyMiddleware: apiKeyMiddleware,
		},
		Station: gatewayrouters.StationRouterDependencies{
			Service: stationService, AuthMiddleware: authMiddleware, APIKeyMiddleware: apiKeyMiddleware,
		},
		UserSetting: gatewayrouters.UserSettingRouterDependencies{Service: userSettingService, AuthMiddleware: authMiddleware},
		UserInfo:    gatewayrouters.UserInfoRouterDependencies{Service: userInfoService, AuthMiddleware: authMiddleware},
		UserAccount: gatewayrouters.UserAccountRouterDependencies{Service: userAccountService, AuthMiddleware: authMiddleware},
		User:        gatewayrouters.UserRouterDependencies{Service: userService, AuthMiddleware: authMiddleware},
		Block: gatewayrouters.BlockRouterDependencies{
			Service: blockService, AuthMiddleware: authMiddleware, APIKeyMiddleware: apiKeyMiddleware,
		},
		Realtime: gatewayrouters.RealtimeRouterDependencies{Service: realtimeService, AuthMiddleware: authMiddleware},
		RoutineTag: gatewayrouters.RoutineTagRouterDependencies{
			Service: routineTagService, AuthMiddleware: authMiddleware, APIKeyMiddleware: apiKeyMiddleware,
		},
		RoutineTaskRecord: gatewayrouters.RoutineTaskRecordRouterDependencies{Service: routineTaskRecordService, AuthMiddleware: authMiddleware},
		SubShelf: gatewayrouters.SubShelfRouterDependencies{
			Service: subShelfService, AuthMiddleware: authMiddleware, APIKeyMiddleware: apiKeyMiddleware,
		},
		BlockPack: gatewayrouters.BlockPackRouterDependencies{
			Service: blockPackService, AuthMiddleware: authMiddleware, APIKeyMiddleware: apiKeyMiddleware,
		},
		Material: gatewayrouters.MaterialRouterDependencies{
			Service: materialService, AuthMiddleware: authMiddleware, APIKeyMiddleware: apiKeyMiddleware,
		},
		Routine: gatewayrouters.RoutineRouterDependencies{
			Service: routineService, AuthMiddleware: authMiddleware, APIKeyMiddleware: apiKeyMiddleware,
		},
		RoutineTask: gatewayrouters.RoutineTaskRouterDependencies{
			Service: routineTaskService, AuthMiddleware: authMiddleware, APIKeyMiddleware: apiKeyMiddleware,
		},
		Theme: gatewayrouters.ThemeRouterDependencies{Service: themeService},
		Item:  gatewayrouters.ItemRouterDependencies{Service: itemService, AuthMiddleware: authMiddleware},
		Badge: gatewayrouters.BadgeRouterDependencies{Service: badgeService, AuthMiddleware: authMiddleware},
	})
	durablejobrouters.ConfigureBlockProjectionRoutes(router, blockService)
	return router
}

func (a *Application) initializeObservability() func() {
	return sobservability.Initialize(
		context.Background(),
		sobservability.LoadConfig("notegic-core"),
	)
}

func (a *Application) initializeDatabase(
	config spostgres.Config,
) (*gorm.DB, error) {
	if config.User != ctypes.Runtime_Core.RoleName() {
		return nil, fmt.Errorf("Core PostgreSQL user must be %q", ctypes.Runtime_Core.RoleName())
	}
	db, err := spostgres.Connect(config)
	if err != nil {
		return nil, fmt.Errorf("failed to connect Core database: %w", err)
	}
	return db, nil
}

func (a *Application) initializeRedis(config sredis.Config) (*sredis.ClientSet, error) {
	redisClientSet, err := sredis.NewClientSet(config)
	if err != nil {
		return nil, err
	}
	return redisClientSet, nil
}

func (a *Application) initializeCacheClients(
	config coreconfig.Config,
	redisClientSet *sredis.ClientSet,
) (*userdata.UserDataCacheClient, *apikeycache.APIKeyCacheClient, error) {
	userDataCacheStore, err := userdata.Register(context.Background(), redisClientSet)
	if err != nil {
		exception := cexceptions.New(
			"ConnectionFailed",
			"Cache",
			"Start",
			"Failed to connect to cache servers",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
		return nil, nil, exception
	}
	apiKeyCacheStore := apikeycache.Register(context.Background(), redisClientSet)
	if err := apiKeyCacheStore.Initialize(context.Background()); err != nil {
		exception := cexceptions.New(
			"ConnectionFailed",
			"Cache",
			"Start",
			"Failed to initialize API key cache",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
		return nil, nil, exception
	}
	return userdata.NewUserDataCacheClient(config.UserDataCache, userDataCacheStore), apikeycache.NewAPIKeyCacheClient(apiKeyCacheStore), nil
}

func (a *Application) initializeYjsClient(config coreconfig.Config) *yjsworkertransport.DocumentInitializationClient {
	return yjsworkertransport.NewDocumentInitializationClient(config.YjsDocumentInitialization)
}

func (a *Application) initializeKafka(
	config skafka.ConnectionConfig,
) (*skafka.Producer, bool) {
	kafkaProducer, err := skafka.NewProducer(skafka.ClientConfig{
		ConnectionConfig: config,
		ClientId:         "notegic-core",
	})
	kafkaReady := err == nil
	if err != nil {
		slogs.NotegicLogger.Warn(context.Background(), "Kafka is unavailable; Core is running in degraded mode", attribute.String("error.message", err.Error()))
	} else if err := kafkaProducer.Ping(context.Background()); err != nil {
		kafkaReady = false
		slogs.NotegicLogger.Warn(context.Background(), "Kafka is unavailable; Core is running in degraded mode", attribute.String("error.message", err.Error()))
	}
	return kafkaProducer, kafkaReady
}

func (a *Application) initializeWorkers(
	config coreconfig.Config,
	db *gorm.DB,
	kafkaConnection skafka.ConnectionConfig,
	kafkaProducer *skafka.Producer,
) func() {
	outboxRelay := coretransports.NewOutboxRelay(
		db,
		srepositories.NewOutboxEventRepository(db),
		kafkaProducer,
		config.OutboxRelay,
	)
	yjsMaintenanceReconciliationWorker := coreworkers.NewYjsMaintenanceReconciliationWorker(
		db,
		srepositories.NewOutboxEventRepository(db),
	)
	yjsMaintenanceWorker := coreworkers.NewYjsMaintenanceWorker(
		db,
		yjsworkerproducers.NewYjsMaintenanceCommandProducer(kafkaProducer),
		config.YjsMaintenanceStrategy,
		skafka.ConsumerConfig{
			ClientConfig: skafka.ClientConfig{
				ConnectionConfig: kafkaConnection,
				ClientId:         "notegic-core-yjs-maintenance",
			},
			ConsumerGroup:       "notegic-core-yjs-maintenance-v1",
			MaximumAttempts:     config.KafkaConsumer.MaximumAttempts,
			InitialRetryBackoff: config.KafkaConsumer.InitialRetryBackoff,
			MaximumRetryBackoff: config.KafkaConsumer.MaximumRetryBackoff,
			MaximumPollRecords:  config.KafkaConsumer.MaximumPollRecords,
		},
	)
	quotaCycleWorker := coreworkers.NewQuotaCycleWorker(
		db,
		config.QuotaCycleWorker,
		srepositories.NewUserQuotaRepository(db),
	)
	yjsCommandConsumer := yjsworkerconsumers.NewYjsCommandConsumer(
		db,
		blockservices.NewYjsPersistenceService(db),
		blockservices.NewBlockService(
			validation.New(),
			db,
			sscopes.NewBlockScope(),
			sscopes.NewBlockPackScope(),
			sscopes.NewSubShelfScope(),
			srepositories.NewBlockPackRepository(db, sscopes.NewBlockPackScope()),
			srepositories.NewBlockRepository(db, sscopes.NewBlockScope()),
		),
		skafka.ConsumerConfig{
			ClientConfig: skafka.ClientConfig{
				ConnectionConfig: kafkaConnection,
				ClientId:         "notegic-core-yjsworker",
			},
			ConsumerGroup:       yjsworkertransport.CommandConsumerGroup,
			MaximumAttempts:     config.KafkaConsumer.MaximumAttempts,
			InitialRetryBackoff: config.KafkaConsumer.InitialRetryBackoff,
			MaximumRetryBackoff: config.KafkaConsumer.MaximumRetryBackoff,
			MaximumPollRecords:  config.KafkaConsumer.MaximumPollRecords,
		},
	)
	shutdownOutboxRelay := outboxRelay.Start(context.Background())
	shutdownYjsMaintenanceReconciliationWorker := yjsMaintenanceReconciliationWorker.Start(context.Background())
	shutdownYjsMaintenanceWorker := yjsMaintenanceWorker.Start(context.Background())
	shutdownQuotaCycleWorker := quotaCycleWorker.Start(context.Background())
	shutdownYjsCommandConsumer := yjsCommandConsumer.Start(context.Background())
	return func() {
		shutdownYjsCommandConsumer()
		shutdownYjsMaintenanceWorker()
		shutdownYjsMaintenanceReconciliationWorker()
		shutdownQuotaCycleWorker()
		shutdownOutboxRelay()
	}
}

func (a *Application) startHTTP(
	config coreconfig.Config,
	router *gin.Engine,
	kafkaReady bool,
) (func(), error) {
	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		return nil, err
	}
	a.healthy.Store(true)
	a.ready.Store(kafkaReady)
	status.ConfigureStartedRouter(router, a.IsHealthy)
	status.ConfigureHealthRouter(router, a.IsReady)
	server := &http.Server{Handler: router}
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
			fmt.Println("Failed to shutdown Core service transport: ", err)
		}
	}, nil
}

func (a *Application) Start() func() {
	shutdownObservability := a.initializeObservability()
	var (
		db              *gorm.DB
		redisClientSet  *sredis.ClientSet
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
			if redisClientSet != nil {
				if err := redisClientSet.Close(); err != nil {
					fmt.Println("Failed to disconnect Core cache servers: ", err)
				}
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

	config, err := coreconfig.LoadConfig()
	if err != nil {
		fail(err)
	}
	redisConfig, err := sredis.LoadConfig()
	if err != nil {
		fail(err)
	}
	kafkaConnectionConfig, err := skafka.LoadConnectionConfig()
	if err != nil {
		fail(err)
	}
	db, err = a.initializeDatabase(config.Postgres)
	if err != nil {
		fail(err)
	}
	redisClientSet, err = a.initializeRedis(redisConfig)
	if err != nil {
		fail(err)
	}
	userDataCacheClient, apiKeyCacheClient, err := a.initializeCacheClients(config, redisClientSet)
	if err != nil {
		fail(err)
	}
	yjsDocumentInitializationClient := a.initializeYjsClient(config)
	kafkaProducer, kafkaReady := a.initializeKafka(kafkaConnectionConfig)
	shutdownWorkers = a.initializeWorkers(config, db, kafkaConnectionConfig, kafkaProducer)
	router := a.buildRouter(config, db, kafkaProducer, userDataCacheClient, yjsDocumentInitializationClient, apiKeyCacheClient)
	shutdownHTTP, err = a.startHTTP(config, router, kafkaReady)
	if err != nil {
		fail(err)
	}
	return shutdown
}

// make sure Application struct followed the ApplicationInterface implementations
var _ ApplicationInterface = (*Application)(nil)
