package repositories

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	coreeventscontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	eventcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	cmodels "github.com/HiIamJeff67/notegic-backend/contracts/types/models"
	crepositories "github.com/HiIamJeff67/notegic-backend/contracts/types/models/repositories"
)

func TestConvertEnvelopeToCreateOutboxEventInputAndSerializePreserveEventContract(t *testing.T) {
	eventId := uuid.New()
	aggregateId := uuid.New()
	occurredAt := time.Now().UTC().Round(0)
	envelope := eventcontract.EventEnvelope[coreeventscontract.BlockPackAccessRevokedData]{
		SchemaVersion: eventcontract.Version,
		EventId:       eventId,
		EventType:     coreeventscontract.EventType_BlockPackAccessRevoked,
		AggregateType: coreeventscontract.AggregateType_BlockPack,
		AggregateId:   aggregateId,
		KafkaKey:      aggregateId.String(),
		OccurredAt:    occurredAt,
		CorrelationId: "request-123",
		Trace: eventcontract.TraceMetadata{
			TraceParent: "00-trace",
		},
		Data: coreeventscontract.BlockPackAccessRevokedData{},
	}

	createInput, err := crepositories.ConvertEnvelopeToCreateOutboxEventInput(coreeventscontract.CoreLifecycleTopic, envelope)
	if err != nil {
		t.Fatalf("failed to create outbox input: %v", err)
	}
	payload, err := crepositories.SerializeOutboxEvent(cmodels.OutboxEvent{
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

	var serialized eventcontract.EventEnvelope[coreeventscontract.BlockPackAccessRevokedData]
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
		coreeventscontract.CoreLifecycleTopic,
		eventcontract.EventEnvelope[coreeventscontract.UserSessionsRevokedData]{
			SchemaVersion: eventcontract.Version,
			EventId:       uuid.New(),
			EventType:     coreeventscontract.EventType_UserSessionsRevoked,
			AggregateType: coreeventscontract.AggregateType_User,
			AggregateId:   uuid.New(),
			KafkaKey:      "another-aggregate",
			OccurredAt:    time.Now(),
			CorrelationId: "request-123",
			Data:          coreeventscontract.UserSessionsRevokedData{},
		},
	)
	if err == nil {
		t.Fatal("expected mismatched Kafka key to be rejected")
	}
}
