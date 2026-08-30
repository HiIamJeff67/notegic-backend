package middlewares

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	scookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"
	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"
)

func JWTMiddleware(accessTokenCookieHandler, refreshTokenCookieHandler *scookies.CookieHandler) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Set(sharedcontexts.ContextFieldName_User_Id.String(), nil)
		ctx.Set(sharedcontexts.ContextFieldName_User_PublicId.String(), nil)
		ctx.Set(sharedcontexts.ContextFieldName_User_Name.String(), nil)
		ctx.Set(sharedcontexts.ContextFieldName_User_Email.String(), nil)
		ctx.Set(sharedcontexts.ContextFieldName_AccessToken.String(), nil)
		ctx.Set(sharedcontexts.ContextFieldName_RefreshToken.String(), nil)
		ctx.Set(sharedcontexts.ContextFieldName_CSRFToken.String(), nil)
		ctx.Set(sharedcontexts.ContextFieldName_IsNewTokens.String(), false)
		if csrfToken := ctx.GetHeader("X-CSRF-Token"); strings.TrimSpace(csrfToken) != "" {
			ctx.Set(sharedcontexts.ContextFieldName_CSRFToken.String(), csrfToken)
		}

		accessToken, _ := accessTokenCookieHandler.Get(ctx)
		if strings.TrimSpace(accessToken) == "" {
			authorizationHeader := ctx.GetHeader("Authorization")
			if strings.HasPrefix(authorizationHeader, "Bearer ") {
				accessToken = strings.TrimPrefix(authorizationHeader, "Bearer ")
			}
		}
		if accessToken != "" {
			if claims, err := sharedtokens.ParseAccessToken(accessToken); err == nil {
				if _, err := uuid.Parse(claims.Subject); err == nil {
					ctx.Set(sharedcontexts.ContextFieldName_User_PublicId.String(), claims.Subject)
					ctx.Set(sharedcontexts.ContextFieldName_User_Name.String(), claims.Name)
					ctx.Set(sharedcontexts.ContextFieldName_User_Email.String(), claims.Email)
					ctx.Set(sharedcontexts.ContextFieldName_AccessToken.String(), accessToken)
					ctx.Next()
					return
				}
			}
		}

		refreshToken, _ := refreshTokenCookieHandler.Get(ctx)
		if refreshToken != "" {
			if claims, err := sharedtokens.ParseRefreshToken(refreshToken); err == nil {
				if _, err := uuid.Parse(claims.Subject); err == nil {
					ctx.Set(sharedcontexts.ContextFieldName_User_PublicId.String(), claims.Subject)
					ctx.Set(sharedcontexts.ContextFieldName_User_Name.String(), claims.Name)
					ctx.Set(sharedcontexts.ContextFieldName_User_Email.String(), claims.Email)
					ctx.Set(sharedcontexts.ContextFieldName_RefreshToken.String(), refreshToken)
				}
			}
		}

		ctx.Next()
	}
}
