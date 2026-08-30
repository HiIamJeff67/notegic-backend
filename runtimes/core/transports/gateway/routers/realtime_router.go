package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/realtime"

	realtimeservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/realtime"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type RealtimeRouterDependencies struct {
	Service        realtimeservices.RealtimeServiceInterface
	AuthMiddleware gin.HandlerFunc
}

func configureRealtimeRoutes(
	router *gin.RouterGroup,
	deps RealtimeRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	endpoint := endpoints.NewRealtimeEndpoint(deps.Service)
	realtimeRoutes := router.Group("/realtime")
	{
		realtimeRoutes.POST(
			"/connection-ticket/create",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateMyRealtimeConnectionTicketOperation,
			),
			authMiddleware,
			endpoint.CreateMyRealtimeConnectionTicket,
		)
		realtimeRoutes.POST(
			"/block-pack-channel-ticket/create",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.CreateMyBlockPackChannelTicketOperation,
			),
			authMiddleware,
			endpoint.CreateMyBlockPackChannelTicket,
		)
	}
}
