package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"

	services "github.com/HiIamJeff67/notegic-backend/internal/notification/services"
)

type notificationServiceStub struct {
	services.NotificationServiceInterface
	consumeCalls          int
	consumeErr            error
	lastNotificationEvent cevent.EventEnvelope[coreevents.NotificationRequestedData]
}

func (s *notificationServiceStub) ConsumeNotificationRequested(
	_ context.Context,
	event cevent.EventEnvelope[coreevents.NotificationRequestedData],
) error {
	s.consumeCalls++
	s.lastNotificationEvent = event
	return s.consumeErr
}

func TestNotificationRequestConsumerConsumesNotificationRequestedEvent(t *testing.T) {
	recipientUserPublicId := uuid.New()
	service := &notificationServiceStub{}
	consumer := &NotificationRequestConsumer{service: service}
	event := cevent.EventEnvelope[json.RawMessage]{
		SchemaVersion: cevent.Version,
		EventId:       uuid.New(),
		EventType:     coreevents.EventType_NotificationRequested,
		AggregateType: coreevents.AggregateType_Notification,
		AggregateId:   recipientUserPublicId,
		KafkaKey:      recipientUserPublicId.String(),
		OccurredAt:    time.Now().UTC(),
		Data: mustMarshalNotificationRequest(t, coreevents.NotificationRequestedData{
			RecipientUserPublicId: recipientUserPublicId,
			Type:                  coreevents.NotificationType_News,
			Priority:              coreevents.NotificationPriority_Normal,
			TemplateKey:           "news",
			TemplateVersion:       1,
			Payload:               json.RawMessage(`{"title":"Release"}`),
			DedupeKey:             "release:" + recipientUserPublicId.String(),
		}),
	}

	if err := consumer.consume(context.Background(), platformkafka.ConsumerRecord{}, event); err != nil {
		t.Fatalf("consume notification event: %v", err)
	}
	if service.consumeCalls != 1 {
		t.Fatalf("consume calls = %d, want 1", service.consumeCalls)
	}
	if service.lastNotificationEvent.Data.RecipientUserPublicId != recipientUserPublicId {
		t.Fatalf("recipient user public ID = %s, want %s", service.lastNotificationEvent.Data.RecipientUserPublicId, recipientUserPublicId)
	}
}

func TestNotificationRequestConsumerClassifiesServiceFailureAsTransient(t *testing.T) {
	service := &notificationServiceStub{consumeErr: errors.New("database unavailable")}
	consumer := &NotificationRequestConsumer{service: service}
	recipientUserPublicId := uuid.New()
	event := cevent.EventEnvelope[json.RawMessage]{
		SchemaVersion: cevent.Version,
		EventId:       uuid.New(),
		EventType:     coreevents.EventType_NotificationRequested,
		AggregateType: coreevents.AggregateType_Notification,
		AggregateId:   recipientUserPublicId,
		KafkaKey:      recipientUserPublicId.String(),
		Data: mustMarshalNotificationRequest(t, coreevents.NotificationRequestedData{
			RecipientUserPublicId: recipientUserPublicId,
			Type:                  coreevents.NotificationType_News,
			Priority:              coreevents.NotificationPriority_Normal,
			TemplateKey:           "news",
			TemplateVersion:       1,
			Payload:               json.RawMessage(`{"title":"Release"}`),
			DedupeKey:             "release:" + recipientUserPublicId.String(),
		}),
	}

	resultErr := consumer.consume(context.Background(), platformkafka.ConsumerRecord{}, event)
	consumerErr, ok := resultErr.(*platformkafka.ConsumerError)
	if !ok {
		t.Fatalf("error type = %T, want *platformkafka.ConsumerError", resultErr)
	}
	if consumerErr.Classification != platformkafka.ErrorClassification_Transient {
		t.Fatalf("classification = %q, want %q", consumerErr.Classification, platformkafka.ErrorClassification_Transient)
	}
}

func mustMarshalNotificationRequest(t *testing.T, data coreevents.NotificationRequestedData) json.RawMessage {
	t.Helper()
	encoded, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal notification request: %v", err)
	}
	return encoded
}
