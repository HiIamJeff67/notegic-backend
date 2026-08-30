package routers

import (
	"time"

	"github.com/gin-gonic/gin"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	realtimelease "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/data/redis/realtimelease"
	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/ratelimit"
	endpoints "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/transports/api/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/transports/api/middlewares"
)

func ConfigureBlockPackRoutes(
	router *gin.RouterGroup,
	realtimeLeaseCache *realtimelease.RealtimeLeaseCacheClient,
	accessTokenCookieHandler *scookies.CookieHandler,
	refreshTokenCookieHandler *scookies.CookieHandler,
	authorizedRateLimiter *ratelimit.HybridRateLimiter,
) {
	endpoint := endpoints.NewBlockPackEndpoint(realtimeLeaseCache)

	router.GET(
		"/block-pack/:block-pack-id/participants",
		middlewares.JWTMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler),
		middlewares.AuthorizedRateLimitMiddleware(authorizedRateLimiter),
		middlewares.TimeoutMiddleware(3*time.Second),
		middlewares.ApplyTracerMiddleware("getRealtimeBlockPackParticipants"),
		middlewares.ApplyMeterMiddleware("server.requests.realtime.blockPackParticipants"),
		endpoint.GetParticipants,
	)
}
