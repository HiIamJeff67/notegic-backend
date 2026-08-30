package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"

	smetrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"
	straces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"
)

func ApplyTracerMiddleware(spanName string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		newContext, span := straces.NotegicTracer.Start(ctx.Request.Context(), "http."+spanName)
		span.SetAttributes(
			attribute.String("http.request.method", ctx.Request.Method),
			attribute.String("http.route", ctx.FullPath()),
		)
		defer func() {
			span.SetAttributes(attribute.Int("http.response.status_code", ctx.Writer.Status()))
			straces.NotegicTracer.End(span, nil)
		}()

		ctx.Request = ctx.Request.WithContext(newContext)
		ctx.Next()
	}
}

func ApplyMeterMiddleware(names ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		isTotalCounted := false
		for _, name := range names {
			if name == "server.requests.total" {
				isTotalCounted = true
			}
			smetrics.NotegicMeter.Count(ctx, name, 1)
		}
		if !isTotalCounted {
			smetrics.NotegicMeter.Count(ctx, "server.requests.total", 1)
		}
		ctx.Next()
	}
}
