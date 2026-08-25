package apigateway

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"

	sobservability "github.com/HiIamJeff67/notegic-backend/shared/platform/observability"
	sredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	gatewayconfig "github.com/HiIamJeff67/notegic-backend/internal/apigateway/configs"
	ratelimitrecord "github.com/HiIamJeff67/notegic-backend/internal/apigateway/data/redis/ratelimitrecord"
	ratelimit "github.com/HiIamJeff67/notegic-backend/internal/apigateway/ratelimit"
	ratelimitmiddlewares "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/api/middlewares"
	developmentroutes "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/api/routes/developmentroutes"
	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/core/adapters"
	status "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/status"
)

type Application struct {
	healthy atomic.Bool
	ready   atomic.Bool
}

type ApplicationInterface interface {
	Start() func()
	IsHealthy() bool
	IsReady() bool
	loadConfig() gatewayconfig.Config
	loadRedisConfig() sredis.Config
	initializeObservability() func()
	initializeRateLimiter(gatewayconfig.Config, *sredis.ClientSet, func()) *ratelimit.HybridRateLimiter
	buildRouter(gatewayconfig.Config, *ratelimit.HybridRateLimiter, *sredis.ClientSet, func()) *gin.Engine
	startHTTP(gatewayconfig.Config, *gin.Engine, *ratelimit.HybridRateLimiter, *sredis.ClientSet, func()) func()
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

func (a *Application) loadConfig() gatewayconfig.Config {
	config, err := gatewayconfig.LoadConfig()
	if err != nil {
		panic(err)
	}
	return config
}

func (a *Application) loadRedisConfig() sredis.Config {
	redisConfig, err := sredis.LoadConfig()
	if err != nil {
		panic(err)
	}
	return redisConfig
}

func (a *Application) initializeObservability() func() {
	return sobservability.Initialize(
		context.Background(),
		sobservability.LoadConfig("notegic-api-gateway"),
	)

}

func (a *Application) initializeRateLimiter(
	config gatewayconfig.Config,
	redisClientSet *sredis.ClientSet,
	shutdownObservability func(),
) *ratelimit.HybridRateLimiter {
	rateLimitRecordCacheStore := ratelimitrecord.Register(context.Background(), redisClientSet)
	if err := rateLimitRecordCacheStore.Initialize(context.Background()); err != nil {
		_ = redisClientSet.Close()
		shutdownObservability()
		panic(err)
	}
	rateLimitRecordCacheClient := ratelimitrecord.NewRateLimitRecordCacheClient(rateLimitRecordCacheStore)
	unauthorizedRateLimitConfig := gatewayconfig.DefaultUnauthorizedRateLimitConfig()
	unauthorizedRateLimitConfig.CacheClient = rateLimitRecordCacheClient
	return ratelimitmiddlewares.InitUnauthorizedRateLimiter(unauthorizedRateLimitConfig)
}

func (a *Application) buildRouter(
	config gatewayconfig.Config,
	unauthorizedRateLimiter *ratelimit.HybridRateLimiter,
	redisClientSet *sredis.ClientSet,
	shutdownObservability func(),
) *gin.Engine {
	router := developmentroutes.NewRouter(developmentroutes.APIRouteDependencies{
		CoreAdapter:    coreadapters.NewCoreAdapter(config.CoreBaseUrl, config.CoreAdapterTimeout),
		AllowedDomains: config.AllowedDomains,
		RateLimiters:   developmentroutes.RateLimiters{Unauthorized: unauthorizedRateLimiter},
	})
	if err := router.SetTrustedProxies(config.TrustedProxies); err != nil {
		unauthorizedRateLimiter.Stop()
		_ = redisClientSet.Close()
		shutdownObservability()
		panic(err)
	}
	status.ConfigureStartedRouter(router, a.IsHealthy)
	status.ConfigureHealthRouter(router, a.IsReady)
	return router
}

func (a *Application) startHTTP(
	config gatewayconfig.Config,
	router *gin.Engine,
	unauthorizedRateLimiter *ratelimit.HybridRateLimiter,
	redisClientSet *sredis.ClientSet,
	shutdownObservability func(),
) func() {
	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		unauthorizedRateLimiter.Stop()
		_ = redisClientSet.Close()
		shutdownObservability()
		panic(err)
	}
	a.healthy.Store(true)
	a.ready.Store(true)
	server := &http.Server{
		Handler: router,
	}

	go func() {
		if err := server.Serve(listener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			panic(err)
		}
	}()

	return func() {
		// Shut down request handling before releasing its shared dependencies.
		a.ready.Store(false)
		a.healthy.Store(false)
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			fmt.Println("Failed to shutdown Gateway server: ", err)
		}
		unauthorizedRateLimiter.Stop()
		if err := redisClientSet.Close(); err != nil {
			fmt.Println("Failed to disconnect Gateway cache servers: ", err)
		}
		shutdownObservability()
	}
}

func (a *Application) Start() func() {
	shutdownObservability := a.initializeObservability()
	config := a.loadConfig()
	redisConfig := a.loadRedisConfig()
	redisClientSet, err := sredis.NewClientSet(redisConfig)
	if err != nil {
		shutdownObservability()
		panic(err)
	}
	unauthorizedRateLimiter := a.initializeRateLimiter(config, redisClientSet, shutdownObservability)
	router := a.buildRouter(config, unauthorizedRateLimiter, redisClientSet, shutdownObservability)
	return a.startHTTP(config, router, unauthorizedRateLimiter, redisClientSet, shutdownObservability)
}

// make sure Application struct followed the ApplicationInterface implementations
var _ ApplicationInterface = (*Application)(nil)
