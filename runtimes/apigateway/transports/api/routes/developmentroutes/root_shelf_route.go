package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	binders "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/core/adapters"
)

type RootShelfRouteDependencies struct {
	CoreAdapter  *coreadapters.CoreAdapter
	RateLimiters RateLimiters
}

func configureDevelopmentRootShelfRoutes(
	router *gin.RouterGroup,
	deps RootShelfRouteDependencies,
) {
	coreAdapter, rateLimiters := deps.CoreAdapter, deps.RateLimiters
	if router == nil {
		router = DevelopmentAPIRouterGroup
	}

	rootShelfBinder := binders.NewRootShelfBinder()
	rootShelfController := controllers.NewRootShelfController(coreAdapter)

	rootShelfRoutes := router.Group("/root-shelves")
	defaultMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(rateLimiters.Unauthorized),
		middlewares.TimeoutMiddleware(1 * time.Second),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		rootShelfRoutes.GET(
			"/:root-shelf-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyRootShelfById"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.getMyRootShelfById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				rootShelfBinder.BindGetMyRootShelfById(rootShelfController.GetMyRootShelfById),
			)...,
		)
		rootShelfRoutes.POST(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createRootShelf"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.createRootShelf"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				rootShelfBinder.BindCreateRootShelf(rootShelfController.CreateRootShelf),
			)...,
		)
		rootShelfRoutes.POST(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createRootShelves"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.createRootShelves"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				rootShelfBinder.BindCreateRootShelves(rootShelfController.CreateRootShelves),
			)...,
		)
		rootShelfRoutes.PUT(
			"/:root-shelf-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyRootShelfById"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.updateMyRootShelfById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				rootShelfBinder.BindUpdateMyRootShelfById(rootShelfController.UpdateMyRootShelfById),
			)...,
		)
		rootShelfRoutes.PUT(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyRootShelvesByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.updateMyRootShelvesByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				rootShelfBinder.BindUpdateMyRootShelvesByIds(rootShelfController.UpdateMyRootShelvesByIds),
			)...,
		)
		rootShelfRoutes.PATCH(
			"/:root-shelf-id/restore",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("restoreMyRootShelfById"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.restoreMyRootShelfById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Owner),
				),
				rootShelfBinder.BindRestoreMyRootShelfById(rootShelfController.RestoreMyRootShelfById),
			)...,
		)
		rootShelfRoutes.PATCH(
			"/batch/restore",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("restoreMyRootShelvesByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.restoreMyRootShelvesByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Owner),
				),
				rootShelfBinder.BindRestoreMyRootShelvesByIds(rootShelfController.RestoreMyRootShelvesByIds),
			)...,
		)
		rootShelfRoutes.DELETE(
			"/:root-shelf-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyRootShelfById"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.deleteMyRootShelfById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				rootShelfBinder.BindDeleteMyRootShelfById(rootShelfController.DeleteMyRootShelfById),
			)...,
		)
		rootShelfRoutes.DELETE(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyRootShelvesByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.deleteMyRootShelvesByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Owner),
				),
				rootShelfBinder.BindDeleteMyRootShelvesByIds(rootShelfController.DeleteMyRootShelvesByIds),
			)...,
		)
		rootShelfRoutes.GET(
			"/:root-shelf-id/permissions/:user-public-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyRootShelfPermission"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.getMyRootShelfPermission"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				rootShelfBinder.BindGetMyRootShelfPermission(rootShelfController.GetMyRootShelfPermission),
			)...,
		)
		rootShelfRoutes.POST(
			"/:root-shelf-id/permissions/:user-public-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createMyRootShelfPermission"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.createMyRootShelfPermission"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				rootShelfBinder.BindCreateMyRootShelfPermission(rootShelfController.CreateMyRootShelfPermission),
			)...,
		)
		rootShelfRoutes.PUT(
			"/:root-shelf-id/permissions/:user-public-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("upsertMyRootShelfPermission"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.upsertMyRootShelfPermission"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				rootShelfBinder.BindUpsertMyRootShelfPermission(rootShelfController.UpsertMyRootShelfPermission),
			)...,
		)
		rootShelfRoutes.PUT(
			"/:root-shelf-id/permissions",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("upsertMyRootShelfPermissions"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.upsertMyRootShelfPermissions"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				rootShelfBinder.BindUpsertMyRootShelfPermissions(rootShelfController.UpsertMyRootShelfPermissions),
			)...,
		)
		rootShelfRoutes.PATCH(
			"/:root-shelf-id/permissions/:user-public-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyRootShelfPermission"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.updateMyRootShelfPermission"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				rootShelfBinder.BindUpdateMyRootShelfPermission(rootShelfController.UpdateMyRootShelfPermission),
			)...,
		)
		rootShelfRoutes.POST(
			"/:root-shelf-id/ownership",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("transferMyRootShelfOwnership"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.transferMyRootShelfOwnership"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Owner),
				),
				rootShelfBinder.BindTransferMyRootShelfOwnership(rootShelfController.TransferMyRootShelfOwnership),
			)...,
		)
		rootShelfRoutes.DELETE(
			"/:root-shelf-id/permissions/:user-public-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyRootShelfPermission"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.deleteMyRootShelfPermission"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				rootShelfBinder.BindDeleteMyRootShelfPermission(rootShelfController.DeleteMyRootShelfPermission),
			)...,
		)
		rootShelfRoutes.DELETE(
			"/:root-shelf-id/permissions",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyRootShelfPermissions"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.deleteMyRootShelfPermissions"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				rootShelfBinder.BindDeleteMyRootShelfPermissions(rootShelfController.DeleteMyRootShelfPermissions),
			)...,
		)
		rootShelfRoutes.DELETE(
			"/:root-shelf-id/memberships/me",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("leaveMyRootShelf"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.leaveMyRootShelf"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				rootShelfBinder.BindLeaveMyRootShelf(rootShelfController.LeaveMyRootShelf),
			)...,
		)
		rootShelfRoutes.DELETE(
			"/memberships/me",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("leaveMyRootShelves"),
					middlewares.ApplyMeterMiddleware("server.requests.rootShelf.leaveMyRootShelves"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				rootShelfBinder.BindLeaveMyRootShelves(rootShelfController.LeaveMyRootShelves),
			)...,
		)
	}
}
