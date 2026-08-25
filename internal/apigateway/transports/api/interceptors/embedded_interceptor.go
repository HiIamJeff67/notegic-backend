package interceptors

import (
	"encoding/json"
	"net/http"

	"github.com/gin-gonic/gin"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"

	responsewriter "github.com/HiIamJeff67/notegic-backend/shared/util/responsewriter"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/apigateway/contexts"
)

// To add additional field to the response with possibly embedded data that is required for the frontend.
// ex. the frontend require a publicId to indicate the user in their local database across APIs
func EmbeddedInterceptor(responseWriterKey string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var writer *responsewriter.ResponseWriter
		existingWriter, exist := ctx.Get(responseWriterKey)
		if !exist || existingWriter == nil {
			cexceptions.New(
				"WrongInterceptorOrder",
				"Context",
				"Interceptor",
				"Cannot find the existing response writer, "+
					"please make sure to call the ShareableResponseWriterInterceptor() and pass EmbeddedInterceptor() as one of the parameters",
				http.StatusInternalServerError,
				true,
			)
			return
		}
		writer = existingWriter.(*responsewriter.ResponseWriter)

		ctx.Next()

		if writer.IsTimeout {
			return
		}

		if writer.ResponseWriter.Written() || writer.Status() >= 400 {
			return
		}

		if ctx.Writer.Status() >= 400 {
			return
		}

		var originalResponse cgateway.ClientResponse[json.RawMessage]
		if err := json.Unmarshal(writer.Body.Bytes(), &originalResponse); err != nil {
			return
		}

		publicId, exception := contexts.GetAndConvertContextFieldToUUID(ctx, sharedcontexts.ContextFieldName_User_PublicId)
		if exception != nil || publicId == nil {
			return
		}

		originalResponse.Embedded = &cgateway.EmbeddedResponse{
			PublicId: *publicId,
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
