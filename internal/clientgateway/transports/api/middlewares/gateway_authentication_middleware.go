package middlewares

import (
	"github.com/gin-gonic/gin"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"
)

func GatewayAuthenticationMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler *scookies.CookieHandler) gin.HandlerFunc {
	return JWTMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler)
}
