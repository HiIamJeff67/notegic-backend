package developmentroutes

import (
	"fmt"

	"github.com/gin-gonic/gin"

	cookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	ratelimit "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/ratelimit"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/core/adapters"
	notificationadapters "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/transports/notification/adapters"
)

var (
	DevelopmentRouter         *gin.Engine
	DevelopmentAPIRouterGroup *gin.RouterGroup
)

type RateLimiters struct {
	Unauthorized *ratelimit.HybridRateLimiter
	Authorized   *ratelimit.HybridRateLimiter
}

type APIRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	NotificationClient        *notificationadapters.NotificationAdapter
	AllowedDomains            []string
	AccessTokenCookieHandler  *cookies.CookieHandler
	RefreshTokenCookieHandler *cookies.CookieHandler
	RateLimiters              RateLimiters
}

func NewRouter(deps APIRouteDependencies) *gin.Engine {
	DevelopmentRouter = logs.WithGinLogger(gin.New())
	coreAdapter, notificationClient := deps.CoreAdapter, deps.NotificationClient
	allowedDomains, accessTokenCookieHandler := deps.AllowedDomains, deps.AccessTokenCookieHandler
	refreshTokenCookieHandler, rateLimiters := deps.RefreshTokenCookieHandler, deps.RateLimiters
	DevelopmentAPIRouterGroup = DevelopmentRouter.Group("/" + cgateway.APIDevelopmentBaseURL) // use in development mode
	DevelopmentAPIRouterGroup.Use(
		middlewares.SanitizeXForwardedForMiddleware(),
		middlewares.CORSMiddleware(),
		middlewares.DomainWhiteListMiddleware(allowedDomains),
	)
	DevelopmentAPIRouterGroup.OPTIONS("/*path", func(ctx *gin.Context) { ctx.Status(200) })
	fmt.Println("API router group path:", DevelopmentAPIRouterGroup.BasePath())

	configureDevelopmentAuthRoutes(DevelopmentAPIRouterGroup, AuthRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentUserRoutes(DevelopmentAPIRouterGroup, UserRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentUserInfoRoutes(DevelopmentAPIRouterGroup, UserInfoRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureUserSettingRoutes(DevelopmentAPIRouterGroup, UserSettingRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentUserAccountRoutes(DevelopmentAPIRouterGroup, UserAccountRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentAPIKeyRoutes(DevelopmentAPIRouterGroup, APIKeyRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})

	configureDevelopmentStationRoutes(DevelopmentAPIRouterGroup, StationRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentRoutineRoutes(DevelopmentAPIRouterGroup, RoutineRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentRoutineTagRoutes(DevelopmentAPIRouterGroup, RoutineTagRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentRoutineTaskRoutes(DevelopmentAPIRouterGroup, RoutineTaskRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentRootShelfRoutes(DevelopmentAPIRouterGroup, RootShelfRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentSubShelfRoutes(DevelopmentAPIRouterGroup, SubShelfRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentMaterialRoutes(DevelopmentAPIRouterGroup, MaterialRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentBlockPackRoutes(DevelopmentAPIRouterGroup, BlockPackRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentBlockRoutes(DevelopmentAPIRouterGroup, BlockRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})

	configureDevelopmentRoutineTaskRecordRoutes(DevelopmentAPIRouterGroup, RoutineTaskRecordRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentRealtimeRoutes(DevelopmentAPIRouterGroup, RealtimeRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentGraphQLRoutes(DevelopmentAPIRouterGroup, GraphQLRouteDependencies{CoreAdapter: coreAdapter, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})
	configureDevelopmentNotificationRoutes(DevelopmentAPIRouterGroup, NotificationRouteDependencies{NotificationClient: notificationClient, AccessTokenCookieHandler: accessTokenCookieHandler, RefreshTokenCookieHandler: refreshTokenCookieHandler, RateLimiters: rateLimiters})

	configureStaticRoutes(DevelopmentAPIRouterGroup, StaticRouteDependencies{RateLimiters: rateLimiters})

	return DevelopmentRouter
}
