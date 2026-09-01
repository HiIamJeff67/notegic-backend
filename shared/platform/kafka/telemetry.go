package kafka

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/attribute"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	metrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"
	traces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"
)

func RecordBrokerPing(ctx context.Context, duration time.Duration, err error) {
	if metrics.NotegicMeter == nil {
		return
	}

	attributes := []attribute.KeyValue{
		attribute.String("messaging.system", "kafka"),
		attribute.Bool("kafka.available", err == nil),
	}
	metrics.NotegicMeter.Duration(ctx, "kafka.broker.ping.duration", duration, attributes...)
	metrics.NotegicMeter.Count(ctx, "kafka.broker.ping.count", 1, attributes...)
}

func RecordPublish(ctx context.Context, topic string, duration time.Duration, err error) {
	if metrics.NotegicMeter == nil {
		return
	}

	attributes := []attribute.KeyValue{
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.destination.name", topic),
		attribute.Bool("kafka.published", err == nil),
	}
	metrics.NotegicMeter.Duration(ctx, "kafka.publish.duration", duration, attributes...)
	metrics.NotegicMeter.Count(ctx, "kafka.publish.count", 1, attributes...)
	if err != nil {
		metrics.NotegicMeter.Count(ctx, "kafka.publish.failure.count", 1, attributes...)
	}
}

func RecordConsume(ctx context.Context, topic string, consumerGroup string, duration time.Duration) {
	if metrics.NotegicMeter == nil {
		return
	}

	attributes := []attribute.KeyValue{
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.destination.name", topic),
		attribute.String("messaging.consumer.group.name", consumerGroup),
	}
	metrics.NotegicMeter.Duration(ctx, "kafka.consume.duration", duration, attributes...)
	metrics.NotegicMeter.Count(ctx, "kafka.consume.count", 1, attributes...)
}

func RecordConsumerLag(ctx context.Context, topic string, consumerGroup string, lag int64) {
	if metrics.NotegicMeter == nil {
		return
	}

	metrics.NotegicMeter.Value(ctx, "kafka.consumer.lag", lag,
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.destination.name", topic),
		attribute.String("messaging.consumer.group.name", consumerGroup),
	)
}

func RecordRetry(ctx context.Context, topic string, consumerGroup string) {
	if metrics.NotegicMeter == nil {
		return
	}

	metrics.NotegicMeter.Count(ctx, "kafka.retry.count", 1,
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.destination.name", topic),
		attribute.String("messaging.consumer.group.name", consumerGroup),
	)
}

func RecordDeadLetter(ctx context.Context, topic string, consumerGroup string) {
	if metrics.NotegicMeter == nil {
		return
	}

	metrics.NotegicMeter.Count(ctx, "kafka.dlq.count", 1,
		attribute.String("messaging.system", "kafka"),
		attribute.String("messaging.destination.name", topic),
		attribute.String("messaging.consumer.group.name", consumerGroup),
	)
}

func RecordFailure(
	ctx context.Context,
	consumerGroup string,
	topic string,
	partition int32,
	offset int64,
	message string,
	err error,
) {
	if traces.NotegicTracer != nil {
		traces.NotegicTracer.RecordError(ctx, err)
	}
	if logs.NotegicLogger != nil {
		logs.NotegicLogger.Error(ctx, err, message,
			attribute.String("messaging.system", "kafka"),
			attribute.String("messaging.destination.name", topic),
			attribute.String("messaging.consumer.group.name", consumerGroup),
			attribute.Int("messaging.kafka.partition", int(partition)),
			attribute.Int64("messaging.kafka.offset", offset),
		)
	}
}
