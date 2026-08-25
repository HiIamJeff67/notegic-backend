package interceptors

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	sresponsewriter "github.com/HiIamJeff67/notegic-backend/shared/util/responsewriter"

	ratelimit "github.com/HiIamJeff67/notegic-backend/internal/apigateway/ratelimit"
)

// use the reusable buffer pool for interceptors which required multiple response writers
var shareableResponseWritersReusableBufferPool *ratelimit.ReusableBufferPool = ratelimit.NewReusableBufferPool()

// This interceptor is required if some interceptors in the current route require a response writer,
// it will initialize, and manage and write with the response writer,
// passing all the other interceptors that require response writer as parameters.
// ex. use `existingWriter, exist := ctx.Get(ratelimiter.SharedResponseWriterKey)` to get the shared response writer
func ShareableResponseWriterInterceptor(interceptors ...func(string) gin.HandlerFunc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		sharedResponseWriterKey := "SharedResponseWriterKey" + uuid.New().String()
		buffer := shareableResponseWritersReusableBufferPool.Get()
		defer func() {
			buffer.Reset()
			shareableResponseWritersReusableBufferPool.Put(buffer)
		}()
		writer := sresponsewriter.NewResponseWriter(ctx.Writer, buffer)

		ctx.Writer = writer // replace the response writer with the declared writer here
		// so that we can re-write the response after the controller sent the response !!
		// we can successfully do this since the interceptor inherit the gin.ResponseWriter interface,
		// and it also implement Write() and WriteString() methods.
		// Note: they write the content into the `originalBody`,
		// so the field of `originalBody` is the original content from the controllers

		ctx.Set(sharedResponseWriterKey, writer)

		ctx.Next()

		for _, interceptorFactory := range interceptors {
			interceptor := interceptorFactory(sharedResponseWriterKey)
			interceptor(ctx)
		}

		writer.Mutex.Lock()
		defer writer.Mutex.Unlock()

		destination := writer.ResponseWriter.Header()
		for key, val := range writer.Headers {
			destination[key] = val
		}

		if writer.Code > 0 {
			writer.ResponseWriter.WriteHeader(writer.Code)
		}

		writer.ResponseWriter.Write(writer.Body.Bytes())
	}
}
