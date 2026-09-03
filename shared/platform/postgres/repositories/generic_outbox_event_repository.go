package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type GenericOutboxEventRepositoryInterface interface {
	CreateMany(createInputs []inputs.CreateOutboxEventInput, opts RepositoryOptionFields) *cexceptions.Exception
	ClaimAvailable(ctx context.Context, workerId string, batchSize int, claimTimeout time.Duration, opts RepositoryOptionFields) ([]schemas.OutboxEvent, *cexceptions.Exception)
	MarkPublishedMany(ctx context.Context, eventIds []uuid.UUID, workerId string, opts RepositoryOptionFields) *cexceptions.Exception
	MarkFailedMany(ctx context.Context, failureInputs []inputs.FailedOutboxEventInput, workerId string, opts RepositoryOptionFields) *cexceptions.Exception
	DeletePublishedBefore(ctx context.Context, publishedBefore time.Time, opts RepositoryOptionFields) (int64, *cexceptions.Exception)
}

type GenericOutboxEventRepository struct {
}

type outboxEventMetadata struct {
	SchemaVersion string               `json:"schemaVersion"`
	CorrelationId string               `json:"correlationId"`
	CausationId   *uuid.UUID           `json:"causationId,omitempty"`
	OccurredAt    time.Time            `json:"occurredAt"`
	Trace         cevent.TraceMetadata `json:"trace"`
}

func NewGenericOutboxEventRepository() GenericOutboxEventRepositoryInterface {
	return &GenericOutboxEventRepository{}
}

func ConvertEnvelopeToCreateOutboxEventInput[D any](
	topic cevent.Topic,
	envelope cevent.EventEnvelope[D],
) (inputs.CreateOutboxEventInput, error) {
	if topic == "" || envelope.EventId == uuid.Nil || envelope.AggregateId == uuid.Nil ||
		envelope.AggregateType == "" || envelope.EventType == "" || envelope.KafkaKey == "" {
		return inputs.CreateOutboxEventInput{}, errors.New("outbox event envelope is incomplete")
	}
	if envelope.KafkaKey != envelope.AggregateId.String() {
		return inputs.CreateOutboxEventInput{}, errors.New("outbox event Kafka key must equal the aggregate ID")
	}

	payload, err := json.Marshal(envelope.Data)
	if err != nil {
		return inputs.CreateOutboxEventInput{}, err
	}
	metadata, err := json.Marshal(outboxEventMetadata{
		SchemaVersion: envelope.SchemaVersion,
		CorrelationId: envelope.CorrelationId,
		CausationId:   envelope.CausationId,
		OccurredAt:    envelope.OccurredAt,
		Trace:         envelope.Trace,
	})
	if err != nil {
		return inputs.CreateOutboxEventInput{}, err
	}

	return inputs.CreateOutboxEventInput{
		Id:            envelope.EventId,
		AggregateType: envelope.AggregateType,
		AggregateId:   envelope.AggregateId,
		EventType:     envelope.EventType,
		Topic:         topic,
		KafkaKey:      envelope.KafkaKey,
		Payload:       payload,
		Metadata:      metadata,
		AvailableAt:   time.Now(),
	}, nil
}

func EnqueueOutboxEvents[D any](
	tx *gorm.DB,
	topic cevent.Topic,
	envelopes []cevent.EventEnvelope[D],
) error {
	if len(envelopes) == 0 {
		return nil
	}

	createInputs := make([]inputs.CreateOutboxEventInput, len(envelopes))
	for index, envelope := range envelopes {
		createInput, err := ConvertEnvelopeToCreateOutboxEventInput(topic, envelope)
		if err != nil {
			return err
		}
		createInputs[index] = createInput
	}

	if exception := (&GenericOutboxEventRepository{}).CreateMany(
		createInputs,
		RepositoryOptionFields{
			DB:                   tx,
			IsTransactionStarted: true,
			BatchSize:            1000,
		},
	); exception != nil {
		return exception
	}
	return nil
}

