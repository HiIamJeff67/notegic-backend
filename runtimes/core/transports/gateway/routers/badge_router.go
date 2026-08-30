package routers

import (
	"github.com/gin-gonic/gin"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/badges"

	otherservices "github.com/HiIamJeff67/notegic-backend/runtimes/core/services/other"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/core/transports/gateway/middlewares"
)

type BadgeRouterDependencies struct {
	Service        otherservices.BadgeServiceInterface
	AuthMiddleware gin.HandlerFunc
}

func configureBadgeRoutes(
	router *gin.RouterGroup,
	deps BadgeRouterDependencies,
) {
	authMiddleware := deps.AuthMiddleware
	endpoint := endpoints.NewBadgeEndpoint(deps.Service)
	badgeRoutes := router.Group("/badges")
	{
		badgeRoutes.POST(
			"/graphql/load",
			middlewares.DelegationAuthenticatedMiddleware(
				capi.LoadUserBadgesOperation,
			),
			authMiddleware,
			endpoint.LoadUserBadges,
		)
	}
}
