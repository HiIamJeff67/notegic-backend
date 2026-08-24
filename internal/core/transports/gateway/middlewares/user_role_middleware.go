package middlewares

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"

	gatewaycontract "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

func UserRoleMiddleware(atLeastUserRole enums.UserRole) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserRoleValue, exists := ctx.Get(sharedcontexts.ContextFieldName_User_Role.String())
		currentUserRole, ok := currentUserRoleValue.(enums.UserRole)
		if !exists || !ok {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gatewaycontract.Response[struct{}]{
				Version: gatewaycontract.Version,
				Metadata: gatewaycontract.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: exceptions.New(
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
				ctx.AbortWithStatusJSON(http.StatusForbidden, gatewaycontract.Response[struct{}]{
					Version: gatewaycontract.Version,
					Metadata: gatewaycontract.ResponseMetadata{
						RequestId:   ctx.GetHeader("X-Request-Id"),
						RespondedAt: time.Now(),
					},
					Data: struct{}{},
					Exception: exceptions.New(
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

		ctx.AbortWithStatusJSON(http.StatusInternalServerError, gatewaycontract.Response[struct{}]{
			Version: gatewaycontract.Version,
			Metadata: gatewaycontract.ResponseMetadata{
				RequestId:   ctx.GetHeader("X-Request-Id"),
				RespondedAt: time.Now(),
			},
			Data: struct{}{},
			Exception: exceptions.New(
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
