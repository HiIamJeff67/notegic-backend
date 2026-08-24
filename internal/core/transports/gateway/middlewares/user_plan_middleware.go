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

func UserPlanMiddleware(atLeastUserPlan enums.UserPlan) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserPlanValue, exists := ctx.Get(sharedcontexts.ContextFieldName_User_Plan.String())
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gatewaycontract.Response[struct{}]{
				Version: gatewaycontract.Version,
				Metadata: gatewaycontract.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: exceptions.New(
					"WrongMiddlewareOrder",
					"Context",
					"Middleware",
					"Cannot find the userPlan; make sure AuthMiddleware runs first",
					http.StatusInternalServerError,
					true,
				),
			})
			return
		}

		currentUserPlan, ok := currentUserPlanValue.(enums.UserPlan)
		if !ok {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gatewaycontract.Response[struct{}]{
				Version: gatewaycontract.Version,
				Metadata: gatewaycontract.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: exceptions.New(
					"InvalidType",
					"User",
					"Authorize",
					"The user plan is not in the correct enum type",
					http.StatusInternalServerError,
					true,
				),
			})
			return
		}

		if currentUserPlan == atLeastUserPlan {
			ctx.Next()
			return
		}
		for _, userPlan := range enums.AllUserPlans {
			if userPlan == currentUserPlan {
				ctx.Next()
				return
			}
			if userPlan == atLeastUserPlan {
				ctx.AbortWithStatusJSON(http.StatusForbidden, gatewaycontract.Response[struct{}]{
					Version: gatewaycontract.Version,
					Metadata: gatewaycontract.ResponseMetadata{
						RequestId:   ctx.GetHeader("X-Request-Id"),
						RespondedAt: time.Now(),
					},
					Data: struct{}{},
					Exception: exceptions.New(
						"PermissionDeniedDueToUserPlan",
						"Auth",
						"Authorize",
						fmt.Sprintf("The current user plan of %v does not have access to this operation", currentUserPlan),
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
				"Authorize",
				"Cannot determine user plan access",
				http.StatusInternalServerError,
				true,
			),
		})
	}
}

func AllowedUserPlanMiddleware(allowedPlans []enums.UserPlan) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserPlanValue, exists := ctx.Get(sharedcontexts.ContextFieldName_User_Plan.String())
		currentUserPlan, ok := currentUserPlanValue.(enums.UserPlan)
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
					"Authorize",
					"The authenticated user plan is unavailable",
					http.StatusInternalServerError,
					true,
				),
			})
			return
		}

		if len(allowedPlans) == 0 {
			ctx.Next()
			return
		}
		for _, allowedPlan := range allowedPlans {
			if allowedPlan == currentUserPlan {
				ctx.Next()
				return
			}
		}

		ctx.AbortWithStatusJSON(http.StatusForbidden, gatewaycontract.Response[struct{}]{
			Version: gatewaycontract.Version,
			Metadata: gatewaycontract.ResponseMetadata{
				RequestId:   ctx.GetHeader("X-Request-Id"),
				RespondedAt: time.Now(),
			},
			Data: struct{}{},
			Exception: exceptions.New(
				"PermissionDeniedDueToUserPlan",
				"Auth",
				"Authorize",
				fmt.Sprintf("The current user plan of %v does not have access to this operation", currentUserPlan),
				http.StatusForbidden,
			),
		})
	}
}
