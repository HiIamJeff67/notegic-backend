package core

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"go.opentelemetry.io/otel/attribute"

	"github.com/gin-gonic/gin"

	types "github.com/HiIamJeff67/notegic-backend/contracts/types"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	authcode "github.com/HiIamJeff67/notegic-backend/shared/lib/authcode"

	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	observability "github.com/HiIamJeff67/notegic-backend/shared/platform/observability"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	platformrepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	platformredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	coreconfig "github.com/HiIamJeff67/notegic-backend/internal/core/configs"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	corescopes "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/scopes"
	seeds "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/seeds"
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
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type Application struct {
	healthy atomic.Bool
	ready   atomic.Bool
}

type ApplicationInterface interface {
	loadConfig() coreconfig.Config
	loadRedisConfig() platformredis.Config
	loadKafkaConnectionConfig() platformkafka.ConnectionConfig
	initializeObservability() func()
	initializeDatabase(platformpostgres.Config, func())
	initializeRedis(platformredis.Config, func()) *platformredis.ClientSet
	initializeCacheClients(coreconfig.Config, *platformredis.ClientSet, func()) (*userdata.UserDataCacheClient, *apikeycache.APIKeyCacheClient)
	initializeYjsClient(coreconfig.Config) *yjsworkertransport.DocumentInitializationClient
	initializeKafka(platformkafka.ConnectionConfig) (*platformkafka.Producer, bool)
	initializeWorkers(coreconfig.Config, platformkafka.ConnectionConfig, *platformkafka.Producer) func()
	buildRouter(coreconfig.Config, *platformkafka.Producer, *userdata.UserDataCacheClient, *yjsworkertransport.DocumentInitializationClient, *apikeycache.APIKeyCacheClient) *gin.Engine
	startHTTP(coreconfig.Config, *platformredis.ClientSet, *platformkafka.Producer, bool, func(), *gin.Engine, func()) func()
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
	kafkaProducer *platformkafka.Producer,
	userDataCacheClient *userdata.UserDataCacheClient,
	yjsDocumentInitializationClient *yjsworkertransport.DocumentInitializationClient,
	apiKeyCacheClient *apikeycache.APIKeyCacheClient,
) *gin.Engine {
	validator := validation.New()

	rootShelfScope := scopes.NewRootShelfScope()
	stationScope := scopes.NewStationScope()
	blockScope := scopes.NewBlockScope()
	blockPackScope := scopes.NewBlockPackScope()
	subShelfScope := scopes.NewSubShelfScope()
	materialScope := scopes.NewMaterialScope()
	routineScope := scopes.NewRoutineScope()
	routineTagScope := corescopes.NewRoutineTagScope()
	routineTaskScope := corescopes.NewRoutineTaskScope()
	routineTaskRecordScope := corescopes.NewRoutineTaskRecordScope()
	itemScope := corescopes.NewItemScope()

	userRepository := repositories.NewUserRepository()
	userInfoRepository := repositories.NewUserInfoRepository()
	userAccountRepository := repositories.NewUserAccountRepository()
	userSettingRepository := repositories.NewUserSettingRepository()
	rootShelfRepository := repositories.NewRootShelfRepository(rootShelfScope)
	stationRepository := repositories.NewStationRepository(stationScope)
	usersToShelvesRepository := repositories.NewUsersToShelvesRepository()
	usersToStationsRepository := repositories.NewUsersToStationsRepository()
	blockRepository := repositories.NewBlockRepository(blockScope)
	blockPackRepository := repositories.NewBlockPackRepository(blockPackScope)
	subShelfRepository := repositories.NewSubShelfRepository(subShelfScope)
	materialRepository := repositories.NewMaterialRepository(materialScope)
	routineRepository := repositories.NewRoutineRepository(routineScope)
	routineTagRepository := repositories.NewRoutineTagRepository(routineTagScope)
	routineTaskRepository := repositories.NewRoutineTaskRepository(routineTaskScope)
	routineTaskRecordRepository := repositories.NewRoutineTaskRecordRepository(routineTaskRecordScope)
	itemRepository := repositories.NewItemRepository(itemScope)
	outboxEventRepository := repositories.NewOutboxEventRepository()
	inMemoryStorage := storage.NewInMemoryStorage()

	oauthService := authservices.NewOAuthService(config.OAuthGoogle.OAuthConfig())
	emailClient := emailtransport.NewClient(
		data.DB,
	)

	authService := authservices.NewAuthService(
		validator,
		data.DB,
		userRepository,
		userInfoRepository,
		userAccountRepository,
		userSettingRepository,
		rootShelfRepository,
		outboxEventRepository,
		oauthService,
		emailClient,
		userDataCacheClient,
		authcode.New(),
	)
	rootShelfService := shelfservices.NewRootShelfService(
		validator,
		data.DB,
		rootShelfScope,
		rootShelfRepository,
		usersToShelvesRepository,
		blockPackRepository,
	)
	stationService := routineservices.NewStationService(
		validator,
		data.DB,
		stationScope,
		stationRepository,
		usersToStationsRepository,
	)
	userSettingService := userservices.NewUserSettingService(
		validator,
		data.DB,
		userSettingRepository,
	)
	userInfoService := userservices.NewUserInfoService(
		validator,
		data.DB,
		userInfoRepository,
		userDataCacheClient,
	)
	userAccountService := userservices.NewUserAccountService(
		validator,
		data.DB,
		userRepository,
		userAccountRepository,
		repositories.NewUserQuotaRepository(),
		oauthService,
	)
	userService := userservices.NewUserService(
		validator,
		data.DB,
		userRepository,
		userDataCacheClient,
	)
	blockService := blockservices.NewBlockService(
		validator,
		data.DB,
		blockScope,
		blockPackScope,
		subShelfScope,
		blockPackRepository,
		blockRepository,
	)
	realtimeService := realtimeservices.NewRealtimeService(
		validator,
		data.DB,
		blockPackRepository,
	)
	routineTagService := routineservices.NewRoutineTagService(
		validator,
		data.DB,
		routineTagRepository,
	)
	routineTaskRecordService := routineservices.NewRoutineTaskRecordService(
		validator,
		data.DB,
		routineTaskRecordRepository,
	)
	subShelfService := shelfservices.NewSubShelfService(
		validator,
		data.DB,
		inMemoryStorage,
		subShelfScope,
		subShelfRepository,
		rootShelfRepository,
		materialRepository,
		blockPackRepository,
	)
	blockPackService := blockservices.NewBlockPackService(
		validator,
		data.DB,
		blockPackScope,
		subShelfRepository,
		blockPackRepository,
	)
	materialService := materialservices.NewMaterialService(
		validator,
		data.DB,
		inMemoryStorage,
		materialScope,
		subShelfRepository,
		materialRepository,
		config.StorageKeySalt,
	)
	routineService := routineservices.NewRoutineService(
		validator,
		data.DB,
		routineScope,
		stationRepository,
		routineRepository,
		routineTagRepository,
		routineTaskRepository,
		itemRepository,
	)
	routineTaskExecutionService := routineservices.NewRoutineTaskExecutionService(
		validator,
		data.DB,
		yjsDocumentInitializationClient,
	)
	routineTaskService := routineservices.NewRoutineTaskService(
		validator,
		data.DB,
		routineTaskScope,
		routineTaskRepository,
		routineTaskRecordRepository,
		repositories.NewUserQuotaRepository(),
		routineTaskExecutionService,
	)
	themeService := otherservices.NewThemeService(data.DB)
	itemService := shelfservices.NewItemService(data.DB, itemScope)
	badgeService := otherservices.NewBadgeService(data.DB)
	apiKeyRepository := repositories.NewAPIKeyRepository()
	apiKeyService := apikeyservices.NewAPIKeyService(validator, data.DB, apiKeyRepository, apiKeyCacheClient)
	authMiddleware := coremiddlewares.AuthMiddleware(userRepository, userDataCacheClient)
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

func (a *Application) loadConfig() coreconfig.Config {
	config, err := coreconfig.LoadConfig()
	if err != nil {
		panic(err)
	}
	return config
}

func (a *Application) loadRedisConfig() platformredis.Config {
	config, err := platformredis.LoadConfig()
	if err != nil {
		panic(err)
	}
	return config
}

func (a *Application) loadKafkaConnectionConfig() platformkafka.ConnectionConfig {
	config, err := platformkafka.LoadConnectionConfig()
	if err != nil {
		panic(err)
	}
	return config
}

func (a *Application) initializeObservability() func() {
	return observability.Initialize(
		context.Background(),
		observability.LoadConfig("notegic-core"),
	)
}

func (a *Application) initializeDatabase(
	config platformpostgres.Config,
	shutdownObservability func(),
) {
	if _, err := data.Connect(config); err != nil {
		shutdownObservability()
		panic(fmt.Errorf("failed to connect Core database: %w", err))
	}
	platformrepositories.SetDefaultDB(data.DB)
	for _, migrate := range []func() error{
		func() error {
			return platformpostgres.Migrate(
				data.DB,
				types.Runtime_Core,
				data.DatabaseMigrationManifest,
			)
		},
		func() error { return platformpostgres.SeedDefaultDataToDatabase(data.DB, seeds.SeedingDefaultDataSQLs) },
	} {
		if err := migrate(); err != nil {
			_ = data.Disconnect(data.DB)
			shutdownObservability()
			panic(fmt.Errorf("failed to initialize Core database schema: %w", err))
		}
	}
}

func (a *Application) initializeRedis(
	config platformredis.Config,
	shutdownObservability func(),
) *platformredis.ClientSet {
	redisClientSet, err := platformredis.NewClientSet(config)
	if err != nil {
		shutdownObservability()
		panic(err)
	}
	return redisClientSet
}

func (a *Application) initializeCacheClients(
	config coreconfig.Config,
	redisClientSet *platformredis.ClientSet,
	shutdownObservability func(),
) (*userdata.UserDataCacheClient, *apikeycache.APIKeyCacheClient) {
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
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(context.Background(), exception.Origin(), exception.String())
		}
		_ = redisClientSet.Close()
		_ = data.Disconnect(data.DB)
		shutdownObservability()
		panic(exception)
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
		_ = redisClientSet.Close()
		_ = data.Disconnect(data.DB)
		shutdownObservability()
		panic(exception)
	}
	return userdata.NewUserDataCacheClient(config.UserDataCache, userDataCacheStore), apikeycache.NewAPIKeyCacheClient(apiKeyCacheStore)
}

