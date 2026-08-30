package middlewares

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
	sharedtokens "github.com/HiIamJeff67/notegic-backend/shared/tokens"
)

func DelegationAuthenticatedMiddleware(operation string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if operation == "" {
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		authorizationHeader := ctx.GetHeader("Authorization")
		if !strings.HasPrefix(authorizationHeader, "Bearer ") {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, err := sharedtokens.ParseDelegationToken(strings.TrimPrefix(authorizationHeader, "Bearer "))
		if err != nil || claims.UserSubject == "" || claims.Operation != operation {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		userPublicId, err := uuid.Parse(claims.UserSubject)
		if err != nil || userPublicId == uuid.Nil {
			ctx.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		ctx.Set(sharedcontexts.ContextFieldName_User_PublicId.String(), userPublicId)
		ctx.Next()
	}
}
