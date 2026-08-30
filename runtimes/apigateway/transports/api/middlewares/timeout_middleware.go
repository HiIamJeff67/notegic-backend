package middlewares

import (
	"context"
	"fmt"
	"net/http"
	"runtime/debug"
	"time"

	"github.com/gin-gonic/gin"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"
	sexceptionwriter "github.com/HiIamJeff67/notegic-backend/shared/util/exceptionwriter"
	sresponsewriter "github.com/HiIamJeff67/notegic-backend/shared/util/responsewriter"

	ratelimit "github.com/HiIamJeff67/notegic-backend/runtimes/apigateway/ratelimit"
)

// use reusable buffer pool for timeout response writer to storing the current response of the handlers
var timeoutReusableBufferPool *ratelimit.ReusableBufferPool = ratelimit.NewReusableBufferPool()

func TimeoutMiddleware(timeout time.Duration) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		originalWriter := ctx.Writer

		currentBufferPool := timeoutReusableBufferPool.Get()
		currentBufferPool.Reset()
		defer func() {
			currentBufferPool.Reset()
			timeoutReusableBufferPool.Put(currentBufferPool)
		}()

		writer := sresponsewriter.NewResponseWriter(originalWriter, currentBufferPool)
		ctx.Writer = writer

		ctxCopy := ctx.Copy()
		ctxCopy.Writer = writer

		timeoutCtx, cancel := context.WithTimeout(ctx.Request.Context(), timeout)
		defer cancel()

		ctxCopy.Request = ctxCopy.Request.WithContext(timeoutCtx)

		// ctx     uses originalWriter
		// ctxCopy uses writer (passing to the next handlers to make them write response in the buffer)

		done := make(chan struct{}, 1)
		panicChannel := make(chan stypes.PanicInfo, 1)

		go func() {
			defer func() {
				if p := recover(); p != nil {
					panicChannel <- stypes.PanicInfo{
						Value: p,
						Stack: debug.Stack(),
					}
				}
			}()

			ctx.Next()
			done <- struct{}{}
		}()

		select {
		case panicInfo := <-panicChannel:
			writer.Mutex.Lock()
			writer.FreeBuffer() // clear the buffer, this will destroy the context field stored by other middlewares
			writer.Mutex.Unlock()
			ctx.Writer = originalWriter

			if gin.IsDebugging() {
				ctx.Error(fmt.Errorf("%v", panicInfo.Value))
				ctx.Writer.WriteHeader(http.StatusInternalServerError)
				fmt.Fprintf(ctx.Writer, "panic caught: %v\n", panicInfo.Value)
				ctx.Writer.Write([]byte("Panic stack trace:\n"))
				ctx.Writer.Write(panicInfo.Stack)
			}

			sexceptionwriter.SafelyAbortAndResponseWithJSON(
				cexceptions.New(
					"FatalPanic",
					"General",
					"Respond",
					"An unexpected panic occurred",
					http.StatusInternalServerError,
					true,
				).WithOrigin(fmt.Errorf("%v", panicInfo.Value)),
				ctx,
			)
		case <-done:
			writer.Mutex.Lock()
			defer writer.Mutex.Unlock()
			if writer.IsTimeout || writer.ResponseWriter.Written() {
				return
			}

			destination := writer.ResponseWriter.Header()
			for key, val := range writer.Headers {
				destination[key] = val
			}

			if writer.Code != 0 {
				writer.ResponseWriter.WriteHeader(writer.Code)
			}

			if currentBufferPool.Len() > 0 {
				if _, err := writer.ResponseWriter.Write(currentBufferPool.Bytes()); err != nil {
					sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
						"FatalPanic",
						"General",
						"Respond",
						"An unexpected panic occurred",
						http.StatusInternalServerError,
						true,
					).WithOrigin(err), ctx)
				}
			}

			return
		case <-timeoutCtx.Done():
			// logs.NotegicLogger.Alert(context.Background(), nil, "Timeout (timeoutCtx.Done())")
			writer.Mutex.Lock()
			writer.IsTimeout = true
			writer.FreeBuffer() // clear the buffer, this will destroy the context field stored by other middlewares
			writer.Mutex.Unlock()

			if !writer.ResponseWriter.Written() {
				writer.ResponseWriter.Header().Set("X-Request-Timeout", timeout.String())
				exception := cexceptions.New(
					"Timeout",
					"General",
					"Respond",
					fmt.Sprintf("Request timed out after %v", timeout),
					http.StatusRequestTimeout,
				)
				writer.ResponseWriter.WriteHeader(exception.HTTPStatusCode())
				timeoutResponseBody, err := sexceptionwriter.GetResponseJSONBytes(exception)
				if err != nil {
					sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
						"FatalPanic",
						"General",
						"Respond",
						"An unexpected panic occurred",
						http.StatusInternalServerError,
						true,
					).WithOrigin(err), ctx)
				}
				if _, err := writer.ResponseWriter.Write(timeoutResponseBody); err != nil {
					sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
						"FatalPanic",
						"General",
						"Respond",
						"An unexpected panic occurred",
						http.StatusInternalServerError,
						true,
					).WithOrigin(err), ctx, "server.responses.failed.timeout")
				}
			}

			return
		case <-time.After(timeout):
			// logs.NotegicLogger.Alert(context.Background(), nil, "Timeout (time.After)")
			writer.Mutex.Lock()
			writer.IsTimeout = true
			writer.FreeBuffer() // clear the buffer, this will destroy the context field stored by other middlewares
			writer.Mutex.Unlock()

			if !writer.ResponseWriter.Written() {
				writer.ResponseWriter.Header().Set("X-Request-Timeout", timeout.String())
				exception := cexceptions.New(
					"Timeout",
					"General",
					"Respond",
					fmt.Sprintf("Request timed out after %v", timeout),
					http.StatusRequestTimeout,
				)
				writer.ResponseWriter.WriteHeader(exception.HTTPStatusCode())
				timeoutResponseBody, err := sexceptionwriter.GetResponseJSONBytes(exception)
				if err != nil {
					sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
						"FatalPanic",
						"General",
						"Respond",
						"An unexpected panic occurred",
						http.StatusInternalServerError,
						true,
					).WithOrigin(err), ctx)
				}
				if _, err := writer.ResponseWriter.Write(timeoutResponseBody); err != nil {
					sexceptionwriter.SafelyAbortAndResponseWithJSON(cexceptions.New(
						"FatalPanic",
						"General",
						"Respond",
						"An unexpected panic occurred",
						http.StatusInternalServerError,
						true,
					).WithOrigin(err), ctx, "server.responses.failed.timeout")
				}
			}

			return
		}
	}
}
