package middlewares

import (
	"bytes"
	"encoding/json"

	"github.com/gin-gonic/gin"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
	sresponsewriter "github.com/HiIamJeff67/notegic-backend/shared/util/responsewriter"
)

func TokenResponseMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		body := &bytes.Buffer{}
		originalWriter := ctx.Writer
		bufferedWriter := sresponsewriter.NewResponseWriter(originalWriter, body)
		ctx.Writer = bufferedWriter

		ctx.Next()

		isNewTokens, exists := ctx.Get(sharedcontexts.ContextFieldName_IsNewTokens.String())
		if exists && isNewTokens == true && bufferedWriter.Status() < 400 {
			accessToken, accessTokenExists := ctx.Get(sharedcontexts.ContextFieldName_AccessToken.String())
			csrfToken, csrfTokenExists := ctx.Get(sharedcontexts.ContextFieldName_CSRFToken.String())
			if accessTokenExists && csrfTokenExists {
				response := &cgateway.Response[json.RawMessage]{}
				if err := json.Unmarshal(body.Bytes(), response); err == nil {
					accessTokenString, accessTokenIsString := accessToken.(string)
					csrfTokenString, csrfTokenIsString := csrfToken.(string)
					if accessTokenIsString && csrfTokenIsString {
						response.Tokens = &cgateway.Tokens{
							AccessToken: accessTokenString,
							CSRFToken:   csrfTokenString,
						}
						if payload, err := json.Marshal(response); err == nil {
							body.Reset()
							_, _ = body.Write(payload)
						}
					}
				}
			}
		}

		ctx.Writer = originalWriter
		_ = bufferedWriter.FlushToOriginalWriter()
	}
}
