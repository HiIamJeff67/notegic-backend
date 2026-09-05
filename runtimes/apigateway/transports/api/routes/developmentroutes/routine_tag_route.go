package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/ratelimit"
	binders "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/core/adapters"
)

type RoutineTagRouteDependencies struct {
	CoreAdapter             *coreadapters.CoreAdapter
	UnauthorizedRateLimiter *ratelimit.HybridRateLimiter
}

func configureDevelopmentRoutineTagRoutes(
	router *gin.RouterGroup,
	deps RoutineTagRouteDependencies,
) {
	coreAdapter, unauthorizedRateLimiter := deps.CoreAdapter, deps.UnauthorizedRateLimiter

	routineTagBinder := binders.NewRoutineTagBinder()
	routineTagController := controllers.NewRoutineTagController(coreAdapter)

	routineTagRoutes := router.Group("/routine-tags")
	defaultMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
		middlewares.TimeoutMiddleware(3 * time.Second),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		routineTagRoutes.GET(
			"/:routine-tag-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyRoutineTagById"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTag.getMyRoutineTagById"),
				},
				defaultMiddlewares,
				routineTagBinder.BindGetMyRoutineTagById(routineTagController.GetMyRoutineTagById),
			)...,
		)
		routineTagRoutes.GET(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getAllMyRoutineTags"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTag.getAllMyRoutineTags"),
				},
				defaultMiddlewares,
				routineTagBinder.BindGetAllMyRoutineTags(routineTagController.GetAllMyRoutineTags),
			)...,
		)
		routineTagRoutes.POST(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createRoutineTag"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTag.createRoutineTag"),
				},
				defaultMiddlewares,
				routineTagBinder.BindCreateRoutineTag(routineTagController.CreateRoutineTag),
			)...,
		)
		routineTagRoutes.POST(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createRoutineTags"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTag.createRoutineTags"),
				},
				defaultMiddlewares,
				routineTagBinder.BindCreateRoutineTags(routineTagController.CreateRoutineTags),
			)...,
		)
		routineTagRoutes.PUT(
			"/:routine-tag-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyRoutineTagById"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTag.updateMyRoutineTagById"),
				},
				defaultMiddlewares,
				routineTagBinder.BindUpdateMyRoutineTagById(routineTagController.UpdateMyRoutineTagById),
			)...,
		)
		routineTagRoutes.PUT(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyRoutineTagsByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTag.updateMyRoutineTagsByIds"),
				},
				defaultMiddlewares,
				routineTagBinder.BindUpdateMyRoutineTagsByIds(routineTagController.UpdateMyRoutineTagsByIds),
			)...,
		)
		routineTagRoutes.DELETE(
			"/:routine-tag-id/permanently",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("hardDeleteMyRoutineTagById"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTag.hardDeleteMyRoutineTagById"),
				},
				defaultMiddlewares,
				routineTagBinder.BindHardDeleteMyRoutineTagById(routineTagController.HardDeleteMyRoutineTagById),
			)...,
		)
		routineTagRoutes.DELETE(
			"/batch/permanently",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("hardDeleteMyRoutineTagsByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.routineTag.hardDeleteMyRoutineTagsByIds"),
				},
				defaultMiddlewares,
				routineTagBinder.BindHardDeleteMyRoutineTagsByIds(routineTagController.HardDeleteMyRoutineTagsByIds),
			)...,
		)
	}
}
