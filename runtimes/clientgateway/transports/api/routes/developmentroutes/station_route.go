package developmentroutes

import (
	"time"

	"github.com/gin-gonic/gin"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	binders "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/binders"
	controllers "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/controllers"
	interceptors "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/interceptors"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/api/middlewares"
	coreadapters "github.com/HiIamJeff67/notegic-backend/runtimes/clientgateway/transports/core/adapters"
)

type StationRouteDependencies struct {
	CoreAdapter               *coreadapters.CoreAdapter
	AccessTokenCookieHandler  *scookies.CookieHandler
	RefreshTokenCookieHandler *scookies.CookieHandler
	RateLimiters              RateLimiters
}

func configureDevelopmentStationRoutes(
	router *gin.RouterGroup,
	deps StationRouteDependencies,
) {
	coreAdapter, accessTokenCookieHandler, refreshTokenCookieHandler, rateLimiters := deps.CoreAdapter, deps.AccessTokenCookieHandler, deps.RefreshTokenCookieHandler, deps.RateLimiters
	if router == nil {
		router = DevelopmentAPIRouterGroup
	}

	stationBinder := binders.NewStationBinder()
	stationController := controllers.NewStationController(coreAdapter)

	stationRoutes := router.Group("/stations")
	defaultMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(rateLimiters.Unauthorized),
		middlewares.TimeoutMiddleware(3 * time.Second),
		middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.RefreshTokenInterceptor(accessTokenCookieHandler),
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		stationRoutes.GET(
			"/:station-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyStationById"),
					middlewares.ApplyMeterMiddleware("server.requests.station.getMyStationById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				stationBinder.BindGetMyStationById(stationController.GetMyStationById),
			)...,
		)
		stationRoutes.GET(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getAllMyStations"),
					middlewares.ApplyMeterMiddleware("server.requests.station.getAllMyStations"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				stationBinder.BindGetAllMyStations(stationController.GetAllMyStations),
			)...,
		)
		stationRoutes.POST(
			"/",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createStation"),
					middlewares.ApplyMeterMiddleware("server.requests.station.createStation"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				stationBinder.BindCreateStation(stationController.CreateStation),
			)...,
		)
		stationRoutes.POST(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createStations"),
					middlewares.ApplyMeterMiddleware("server.requests.station.createStations"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				stationBinder.BindCreateStations(stationController.CreateStations),
			)...,
		)
		stationRoutes.PUT(
			"/:station-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyStationById"),
					middlewares.ApplyMeterMiddleware("server.requests.station.updateMyStationById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				stationBinder.BindUpdateMyStationById(stationController.UpdateMyStationById),
			)...,
		)
		stationRoutes.PUT(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyStationsByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.station.updateMyStationsByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				stationBinder.BindUpdateMyStationsByIds(stationController.UpdateMyStationsByIds),
			)...,
		)
		stationRoutes.PATCH(
			"/:station-id/restore",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("restoreMyStationById"),
					middlewares.ApplyMeterMiddleware("server.requests.station.restoreMyStationById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Owner),
				),
				stationBinder.BindRestoreMyStationById(stationController.RestoreMyStationById),
			)...,
		)
		stationRoutes.PATCH(
			"/batch/restore",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("restoreMyStationsByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.station.restoreMyStationsByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Owner),
				),
				stationBinder.BindRestoreMyStationsByIds(stationController.RestoreMyStationsByIds),
			)...,
		)
		stationRoutes.DELETE(
			"/:station-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyStationById"),
					middlewares.ApplyMeterMiddleware("server.requests.station.deleteMyStationById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				stationBinder.BindDeleteMyStationById(stationController.DeleteMyStationById),
			)...,
		)
		stationRoutes.DELETE(
			"/batch",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyStationsByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.station.deleteMyStationsByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Owner),
				),
				stationBinder.BindDeleteMyStationsByIds(stationController.DeleteMyStationsByIds),
			)...,
		)
		stationRoutes.DELETE(
			"/:station-id/permanently",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("hardDeleteMyStationById"),
					middlewares.ApplyMeterMiddleware("server.requests.station.hardDeleteMyStationById"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Owner),
				),
				stationBinder.BindHardDeleteMyStationById(stationController.HardDeleteMyStationById),
			)...,
		)
		stationRoutes.DELETE(
			"/batch/permanently",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("hardDeleteMyStationsByIds"),
					middlewares.ApplyMeterMiddleware("server.requests.station.hardDeleteMyStationsByIds"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Owner),
				),
				stationBinder.BindHardDeleteMyStationsByIds(stationController.HardDeleteMyStationsByIds),
			)...,
		)
		/* ============================== Routes for Station Permissions ============================== */

		stationRoutes.GET(
			"/:station-id/permissions/:user-public-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("getMyStationPermission"),
					middlewares.ApplyMeterMiddleware("server.requests.station.getMyStationPermission"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				stationBinder.BindGetMyStationPermission(stationController.GetMyStationPermission),
			)...,
		)
		stationRoutes.POST(
			"/:station-id/permissions/:user-public-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("createMyStationPermission"),
					middlewares.ApplyMeterMiddleware("server.requests.station.createMyStationPermission"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				stationBinder.BindCreateMyStationPermission(stationController.CreateMyStationPermission),
			)...,
		)
		stationRoutes.PUT(
			"/:station-id/permissions/:user-public-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("upsertMyStationPermission"),
					middlewares.ApplyMeterMiddleware("server.requests.station.upsertMyStationPermission"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				stationBinder.BindUpsertMyStationPermission(stationController.UpsertMyStationPermission),
			)...,
		)
		stationRoutes.PUT(
			"/:station-id/permissions",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("upsertMyStationPermissions"),
					middlewares.ApplyMeterMiddleware("server.requests.station.upsertMyStationPermissions"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				stationBinder.BindUpsertMyStationPermissions(stationController.UpsertMyStationPermissions),
			)...,
		)
		stationRoutes.PATCH(
			"/:station-id/permissions/:user-public-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("updateMyStationPermission"),
					middlewares.ApplyMeterMiddleware("server.requests.station.updateMyStationPermission"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				stationBinder.BindUpdateMyStationPermission(stationController.UpdateMyStationPermission),
			)...,
		)
		stationRoutes.POST(
			"/:station-id/ownership",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("transferMyStationOwnership"),
					middlewares.ApplyMeterMiddleware("server.requests.station.transferMyStationOwnership"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Owner),
				),
				stationBinder.BindTransferMyStationOwnership(stationController.TransferMyStationOwnership),
			)...,
		)
		stationRoutes.DELETE(
			"/:station-id/permissions/:user-public-id",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyStationPermission"),
					middlewares.ApplyMeterMiddleware("server.requests.station.deleteMyStationPermission"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				stationBinder.BindDeleteMyStationPermission(stationController.DeleteMyStationPermission),
			)...,
		)
		stationRoutes.DELETE(
			"/:station-id/permissions",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("deleteMyStationPermissions"),
					middlewares.ApplyMeterMiddleware("server.requests.station.deleteMyStationPermissions"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Admin),
				),
				stationBinder.BindDeleteMyStationPermissions(stationController.DeleteMyStationPermissions),
			)...,
		)
		stationRoutes.DELETE(
			"/:station-id/memberships/me",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("leaveMyStation"),
					middlewares.ApplyMeterMiddleware("server.requests.station.leaveMyStation"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				stationBinder.BindLeaveMyStation(stationController.LeaveMyStation),
			)...,
		)
		stationRoutes.DELETE(
			"/memberships/me",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("leaveMyStations"),
					middlewares.ApplyMeterMiddleware("server.requests.station.leaveMyStations"),
				},
				append(
					defaultMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				stationBinder.BindLeaveMyStations(stationController.LeaveMyStations),
			)...,
		)
	}

	/* ============================== Routes for Visualization ============================== */

	visualizationRoutes := router.Group("/stations/visualizations")
	visualizationMiddlewares := []gin.HandlerFunc{
		middlewares.UnauthorizedRateLimitMiddleware(rateLimiters.Unauthorized),
		middlewares.TimeoutMiddleware(3 * time.Second),
		middlewares.GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
		interceptors.ShareableResponseWriterInterceptor(
			interceptors.RefreshTokenInterceptor(accessTokenCookieHandler),
			interceptors.EmbeddedInterceptor,
		),
	}
	{
		visualizationRoutes.GET(
			"/total-count",
			middlewares.Reposition(
				[]gin.HandlerFunc{
					middlewares.ApplyTracerMiddleware("visualizeMyTotalCount"),
					middlewares.ApplyMeterMiddleware("server.requests.station.visualizeMyTotalCount"),
				},
				append(
					visualizationMiddlewares,
					middlewares.AllowedPermissionsAbove(cenums.AccessControlPermission_Read),
				),
				stationBinder.BindVisualizeMyTotalCount(stationController.VisualizeMyTotalCount),
			)...,
		)
	}
}
