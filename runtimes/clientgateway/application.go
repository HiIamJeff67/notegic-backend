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

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"
	splatform "github.com/HiIamJeff67/notegic-backend/shared/platform"
	sobservability "github.com/HiIamJeff67/notegic-backend/shared/platform/observability"
	sredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"
	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"

	gatewayconfig "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/configs"
	ratelimitrecord "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/data/redis/ratelimitrecord"
	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/ratelimit"
	ratelimitmiddlewares "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/middlewares"
	developmentroutes "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/routes/developmentroutes"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
	notificationadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/notification/adapters"
	status "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/status"
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
	initializeRateLimiters(*sredis.ClientSet, func()) (*ratelimit.HybridRateLimiter, *ratelimit.HybridRateLimiter)
	buildRouter(gatewayconfig.Config, *ratelimit.HybridRateLimiter, *ratelimit.HybridRateLimiter, *sredis.ClientSet, func()) *gin.Engine
	startHTTP(gatewayconfig.Config, *gin.Engine, *ratelimit.HybridRateLimiter, *ratelimit.HybridRateLimiter, *sredis.ClientSet, func()) func()
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
		sobservability.LoadConfig("notegic-client-gateway"),
	)

}

func (a *Application) initializeRateLimiters(
	redisClientSet *sredis.ClientSet,
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
	redisClientSet *sredis.ClientSet,
	shutdownObservability func(),
) *gin.Engine {
	accessTokenCookieHandler := scookies.New(scookies.Config{
		Name:     scookies.ValidCookieName_AccessToken,
		Path:     "/",
		Duration: 30 * time.Minute,
		Secure:   splatform.CurrentEnvironment == stypes.Environment_Production,
		HTTPOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	refreshTokenCookieHandler := scookies.New(scookies.Config{
		Name:     scookies.ValidCookieName_RefreshToken,
		Path:     "/",
		Duration: 14 * 24 * time.Hour,
		Secure:   splatform.CurrentEnvironment == stypes.Environment_Production,
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
	redisClientSet *sredis.ClientSet,
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
	redisClientSet, err := sredis.NewClientSet(redisConfig)
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
