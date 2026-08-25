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

func UserPlanMiddleware(atLeastUserPlan enums.UserPlan) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		currentUserPlanValue, exists := ctx.Get(sharedcontexts.ContextFieldName_User_Plan.String())
		if !exists {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, cgateway.Response[struct{}]{
				Version: cgateway.Version,
				Metadata: cgateway.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: cexceptions.New(
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
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, cgateway.Response[struct{}]{
				Version: cgateway.Version,
				Metadata: cgateway.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: cexceptions.New(
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
				ctx.AbortWithStatusJSON(http.StatusForbidden, cgateway.Response[struct{}]{
					Version: cgateway.Version,
					Metadata: cgateway.ResponseMetadata{
						RequestId:   ctx.GetHeader("X-Request-Id"),
						RespondedAt: time.Now(),
					},
					Data: struct{}{},
					Exception: cexceptions.New(
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

		ctx.AbortWithStatusJSON(http.StatusForbidden, cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId:   ctx.GetHeader("X-Request-Id"),
				RespondedAt: time.Now(),
			},
			Data: struct{}{},
			Exception: cexceptions.New(
				"PermissionDeniedDueToUserPlan",
				"Auth",
				"Authorize",
				fmt.Sprintf("The current user plan of %v does not have access to this operation", currentUserPlan),
				http.StatusForbidden,
			),
		})
	}
}
