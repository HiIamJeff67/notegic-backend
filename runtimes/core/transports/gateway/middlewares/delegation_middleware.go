package middlewares

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
)

func DelegationMiddleware(expectedOperation string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		delegationClaims, err := sharedtokens.ParseDelegationToken(strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer "))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, cgateway.Response[struct{}]{
				Version: cgateway.Version,
				Metadata: cgateway.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: cexceptions.New(
					"Unauthorized",
					"Core",
					"VerifyDelegation",
					"invalid internal delegation credential",
					http.StatusUnauthorized,
				),
			})
			return
		}

		request := &cgateway.Request[json.RawMessage]{}
		if ctx.Request.ContentLength != 0 {
			bodyBindingError := ctx.ShouldBindBodyWithJSON(request)
			if bodyBindingError != nil ||
				request.GetVersion() != cgateway.Version ||
				(expectedOperation != "" && request.GetOperation() != expectedOperation) ||
				delegationClaims.Operation != request.GetOperation() ||
				delegationClaims.RequestId != request.GetMetadata().RequestId {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, cgateway.Response[struct{}]{
					Version: cgateway.Version,
					Metadata: cgateway.ResponseMetadata{
						RequestId:   ctx.GetHeader("X-Request-Id"),
						RespondedAt: time.Now(),
					},
					Data: struct{}{},
					Exception: cexceptions.New(
						"InvalidDelegation",
						"Core",
						"VerifyDelegation",
						"delegation credential does not match the request",
						http.StatusUnauthorized,
					),
				})
				return
			}
		}

		permissions := make([]cenums.AccessControlPermission, 0, len(delegationClaims.AllowedPermissions))
		for _, permissionString := range delegationClaims.AllowedPermissions {
			permission, err := cenums.ConvertStringToAccessControlPermission(permissionString)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, cgateway.Response[struct{}]{
					Version: cgateway.Version,
					Metadata: cgateway.ResponseMetadata{
						RequestId:   ctx.GetHeader("X-Request-Id"),
						RespondedAt: time.Now(),
					},
					Data: struct{}{},
					Exception: cexceptions.New(
						"InvalidDelegation",
						"Core",
						"VerifyDelegation",
						"invalid delegated permission",
						http.StatusUnauthorized,
					),
				})
				return
			}
			permissions = append(permissions, *permission)
		}

		ctx.Request = ctx.Request.WithContext(
			contexts.WithDelegationMetadata(ctx.Request.Context(), delegationClaims),
		)
		if len(permissions) > 0 {
			ctx.Request = ctx.Request.WithContext(
				contexts.WithAllowedPermissions(ctx.Request.Context(), permissions),
			)
		}
		ctx.Next()
	}
}
