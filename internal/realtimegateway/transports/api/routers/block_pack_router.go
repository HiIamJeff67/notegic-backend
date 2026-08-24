package routers

import (
	"time"

	"github.com/gin-gonic/gin"

	cookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	realtimelease "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/data/redis/realtimelease"
	ratelimit "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/ratelimit"
	endpoints "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/transports/api/endpoints"
	middlewares "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/transports/api/middlewares"
)

func ConfigureBlockPackRoutes(
	router *gin.RouterGroup,
	realtimeLeaseCache *realtimelease.RealtimeLeaseCacheClient,
	accessTokenCookieHandler *cookies.CookieHandler,
	refreshTokenCookieHandler *cookies.CookieHandler,
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
