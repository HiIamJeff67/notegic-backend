package routers

import (
	"github.com/gin-gonic/gin"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	realtimelease "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/data/redis/realtimelease"
	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/ratelimit"
)

func ConfigureRoutes(
	router *gin.RouterGroup,
	realtimeLeaseCache *realtimelease.RealtimeLeaseCacheClient,
	accessTokenCookieHandler *scookies.CookieHandler,
	refreshTokenCookieHandler *scookies.CookieHandler,
	authorizedRateLimiter *ratelimit.HybridRateLimiter,
) {
	ConfigureBlockPackRoutes(router, realtimeLeaseCache, accessTokenCookieHandler, refreshTokenCookieHandler, authorizedRateLimiter)
}
