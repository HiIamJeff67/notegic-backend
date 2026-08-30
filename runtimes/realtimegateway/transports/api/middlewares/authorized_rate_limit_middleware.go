package middlewares

import (
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/ratelimit"
)

func AuthorizedRateLimitMiddleware(rateLimiter *ratelimit.HybridRateLimiter) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if rateLimiter == nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
				"RateLimiterRequired",
				"RealtimeGateway",
				"RateLimit",
				"The authorized rate limiter is not configured",
				http.StatusInternalServerError,
				true,
			), ctx)
			return
		}

		userPublicId, exists := ctx.Get(sharedcontexts.ContextFieldName_User_PublicId.String())
		if !exists || userPublicId == nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
				"WrongMiddlewareOrder",
				"Context",
				"Middleware",
				"Cannot find the user public ID; JWTMiddleware() must run first",
				http.StatusInternalServerError,
				true,
			), ctx)
			return
		}

		publicId, err := uuid.Parse(fmt.Sprint(userPublicId))
		if err != nil {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.InvalidInput("User").WithOrigin(err), ctx)
			return
		}

		allowed, remaining := rateLimiter.AllowByUserId(publicId)
		ctx.Header("X-RateLimit-Limit", strconv.Itoa(int(rateLimiter.UserLimit)))
		ctx.Header("X-RateLimit-Remaining", strconv.Itoa(int(remaining)))
		ctx.Header("X-RateLimit-Reset", strconv.FormatInt(time.Now().Add(rateLimiter.WindowDuration).Unix(), 10))
		ctx.Header("X-RateLimit-Window", rateLimiter.WindowDuration.String())
		ctx.Header("X-RateLimit-Policy", "hybrid-token-bucket")
		if !allowed {
			sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
				"PermissionDeniedDueToTooManyRequests",
				"RealtimeGateway",
				"Authorize",
				"Too many requests; please wait before trying again",
				http.StatusTooManyRequests,
			), ctx, "server.responses.failed.rateLimit")
			return
		}

		ctx.Next()
	}
}
