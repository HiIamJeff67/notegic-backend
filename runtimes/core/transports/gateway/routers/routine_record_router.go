package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-records"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	routineservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/routines"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type RoutineRecordRouterDependencies struct {
	Service          routineservices.RoutineRecordServiceInterface
	AuthMiddleware   gin.HandlerFunc
	APIKeyMiddleware gin.HandlerFunc
}

func configureRoutineRecordRoutes(
	router *gin.RouterGroup,
	deps RoutineRecordRouterDependencies,
) {
	apiCompatibleAuthMiddleware := middlewares.EitherMiddleware(
		[]gin.HandlerFunc{deps.AuthMiddleware},
		[]gin.HandlerFunc{deps.APIKeyMiddleware},
		func(ctx *gin.Context) bool { return contexts.IsClientGateway(ctx.Request.Context()) },
	)[0]
	endpoint := endpoints.NewRoutineRecordEndpoint(deps.Service)
	routineRecordRoutes := router.Group("/routine-records")
	{
		routineRecordRoutes.POST(
			"/graphql/search",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.SearchRoutineRecordsOperation,
			),
			apiCompatibleAuthMiddleware,
			endpoint.SearchRoutineRecords,
		)
	}
}
