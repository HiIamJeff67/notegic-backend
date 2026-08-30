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

type MaterialRouteDependencies struct {
	CoreAdapter  *coreadapters.CoreAdapter
	RateLimiters RateLimiters
}

func configureDevelopmentMaterialRoutes(
	router *gin.RouterGroup,
	deps MaterialRouteDependencies,
) {
	coreAdapter, rateLimiters := deps.CoreAdapter, deps.RateLimiters
	if router == nil {
		router = DevelopmentAPIRouterGroup
	}

	materialBinder := binders.NewMaterialBinder()
	materialController := controllers.NewMaterialController(coreAdapter)

	materialRoutes := router.Group("/materials")
	defaultMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(rateLimiters.Unauthorized),
		middlewares.TimeoutMiddleware(3 * time.Second),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		materialRoutes.GET(
			"/:material-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyMaterialById"),
					middlewares.ApplyMeterMiddleware("server.requests.material.getMyMaterialById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				materialBinder.BindGetMyMaterialById(materialController.GetMyMaterialById),
			)...,
		)
		materialRoutes.GET(
			"/:material-id/parent",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyMaterialAndItsParentById"),
					middlewares.ApplyMeterMiddleware("server.requests.material.getMyMaterialAndItsParentById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				materialBinder.BindGetMyMaterialAndItsParentById(materialController.GetMyMaterialAndItsParentById),
			)...,
		)
		materialRoutes.GET(
			"/sub-shelf/:parent-sub-shelf-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyMaterialsByParentSubShelfId"),
					middlewares.ApplyMeterMiddleware("server.requests.material.getMyMaterialsByParentSubShelfId"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				materialBinder.BindGetMyMaterialsByParentSubShelfId(materialController.GetMyMaterialsByParentSubShelfId),
			)...,
		)
		materialRoutes.GET(
			"/root-shelf/:root-shelf-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getAllMyMaterialsByRootShelfId"),
					middlewares.ApplyMeterMiddleware("server.requests.material.getAllMyMaterialsByRootShelfId"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				materialBinder.BindGetAllMyMaterialsByRootShelfId(materialController.GetAllMyMaterialsByRootShelfId),
			)...,
		)
		materialRoutes.POST(
			"/sub-shelf/:parent-sub-shelf-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createMyMaterial"),
					middlewares.ApplyMeterMiddleware("server.requests.material.createMyMaterial"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				materialBinder.BindCreateMyMaterial(materialController.CreateMyMaterial),
			)...,
		)
		materialRoutes.PUT(
			"/:material-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyMaterialById"),
					middlewares.ApplyMeterMiddleware("server.requests.material.updateMyMaterialById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				materialBinder.BindUpdateMyMaterialById(materialController.UpdateMyMaterialById),
			)...,
		)
		materialRoutes.PUT(
			"/:material-id/content",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("saveMyMaterialById"),
					middlewares.ApplyMeterMiddleware("server.requests.material.saveMyMaterialById"),
					middlewares.MultipartMiddleware(),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				materialBinder.BindSaveMyMaterialById(materialController.SaveMyMaterialById),
			)...,
		)
		materialRoutes.PUT(
			"/:material-id/parent",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("moveMyMaterialById"),
					middlewares.ApplyMeterMiddleware("server.requests.material.moveMyMaterialById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				materialBinder.BindMoveMyMaterialById(materialController.MoveMyMaterialById),
			)...,
		)
		materialRoutes.PUT(
			"/batch/parent",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("moveMyMaterialsByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.material.moveMyMaterialsByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				materialBinder.BindMoveMyMaterialsByIds(materialController.MoveMyMaterialsByIds),
			)...,
		)
		materialRoutes.PATCH(
			"/:material-id/restore",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("restoreMyMaterialById"),
					middlewares.ApplyMeterMiddleware("server.requests.material.restoreMyMaterialById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				materialBinder.BindRestoreMyMaterialById(materialController.RestoreMyMaterialById),
			)...,
		)
		materialRoutes.PATCH(
			"/batch/restore",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("restoreMyMaterialsByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.material.restoreMyMaterialsByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				materialBinder.BindRestoreMyMaterialsByIds(materialController.RestoreMyMaterialsByIds),
			)...,
		)
		materialRoutes.DELETE(
			"/:material-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyMaterialById"),
					middlewares.ApplyMeterMiddleware("server.requests.material.deleteMyMaterialById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				materialBinder.BindDeleteMyMaterialById(materialController.DeleteMyMaterialById),
			)...,
		)
		materialRoutes.DELETE(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyMaterialsByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.material.deleteMyMaterialsByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				materialBinder.BindDeleteMyMaterialsByIds(materialController.DeleteMyMaterialsByIds),
			)...,
		)
	}
}
