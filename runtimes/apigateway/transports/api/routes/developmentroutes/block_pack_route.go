package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/ratelimit"
	binders "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/transports/core/adapters"
)

type BlockPackRouteDependencies struct {
	CoreAdapter             *coreadapters.CoreAdapter
	UnauthorizedRateLimiter *ratelimit.HybridRateLimiter
}

func configureDevelopmentBlockPackRoutes(
	router *gin.RouterGroup,
	deps BlockPackRouteDependencies,
) {
	coreAdapter, unauthorizedRateLimiter := deps.CoreAdapter, deps.UnauthorizedRateLimiter

	blockPackBinder := binders.NewBlockPackBinder()
	blockPackController := controllers.NewBlockPackController(coreAdapter)

	blockPackRoutes := router.Group("/block-packs")
	defaultMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(unauthorizedRateLimiter),
		middlewares.TimeoutMiddleware(3 * time.Second),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		blockPackRoutes.GET(
			"/:block-pack-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyBlockPackById"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.getMyBlockPackById"),
				},
				defaultMiddlewares,
				middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				blockPackBinder.BindGetMyBlockPackById(blockPackController.GetMyBlockPackById),
			)...,
		)
		blockPackRoutes.GET(
			"/:block-pack-id/parent",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyBlockPackAndItsParentById"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.getMyBlockPackAndItsParentById"),
				},
				defaultMiddlewares,
				middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				blockPackBinder.BindGetMyBlockPackAndItsParentById(blockPackController.GetMyBlockPackAndItsParentById),
			)...,
		)
		blockPackRoutes.GET(
			"/sub-shelf/:parent-sub-shelf-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyBlockPacksByParentSubShelfId"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.getMyBlockPacksByParentSubShelfId"),
				},
				defaultMiddlewares,
				middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				blockPackBinder.BindGetMyBlockPacksByParentSubShelfId(blockPackController.GetMyBlockPacksByParentSubShelfId),
			)...,
		)
		blockPackRoutes.GET(
			"/root-shelf/:root-shelf-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyBlockPacksByRootShelfId"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.getMyBlockPacksByRootShelfId"),
				},
				defaultMiddlewares,
				middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				blockPackBinder.BindGetMyBlockPacksByRootShelfId(blockPackController.GetMyBlockPacksByRootShelfId),
			)...,
		)
		blockPackRoutes.POST(
			"/sub-shelf/:parent-sub-shelf-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createBlockPack"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.createBlockPack"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				blockPackBinder.BindCreateBlockPack(blockPackController.CreateBlockPack),
			)...,
		)
		blockPackRoutes.POST(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createBlockPacks"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.CreateBlockPacks"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				blockPackBinder.BindCreateBlockPacks(blockPackController.CreateBlockPacks),
			)...,
		)
		blockPackRoutes.PUT(
			"/:block-pack-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyBlockPackById"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.updateMyBlockPackById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				blockPackBinder.BindUpdateMyBlockPackById(blockPackController.UpdateMyBlockPackById),
			)...,
		)
		blockPackRoutes.PUT(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyBlockPacksByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.UpdateMyBlockPacksByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				blockPackBinder.BindUpdateMyBlockPacksByIds(blockPackController.UpdateMyBlockPacksByIds),
			)...,
		)
		blockPackRoutes.PUT(
			"/:block-pack-id/position",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("moveMyBlockPackById"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.moveMyBlockPackById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				blockPackBinder.BindMoveMyBlockPackByParentSubShelfId(blockPackController.MoveMyBlockPackByParentSubShelfId),
			)...,
		)
		blockPackRoutes.PUT(
			"/position",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("moveMyBlockPacksByParentSubShelfId"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.moveMyBlockPacksByParentSubShelfId"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				blockPackBinder.BindMoveMyBlockPacksByParentSubShelfId(blockPackController.MoveMyBlockPacksByParentSubShelfId),
			)...,
		)
		blockPackRoutes.PUT(
			"/batch/position",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("moveMyBlockPacksByParentSubShelfIds"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.moveMyBlockPacksByParentSubShelfIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				blockPackBinder.BindMoveMyBlockPacksByParentSubShelfIds(blockPackController.MoveMyBlockPacksByParentSubShelfIds),
			)...,
		)
		blockPackRoutes.PATCH(
			"/:block-pack-id/restore",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("restoreMyBlockPackById"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.restoreMyBlockPackById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				blockPackBinder.BindRestoreMyBlockPackById(blockPackController.RestoreMyBlockPackById),
			)...,
		)
		blockPackRoutes.PATCH(
			"/batch/restore",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("restoreMyBlockPacksByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.restoreMyBlockPacksByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				blockPackBinder.BindRestoreMyBlockPacksByIds(blockPackController.RestoreMyBlockPacksByIds),
			)...,
		)
		blockPackRoutes.DELETE(
			"/:block-pack-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyBlockPackById"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.deleteMyBlockPackById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				blockPackBinder.BindDeleteMyBlockPackById(blockPackController.DeleteMyBlockPackById),
			)...,
		)
		blockPackRoutes.DELETE(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyBlockPacksByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.blockPack.deleteMyBlockPacksByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Write),
				),
				blockPackBinder.BindDeleteMyBlockPacksByIds(blockPackController.DeleteMyBlockPacksByIds),
			)...,
		)
	}
}
