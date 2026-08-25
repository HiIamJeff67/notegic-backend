package repositories

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	crepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	platformschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

func TestConvertEnvelopeToCreateOutboxEventInputAndSerializePreserveEventContract(t *testing.T) {
	eventId := uuid.New()
	aggregateId := uuid.New()
	occurredAt := time.Now().UTC().Round(0)
	envelope := cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData]{
		SchemaVersion: cevent.Version,
		EventId:       eventId,
		EventType:     coreevents.EventType_BlockPackAccessRevoked,
		AggregateType: coreevents.AggregateType_BlockPack,
		AggregateId:   aggregateId,
		KafkaKey:      aggregateId.String(),
		OccurredAt:    occurredAt,
		CorrelationId: "request-123",
		Trace: cevent.TraceMetadata{
			TraceParent: "00-trace",
		},
		Data: coreevents.BlockPackAccessRevokedData{},
	}

	createInput, err := crepositories.ConvertEnvelopeToCreateOutboxEventInput(coreevents.CoreLifecycleTopic, envelope)
	if err != nil {
		t.Fatalf("failed to create outbox input: %v", err)
	}
	payload, err := crepositories.SerializeOutboxEvent(platformschemas.OutboxEvent{
		Id:            createInput.Id,
		AggregateType: createInput.AggregateType,
		AggregateId:   createInput.AggregateId,
		EventType:     createInput.EventType,
		Topic:         createInput.Topic,
		KafkaKey:      createInput.KafkaKey,
		Payload:       datatypes.JSON(createInput.Payload),
		Metadata:      datatypes.JSON(createInput.Metadata),
	})
	if err != nil {
		t.Fatalf("failed to serialize outbox event: %v", err)
	}

	var serialized cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData]
	if err := json.Unmarshal(payload, &serialized); err != nil {
		t.Fatalf("failed to decode serialized event: %v", err)
	}
	if serialized.EventId != eventId || serialized.AggregateId != aggregateId ||
		serialized.KafkaKey != aggregateId.String() || serialized.CorrelationId != "request-123" ||
		serialized.Trace.TraceParent != "00-trace" {
		t.Fatalf("serialized event lost contract fields: %#v", serialized)
	}
}

func TestConvertEnvelopeToCreateOutboxEventInputRejectsMismatchedKafkaKey(t *testing.T) {
	_, err := crepositories.ConvertEnvelopeToCreateOutboxEventInput(
		coreevents.CoreLifecycleTopic,
		cevent.EventEnvelope[coreevents.UserSessionsRevokedData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_UserSessionsRevoked,
			AggregateType: coreevents.AggregateType_User,
			AggregateId:   uuid.New(),
			KafkaKey:      "another-aggregate",
			OccurredAt:    time.Now(),
			CorrelationId: "request-123",
			Data:          coreevents.UserSessionsRevokedData{},
		},
	)
	if err == nil {
		t.Fatal("expected mismatched Kafka key to be rejected")
	}
}