func (a *Application) initializeYjsClient(config coreconfig.Config) *yjsworkertransport.DocumentInitializationClient {
	return yjsworkertransport.NewDocumentInitializationClient(config.YjsDocumentInitialization)
}

func (a *Application) initializeKafka(
	config platformkafka.ConnectionConfig,
) (*platformkafka.Producer, bool) {
	kafkaProducer, err := platformkafka.NewProducer(platformkafka.ClientConfig{
		ConnectionConfig: config,
		ClientId:         "notegic-core",
	})
	kafkaReady := err == nil
	if err != nil {
		logs.NotegicLogger.Warn(context.Background(), "Kafka is unavailable; Core is running in degraded mode", attribute.String("error.message", err.Error()))
	} else if err := kafkaProducer.Ping(context.Background()); err != nil {
		kafkaReady = false
		logs.NotegicLogger.Warn(context.Background(), "Kafka is unavailable; Core is running in degraded mode", attribute.String("error.message", err.Error()))
	}
	return kafkaProducer, kafkaReady
}

func (a *Application) initializeWorkers(
	config coreconfig.Config,
	kafkaConnection platformkafka.ConnectionConfig,
	kafkaProducer *platformkafka.Producer,
) func() {
	outboxRelay := coretransports.NewOutboxRelay(
		data.DB,
		repositories.NewOutboxEventRepository(),
		kafkaProducer,
		config.OutboxRelay,
	)
	yjsMaintenanceReconciliationWorker := coreworkers.NewYjsMaintenanceReconciliationWorker(
		data.DB,
		repositories.NewOutboxEventRepository(),
	)
	yjsMaintenanceWorker := coreworkers.NewYjsMaintenanceWorker(
		data.DB,
		yjsworkerproducers.NewYjsMaintenanceCommandProducer(kafkaProducer),
		config.YjsMaintenanceStrategy,
		platformkafka.ConsumerConfig{
			ClientConfig: platformkafka.ClientConfig{
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
		data.DB,
		config.QuotaCycleWorker,
		repositories.NewUserQuotaRepository(),
	)
	yjsCommandConsumer := yjsworkerconsumers.NewYjsCommandConsumer(
		data.DB,
		blockservices.NewYjsPersistenceService(data.DB),
		blockservices.NewBlockService(
			validation.New(),
			data.DB,
			scopes.NewBlockScope(),
			scopes.NewBlockPackScope(),
			scopes.NewSubShelfScope(),
			repositories.NewBlockPackRepository(scopes.NewBlockPackScope()),
			repositories.NewBlockRepository(scopes.NewBlockScope()),
		),
		platformkafka.ConsumerConfig{
			ClientConfig: platformkafka.ClientConfig{
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
	redisClientSet *platformredis.ClientSet,
	kafkaProducer *platformkafka.Producer,
	kafkaReady bool,
	shutdownWorkers func(),
	router *gin.Engine,
	shutdownObservability func(),
) func() {
	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		shutdownWorkers()
		if kafkaProducer != nil {
			kafkaProducer.Close()
		}
		_ = redisClientSet.Close()
		_ = data.Disconnect(data.DB)
		shutdownObservability()
		panic(err)
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
		shutdownWorkers()
		if kafkaProducer != nil {
			kafkaProducer.Close()
		}
		if err := redisClientSet.Close(); err != nil {
			fmt.Println("Failed to disconnect Core cache servers: ", err)
		}
		platformrepositories.SetDefaultDB(nil)
		_ = data.Disconnect(data.DB)
		shutdownObservability()
	}
}

func (a *Application) Start() func() {
	shutdownObservability := a.initializeObservability()
	config := a.loadConfig()
	redisConfig := a.loadRedisConfig()
	kafkaConnectionConfig := a.loadKafkaConnectionConfig()
	a.initializeDatabase(config.Postgres, shutdownObservability)
	redisClientSet := a.initializeRedis(redisConfig, shutdownObservability)
	userDataCacheClient, apiKeyCacheClient := a.initializeCacheClients(config, redisClientSet, shutdownObservability)
	yjsDocumentInitializationClient := a.initializeYjsClient(config)
	kafkaProducer, kafkaReady := a.initializeKafka(kafkaConnectionConfig)
	shutdownWorkers := a.initializeWorkers(config, kafkaConnectionConfig, kafkaProducer)
	router := a.buildRouter(config, kafkaProducer, userDataCacheClient, yjsDocumentInitializationClient, apiKeyCacheClient)
	return a.startHTTP(config, redisClientSet, kafkaProducer, kafkaReady, shutdownWorkers, router, shutdownObservability)
}

// make sure Application struct followed the ApplicationInterface implementations
var _ ApplicationInterface = (*Application)(nil)
