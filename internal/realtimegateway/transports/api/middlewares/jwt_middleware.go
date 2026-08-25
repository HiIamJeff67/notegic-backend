package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"
	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"
)

func JWTMiddleware(
	accessTokenCookieHandler *scookies.CookieHandler,
	refreshTokenCookieHandler *scookies.CookieHandler,
) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		accessToken, _ := accessTokenCookieHandler.Get(ctx)
		if strings.TrimSpace(accessToken) == "" {
			authorizationHeader := ctx.GetHeader("Authorization")
			if strings.HasPrefix(authorizationHeader, "Bearer ") {
				accessToken = strings.TrimPrefix(authorizationHeader, "Bearer ")
			}
		}
		if claims, err := sharedtokens.ParseAccessToken(accessToken); err == nil {
			if _, err := uuid.Parse(claims.Subject); err == nil {
				ctx.Set(sharedcontexts.ContextFieldName_User_PublicId.String(), claims.Subject)
				ctx.Next()
				return
			}
		}

		refreshToken, _ := refreshTokenCookieHandler.Get(ctx)
		if claims, err := sharedtokens.ParseRefreshToken(refreshToken); err == nil {
			if _, err := uuid.Parse(claims.Subject); err == nil {
				ctx.Set(sharedcontexts.ContextFieldName_User_PublicId.String(), claims.Subject)
				ctx.Next()
				return
			}
		}

		ctx.AbortWithStatus(http.StatusUnauthorized)
	}
}
