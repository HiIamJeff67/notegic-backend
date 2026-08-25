package interceptors

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	cookies "github.com/HiIamJeff67/notegic-backend/shared/cookies"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"

	responsewriter "github.com/HiIamJeff67/notegic-backend/shared/util/responsewriter"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/clientgateway/contexts"
)

// Adds non-sensitive refresh metadata to the public response.
// The access token is written only to the Gateway cookie and never to JSON.
// Note: this interceptor should be placed below the `JWTMiddleware`,
// so that it can access the `AccessToken` and `CSRFToken` in the context field
func RefreshTokenInterceptor(accessTokenCookieHandler *cookies.CookieHandler) func(string) gin.HandlerFunc {
	return func(responseWriterKey string) gin.HandlerFunc {
		return func(ctx *gin.Context) {
			var writer *responsewriter.ResponseWriter
			existingWriter, exist := ctx.Get(responseWriterKey)
			if !exist || existingWriter == nil {
				cexceptions.New(
					"WrongInterceptorOrder",
					"Context",
					"Interceptor",
					"Cannot find the existing response writer, "+
						"please make sure to call the ShareableResponseWriterInterceptor() and pass RefreshTokenInterceptor() as one of the parameters",
					http.StatusInternalServerError,
					true,
				)
				return
			}
			writer = existingWriter.(*responsewriter.ResponseWriter)

			ctx.Next() // execute the following first

			if writer.IsTimeout {
				return
			}

			if writer.ResponseWriter.Written() || writer.Status() >= 400 {
				return
			}

			if ctx.Writer.Status() >= 400 {
				return
			}

			IsNewTokens, exception := contexts.GetAndConvertContextFieldToBoolean(ctx, sharedcontexts.ContextFieldName_IsNewTokens)
			if exception != nil || IsNewTokens == nil || !*IsNewTokens {
				return
			}

			var originalResponse cgateway.ClientResponse[json.RawMessage]
			if err := json.Unmarshal(writer.Body.Bytes(), &originalResponse); err != nil {
				return
			}

			accessToken, exceptionOfAccessToken := contexts.GetAndConvertContextFieldToString(ctx, sharedcontexts.ContextFieldName_AccessToken)
			csrfToken, exceptionOfCSRFToken := contexts.GetAndConvertContextFieldToString(ctx, sharedcontexts.ContextFieldName_CSRFToken)
			if exceptionOfAccessToken != nil || exceptionOfCSRFToken != nil || accessToken == nil || csrfToken == nil {
				return
			}

			accessTokenCookieHandler.Set(ctx, *accessToken)
			ctx.Header("X-CSRF-Token", *csrfToken)
			originalResponse.RefreshableTokens = &cgateway.RefreshableTokens{
				NewCSRFToken: *csrfToken,
			}
			modifiedResponse, err := json.Marshal(originalResponse)
			if err != nil {
				return
			}

			writer.Mutex.Lock()
			writer.Body.Reset()
			writer.Body.Write(modifiedResponse)
			writer.Mutex.Unlock()
		}
	}
}
