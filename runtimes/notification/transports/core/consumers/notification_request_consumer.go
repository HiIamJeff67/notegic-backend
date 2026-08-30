package consumers

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	services "github.com/HiIamJeff67/notegic-backend/runtimes/notification/services"
)

type NotificationRequestConsumer struct {
	service     services.NotificationServiceInterface
	kafkaConfig skafka.ConsumerConfig
}

func NewNotificationRequestConsumer(
	service services.NotificationServiceInterface,
	kafkaConfig skafka.ConsumerConfig,
) *NotificationRequestConsumer {
	return &NotificationRequestConsumer{
		service:     service,
		kafkaConfig: kafkaConfig,
	}
}

func (c *NotificationRequestConsumer) Start(ctx context.Context) func() {
	workerCtx, cancel := context.WithCancel(ctx)
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		c.run(workerCtx)
	}()

	return func() {
		cancel()
		waitGroup.Wait()
	}
}

func (c *NotificationRequestConsumer) run(ctx context.Context) {
	for ctx.Err() == nil {
		consumer, err := skafka.NewConsumer(
			c.kafkaConfig,
			coreevents.CoreNotificationTopic.String(),
		)
		if err == nil {
			err = consumer.Run(ctx, c.consume)
			consumer.Close()
		}
		if ctx.Err() != nil {
			return
		}
		if slogs.NotegicLogger != nil {
			slogs.NotegicLogger.Error(ctx, err, "Notification request consumer stopped")
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func (c *NotificationRequestConsumer) consume(
	ctx context.Context,
	_ skafka.ConsumerRecord,
	event cevent.EventEnvelope[json.RawMessage],
) error {
	if event.EventType != coreevents.EventType_NotificationRequested {
		return nil
	}
	var data coreevents.NotificationRequestedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return &skafka.ConsumerError{
			Classification: skafka.ErrorClassification_SchemaIncompatible,
			Origin:         err,
		}
	}
	eventWithData := cevent.EventEnvelope[coreevents.NotificationRequestedData]{
		SchemaVersion: event.SchemaVersion,
		EventId:       event.EventId,
		EventType:     event.EventType,
		AggregateType: event.AggregateType,
		AggregateId:   event.AggregateId,
		KafkaKey:      event.KafkaKey,
		OccurredAt:    event.OccurredAt,
		CorrelationId: event.CorrelationId,
		CausationId:   event.CausationId,
		Trace:         event.Trace,
		Data:          data,
	}
	if err := c.service.ConsumeNotificationRequested(ctx, eventWithData); err != nil {
		return &skafka.ConsumerError{
			Classification: skafka.ErrorClassification_Transient,
			Origin:         err,
		}
	}

	return nil
}
