package middlewares

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"

	gatewaycontract "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	userdata "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/userdata"
)

func CSRFMiddleware(userDataCacheClient *userdata.UserDataCacheClient) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		actorUserName, exception := contexts.GetActorUserName(ctx.Request.Context())
		if exception != nil {
			ctx.AbortWithStatusJSON(http.StatusInternalServerError, gatewaycontract.Response[struct{}]{
				Version:   gatewaycontract.Version,
				Metadata:  gatewaycontract.ResponseMetadata{RequestId: ctx.GetHeader("X-Request-Id"), RespondedAt: time.Now()},
				Data:      struct{}{},
				Exception: exception.ToPublic(),
			})
			return
		}
		request := &gatewaycontract.Request[json.RawMessage]{}
		if ctx.Request.ContentLength != 0 {
			_ = ctx.ShouldBindBodyWithJSON(request)
		}
		csrfToken := request.Tokens.CSRFToken
		if strings.TrimSpace(csrfToken) == "" {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gatewaycontract.Response[struct{}]{
				Version:   gatewaycontract.Version,
				Metadata:  gatewaycontract.ResponseMetadata{RequestId: ctx.GetHeader("X-Request-Id"), RespondedAt: time.Now()},
				Data:      struct{}{},
				Exception: exceptions.New("InvalidCSRFToken", "Token", "ValidateCSRFToken", "the CSRF token is missing or invalid", http.StatusUnauthorized),
			})
			return
		}
		userDataCache, exception := userDataCacheClient.Get(actorUserName)
		if exception != nil {
			ctx.AbortWithStatusJSON(exception.HTTPStatusCode(), gatewaycontract.Response[struct{}]{
				Version:   gatewaycontract.Version,
				Metadata:  gatewaycontract.ResponseMetadata{RequestId: ctx.GetHeader("X-Request-Id"), RespondedAt: time.Now()},
				Data:      struct{}{},
				Exception: exception.ToPublic(),
			})
			return
		}
		expectedToken := userDataCache.CSRFToken
		isPreviousToken := userDataCache.PreviousCSRFToken != "" && csrfToken == userDataCache.PreviousCSRFToken
		if csrfToken != userDataCache.CSRFToken && !isPreviousToken {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gatewaycontract.Response[struct{}]{
				Version:   gatewaycontract.Version,
				Metadata:  gatewaycontract.ResponseMetadata{RequestId: ctx.GetHeader("X-Request-Id"), RespondedAt: time.Now()},
				Data:      struct{}{},
				Exception: exceptions.New("InvalidCSRFToken", "Token", "ValidateCSRFToken", "the CSRF token is invalid", http.StatusUnauthorized),
			})
			return
		}
		if isPreviousToken {
			expectedToken = userDataCache.PreviousCSRFToken
		}

		claims, err := sharedtokens.ValidateCSRFToken(csrfToken, expectedToken)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnauthorized, gatewaycontract.Response[struct{}]{
				Version:   gatewaycontract.Version,
				Metadata:  gatewaycontract.ResponseMetadata{RequestId: ctx.GetHeader("X-Request-Id"), RespondedAt: time.Now()},
				Data:      struct{}{},
				Exception: exceptions.New("InvalidCSRFToken", "Token", "ValidateCSRFToken", "the CSRF token is invalid", http.StatusUnauthorized).WithOrigin(err),
			})
			return
		}
		if isPreviousToken {
			ctx.Set(sharedcontexts.ContextFieldName_IsNewTokens.String(), true)
			ctx.Set(sharedcontexts.ContextFieldName_CSRFToken.String(), userDataCache.CSRFToken)
		} else if sharedtokens.IsCSRFTokenExpiringSoon(claims) {
			newCSRFToken, err := sharedtokens.GenerateCSRFToken(sharedtokens.CSRFTokenClaims{})
			if err != nil {
				ctx.AbortWithStatusJSON(http.StatusInternalServerError, gatewaycontract.Response[struct{}]{
					Version:   gatewaycontract.Version,
					Metadata:  gatewaycontract.ResponseMetadata{RequestId: ctx.GetHeader("X-Request-Id"), RespondedAt: time.Now()},
					Data:      struct{}{},
					Exception: exceptions.New("GenerationFailed", "Token", "GenerateCSRFToken", "failed to generate a CSRF token", http.StatusInternalServerError, true).WithOrigin(err),
				})
				return
			}
			currentCSRFToken, _, exception := userDataCacheClient.RotateCSRFToken(
				actorUserName,
				userDataCache.CSRFToken,
				*newCSRFToken,
			)
			if exception != nil {
				ctx.AbortWithStatusJSON(exception.HTTPStatusCode(), gatewaycontract.Response[struct{}]{
					Version:   gatewaycontract.Version,
					Metadata:  gatewaycontract.ResponseMetadata{RequestId: ctx.GetHeader("X-Request-Id"), RespondedAt: time.Now()},
					Data:      struct{}{},
					Exception: exception.ToPublic(),
				})
				return
			}
			ctx.Set(sharedcontexts.ContextFieldName_IsNewTokens.String(), true)
			if strings.TrimSpace(request.Tokens.AccessToken) != "" {
				ctx.Set(sharedcontexts.ContextFieldName_AccessToken.String(), request.Tokens.AccessToken)
			}
			ctx.Set(sharedcontexts.ContextFieldName_CSRFToken.String(), currentCSRFToken)
		}

		ctx.Next()
	}
}
