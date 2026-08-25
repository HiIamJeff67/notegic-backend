package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

func UserRoleMiddleware(atLeastUserRole enums.UserRole) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserRoleValue, exists := ctx.Get(sharedcontexts.ContextFieldName_User_Role.String())
		currentUserRole, ok := currentUserRoleValue.(enums.UserRole)
		if !exists || !ok {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, cgateway.Response[struct{}]{
				Version: cgateway.Version,
				Metadata: cgateway.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: cexceptions.New(
					"AuthenticationContextInvalid",
					"Core",
					"AuthorizeRequest",
					"the authenticated user role is unavailable",
					http.StatusInternalServerError,
					true,
				),
			})
			return
		}

		if currentUserRole == atLeastUserRole {
			ctx.Next()
			return
		}
		for _, userRole := range enums.AllUserRoles {
			if userRole == currentUserRole {
				ctx.Next()
				return
			}
			if userRole == atLeastUserRole {
				ctx.AbortWithStatusJSON(http.StatusForbidden, cgateway.Response[struct{}]{
					Version: cgateway.Version,
					Metadata: cgateway.ResponseMetadata{
						RequestId:   ctx.GetHeader("X-Request-Id"),
						RespondedAt: time.Now(),
					},
					Data: struct{}{},
					Exception: cexceptions.New(
						"PermissionDeniedDueToUserRole",
						"Auth",
						"AuthorizeRequest",
						fmt.Sprintf("the current user role of %v cannot access this operation", currentUserRole),
						http.StatusForbidden,
					),
				})
				return
			}
		}

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   ctx.GetHeader("X-Request-Id"),
				RespondedAt: time.Now(),
			},
			Data: struct{}{},
			Exception: cexceptions.New(
				"AuthenticationContextInvalid",
				"Core",
				"AuthorizeRequest",
				"the authenticated user role is invalid",
				http.StatusInternalServerError,
				true,
			),
		})
	}
}
