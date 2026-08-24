package middlewares

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	gatewaycontract "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

func DelegationMiddleware(expectedOperation string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		delegationClaims, err := sharedtokens.ParseDelegationToken(strings.TrimPrefix(ctx.GetHeader("Authorization"), "Bearer "))
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gatewaycontract.Response[struct{}]{
				Version: gatewaycontract.Version,
				Metadata: gatewaycontract.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: exceptions.New(
					"Unauthorized",
					"Core",
					"VerifyDelegation",
					"invalid internal delegation credential",
					http.StatusUnauthorized,
				),
			})
			return
		}

		request := &gatewaycontract.Request[json.RawMessage]{}
		if ctx.Request.ContentLength != 0 {
			bodyBindingError := ctx.ShouldBindBodyWithJSON(request)
			if bodyBindingError != nil ||
				request.GetVersion() != gatewaycontract.Version ||
				(expectedOperation != "" && request.GetOperation() != expectedOperation) ||
				delegationClaims.Operation != request.GetOperation() ||
				delegationClaims.RequestId != request.GetMetadata().RequestId {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gatewaycontract.Response[struct{}]{
					Version: gatewaycontract.Version,
					Metadata: gatewaycontract.ResponseMetadata{
						RequestId:   ctx.GetHeader("X-Request-Id"),
						RespondedAt: time.Now(),
					},
					Data: struct{}{},
					Exception: exceptions.New(
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

		permissions := make([]enums.AccessControlPermission, 0, len(delegationClaims.AllowedPermissions))
		for _, permissionString := range delegationClaims.AllowedPermissions {
			permission, err := enums.ConvertStringToAccessControlPermission(permissionString)
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusUnauthorized, gatewaycontract.Response[struct{}]{
					Version: gatewaycontract.Version,
					Metadata: gatewaycontract.ResponseMetadata{
						RequestId:   ctx.GetHeader("X-Request-Id"),
						RespondedAt: time.Now(),
					},
					Data: struct{}{},
					Exception: exceptions.New(
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
