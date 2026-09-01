package kafka

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	franzkgo "github.com/twmb/franz-go/pkg/kgo"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	traces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"
)

type Consumer struct {
	client        *franzkgo.Client
	consumerGroup string
	config        ConsumerConfig
}

func NewConsumer(
	kafkaConfig ConsumerConfig,
	topics ...string,
) (*Consumer, error) {
	if kafkaConfig.ConsumerGroup == "" {
		return nil, errors.New("Kafka consumer group is required")
	}
	if len(topics) == 0 {
		return nil, errors.New("at least one Kafka consumer topic is required")
	}

	options, err := newConnectionOptions(kafkaConfig.ClientConfig)
	if err != nil {
		return nil, err
	}
	options = append(options,
		franzkgo.ConsumerGroup(kafkaConfig.ConsumerGroup),
		franzkgo.ConsumeTopics(topics...),
		franzkgo.DisableAutoCommit(),
		franzkgo.BlockRebalanceOnPoll(),
	)
	client, err := franzkgo.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("create Kafka consumer: %w", err)
	}

	return &Consumer{
		client:        client,
		consumerGroup: kafkaConfig.ConsumerGroup,
		config:        kafkaConfig,
	}, nil
}

/* ============================== Auxiliary Methods ============================== */

func (c *Consumer) consumeFetches(
	ctx context.Context,
	fetches franzkgo.Fetches,
	handler ConsumerHandler,
) {
	fetches.EachPartition(func(fetchTopicPartition franzkgo.FetchTopicPartition) {
		if fetchTopicPartition.Err != nil {
			RecordFailure(ctx, c.consumerGroup, fetchTopicPartition.Topic, fetchTopicPartition.Partition, 0, "Kafka consumer partition fetch failed", fetchTopicPartition.Err)
			return
		}

		recordsToCommit := make([]*franzkgo.Record, 0, len(fetchTopicPartition.Records))
		for _, record := range fetchTopicPartition.Records {
			RecordConsumerLag(
				ctx,
				record.Topic,
				c.consumerGroup,
				fetchTopicPartition.HighWatermark-record.Offset,
			)
			if !c.consumeRecord(ctx, record, handler) {
				break
			}
			recordsToCommit = append(recordsToCommit, record)
		}
		if len(recordsToCommit) == 0 {
			return
		}
		if err := c.client.CommitRecords(ctx, recordsToCommit[len(recordsToCommit)-1]); err != nil {
			RecordFailure(
				ctx,
				c.consumerGroup,
				fetchTopicPartition.Topic,
				fetchTopicPartition.Partition,
				recordsToCommit[len(recordsToCommit)-1].Offset,
				"Failed to commit Kafka consumer offset",
				err,
			)
		}
	})
}

func (c *Consumer) consumeRecord(
	ctx context.Context,
	record *franzkgo.Record,
	handler ConsumerHandler,
) bool {
	var envelope cevent.EventEnvelope[json.RawMessage]
	if err := json.Unmarshal(record.Value, &envelope); err != nil {
		return c.publishDeadLetter(
			ctx,
			record,
			nil,
			ErrorClassification_SchemaIncompatible,
			1,
			fmt.Errorf("decode Kafka event envelope: %w", err),
		)
	}
	if envelope.SchemaVersion != cevent.Version {
		return c.publishDeadLetter(
			ctx,
			record,
			nil,
			ErrorClassification_SchemaIncompatible,
			1,
			fmt.Errorf("unsupported Kafka event schema version %q", envelope.SchemaVersion),
		)
	}
	if envelope.EventId == uuid.Nil || envelope.EventType == "" || envelope.AggregateType == "" ||
		envelope.AggregateId == uuid.Nil || envelope.KafkaKey == "" {
		return c.publishDeadLetter(
			ctx,
			record,
			nil,
			ErrorClassification_SchemaIncompatible,
			1,
			errors.New("Kafka event envelope is incomplete"),
		)
	}
	if envelope.KafkaKey != envelope.AggregateId.String() || envelope.KafkaKey != string(record.Key) {
		return c.publishDeadLetter(
			ctx,
			record,
			nil,
			ErrorClassification_SchemaIncompatible,
			1,
			errors.New("Kafka event envelope key does not match the aggregate ID"),
		)
	}

	if envelope.Trace.TraceParent != "" {
		ctx = otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier{
			"traceparent": envelope.Trace.TraceParent,
			"tracestate":  envelope.Trace.TraceState,
		})
	}
	if traces.NotegicTracer != nil {
		consumerCtx, span := traces.NotegicTracer.Start(ctx, "kafka.consume")
		defer traces.NotegicTracer.End(span, nil)
		ctx = consumerCtx
	}

	var err error
	for attempt := 1; attempt <= c.config.MaximumAttempts; attempt++ {
		startedAt := time.Now()
		err = handler(ctx, ConsumerRecord{
			Topic:     record.Topic,
			Partition: record.Partition,
			Offset:    record.Offset,
			Key:       string(record.Key),
		}, envelope)
		if err == nil {
			RecordConsume(ctx, record.Topic, c.consumerGroup, time.Since(startedAt))
			return true
		}

		classification := ErrorClassification_Transient
		var consumerErr *ConsumerError
		if errors.As(err, &consumerErr) {
			switch consumerErr.Classification {
			case ErrorClassification_PoisonMessage, ErrorClassification_SchemaIncompatible:
				classification = consumerErr.Classification
			}
		}
		if classification != ErrorClassification_Transient {
			return c.publishDeadLetter(ctx, record, &envelope.EventId, classification, attempt, err)
		}
		if attempt == c.config.MaximumAttempts {
			return c.publishDeadLetter(ctx, record, &envelope.EventId, classification, attempt, err)
		}

		RecordRetry(ctx, record.Topic, c.consumerGroup)
		RecordFailure(ctx, c.consumerGroup, record.Topic, record.Partition, record.Offset, "Retrying Kafka consumer event", err)
		backoff := c.config.InitialRetryBackoff
		for index := 1; index < attempt && backoff < c.config.MaximumRetryBackoff; index++ {
			backoff *= 2
		}
		if backoff > c.config.MaximumRetryBackoff {
			backoff = c.config.MaximumRetryBackoff
		}
		select {
		case <-ctx.Done():
			return false
		case <-time.After(backoff):
		}
	}

	return false
}

/* ============================== Consumer Methods ============================== */

func (c *Consumer) Run(ctx context.Context, handler ConsumerHandler) error {
	if handler == nil {
		return errors.New("Kafka consumer handler is required")
	}

	pollRetryBackoff := c.config.InitialRetryBackoff
	for ctx.Err() == nil {
		fetches := c.client.PollRecords(ctx, c.config.MaximumPollRecords)
		if err := fetches.Err(); err != nil {
			if ctx.Err() != nil {
				return nil
			}

			c.client.AllowRebalance()
			RecordFailure(
				ctx,
				c.consumerGroup,
				"",
				0,
				0,
				"Kafka consumer poll failed; retrying",
				err,
			)
			select {
			case <-ctx.Done():
				return nil
			case <-time.After(pollRetryBackoff):
			}
			pollRetryBackoff *= 2
			if pollRetryBackoff > c.config.MaximumRetryBackoff {
				pollRetryBackoff = c.config.MaximumRetryBackoff
			}
			continue
		}

		pollRetryBackoff = c.config.InitialRetryBackoff
		c.consumeFetches(ctx, fetches, handler)
		c.client.AllowRebalance()
	}

	return nil
}

func (c *Consumer) Close() {
	c.client.Close()
}
