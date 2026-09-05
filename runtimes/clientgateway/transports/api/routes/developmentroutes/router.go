package developmentroutes

import (
	"fmt"

	"github.com/gin-gonic/gin"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/ratelimit"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
	notificationadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/notification/adapters"
)

type APIRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	NotificationClient        *notificationadapters.NotificationAdapter
	AllowedDomains            []string
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	UnauthorizedRateLimiter   *ratelimit.HybridRateLimiter
}

func NewRouter(deps APIRouteDependencies) *gin.Engine {
	developmentRouter := slogs.WithGinLogger(gin.New())
	coreAdapter, notificationClient := deps.CoreAdapter, deps.NotificationClient
	allowedDomains, accessTokenCookieHandler := deps.AllowedDomains, deps.AccessTokenCookieHandler
	refreshTokenCookieHandler, unauthorizedRateLimiter := deps.RefreshTokenCookieHandler, deps.UnauthorizedRateLimiter
	developmentAPIRouterGroup := developmentRouter.Group("/" + cgateway.APIDevelopmentBaseURL) // use in development mode
	developmentAPIRouterGroup.Use(
		middlewares.SanitizeXForwardedForMiddleware(),
		middlewares.CORSMiddleware(),
		middlewares.DomainWhiteListMiddleware(allowedDomains),
	)
	developmentAPIRouterGroup.OPTIONS("/*path", func(ctx *gin.Context) { ctx.Status(200) })
	fmt.Println("API router group path:", developmentAPIRouterGroup.BasePath())

	configureDevelopmentAuthRoutes(
		developmentAPIRouterGroup,
		AuthRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentUserRoutes(
		developmentAPIRouterGroup,
		UserRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentUserInfoRoutes(
		developmentAPIRouterGroup,
		UserInfoRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureUserSettingRoutes(
		developmentAPIRouterGroup,
		UserSettingRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentUserAccountRoutes(
		developmentAPIRouterGroup,
		UserAccountRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentAPIKeyRoutes(
		developmentAPIRouterGroup,
		APIKeyRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)

	configureDevelopmentStationRoutes(
		developmentAPIRouterGroup,
		StationRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentRoutineRoutes(
		developmentAPIRouterGroup,
		RoutineRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentRoutineTagRoutes(
		developmentAPIRouterGroup,
		RoutineTagRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentRoutineTaskRoutes(
		developmentAPIRouterGroup,
		RoutineTaskRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentRoutineTaskDependencyRoutes(
		developmentAPIRouterGroup,
		RoutineTaskDependencyRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentRootShelfRoutes(
		developmentAPIRouterGroup,
		RootShelfRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentSubShelfRoutes(
		developmentAPIRouterGroup,
		SubShelfRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentMaterialRoutes(
		developmentAPIRouterGroup,
		MaterialRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentBlockPackRoutes(
		developmentAPIRouterGroup,
		BlockPackRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentBlockRoutes(
		developmentAPIRouterGroup,
		BlockRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)

	configureDevelopmentRoutineTaskRecordRoutes(
		developmentAPIRouterGroup,
		RoutineTaskRecordRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentRealtimeRoutes(
		developmentAPIRouterGroup,
		RealtimeRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentGraphQLRoutes(
		developmentAPIRouterGroup,
		GraphQLRouteDependencies{
			CoreAdapter:               coreAdapter,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)
	configureDevelopmentNotificationRoutes(
		developmentAPIRouterGroup,
		NotificationRouteDependencies{
			NotificationClient:        notificationClient,
			AccessTokenCookieHandler:  accessTokenCookieHandler,
			RefreshTokenCookieHandler: refreshTokenCookieHandler,
			UnauthorizedRateLimiter:   unauthorizedRateLimiter,
		},
	)

	configureStaticRoutes(developmentAPIRouterGroup)

	return developmentRouter
}
