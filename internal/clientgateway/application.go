package clientgateway

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

	observability "github.com/HiIamJeff67/notegic-backend/shared/platform/observability"
	platformredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	gatewayconfig "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/configs"
	ratelimitrecord "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/data/redis/ratelimitrecord"
	ratelimit "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/ratelimit"
	ratelimitmiddlewares "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/middlewares"
	developmentroutes "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/routes/developmentroutes"
	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
	notificationadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/notification/adapters"
	status "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/status"
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
	loadRedisConfig() platformredis.Config
	initializeObservability() func()
	initializeRateLimiters(*platformredis.ClientSet, func()) (*ratelimit.HybridRateLimiter, *ratelimit.HybridRateLimiter)
	buildRouter(gatewayconfig.Config, *ratelimit.HybridRateLimiter, *ratelimit.HybridRateLimiter, *platformredis.ClientSet, func()) *gin.Engine
	startHTTP(gatewayconfig.Config, *gin.Engine, *ratelimit.HybridRateLimiter, *ratelimit.HybridRateLimiter, *platformredis.ClientSet, func()) func()
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

func (a *Application) loadRedisConfig() platformredis.Config {
	redisConfig, err := platformredis.LoadConfig()
	if err != nil {
		panic(err)
	}
	return redisConfig
}

func (a *Application) initializeObservability() func() {
	return observability.Initialize(
		context.Background(),
		observability.LoadConfig("notegic-client-gateway"),
	)

}

func (a *Application) initializeRateLimiters(
	redisClientSet *platformredis.ClientSet,
	shutdownObservability func(),
) (*ratelimit.HybridRateLimiter, *ratelimit.HybridRateLimiter) {
	rateLimitRecordCacheStore := ratelimitrecord.Register(context.Background(), redisClientSet)
	if err := rateLimitRecordCacheStore.Initialize(context.Background()); err != nil {
		_ = redisClientSet.Close()
		shutdownObservability()
		panic(err)
	}
	rateLimitRecordCacheClient := ratelimitrecord.NewRateLimitRecordCacheClient(rateLimitRecordCacheStore)
	unauthorizedRateLimitConfig := gatewayconfig.DefaultUnauthorizedRateLimitConfig()
	unauthorizedRateLimitConfig.CacheClient = rateLimitRecordCacheClient
	authorizedRateLimitConfig := gatewayconfig.DefaultAuthorizedRateLimitConfig()
	authorizedRateLimitConfig.CacheClient = rateLimitRecordCacheClient
	return ratelimitmiddlewares.InitUnauthorizedRateLimiter(unauthorizedRateLimitConfig), ratelimitmiddlewares.InitAuthorizedRateLimiter(authorizedRateLimitConfig)
}

func (a *Application) buildRouter(
	config gatewayconfig.Config,
	unauthorizedRateLimiter *ratelimit.HybridRateLimiter,
	authorizedRateLimiter *ratelimit.HybridRateLimiter,
	redisClientSet *platformredis.ClientSet,
	shutdownObservability func(),
) *gin.Engine {
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
	router := developmentroutes.NewRouter(developmentroutes.APIRouteDependencies{
		CoreAdapter:               coreadapters.NewCoreAdapter(config.CoreBaseUrl, config.CoreAdapterTimeout),
		NotificationClient:        notificationadapters.NewNotificationAdapter(config.NotificationBaseUrl, config.NotificationAdapterTimeout),
		AllowedDomains:            config.AllowedDomains,
		AccessTokenCookieHandler:  accessTokenCookieHandler,
		RefreshTokenCookieHandler: refreshTokenCookieHandler,
		RateLimiters: developmentroutes.RateLimiters{
			Unauthorized: unauthorizedRateLimiter,
			Authorized:   authorizedRateLimiter,
		},
	})
	if err := router.SetTrustedProxies(config.TrustedProxies); err != nil {
		unauthorizedRateLimiter.Stop()
		authorizedRateLimiter.Stop()
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
	authorizedRateLimiter *ratelimit.HybridRateLimiter,
	redisClientSet *platformredis.ClientSet,
	shutdownObservability func(),
) func() {
	listener, err := net.Listen("tcp", config.ListenAddress)
	if err != nil {
		unauthorizedRateLimiter.Stop()
		authorizedRateLimiter.Stop()
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
		authorizedRateLimiter.Stop()
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
	redisClientSet, err := platformredis.NewClientSet(redisConfig)
	if err != nil {
		shutdownObservability()
		panic(err)
	}
	unauthorizedRateLimiter, authorizedRateLimiter := a.initializeRateLimiters(redisClientSet, shutdownObservability)
	router := a.buildRouter(config, unauthorizedRateLimiter, authorizedRateLimiter, redisClientSet, shutdownObservability)
	return a.startHTTP(config, router, unauthorizedRateLimiter, authorizedRateLimiter, redisClientSet, shutdownObservability)
}

// make sure Application struct followed the ApplicationInterface implementations
var _ ApplicationInterface = (*Application)(nil)
