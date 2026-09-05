package developmentroutes

import (
	"fmt"

	"github.com/gin-gonic/gin"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/ratelimit"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/core/adapters"
)

type APIRouteDependencies struct {
	CoreAdapter             *coreadapters.CoreAdapter
	AllowedDomains          []string
	UnauthorizedRateLimiter *ratelimit.HybridRateLimiter
}

func NewRouter(deps APIRouteDependencies) *gin.Engine {
	developmentRouter := slogs.WithGinLogger(gin.New())
	coreAdapter, allowedDomains, unauthorizedRateLimiter := deps.CoreAdapter, deps.AllowedDomains, deps.UnauthorizedRateLimiter
	developmentAPIRouterGroup := developmentRouter.Group("/" + cgateway.APIDevelopmentBaseURL) // use in development mode
	developmentAPIRouterGroup.Use(
		middlewares.SanitizeXForwardedForMiddleware(),
		middlewares.CORSMiddleware(),
		middlewares.DomainWhiteListMiddleware(allowedDomains),
	)
	developmentAPIRouterGroup.Use(middlewares.KeyMiddleware())
	developmentAPIRouterGroup.OPTIONS("/*path", func(ctx *gin.Context) { ctx.Status(200) })
	fmt.Println("API router group path:", developmentAPIRouterGroup.BasePath())

	// APIGateway deliberately exposes only the first-party resource domains
	// that are stable enough for external integrations.
	configureDevelopmentStationRoutes(
		developmentAPIRouterGroup,
		StationRouteDependencies{
			CoreAdapter:             coreAdapter,
			UnauthorizedRateLimiter: unauthorizedRateLimiter,
		},
	)
	configureDevelopmentRoutineRoutes(
		developmentAPIRouterGroup,
		RoutineRouteDependencies{
			CoreAdapter:             coreAdapter,
			UnauthorizedRateLimiter: unauthorizedRateLimiter,
		},
	)
	configureDevelopmentRoutineTagRoutes(
		developmentAPIRouterGroup,
		RoutineTagRouteDependencies{
			CoreAdapter:             coreAdapter,
			UnauthorizedRateLimiter: unauthorizedRateLimiter,
		},
	)
	configureDevelopmentRoutineTaskRoutes(
		developmentAPIRouterGroup,
		RoutineTaskRouteDependencies{
			CoreAdapter:             coreAdapter,
			UnauthorizedRateLimiter: unauthorizedRateLimiter,
		},
	)
	configureDevelopmentRoutineTaskDependencyRoutes(
		developmentAPIRouterGroup,
		RoutineTaskDependencyRouteDependencies{
			CoreAdapter:             coreAdapter,
			UnauthorizedRateLimiter: unauthorizedRateLimiter,
		},
	)
	configureDevelopmentRootShelfRoutes(
		developmentAPIRouterGroup,
		RootShelfRouteDependencies{
			CoreAdapter:             coreAdapter,
			UnauthorizedRateLimiter: unauthorizedRateLimiter,
		},
	)
	configureDevelopmentSubShelfRoutes(
		developmentAPIRouterGroup,
		SubShelfRouteDependencies{
			CoreAdapter:             coreAdapter,
			UnauthorizedRateLimiter: unauthorizedRateLimiter,
		},
	)
	configureDevelopmentMaterialRoutes(
		developmentAPIRouterGroup,
		MaterialRouteDependencies{
			CoreAdapter:             coreAdapter,
			UnauthorizedRateLimiter: unauthorizedRateLimiter,
		},
	)
	configureDevelopmentBlockPackRoutes(
		developmentAPIRouterGroup,
		BlockPackRouteDependencies{
			CoreAdapter:             coreAdapter,
			UnauthorizedRateLimiter: unauthorizedRateLimiter,
		},
	)
	configureDevelopmentBlockRoutes(
		developmentAPIRouterGroup,
		BlockRouteDependencies{
			CoreAdapter:             coreAdapter,
			UnauthorizedRateLimiter: unauthorizedRateLimiter,
		},
	)

	return developmentRouter
}
