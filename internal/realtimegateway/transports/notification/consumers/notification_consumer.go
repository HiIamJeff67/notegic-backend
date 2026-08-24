package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	notificationeventscontract "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/events"
	eventcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	realtimeleasecache "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/data/redis/realtimelease"
)

type NotificationConsumer struct {
	leaseStore  *realtimeleasecache.RealtimeLeaseCacheClient
	kafkaConfig platformkafka.ConsumerConfig
}

func NewNotificationConsumer(
	leaseStore *realtimeleasecache.RealtimeLeaseCacheClient,
	kafkaConfig platformkafka.ConsumerConfig,
) *NotificationConsumer {
	return &NotificationConsumer{
		leaseStore:  leaseStore,
		kafkaConfig: kafkaConfig,
	}
}

func (c *NotificationConsumer) Start(ctx context.Context) func() {
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

func (c *NotificationConsumer) run(ctx context.Context) {
	for ctx.Err() == nil {
		consumer, err := platformkafka.NewConsumer(
			c.kafkaConfig,
			notificationeventscontract.NotificationTopic.String(),
		)
		if err == nil {
			err = consumer.Run(ctx, c.consume)
			consumer.Close()
		}
		if ctx.Err() != nil {
			return
		}
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(ctx, err, "RealtimeGateway notification consumer stopped")
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func (c *NotificationConsumer) consume(
	ctx context.Context,
	_ platformkafka.ConsumerRecord,
	event eventcontract.EventEnvelope[json.RawMessage],
) error {
	if event.EventType != notificationeventscontract.EventType_NotificationCreated {
		return &platformkafka.ConsumerError{
			Classification: platformkafka.ErrorClassification_PoisonMessage,
			Origin:         errors.New("unsupported Notification event type"),
		}
	}
	var data notificationeventscontract.NotificationCreatedData
	if err := json.Unmarshal(event.Data, &data); err != nil {
		return &platformkafka.ConsumerError{
			Classification: platformkafka.ErrorClassification_SchemaIncompatible,
			Origin:         err,
		}
	}
	if data.RecipientUserPublicId == uuid.Nil || data.NotificationId == uuid.Nil {
		return &platformkafka.ConsumerError{
			Classification: platformkafka.ErrorClassification_SchemaIncompatible,
			Origin:         errors.New("Notification event recipient or notification ID is missing"),
		}
	}
	claimed, err := c.leaseStore.MarkLifecycleEventProcessed(event.EventId)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	if err := c.leaseStore.PublishNotification(realtimeleasecache.NotificationEvent{
		EventId:               event.EventId,
		NotificationId:        data.NotificationId,
		RecipientUserPublicId: data.RecipientUserPublicId,
		Type:                  data.Type,
		Priority:              data.Priority,
		TemplateKey:           data.TemplateKey,
		TemplateVersion:       data.TemplateVersion,
		Payload:               data.Payload,
		CreatedAt:             data.CreatedAt,
		ExpiresAt:             data.ExpiresAt,
	}); err != nil {
		_ = c.leaseStore.ReleaseLifecycleEvent(event.EventId)
		return err
	}

	return nil
}
