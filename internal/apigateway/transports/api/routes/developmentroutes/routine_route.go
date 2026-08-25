package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	binders "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/core/adapters"
)

type RoutineRouteDependencies struct {
	CoreAdapter  *coreadapters.CoreAdapter
	RateLimiters RateLimiters
}

func configureDevelopmentRoutineRoutes(
	router *gin.RouterGroup,
	deps RoutineRouteDependencies,
) {
	coreAdapter, rateLimiters := deps.CoreAdapter, deps.RateLimiters
	if router == nil {
		router = DevelopmentAPIRouterGroup
	}

	routineBinder := binders.NewRoutineBinder()
	routineController := controllers.NewRoutineController(coreAdapter)

	routineRoutes := router.Group("/routines")
	defaultMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(rateLimiters.Unauthorized),
		middlewares.TimeoutMiddleware(3 * time.Second),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		routineRoutes.GET(
			"/:routine-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyRoutineById"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.getMyRoutineById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineBinder.BindGetMyRoutineById(routineController.GetMyRoutineById),
			)...,
		)
		routineRoutes.GET(
			"/station/:station-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyRoutinesByStationId"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.getMyRoutinesByStationId"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineBinder.BindGetMyRoutinesByStationId(routineController.GetMyRoutinesByStationId),
			)...,
		)
		routineRoutes.GET(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getAllMyRoutinesByTimeRange"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.getAllMyRoutinesByTimeRange"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineBinder.BindGetAllMyRoutinesByTimeRange(routineController.GetAllMyRoutinesByTimeRange),
			)...,
		)
		routineRoutes.POST(
			"/station/:station-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createRoutineByStationId"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.createRoutineByStationId"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindCreateRoutineByStationId(routineController.CreateRoutineByStationId),
			)...,
		)
		routineRoutes.POST(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createRoutinesByStationIds"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.createRoutinesByStationIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindCreateRoutinesByStationIds(routineController.CreateRoutinesByStationIds),
			)...,
		)
		routineRoutes.PUT(
			"/:routine-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyRoutineById"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.updateMyRoutineById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindUpdateMyRoutineById(routineController.UpdateMyRoutineById),
			)...,
		)
		routineRoutes.PUT(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyRoutinesByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.updateMyRoutinesByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindUpdateMyRoutinesByIds(routineController.UpdateMyRoutinesByIds),
			)...,
		)
		routineRoutes.POST(
			"/:routine-id/tags/:routine-tag-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("linkRoutineTagById"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.linkRoutineTagById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindLinkRoutineTagById(routineController.LinkRoutineTagById),
			)...,
		)
		routineRoutes.POST(
			"/tags",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("linkRoutineTagsByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.linkRoutineTagsByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindLinkRoutineTagsByIds(routineController.LinkRoutineTagsByIds),
			)...,
		)
		routineRoutes.POST(
			"/:routine-id/items/:item-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("linkRoutineItemById"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.linkRoutineItemById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindLinkRoutineItemById(routineController.LinkRoutineItemById),
			)...,
		)
		routineRoutes.POST(
			"/items",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("linkRoutineItemsByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.linkRoutineItemsByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindLinkRoutineItemsByIds(routineController.LinkRoutineItemsByIds),
			)...,
		)
		routineRoutes.PATCH(
			"/:routine-id/restore",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("restoreMyRoutineById"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.restoreMyRoutineById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindRestoreMyRoutineById(routineController.RestoreMyRoutineById),
			)...,
		)
		routineRoutes.PATCH(
			"/batch/restore",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("restoreMyRoutinesByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.restoreMyRoutinesByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindRestoreMyRoutinesByIds(routineController.RestoreMyRoutinesByIds),
			)...,
		)
		routineRoutes.DELETE(
			"/:routine-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyRoutineById"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.deleteMyRoutineById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindDeleteMyRoutineById(routineController.DeleteMyRoutineById),
			)...,
		)
		routineRoutes.DELETE(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyRoutinesByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.deleteMyRoutinesByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindDeleteMyRoutinesByIds(routineController.DeleteMyRoutinesByIds),
			)...,
		)
		routineRoutes.DELETE(
			"/:routine-id/permanently",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("hardDeleteMyRoutineById"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.hardDeleteMyRoutineById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindHardDeleteMyRoutineById(routineController.HardDeleteMyRoutineById),
			)...,
		)
		routineRoutes.DELETE(
			"/batch/permanently",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("hardDeleteMyRoutinesByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.hardDeleteMyRoutinesByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				routineBinder.BindHardDeleteMyRoutinesByIds(routineController.HardDeleteMyRoutinesByIds),
			)...,
		)
	}

	/* ============================== Routes for Visualization ============================== */

	visualizationRoutes := router.Group("/routines/visualizations")
	visualizationMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(rateLimiters.Unauthorized),
		middlewares.TimeoutMiddleware(3 * time.Second),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		visualizationRoutes.GET(
			"/status-count",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("visualizeMyRoutineStatusCount"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.visualizeMyRoutineStatusCount"),
				},
				append(
					visualizationMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineBinder.BindVisualizeMyRoutineStatusCount(routineController.VisualizeMyRoutineStatusCount),
			)...,
		)
		visualizationRoutes.GET(
			"/period-count",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("visualizeMyRoutinePeriodCount"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.visualizeMyRoutinePeriodCount"),
				},
				append(
					visualizationMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineBinder.BindVisualizeMyRoutinePeriodCount(routineController.VisualizeMyRoutinePeriodCount),
			)...,
		)
		visualizationRoutes.GET(
			"/scheduled-start-at-count",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("visualizeMyRoutineScheduledStartAtCount"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.visualizeMyRoutineScheduledStartAtCount"),
				},
				append(
					visualizationMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineBinder.BindVisualizeMyRoutineScheduledStartAtCount(routineController.VisualizeMyRoutineScheduledStartAtCount),
			)...,
		)
		visualizationRoutes.GET(
			"/scheduled-end-at-count",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("visualizeMyRoutineScheduledEndAtCount"),
					middlewares.ApplyMeterMiddleware("server.requests.routine.visualizeMyRoutineScheduledEndAtCount"),
				},
				append(
					visualizationMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				routineBinder.BindVisualizeMyRoutineScheduledEndAtCount(routineController.VisualizeMyRoutineScheduledEndAtCount),
			)...,
		)
	}
}
