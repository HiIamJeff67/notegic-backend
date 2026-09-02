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

var (
	DevelopmentRouter         *gin.Engine
	DevelopmentAPIRouterGroup *gin.RouterGroup
)

type RateLimiters struct {
	Unauthorized *ratelimit.HybridRateLimiter
}

type APIRouteDependencies struct {
	CoreAdapter    *coreadapters.CoreAdapter
	AllowedDomains []string
	RateLimiters   RateLimiters
}

func NewRouter(deps APIRouteDependencies) *gin.Engine {
	DevelopmentRouter = slogs.WithGinLogger(gin.New())
	coreAdapter, allowedDomains, rateLimiters := deps.CoreAdapter, deps.AllowedDomains, deps.RateLimiters
	DevelopmentAPIRouterGroup = DevelopmentRouter.Group("/" + cgateway.APIDevelopmentBaseURL) // use in development mode
	DevelopmentAPIRouterGroup.Use(
		middlewares.SanitizeXForwardedForMiddleware(),
		middlewares.CORSMiddleware(),
		middlewares.DomainWhiteListMiddleware(allowedDomains),
	)
	DevelopmentAPIRouterGroup.Use(middlewares.KeyMiddleware())
	DevelopmentAPIRouterGroup.OPTIONS("/*path", func(ctx *gin.Context) { ctx.Status(200) })
	fmt.Println("API router group path:", DevelopmentAPIRouterGroup.BasePath())

	// APIGateway deliberately exposes only the first-party resource domains
	// that are stable enough for external integrations.
	configureDevelopmentStationRoutes(DevelopmentAPIRouterGroup, StationRouteDependencies{CoreAdapter: coreAdapter, RateLimiters: rateLimiters})
	configureDevelopmentRoutineRoutes(DevelopmentAPIRouterGroup, RoutineRouteDependencies{CoreAdapter: coreAdapter, RateLimiters: rateLimiters})
	configureDevelopmentRoutineTagRoutes(DevelopmentAPIRouterGroup, RoutineTagRouteDependencies{CoreAdapter: coreAdapter, RateLimiters: rateLimiters})
	configureDevelopmentRoutineTaskRoutes(DevelopmentAPIRouterGroup, RoutineTaskRouteDependencies{CoreAdapter: coreAdapter, RateLimiters: rateLimiters})
	configureDevelopmentRoutineTaskDependencyRoutes(DevelopmentAPIRouterGroup, RoutineTaskDependencyRouteDependencies{CoreAdapter: coreAdapter, RateLimiters: rateLimiters})
	configureDevelopmentRootShelfRoutes(DevelopmentAPIRouterGroup, RootShelfRouteDependencies{CoreAdapter: coreAdapter, RateLimiters: rateLimiters})
	configureDevelopmentSubShelfRoutes(DevelopmentAPIRouterGroup, SubShelfRouteDependencies{CoreAdapter: coreAdapter, RateLimiters: rateLimiters})
	configureDevelopmentMaterialRoutes(DevelopmentAPIRouterGroup, MaterialRouteDependencies{CoreAdapter: coreAdapter, RateLimiters: rateLimiters})
	configureDevelopmentBlockPackRoutes(DevelopmentAPIRouterGroup, BlockPackRouteDependencies{CoreAdapter: coreAdapter, RateLimiters: rateLimiters})
	configureDevelopmentBlockRoutes(DevelopmentAPIRouterGroup, BlockRouteDependencies{CoreAdapter: coreAdapter, RateLimiters: rateLimiters})

	return DevelopmentRouter
}
