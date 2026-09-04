package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-task-records"

	routineservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/routines"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type RoutineTaskRecordRouterDependencies struct {
	Service        routineservices.RoutineTaskRecordServiceInterface
	AuthMiddleware gin.HandlerFunc
}

func configureRoutineTaskRecordRoutes(
	router *gin.RouterGroup,
	deps RoutineTaskRecordRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	endpoint := endpoints.NewRoutineTaskRecordEndpoint(deps.Service)
	routineTaskRecordRoutes := router.Group("/routine-task-records")
	{
		routineTaskRecordRoutes.POST(
			"/get-by-routine-task-id",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.GetMyRoutineTaskRecordsByRoutineTaskIdOperation,
			),
			authMiddleware,
			endpoint.GetMyRoutineTaskRecordsByRoutineTaskId,
		)
		routineTaskRecordRoutes.POST(
			"/graphql/search",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.SearchRoutineTaskRecordsOperation,
			),
			authMiddleware,
			endpoint.SearchRoutineTaskRecords,
		)
	}
	visualizationRoutes := routineTaskRecordRoutes.Group("/visualizations")
	{
		visualizationRoutes.POST(
			"/status-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutineTaskRecordStatusCountOperation,
			),
			authMiddleware,
			endpoint.VisualizeMyRoutineTaskRecordStatusCount,
		)
		visualizationRoutes.POST(
			"/purpose-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutineTaskRecordPurposeCountOperation,
			),
			authMiddleware,
			endpoint.VisualizeMyRoutineTaskRecordPurposeCount,
		)
		visualizationRoutes.POST(
			"/scheduled-at-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutineTaskRecordScheduledAtCountOperation,
			),
			authMiddleware,
			endpoint.VisualizeMyRoutineTaskRecordScheduledAtCount,
		)
		visualizationRoutes.POST(
			"/actual-started-at-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutineTaskRecordActualStartedAtCountOperation,
			),
			authMiddleware,
			endpoint.VisualizeMyRoutineTaskRecordActualStartedAtCount,
		)
		visualizationRoutes.POST(
			"/actual-ended-at-count",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.VisualizeMyRoutineTaskRecordActualEndedAtCountOperation,
			),
			authMiddleware,
			endpoint.VisualizeMyRoutineTaskRecordActualEndedAtCount,
		)
	}
}
