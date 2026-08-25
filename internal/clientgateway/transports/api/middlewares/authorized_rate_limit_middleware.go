package middlewares

import (
	"context"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"

	exceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	gatewayconfig "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/configs"
	contexts "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/contexts"
	ratelimit "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/ratelimit"
)

func InitAuthorizedRateLimiter(config gatewayconfig.RateLimitConfig) *ratelimit.HybridRateLimiter {
	limiter := ratelimit.NewHybridRateLimiter(config, true)
	logs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Authorized rate limiter initialized with rate: %v, burst: %d, user limit: %d, window: %v", config.RateLimit, config.Burst, config.UserLimit, config.WindowDuration))
	return limiter
}

func AuthorizedRateLimitMiddleware(rateLimiter *ratelimit.HybridRateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if rateLimiter == nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
				"RateLimiterRequired",
				"Gateway",
				"RateLimit",
				"The authorized rate limiter is not configured",
				http.StatusInternalServerError,
				true,
			), ctx)
			return
		}

		userId, exception := contexts.GetAndConvertContextFieldToUUID(ctx, sharedcontexts.ContextFieldName_User_Id)
		if exception != nil || userId == nil {
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
				"WrongMiddlewareOrder",
				"Context",
				"Middleware",
				"Cannot find the userId, "+
					"please make sure JWTMiddleware() is placed before AuthorizedRateLimitMiddleware()",
				http.StatusInternalServerError,
				true,
			), ctx)
			return
		}

		allowed, remaining := rateLimiter.AllowByUserId(*userId)
		if !allowed {
			setRateLimitHeaders(ctx, remaining, rateLimiter)
			logs.NotegicLogger.Debug(ctx.Request.Context(), fmt.Sprintf("Rate limit exceeded for user: %s", userId.String()))
			exceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
				"PermissionDeniedDueToTooManyRequests",
				"Auth",
				"Authorize",
				"Too many requests; please wait before trying again",
				http.StatusTooManyRequests,
			), ctx, "server.responses.failed.rateLimit")
			return
		}

		setRateLimitHeaders(ctx, remaining, rateLimiter)

		ctx.Next()
	}
}
