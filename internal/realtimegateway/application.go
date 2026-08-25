package realtimegateway

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	cookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"
	platform "github.com/HiIamJeff67/notegic-backend/shared/platform"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	crealtimegateway "github.com/HiIamJeff67/notegic-backend/contracts/realtime-gateway/v1"

	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	observability "github.com/HiIamJeff67/notegic-backend/shared/platform/observability"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	platformredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	realtimeconfig "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/configs"
	ratelimitrecord "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/data/redis/ratelimitrecord"
	realtimelease "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/data/redis/realtimelease"
	ratelimit "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/ratelimit"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/transports/api/middlewares"
	apiRouters "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/transports/api/routers"
	coreconsumers "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/transports/core/consumers"
	durablejobconsumers "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/transports/durablejob/consumers"
	notificationconsumers "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/transports/notification/consumers"
	status "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/transports/status"
	websockettransport "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/transports/websocket"
)

type Application struct {
	healthy atomic.Bool
	ready   atomic.Bool
}

type ApplicationInterface interface {
	Start() func()
	IsHealthy() bool
	IsReady() bool
	loadConfig() realtimeconfig.Config
	loadRedisConfig() platformredis.Config
	loadKafkaConnectionConfig() platformkafka.ConnectionConfig
	initializeObservability() func()
	initializeCaches(platformredis.Config, func()) (*platformredis.ClientSet, *realtimelease.RealtimeLeaseCacheClient, *ratelimit.HybridRateLimiter, *ratelimit.HybridRateLimiter)
	buildRouter(realtimeconfig.Config, *platformredis.ClientSet, *realtimelease.RealtimeLeaseCacheClient, *ratelimit.HybridRateLimiter, *ratelimit.HybridRateLimiter, func()) (*gin.Engine, *websockettransport.WebSocketAdapter)
	initializeConsumers(realtimeconfig.Config, platformkafka.ConnectionConfig, *realtimelease.RealtimeLeaseCacheClient) func()
	startHTTP(realtimeconfig.Config, *platformredis.ClientSet, *gin.Engine, *websockettransport.WebSocketAdapter, *ratelimit.HybridRateLimiter, *ratelimit.HybridRateLimiter, func(), func()) func()
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

func (a *Application) loadConfig() realtimeconfig.Config {
	config, err := realtimeconfig.LoadConfig()
	if err != nil {
		panic(err)
	}
	return config
}

func (a *Application) loadRedisConfig() platformredis.Config {
	redisConfig, err := platformredis.LoadConfig()
	if err != nil {
		panic(err)
	}
	return redisConfig
}

func (a *Application) loadKafkaConnectionConfig() platformkafka.ConnectionConfig {
	kafkaConnectionConfig, err := platformkafka.LoadConnectionConfig()
	if err != nil {
		panic(err)
	}
	return kafkaConnectionConfig
}

func (a *Application) initializeObservability() func() {
	return observability.Initialize(
		context.Background(),
		observability.LoadConfig("notegic-realtime-gateway"),
	)
}

func (a *Application) initializeCaches(
	redisConfig platformredis.Config,
	shutdownObservability func(),
) (*platformredis.ClientSet, *realtimelease.RealtimeLeaseCacheClient, *ratelimit.HybridRateLimiter, *ratelimit.HybridRateLimiter) {
	redisClientSet, err := platformredis.NewClientSet(redisConfig)
	if err != nil {
		shutdownObservability()
		panic(err)
	}
	realtimeLeaseCacheStore := realtimelease.Register(context.Background(), redisClientSet)
	if err := realtimeLeaseCacheStore.Initialize(context.Background()); err != nil {
		_ = redisClientSet.Close()
		shutdownObservability()
		panic(err)
	}
	rateLimitRecordCacheStore := ratelimitrecord.Register(context.Background(), redisClientSet)
	if err := rateLimitRecordCacheStore.Initialize(context.Background()); err != nil {
		_ = redisClientSet.Close()
		shutdownObservability()
		panic(err)
	}
	rateLimitRecordCacheClient := ratelimitrecord.NewRateLimitRecordCacheClient(rateLimitRecordCacheStore)
	upgradeRateLimitConfig := realtimeconfig.DefaultUpgradeRateLimitConfig()
	upgradeRateLimitConfig.CacheClient = rateLimitRecordCacheClient
	return redisClientSet,
		realtimelease.NewRealtimeLeaseCacheClient(realtimeLeaseCacheStore),
		ratelimit.NewHybridRateLimiter(upgradeRateLimitConfig, false),
		ratelimit.NewHybridRateLimiter(upgradeRateLimitConfig, true)
}

func (a *Application) buildRouter(
	config realtimeconfig.Config,
	redisClientSet *platformredis.ClientSet,
	realtimeLeaseClient *realtimelease.RealtimeLeaseCacheClient,
	unauthorizedLimiter *ratelimit.HybridRateLimiter,
	authorizedLimiter *ratelimit.HybridRateLimiter,
	shutdownObservability func(),
) (*gin.Engine, *websockettransport.WebSocketAdapter) {
	router := logs.WithGinLogger(gin.New())
	if err := router.SetTrustedProxies(config.TrustedProxies); err != nil {
		unauthorizedLimiter.Stop()
		authorizedLimiter.Stop()
		_ = redisClientSet.Close()
		shutdownObservability()
		panic(err)
	}
	status.ConfigureStartedRouter(router, a.IsHealthy)
	status.ConfigureHealthRouter(router, a.IsReady)
	routes := router.Group("/" + crealtimegateway.RealtimeDevelopmentBaseURL)
	routes.Use(
		middlewares.SanitizeXForwardedForMiddleware(),
		middlewares.CORSMiddleware(),
		middlewares.DomainWhiteListMiddleware(config.AllowedDomains),
		middlewares.UnauthorizedRateLimitMiddleware(unauthorizedLimiter),
	)
	routes.OPTIONS("/*path", func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })
	accessTokenCookieHandler := cookies.New(cookies.Config{
		Name:     cookies.ValidCookieName_AccessToken,
		Path:     "/",
		Duration: 30 * time.Minute,
		Secure:   platform.CurrentEnvironment == types.Environment_Production,
		HTTPOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	refreshTokenCookieHandler := cookies.New(cookies.Config{
		Name:     cookies.ValidCookieName_RefreshToken,
		Path:     "/",
		Duration: 14 * 24 * time.Hour,
		Secure:   platform.CurrentEnvironment == types.Environment_Production,
		HTTPOnly: true,
		SameSite: http.SameSiteStrictMode,
	})
	apiRouters.ConfigureRoutes(routes, realtimeLeaseClient, accessTokenCookieHandler, refreshTokenCookieHandler, authorizedLimiter)
	websocketAdapter := websockettransport.NewWebSocketAdapter(config, realtimeLeaseClient)
	routes.GET("", websocketAdapter.Handle)
	return router, websocketAdapter
}

func (a *Application) initializeConsumers(
	config realtimeconfig.Config,
	kafkaConnection platformkafka.ConnectionConfig,
	realtimeLeaseClient *realtimelease.RealtimeLeaseCacheClient,
) func() {
	lifecycleConsumer := coreconsumers.NewLifecycleConsumer(
		realtimeLeaseClient,
		platformkafka.ConsumerConfig{
			ClientConfig: platformkafka.ClientConfig{
				ConnectionConfig: kafkaConnection,
				ClientId:         "notegic-realtime-gateway-lifecycle",
			},
			ConsumerGroup:       "notegic-realtime-gateway-lifecycle-v1",
			MaximumAttempts:     config.KafkaConsumer.MaximumAttempts,
			InitialRetryBackoff: config.KafkaConsumer.InitialRetryBackoff,
			MaximumRetryBackoff: config.KafkaConsumer.MaximumRetryBackoff,
			MaximumPollRecords:  config.KafkaConsumer.MaximumPollRecords,
		},
	)
	routineTaskLifecycleConsumer := durablejobconsumers.NewRoutineTaskLifecycleConsumer(
		realtimeLeaseClient,
		platformkafka.ConsumerConfig{
			ClientConfig: platformkafka.ClientConfig{
				ConnectionConfig: kafkaConnection,
				ClientId:         "notegic-realtime-gateway-durable-job-routine-task-lifecycle",
			},
			ConsumerGroup:       "notegic-realtime-gateway-durable-job-routine-task-lifecycle-v1",
			MaximumAttempts:     config.KafkaConsumer.MaximumAttempts,
			InitialRetryBackoff: config.KafkaConsumer.InitialRetryBackoff,
			MaximumRetryBackoff: config.KafkaConsumer.MaximumRetryBackoff,
			MaximumPollRecords:  config.KafkaConsumer.MaximumPollRecords,
		},
	)
	notificationConsumer := notificationconsumers.NewNotificationConsumer(
		realtimeLeaseClient,
		platformkafka.ConsumerConfig{
			ClientConfig: platformkafka.ClientConfig{
				ConnectionConfig: kafkaConnection,
				ClientId:         "notegic-realtime-gateway-notification",
			},
			ConsumerGroup:       "notegic-realtime-gateway-notification-v1",
			MaximumAttempts:     config.KafkaConsumer.MaximumAttempts,
			InitialRetryBackoff: config.KafkaConsumer.InitialRetryBackoff,
			MaximumRetryBackoff: config.KafkaConsumer.MaximumRetryBackoff,
			MaximumPollRecords:  config.KafkaConsumer.MaximumPollRecords,
		},
	)
	shutdownLifecycle := lifecycleConsumer.Start(context.Background())
	shutdownRoutineTaskLifecycle := routineTaskLifecycleConsumer.Start(context.Background())
	shutdownNotification := notificationConsumer.Start(context.Background())
	return func() {
		shutdownLifecycle()
		shutdownRoutineTaskLifecycle()
		shutdownNotification()
	}
}

func (a *Application) startHTTP(
	config realtimeconfig.Config,
	redisClientSet *platformredis.ClientSet,
	router *gin.Engine,
	websocketAdapter *websockettransport.WebSocketAdapter,
	unauthorizedLimiter *ratelimit.HybridRateLimiter,
	authorizedLimiter *ratelimit.HybridRateLimiter,
	shutdownConsumers func(),
	shutdownObservability func(),
) func() {
	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		shutdownConsumers()
		websocketAdapter.Shutdown()
		unauthorizedLimiter.Stop()
		authorizedLimiter.Stop()
		_ = redisClientSet.Close()
		shutdownObservability()
		panic(err)
	}
	a.healthy.Store(true)
	a.ready.Store(true)
	server := &http.Server{Handler: router}
	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()
	return func() {
		a.ready.Store(false)
		a.healthy.Store(false)
		shutdownConsumers()
		websocketAdapter.Shutdown()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			fmt.Println("Failed to shutdown WebSocket server: ", err)
		}
		unauthorizedLimiter.Stop()
		authorizedLimiter.Stop()
		if err := redisClientSet.Close(); err != nil {
			fmt.Println("Failed to disconnect WebSocket cache servers: ", err)
		}
		shutdownObservability()
	}
}

func (a *Application) Start() func() {
	shutdownObservability := a.initializeObservability()
	config := a.loadConfig()
	redisConfig := a.loadRedisConfig()
	kafkaConnection := a.loadKafkaConnectionConfig()
	redisClientSet, realtimeLeaseClient, unauthorizedLimiter, authorizedLimiter := a.initializeCaches(redisConfig, shutdownObservability)
	router, websocketAdapter := a.buildRouter(config, redisClientSet, realtimeLeaseClient, unauthorizedLimiter, authorizedLimiter, shutdownObservability)
	shutdownConsumers := a.initializeConsumers(config, kafkaConnection, realtimeLeaseClient)
	return a.startHTTP(config, redisClientSet, router, websocketAdapter, unauthorizedLimiter, authorizedLimiter, shutdownConsumers, shutdownObservability)
}

// make sure Application struct followed the ApplicationInterface implementations
var _ ApplicationInterface = (*Application)(nil)
