package observability

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	logglobal "go.opentelemetry.io/otel/log/global"
	"go.opentelemetry.io/otel/propagation"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	metrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"
	traces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"
)

func Initialize(ctx context.Context, config Config) func() {

	response, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.DeploymentEnvironment),
			semconv.ServiceInstanceID(config.ServiceInstanceId),
		),
	)
	if err != nil {
		fmt.Println("Failed to create OpenTelemetry resource: ", err)
		response = resource.Default()
	}

	traceProviderOptions := []sdktrace.TracerProviderOption{
		sdktrace.WithResource(response),
	}
	traceExporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(config.CollectorEndpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		fmt.Println("Failed to create OpenTelemetry trace exporter: ", err)
	} else {
		traceProviderOptions = append(traceProviderOptions, sdktrace.WithBatcher(traceExporter))
	}
	traceProvider := sdktrace.NewTracerProvider(traceProviderOptions...)
	otel.SetTracerProvider(traceProvider)

	meterProviderOptions := []sdkmetric.Option{
		sdkmetric.WithResource(response),
	}
	metricExporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(config.CollectorEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		fmt.Println("Failed to create OpenTelemetry metric exporter: ", err)
	} else {
		meterProviderOptions = append(meterProviderOptions, sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(
				metricExporter,
				sdkmetric.WithInterval(15*time.Second),
			),
		))
	}
	meterProvider := sdkmetric.NewMeterProvider(meterProviderOptions...)
	otel.SetMeterProvider(meterProvider)

	logProviderOptions := []sdklog.LoggerProviderOption{
		sdklog.WithResource(response),
	}
	logExporter, err := otlploggrpc.New(
		ctx,
		otlploggrpc.WithEndpoint(config.CollectorEndpoint),
		otlploggrpc.WithInsecure(),
	)
	if err != nil {
		fmt.Println("Failed to create OpenTelemetry log exporter: ", err)
	} else {
		logProviderOptions = append(logProviderOptions, sdklog.WithProcessor(sdklog.NewBatchProcessor(logExporter)))
	}
	logProvider := sdklog.NewLoggerProvider(logProviderOptions...)
	logglobal.SetLoggerProvider(logProvider)

	otel.SetTextMapPropagator(propagation.TraceContext{})

	logs.NotegicLogger = logs.NewLogger(true)
	metrics.NotegicMeter = metrics.NewMeter(otel.Meter(config.ServiceName))
	traces.NotegicTracer = traces.NewTracer(otel.Tracer(config.ServiceName))

	return func() {
		if err := traceProvider.Shutdown(ctx); err != nil {
			fmt.Println("Failed to shutdown OpenTelemetry trace provider: ", err)
		}
		if err := meterProvider.Shutdown(ctx); err != nil {
			fmt.Println("Failed to shutdown OpenTelemetry meter provider: ", err)
		}
		if err := logProvider.Shutdown(ctx); err != nil {
			fmt.Println("Failed to shutdown OpenTelemetry log provider: ", err)
		}
	}
}
