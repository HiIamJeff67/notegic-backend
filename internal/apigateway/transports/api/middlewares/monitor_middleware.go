package middlewares

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"

	smetrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"
	straces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"
)

func ApplyTracerMiddleware(spanName string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		newCtx, span := straces.NotegicTracer.Start(ctx.Request.Context(), "http."+spanName)
		span.SetAttributes(
			attribute.String("http.request.method", ctx.Request.Method),
			attribute.String("http.route", ctx.FullPath()),
			attribute.String("gateway.surface", "api-gateway"),
			attribute.String("gateway.auth_method", "api-key"),
			attribute.String("gateway.operation", spanName),
		)
		defer func() {
			span.SetAttributes(attribute.Int("http.response.status_code", ctx.Writer.Status()))
			straces.NotegicTracer.End(span, nil)
		}()

		ctx.Request = ctx.Request.WithContext(newCtx)
		ctx.Next()
	}
}

func ApplyMeterMiddleware(names ...string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()
		status := ctx.Writer.Status()
		isTotalCounted := false
		for _, name := range names {
			if name == "server.requests.total" {
				isTotalCounted = true
			}
			smetrics.NotegicMeter.Count(ctx, name, 1,
				attribute.String("gateway.surface", "api-gateway"),
				attribute.String("gateway.auth_method", "api-key"),
				attribute.String("gateway.operation", name),
				attribute.Int("http.response.status_code", status),
			)
		}
		if !isTotalCounted {
			smetrics.NotegicMeter.Count(ctx, "server.requests.total", 1,
				attribute.String("gateway.surface", "api-gateway"),
				attribute.String("gateway.auth_method", "api-key"),
				attribute.String("gateway.operation", "server.requests.total"),
				attribute.Int("http.response.status_code", status),
			)
		}
	}
}
