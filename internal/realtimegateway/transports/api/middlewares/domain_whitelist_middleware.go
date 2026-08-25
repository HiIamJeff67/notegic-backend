package middlewares

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
)

func isAllowedOrigin(origin string, allowedDomains []string) bool {
	for _, allowed := range allowedDomains {
		if origin == allowed || (origin[len(origin)-1] == '/' && origin[0:len(origin)-1] == allowed) {
			return true
		}
	}
	return false
}

func isAllowedReferer(referer string, allowedDomains []string) bool {
	for _, allowed := range allowedDomains {
		if referer == allowed || (referer[len(referer)-1] == '/' && referer[0:len(referer)-1] == allowed) {
			return true
		}
	}
	return false
}

func DomainWhiteListMiddleware(allowedDomains []string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		origin := ctx.GetHeader("Origin")
		if origin != "" {
			if !isAllowedOrigin(origin, allowedDomains) {
				logs.NotegicLogger.Alert(ctx.Request.Context(), nil, fmt.Sprintf("Blocked Origin: %s, allowed origins: ", origin))
				for _, domain := range allowedDomains {
					logs.NotegicLogger.Alert(ctx.Request.Context(), nil, domain)
				}
				ctx.AbortWithStatusJSON(http.StatusForbidden,
					exceptionwriter.GetGinH(cexceptions.New(
						"PermissionDeniedDueToInvalidRequestOriginDomain",
						"Auth",
						"Authorize",
						fmt.Sprintf("The current request origin domain of %s is invalid", origin),
						http.StatusForbidden,
					)))
				return
			}
		}

		referer := ctx.GetHeader("Referer")
		if referer != "" && origin == "" {
			if !isAllowedReferer(referer, allowedDomains) {
				logs.NotegicLogger.Alert(ctx.Request.Context(), nil, fmt.Sprintf("Blocked Referer: %s", referer))
				ctx.AbortWithStatusJSON(http.StatusForbidden,
					exceptionwriter.GetGinH(cexceptions.New(
						"PermissionDeniedDueToInvalidRequestOriginDomain",
						"Auth",
						"Authorize",
						fmt.Sprintf("The current request origin domain of %s is invalid", referer),
						http.StatusForbidden,
					)))
				return
			}
		}

		ctx.Next()
	}
}