func (r *GenericOutboxEventRepository) CreateMany(
	createInputs []inputs.CreateOutboxEventInput,
	parsedOptions RepositoryOptionFields,
) *cexceptions.Exception {
	if len(createInputs) == 0 {
		return nil
	}

	if !parsedOptions.IsTransactionStarted {
		return cexceptions.New("TransactionRequired", "Outbox", "Create", "Outbox events must be created in the domain transaction", http.StatusInternalServerError)
	}

	events := make([]schemas.OutboxEvent, len(createInputs))
	for index, createInput := range createInputs {
		events[index] = schemas.OutboxEvent{
			Id:            createInput.Id,
			AggregateType: createInput.AggregateType,
			AggregateId:   createInput.AggregateId,
			EventType:     createInput.EventType,
			Topic:         createInput.Topic,
			KafkaKey:      createInput.KafkaKey,
			Payload:       datatypes.JSON(createInput.Payload),
			Metadata:      datatypes.JSON(createInput.Metadata),
			AvailableAt:   createInput.AvailableAt,
		}
	}

	result := parsedOptions.DB.CreateInBatches(&events, parsedOptions.BatchSize)
	if result.Error != nil {
		return cexceptions.New("FailedToCreate", "Outbox", "Create", "Failed to create outbox events", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	return nil
}

func SerializeOutboxEvent(event schemas.OutboxEvent) ([]byte, error) {
	var metadata outboxEventMetadata
	if err := json.Unmarshal(event.Metadata, &metadata); err != nil {
		return nil, err
	}

	var payload json.RawMessage
	if err := json.Unmarshal(event.Payload, &payload); err != nil {
		return nil, err
	}

	return json.Marshal(cevent.EventEnvelope[json.RawMessage]{
		SchemaVersion: metadata.SchemaVersion,
		EventId:       event.Id,
		EventType:     event.EventType,
		AggregateType: event.AggregateType,
		AggregateId:   event.AggregateId,
		KafkaKey:      event.KafkaKey,
		OccurredAt:    metadata.OccurredAt,
		CorrelationId: metadata.CorrelationId,
		CausationId:   metadata.CausationId,
		Trace:         metadata.Trace,
		Data:          payload,
	})
}

func (r *GenericOutboxEventRepository) ClaimAvailable(
	ctx context.Context,
	workerId string,
	batchSize int,
	claimTimeout time.Duration,
	parsedOptions RepositoryOptionFields,
) ([]schemas.OutboxEvent, *cexceptions.Exception) {
	now := time.Now()
	expiredAt := now.Add(-claimTimeout)
	tx := parsedOptions.DB.WithContext(ctx).Begin()
	lockingStrength := "UPDATE"

	var events []schemas.OutboxEvent
	result := tx.
		Model(&schemas.OutboxEvent{}).
		Where("published_at IS NULL").
		Where("available_at <= ?", now).
		Where("claimed_at IS NULL OR claimed_at <= ?", expiredAt).
		Order("created_at ASC").
		Limit(batchSize).
		Scopes(scopes.Locking(&lockingStrength, "SKIP LOCKED")).
		Find(&events)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New("FailedToGet", "Outbox", "Claim", "Failed to claim available outbox events", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	if len(events) == 0 {
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return nil, cexceptions.New("TransactionCommitFailed", "Outbox", "Claim", "Failed to commit the empty outbox claim transaction", http.StatusInternalServerError, true).WithOrigin(err)
		}
		return events, nil
	}

	eventIds := make([]uuid.UUID, len(events))
	for index, event := range events {
		eventIds[index] = event.Id
	}
	result = tx.Model(&schemas.OutboxEvent{}).
		Where("id IN ?", eventIds).
		Updates(map[string]any{"claimed_by": workerId, "claimed_at": now})
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New("FailedToUpdate", "Outbox", "Claim", "Failed to claim available outbox events", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New("TransactionCommitFailed", "Outbox", "Claim", "Failed to commit the outbox claim transaction", http.StatusInternalServerError, true).WithOrigin(err)
	}

	for index := range events {
		events[index].ClaimedBy = &workerId
		events[index].ClaimedAt = &now
	}
	return events, nil
}

func (r *GenericOutboxEventRepository) MarkPublishedMany(
	ctx context.Context,
	eventIds []uuid.UUID,
	workerId string,
	parsedOptions RepositoryOptionFields,
) *cexceptions.Exception {
	if len(eventIds) == 0 {
		return nil
	}
	result := parsedOptions.DB.WithContext(ctx).
		Model(&schemas.OutboxEvent{}).
		Where("id IN ? AND claimed_by = ? AND published_at IS NULL", eventIds, workerId).
		Updates(map[string]any{"published_at": time.Now(), "last_error": nil, "claimed_by": nil, "claimed_at": nil})
	if result.Error != nil {
		return cexceptions.New("FailedToUpdate", "Outbox", "MarkPublished", "Failed to mark outbox events as published", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	return nil
}

func (r *GenericOutboxEventRepository) MarkFailedMany(
	ctx context.Context,
	failureInputs []inputs.FailedOutboxEventInput,
	workerId string,
	parsedOptions RepositoryOptionFields,
) *cexceptions.Exception {
	if len(failureInputs) == 0 {
		return nil
	}
	valuePlaceholders := make([]string, 0, len(failureInputs))
	valueArguments := make([]any, 0, len(failureInputs)*3+1)
	for _, failureInput := range failureInputs {
		valuePlaceholders = append(valuePlaceholders, "(?::uuid, ?::text, ?::timestamptz)")
		valueArguments = append(valueArguments, failureInput.Id, failureInput.LastError, failureInput.AvailableAt)
	}
	valueArguments = append(valueArguments, workerId)
	query := fmt.Sprintf(`
		UPDATE "OutboxEventTable" AS outbox_event
		SET available_at = value.available_at,
			publish_count = outbox_event.publish_count + 1,
			last_error = value.last_error,
			claimed_by = NULL,
			claimed_at = NULL
		FROM (VALUES %s) AS value(id, last_error, available_at)
		WHERE outbox_event.id = value.id
			AND outbox_event.claimed_by = ?
			AND outbox_event.published_at IS NULL
	`, strings.Join(valuePlaceholders, ","))
	result := parsedOptions.DB.WithContext(ctx).Exec(query, valueArguments...)
	if result.Error != nil {
		return cexceptions.New("FailedToUpdate", "Outbox", "MarkFailed", "Failed to schedule outbox event retries", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	return nil
}

func (r *GenericOutboxEventRepository) DeletePublishedBefore(
	ctx context.Context,
	publishedBefore time.Time,
	parsedOptions RepositoryOptionFields,
) (int64, *cexceptions.Exception) {
	result := parsedOptions.DB.WithContext(ctx).
		Where("published_at IS NOT NULL AND published_at < ?", publishedBefore).
		Delete(&schemas.OutboxEvent{})
	if result.Error != nil {
		return 0, cexceptions.New("FailedToDelete", "Outbox", "Cleanup", "Failed to delete published outbox events", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	return result.RowsAffected, nil
}
