package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	gatewaycontexts "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/contexts"
)

// KeyMiddleware is the APIGateway edge check. It intentionally verifies only
// presence and format; ownership and revocation are authoritative in Core's
// APIKeyMiddleware after the delegation credential is verified.
func KeyMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if ctx.Request.Method == http.MethodOptions {
			ctx.Next()
			return
		}
		key := strings.TrimSpace(ctx.GetHeader("X-API-Key"))
		if key == "" || sharedtokens.ValidateAPIKeyFormat(key) != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, cgateway.Response[struct{}]{
				Version: cgateway.Version,
				Metadata: cgateway.ResponseMetadata{
					RequestId:   ctx.GetHeader("X-Request-Id"),
					RespondedAt: time.Now(),
				},
				Data: struct{}{},
				Exception: cexceptions.New(
					"Unauthorized",
					"Gateway",
					"AuthenticateAPIKey",
					"a valid API key is required",
					http.StatusUnauthorized,
				),
			})
			return
		}
		gatewaycontexts.SetGatewaySource(ctx, sharedtokens.GatewaySourceAPI)
		ctx.Next()
	}
}
