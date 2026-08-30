package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	gatewayconfig "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/configs"
	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/ratelimit"
)

func InitUnauthorizedRateLimiter(config gatewayconfig.RateLimitConfig) *ratelimit.HybridRateLimiter {
	limiter := ratelimit.NewHybridRateLimiter(config, false)
	slogs.NotegicLogger.Info(context.Background(), fmt.Sprintf("Unauthorized rate limiter initialized with rate: %v, burst: %d, user limit: %d, window: %v", config.RateLimit, config.Burst, config.UserLimit, config.WindowDuration))
	return limiter
}

func UnauthorizedRateLimitMiddleware(rateLimiter *ratelimit.HybridRateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if rateLimiter == nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
				"RateLimiterRequired",
				"Gateway",
				"RateLimit",
				"The unauthorized rate limiter is not configured",
				http.StatusInternalServerError,
				true,
			), ctx)
			return
		}

		fingerprint := ctx.ClientIP()

		allowed, remaining := rateLimiter.AllowByFingerprint(fingerprint)
		if !allowed {
			setRateLimitHeaders(ctx, remaining, rateLimiter)
			slogs.NotegicLogger.Debug(ctx.Request.Context(), fmt.Sprintf("Rate limit exceeded for fingerprint: %s", fingerprint))
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
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

func setRateLimitHeaders(ctx *gin.Context, remaining int32, limiter *ratelimit.HybridRateLimiter) {
	// standard information
	ctx.Header("X-RateLimit-Limit", strconv.Itoa(int(limiter.UserLimit)))
	ctx.Header("X-RateLimit-Remaining", strconv.Itoa(int(remaining)))
	ctx.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(limiter.WindowDuration).Unix(), 10))

	// extra information
	ctx.Header("X-RateLimit-Window", limiter.WindowDuration.String())
	ctx.Header("X-RateLimit-Policy", "hybrid-token-bucket")
}
