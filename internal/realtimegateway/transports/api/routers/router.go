package routers

import (
	"github.com/gin-gonic/gin"

	realtimelease "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/data/redis/realtimelease"
	ratelimit "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/ratelimit"
	cookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"
)

func ConfigureRoutes(
	router *gin.RouterGroup,
	realtimeLeaseCache *realtimelease.RealtimeLeaseCacheClient,
	accessTokenCookieHandler *cookies.CookieHandler,
	refreshTokenCookieHandler *cookies.CookieHandler,
	authorizedRateLimiter *ratelimit.HybridRateLimiter,
) {
	ConfigureBlockPackRoutes(router, realtimeLeaseCache, accessTokenCookieHandler, refreshTokenCookieHandler, authorizedRateLimiter)
}
