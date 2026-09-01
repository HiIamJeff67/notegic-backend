package kafka

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/google/uuid"
	franzkgo "github.com/twmb/franz-go/pkg/kgo"

	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	traces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"
)

type DeadLetter struct {
	SchemaVersion   string              `json:"schemaVersion"`
	ConsumerGroup   string              `json:"consumerGroup"`
	SourceTopic     string              `json:"sourceTopic"`
	SourcePartition int32               `json:"sourcePartition"`
	SourceOffset    int64               `json:"sourceOffset"`
	Key             string              `json:"key"`
	EventId         *uuid.UUID          `json:"eventId,omitempty"`
	Classification  ErrorClassification `json:"classification"`
	Attempts        int                 `json:"attempts"`
	Error           string              `json:"error"`
	Value           []byte              `json:"value"`
	FailedAt        time.Time           `json:"failedAt"`
}

func GetDeadLetterTopic(topic string) string {
	return topic + ".dlq"
}

func (c *Consumer) publishDeadLetter(
	ctx context.Context,
	record *franzkgo.Record,
	eventId *uuid.UUID,
	classification ErrorClassification,
	attempts int,
	err error,
) bool {
	if traces.NotegicTracer != nil {
		traces.NotegicTracer.RecordError(ctx, err)
	}

	deadLetter := DeadLetter{
		SchemaVersion:   cevent.Version,
		ConsumerGroup:   c.consumerGroup,
		SourceTopic:     record.Topic,
		SourcePartition: record.Partition,
		SourceOffset:    record.Offset,
		Key:             string(record.Key),
		EventId:         eventId,
		Classification:  classification,
		Attempts:        attempts,
		Error:           err.Error(),
		Value:           record.Value,
		FailedAt:        time.Now().UTC(),
	}
	payload, marshalErr := json.Marshal(deadLetter)
	if marshalErr != nil {
		RecordFailure(ctx, c.consumerGroup, record.Topic, record.Partition, record.Offset, "Failed to serialize Kafka dead-letter record", marshalErr)
		return false
	}

	startedAt := time.Now()
	deadLetterTopic := GetDeadLetterTopic(record.Topic)
	result := c.client.ProduceSync(ctx, &franzkgo.Record{
		Topic: deadLetterTopic,
		Key:   record.Key,
		Value: payload,
	})
	if publishErr := result.FirstErr(); publishErr != nil {
		RecordPublish(ctx, deadLetterTopic, time.Since(startedAt), publishErr)
		RecordFailure(ctx, c.consumerGroup, record.Topic, record.Partition, record.Offset, "Failed to publish Kafka dead-letter record", publishErr)
		return false
	}
	RecordPublish(ctx, deadLetterTopic, time.Since(startedAt), nil)
	RecordDeadLetter(ctx, record.Topic, c.consumerGroup)
	if logs.NotegicLogger != nil {
		deadLetter.Value = nil
		_ = logs.NotegicLogger.JSON(ctx, slog.LevelError, "Kafka event sent to dead-letter topic", deadLetter)
	}

	return true
}
