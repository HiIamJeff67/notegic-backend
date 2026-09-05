package general

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

func TestOutboxEventRepositoryConvertEnvelopeToCreateInputAndSerializePreserveEventContract(t *testing.T) {
	t.Parallel()

	aggregateId := uuid.New()
	eventId := uuid.New()
	occurredAt := time.Now().UTC().Round(0)
	envelope := cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData]{
		SchemaVersion: cevent.Version,
		EventId:       eventId,
		EventType:     coreevents.EventType_BlockPackAccessRevoked,
		AggregateType: coreevents.AggregateType_BlockPack,
		AggregateId:   aggregateId,
		KafkaKey:      aggregateId.String(),
		OccurredAt:    occurredAt,
		CorrelationId: "correlation-id",
		Data: coreevents.BlockPackAccessRevokedData{
			Reason: coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
		},
	}
	repository := NewOutboxEventRepository[coreevents.BlockPackAccessRevokedData]()

	createInput, err := repository.(*OutboxEventRepository[coreevents.BlockPackAccessRevokedData]).convertEnvelopeToCreateInput(
		coreevents.CoreLifecycleTopic,
		envelope,
	)
	if err != nil {
		t.Fatalf("convert envelope: %v", err)
	}

	payload, err := repository.SerializeOutboxEvent(schemas.OutboxEvent{
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
		t.Fatalf("serialize outbox event: %v", err)
	}

	var serialized cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData]
	if err := json.Unmarshal(payload, &serialized); err != nil {
		t.Fatalf("decode serialized payload: %v", err)
	}
	if serialized != envelope {
		t.Fatalf("serialized envelope mismatch: got %#v, want %#v", serialized, envelope)
	}
}

func TestOutboxEventRepositoryConvertEnvelopeToCreateInputRejectsMismatchedKafkaKey(t *testing.T) {
	t.Parallel()

	repository := NewOutboxEventRepository[coreevents.ResourceChangedData]()
	_, err := repository.(*OutboxEventRepository[coreevents.ResourceChangedData]).convertEnvelopeToCreateInput(
		coreevents.CoreLifecycleTopic,
		cevent.EventEnvelope[coreevents.ResourceChangedData]{
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_BlockPackChanged,
			AggregateType: coreevents.AggregateType_BlockPack,
			AggregateId:   uuid.New(),
			KafkaKey:      "unexpected-key",
			OccurredAt:    time.Now().UTC(),
			Data: coreevents.ResourceChangedData{
				Change: coreevents.ResourceEventChange_Updated,
			},
		},
	)
	if err == nil {
		t.Fatal("expected mismatched Kafka key error")
	}
}
