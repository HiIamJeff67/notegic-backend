package middlewares

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	gatewaycontract "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	apikeycache "github.com/HiIamJeff67/notegic-backend/internal/core/data/redis/apikey"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"
)

const APIKeyHeader = "X-API-Key"

// APIKeyMiddleware authenticates an API request after the internal delegation
// credential has been verified. The raw key is hashed in memory and never
// enters logs, context, Redis values, or the database.
func APIKeyMiddleware(
	apiKeyRepository repositories.APIKeyRepositoryInterface,
	userRepository repositories.UserRepositoryInterface,
	cacheClients ...*apikeycache.APIKeyCacheClient,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if apiKeyRepository == nil || userRepository == nil {
			abortAPIKey(ctx, "API key authentication is not configured", http.StatusInternalServerError)
			return
		}

		if source, exception := contexts.GetGatewaySource(ctx.Request.Context()); exception == nil && source != sharedtokens.GatewaySourceAPI {
			abortAPIKey(ctx, "the request source does not allow API key authentication", http.StatusUnauthorized)
			return
		}
		rawKey := strings.TrimSpace(ctx.GetHeader(APIKeyHeader))
		if rawKey == "" {
			abortAPIKey(ctx, "an API key is required", http.StatusUnauthorized)
			return
		}

		keyHash := sharedtokens.HashAPIKey(rawKey)
		var cacheClient *apikeycache.APIKeyCacheClient
		if len(cacheClients) > 0 {
			cacheClient = cacheClients[0]
		}
		if cacheClient != nil {
			if cached, cacheException := cacheClient.Get(keyHash); cacheException == nil && cached != nil {
				if !isAPIKeyActive(cached.RevokedAt, cached.ExpiresAt, time.Now()) {
					abortAPIKey(ctx, "the API key is expired or revoked", http.StatusUnauthorized)
					return
				}
				setAPIKeyContext(ctx, cached.Id, cached.UserId, cached.UserPublicId)
				ctx.Next()
				return
			}
		}
		apiKey, exception := apiKeyRepository.GetOneByKeyHash(keyHash)
		if exception != nil || apiKey == nil {
			abortAPIKey(ctx, "the API key is invalid", http.StatusUnauthorized)
			return
		}
		now := time.Now()
		if !isAPIKeyActive(apiKey.RevokedAt, apiKey.ExpiresAt, now) {
			abortAPIKey(ctx, "the API key is expired or revoked", http.StatusUnauthorized)
			return
		}

		user, exception := userRepository.GetOneById(apiKey.UserId, nil)
		if exception != nil || user == nil || user.Id == uuid.Nil || user.PublicId == uuid.Nil {
			abortAPIKey(ctx, "the API key owner is invalid", http.StatusUnauthorized)
			return
		}
		setAPIKeyContext(ctx, apiKey.Id, user.Id, user.PublicId)
		if cacheClient != nil {
			_ = cacheClient.Set(keyHash, apikeycache.APIKeyCache{
				Id:           apiKey.Id,
				UserId:       user.Id,
				UserPublicId: user.PublicId,
				ExpiresAt:    apiKey.ExpiresAt,
				RevokedAt:    apiKey.RevokedAt,
			})
		}
		_ = apiKeyRepository.MarkUsed(apiKey.Id, now)
		ctx.Next()
	}
}

func isAPIKeyActive(revokedAt, expiresAt *time.Time, now time.Time) bool {
	return revokedAt == nil && (expiresAt == nil || expiresAt.After(now))
}

func setAPIKeyContext(ctx *gin.Context, apiKeyId, userId, userPublicId uuid.UUID) {
	requestContext := contexts.WithGatewaySource(ctx.Request.Context(), sharedtokens.GatewaySourceAPI)
	requestContext = contexts.WithAuthMethod(requestContext, sharedtokens.AuthMethodAPIKey)
	requestContext = contexts.WithAPIKeyId(requestContext, apiKeyId.String())
	requestContext = contexts.WithActorUserId(requestContext, userId)
	requestContext = contexts.WithActorUserPublicId(requestContext, userPublicId)
	ctx.Request = ctx.Request.WithContext(requestContext)
}

func abortAPIKey(ctx *gin.Context, message string, status int) {
	ctx.AbortWithStatusJSON(status, gatewaycontract.Response[struct{}]{
		Version: gatewaycontract.Version,
		Metadata: gatewaycontract.ResponseMetadata{
			RequestId:   ctx.GetHeader("X-Request-Id"),
			RespondedAt: time.Now(),
		},
		Data: struct{}{},
		Exception: exceptions.New(
			"Unauthorized",
			"Core",
			"AuthenticateAPIKey",
			message,
			status,
		),
	})
}
