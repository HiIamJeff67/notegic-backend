package exceptionwriter

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/99designs/gqlgen/graphql"
	"github.com/gin-gonic/gin"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/gqlerror"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	metrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"
	traces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"
)

func IncrementMeter(exception *cexceptions.Exception, ctx *gin.Context, names ...string) {
	if metrics.NotegicMeter == nil {
		return
	}

	isTotalCounted := false
	for _, name := range names {
		if name == "server.responses.failed.total" {
			isTotalCounted = true
		}
		metrics.NotegicMeter.Count(ctx, name, 1)
	}
	if !isTotalCounted {
		metrics.NotegicMeter.Count(ctx, "server.responses.failed.total", 1)
	}
}

func GetGinH(exception *cexceptions.Exception) gin.H {
	exception = exception.ToPublic()

	return gin.H{
		"reason":    exception.Reason,
		"domain":    exception.Domain,
		"operation": exception.Operation,
		"message":   exception.Message,
		"retryable": exception.Retryable,
	}
}

func GetGinHBytes(exception *cexceptions.Exception) ([]byte, error) {
	return json.Marshal(GetGinH(exception))
}

func GetResponseJSONBytes(exception *cexceptions.Exception) ([]byte, error) {
	return json.Marshal(cgateway.ClientResponse[any]{
		Success:   false,
		Data:      nil,
		Exception: ToPublic(context.Background(), exception),
	})
}

func getRequestContext(ctx *gin.Context) context.Context {
	if ctx != nil && ctx.Request != nil {
		return ctx.Request.Context()
	}

	return context.Background()
}

func ResponseWithJSON(exception *cexceptions.Exception, ctx *gin.Context, names ...string) {
	publicException := ToPublic(getRequestContext(ctx), exception)
	IncrementMeter(publicException, ctx, names...)
	if exception != nil &&
		exception.HTTPStatusCode() >= http.StatusInternalServerError &&
		traces.NotegicTracer != nil {
		traces.NotegicTracer.RecordError(ctx, exception)
	}

	ctx.JSON(publicException.HTTPStatusCode(), cgateway.ClientResponse[any]{
		Success:   false,
		Data:      nil,
		Exception: publicException,
	})
}

func SafelyResponseWithJSON(exception *cexceptions.Exception, ctx *gin.Context, names ...string) {
	publicException := ToPublic(getRequestContext(ctx), exception)
	IncrementMeter(publicException, ctx, names...)
	if exception != nil &&
		exception.HTTPStatusCode() >= http.StatusInternalServerError &&
		traces.NotegicTracer != nil {
		traces.NotegicTracer.RecordError(ctx, exception)
	}

	ctx.JSON(publicException.HTTPStatusCode(), cgateway.ClientResponse[any]{
		Success:   false,
		Data:      nil,
		Exception: publicException,
	})
}

func SafelyAbortAndResponseWithJSON(exception *cexceptions.Exception, ctx *gin.Context, names ...string) {
	publicException := ToPublic(getRequestContext(ctx), exception)
	IncrementMeter(publicException, ctx, names...)
	if exception != nil &&
		exception.HTTPStatusCode() >= http.StatusInternalServerError &&
		traces.NotegicTracer != nil {
		traces.NotegicTracer.RecordError(ctx, exception)
	}

	ctx.AbortWithStatusJSON(publicException.HTTPStatusCode(), cgateway.ClientResponse[any]{
		Success:   false,
		Data:      nil,
		Exception: publicException,
	})
}

func ToGraphQLError(exception *cexceptions.Exception, ctx context.Context) *gqlerror.Error {
	exception = ToPublic(ctx, exception)
	extensions := map[string]any{
		"reason":     exception.Reason,
		"domain":     exception.Domain,
		"operation":  exception.Operation,
		"httpStatus": exception.HTTPStatusCode(),
		"retryable":  exception.Retryable,
		"timestamp":  time.Now().Format(time.RFC3339),
	}

	var path ast.Path
	var locations []gqlerror.Location
	if ctx != nil {
		if fieldContext := graphql.GetFieldContext(ctx); fieldContext != nil {
			path = fieldContext.Path()
			if fieldContext.Field.Position != nil {
				locations = []gqlerror.Location{{
					Line:   fieldContext.Field.Position.Line,
					Column: fieldContext.Field.Position.Column,
				}}
			}
		}
		if operationContext := graphql.GetOperationContext(ctx); operationContext != nil && operationContext.OperationName != "" {
			extensions["operationName"] = operationContext.OperationName
		}
	}

	gqlError := &gqlerror.Error{
		Message:    exception.Message,
		Path:       path,
		Locations:  locations,
		Extensions: extensions,
	}
	return gqlError
}

func ToPublic(ctx context.Context, exception *cexceptions.Exception) *cexceptions.Exception {
	if exception == nil {
		return cexceptions.InternalServerError()
	}
	if logs.NotegicLogger != nil {
		if err := logs.NotegicLogger.JSON(
			ctx,
			slog.LevelError,
			exception.String(),
			exception,
		); err != nil {
			logs.NotegicLogger.Error(
				ctx,
				err,
				"failed to marshal exception for logging",
			)
		}
	}

	return exception.ToPublic()
}
