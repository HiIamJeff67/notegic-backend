package repositories

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptions "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type OutboxEventRepositoryInterface interface {
	CreateMany(createInputs []inputs.CreateOutboxEventInput, opts ...RepositoryOptions) *cexceptions.Exception
	EnqueueBlockPackAccessRevocations(tx *gorm.DB, correlationId string, blockPackIds []uuid.UUID, targetUserPublicIds []uuid.UUID, reason coreevents.BlockPackAccessRevocationReason) error
	EnqueueRootShelfPermissionChanged(tx *gorm.DB, correlationId string, rootShelfId uuid.UUID, targetUserPublicId uuid.UUID, permission string) error
	EnqueueManyRootShelfPermissionChanges(tx *gorm.DB, correlationId string, rootShelfId uuid.UUID, permissions []schemas.UsersToShelves, userPublicIdByUserId map[uuid.UUID]uuid.UUID) error
	EnqueueRootShelfPermissionRevoked(tx *gorm.DB, correlationId string, rootShelfId uuid.UUID, targetUserPublicId uuid.UUID) error
	EnqueueManyRootShelfPermissionRevocations(tx *gorm.DB, correlationId string, rootShelfIds []uuid.UUID, targetUserPublicIds []uuid.UUID) error
	EnqueueRootShelfDeleted(tx *gorm.DB, correlationId string, rootShelfId uuid.UUID, targetUserPublicIds []uuid.UUID) error
	EnqueueManyRootShelfDeleted(tx *gorm.DB, correlationId string, rootShelfIds []uuid.UUID, targetUserPublicIdsByRootShelfId map[uuid.UUID][]uuid.UUID) error
	EnqueueBlockPackChanged(tx *gorm.DB, correlationId string, blockPackIds []uuid.UUID) error
	EnqueueBlockPackDeleted(tx *gorm.DB, correlationId string, blockPackIds []uuid.UUID) error
	EnqueueUserSessionsRevoked(tx *gorm.DB, correlationId string, userPublicId uuid.UUID) error
	EnqueueNotificationRequested(tx *gorm.DB, correlationId string, data coreevents.NotificationRequestedData) error
	EnqueueYjsMaintenanceHint(tx *gorm.DB, correlationId string, blockPackId uuid.UUID, reason string) error
	EnqueueManyYjsMaintenanceHints(tx *gorm.DB, correlationId string, blockPackIds []uuid.UUID, reason string) error
	ClaimAvailable(ctx context.Context, workerId string, batchSize int, claimTimeout time.Duration, opts ...RepositoryOptions) ([]schemas.OutboxEvent, *cexceptions.Exception)
	MarkPublishedMany(ctx context.Context, eventIds []uuid.UUID, workerId string, opts ...RepositoryOptions) *cexceptions.Exception
	MarkFailedMany(ctx context.Context, failureInputs []inputs.FailedOutboxEventInput, workerId string, opts ...RepositoryOptions) *cexceptions.Exception
	DeletePublishedBefore(ctx context.Context, publishedBefore time.Time, opts ...RepositoryOptions) (int64, *cexceptions.Exception)
}

type OutboxEventRepository struct {
	db                 *gorm.DB
	contractRepository GenericOutboxEventRepositoryInterface
	exceptions         exceptions.OutboxException
}

func NewOutboxEventRepository(dbs ...*gorm.DB) OutboxEventRepositoryInterface {
	var db *gorm.DB
	if len(dbs) > 0 {
		db = dbs[0]
	}

	return &OutboxEventRepository{
		db:                 db,
		contractRepository: NewGenericOutboxEventRepository(),
		exceptions:         exceptions.NewOutboxException(),
	}
}

func (r *OutboxEventRepository) CreateMany(
	createInputs []inputs.CreateOutboxEventInput,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	return r.contractRepository.CreateMany(
		createInputs,
		parsedOptions,
	)
}

func (r *OutboxEventRepository) EnqueueBlockPackAccessRevocations(
	tx *gorm.DB,
	correlationId string,
	blockPackIds []uuid.UUID,
	targetUserPublicIds []uuid.UUID,
	reason coreevents.BlockPackAccessRevocationReason,
) error {
	if len(blockPackIds) == 0 {
		return nil
	}

	targetCount := len(targetUserPublicIds)
	if targetCount == 0 {
		targetCount = 1
	}
	events := make(
		[]cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData],
		0,
		len(blockPackIds)*targetCount,
	)
	occurredAt := time.Now().UTC()
	for _, blockPackId := range blockPackIds {
		if len(targetUserPublicIds) == 0 {
			events = append(events, cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData]{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     coreevents.EventType_BlockPackAccessRevoked,
				AggregateType: coreevents.AggregateType_BlockPack,
				AggregateId:   blockPackId,
				KafkaKey:      blockPackId.String(),
				OccurredAt:    occurredAt,
				CorrelationId: correlationId,
				Data: coreevents.BlockPackAccessRevokedData{
					Reason: reason,
				},
			})
			continue
		}

		for _, targetUserPublicId := range targetUserPublicIds {
			targetUserPublicId := targetUserPublicId
			events = append(events, cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData]{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     coreevents.EventType_BlockPackAccessRevoked,
				AggregateType: coreevents.AggregateType_BlockPack,
				AggregateId:   blockPackId,
				KafkaKey:      blockPackId.String(),
				OccurredAt:    occurredAt,
				CorrelationId: correlationId,
				Data: coreevents.BlockPackAccessRevokedData{
					TargetUserPublicId: &targetUserPublicId,
					Reason:             reason,
				},
			})
		}
	}

	return EnqueueOutboxEvents(tx, coreevents.CoreLifecycleTopic, events)
}

func (r *OutboxEventRepository) EnqueueRootShelfPermissionChanged(
	tx *gorm.DB,
	correlationId string,
	rootShelfId uuid.UUID,
	targetUserPublicId uuid.UUID,
	permission string,
) error {
	return EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		[]cevent.EventEnvelope[coreevents.ResourceChangedData]{
			{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     coreevents.EventType_RootShelfPermissionChanged,
				AggregateType: coreevents.AggregateType_RootShelf,
				AggregateId:   rootShelfId,
				KafkaKey:      rootShelfId.String(),
				OccurredAt:    time.Now().UTC(),
				CorrelationId: correlationId,
				Data: coreevents.ResourceChangedData{
					ResourceId:         rootShelfId,
					TargetUserPublicId: &targetUserPublicId,
					Change:             coreevents.ResourceEventChange_PermissionUpdated,
					Permission:         permission,
				},
			},
		},
	)
}

func (r *OutboxEventRepository) EnqueueManyRootShelfPermissionChanges(
	tx *gorm.DB,
	correlationId string,
	rootShelfId uuid.UUID,
	permissions []schemas.UsersToShelves,
	userPublicIdByUserId map[uuid.UUID]uuid.UUID,
) error {
	events := make([]cevent.EventEnvelope[coreevents.ResourceChangedData], 0, len(permissions))
	occurredAt := time.Now().UTC()
	for _, permission := range permissions {
		targetUserPublicId, exists := userPublicIdByUserId[permission.UserId]
		if !exists {
			return errors.New("root shelf permission event target user is unavailable")
		}

		events = append(events, cevent.EventEnvelope[coreevents.ResourceChangedData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_RootShelfPermissionChanged,
			AggregateType: coreevents.AggregateType_RootShelf,
			AggregateId:   rootShelfId,
			KafkaKey:      rootShelfId.String(),
			OccurredAt:    occurredAt,
			CorrelationId: correlationId,
			Data: coreevents.ResourceChangedData{
				ResourceId:         rootShelfId,
				TargetUserPublicId: &targetUserPublicId,
				Change:             coreevents.ResourceEventChange_PermissionUpdated,
				Permission:         permission.Permission.String(),
			},
		})
	}

	return EnqueueOutboxEvents(tx, coreevents.CoreLifecycleTopic, events)
}

func (r *OutboxEventRepository) EnqueueRootShelfPermissionRevoked(
	tx *gorm.DB,
	correlationId string,
	rootShelfId uuid.UUID,
	targetUserPublicId uuid.UUID,
) error {
	return EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		[]cevent.EventEnvelope[coreevents.ResourceChangedData]{
			{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     coreevents.EventType_RootShelfPermissionRevoked,
				AggregateType: coreevents.AggregateType_RootShelf,
				AggregateId:   rootShelfId,
				KafkaKey:      rootShelfId.String(),
				OccurredAt:    time.Now().UTC(),
				CorrelationId: correlationId,
				Data: coreevents.ResourceChangedData{
					ResourceId:         rootShelfId,
					TargetUserPublicId: &targetUserPublicId,
					Change:             coreevents.ResourceEventChange_PermissionRevoked,
				},
			},
		},
	)
}

func (r *OutboxEventRepository) EnqueueManyRootShelfPermissionRevocations(
	tx *gorm.DB,
	correlationId string,
	rootShelfIds []uuid.UUID,
	targetUserPublicIds []uuid.UUID,
) error {
	events := make([]cevent.EventEnvelope[coreevents.ResourceChangedData], 0, len(rootShelfIds)*len(targetUserPublicIds))
	occurredAt := time.Now().UTC()
	for _, rootShelfId := range rootShelfIds {
		for _, targetUserPublicId := range targetUserPublicIds {
			targetUserPublicId := targetUserPublicId
			events = append(events, cevent.EventEnvelope[coreevents.ResourceChangedData]{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     coreevents.EventType_RootShelfPermissionRevoked,
				AggregateType: coreevents.AggregateType_RootShelf,
				AggregateId:   rootShelfId,
				KafkaKey:      rootShelfId.String(),
				OccurredAt:    occurredAt,
				CorrelationId: correlationId,
				Data: coreevents.ResourceChangedData{
					ResourceId:         rootShelfId,
					TargetUserPublicId: &targetUserPublicId,
					Change:             coreevents.ResourceEventChange_PermissionRevoked,
				},
			})
		}
	}

	return EnqueueOutboxEvents(tx, coreevents.CoreLifecycleTopic, events)
}

func (r *OutboxEventRepository) EnqueueRootShelfDeleted(
	tx *gorm.DB,
	correlationId string,
	rootShelfId uuid.UUID,
	targetUserPublicIds []uuid.UUID,
) error {
	events := make([]cevent.EventEnvelope[coreevents.ResourceChangedData], 0, len(targetUserPublicIds))
	for _, targetUserPublicId := range targetUserPublicIds {
		targetUserPublicId := targetUserPublicId
		events = append(events, cevent.EventEnvelope[coreevents.ResourceChangedData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_RootShelfDeleted,
			AggregateType: coreevents.AggregateType_RootShelf,
			AggregateId:   rootShelfId,
			KafkaKey:      rootShelfId.String(),
			OccurredAt:    time.Now().UTC(),
			CorrelationId: correlationId,
			Data: coreevents.ResourceChangedData{
				ResourceId:         rootShelfId,
				TargetUserPublicId: &targetUserPublicId,
				Change:             coreevents.ResourceEventChange_Deleted,
			},
		})
	}

	return EnqueueOutboxEvents(tx, coreevents.CoreLifecycleTopic, events)
}

func (r *OutboxEventRepository) EnqueueManyRootShelfDeleted(
	tx *gorm.DB,
	correlationId string,
	rootShelfIds []uuid.UUID,
	targetUserPublicIdsByRootShelfId map[uuid.UUID][]uuid.UUID,
) error {
	events := make([]cevent.EventEnvelope[coreevents.ResourceChangedData], 0)
	occurredAt := time.Now().UTC()
	for _, rootShelfId := range rootShelfIds {
		for _, targetUserPublicId := range targetUserPublicIdsByRootShelfId[rootShelfId] {
			targetUserPublicId := targetUserPublicId
			events = append(events, cevent.EventEnvelope[coreevents.ResourceChangedData]{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     coreevents.EventType_RootShelfDeleted,
				AggregateType: coreevents.AggregateType_RootShelf,
				AggregateId:   rootShelfId,
				KafkaKey:      rootShelfId.String(),
				OccurredAt:    occurredAt,
				CorrelationId: correlationId,
				Data: coreevents.ResourceChangedData{
					ResourceId:         rootShelfId,
					TargetUserPublicId: &targetUserPublicId,
					Change:             coreevents.ResourceEventChange_Deleted,
				},
			})
		}
	}

	return EnqueueOutboxEvents(tx, coreevents.CoreLifecycleTopic, events)
}

func (r *OutboxEventRepository) EnqueueBlockPackChanged(
	tx *gorm.DB,
	correlationId string,
	blockPackIds []uuid.UUID,
) error {
	events := make([]cevent.EventEnvelope[coreevents.ResourceChangedData], 0, len(blockPackIds))
	for _, blockPackId := range blockPackIds {
		events = append(events, cevent.EventEnvelope[coreevents.ResourceChangedData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_BlockPackChanged,
			AggregateType: coreevents.AggregateType_BlockPack,
			AggregateId:   blockPackId,
			KafkaKey:      blockPackId.String(),
			OccurredAt:    time.Now().UTC(),
			CorrelationId: correlationId,
			Data: coreevents.ResourceChangedData{
				ResourceId: blockPackId,
				Change:     coreevents.ResourceEventChange_Updated,
			},
		})
	}

	return EnqueueOutboxEvents(tx, coreevents.CoreLifecycleTopic, events)
}

func (r *OutboxEventRepository) EnqueueBlockPackDeleted(
	tx *gorm.DB,
	correlationId string,
	blockPackIds []uuid.UUID,
) error {
	events := make([]cevent.EventEnvelope[coreevents.ResourceChangedData], 0, len(blockPackIds))
	for _, blockPackId := range blockPackIds {
		events = append(events, cevent.EventEnvelope[coreevents.ResourceChangedData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_BlockPackDeleted,
			AggregateType: coreevents.AggregateType_BlockPack,
			AggregateId:   blockPackId,
			KafkaKey:      blockPackId.String(),
			OccurredAt:    time.Now().UTC(),
			CorrelationId: correlationId,
			Data: coreevents.ResourceChangedData{
				ResourceId: blockPackId,
				Change:     coreevents.ResourceEventChange_Deleted,
			},
		})
	}

	return EnqueueOutboxEvents(tx, coreevents.CoreLifecycleTopic, events)
}

func (r *OutboxEventRepository) EnqueueUserSessionsRevoked(
	tx *gorm.DB,
	correlationId string,
	userPublicId uuid.UUID,
) error {
	return EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		[]cevent.EventEnvelope[coreevents.UserSessionsRevokedData]{
			{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     coreevents.EventType_UserSessionsRevoked,
				AggregateType: coreevents.AggregateType_User,
				AggregateId:   userPublicId,
				KafkaKey:      userPublicId.String(),
				OccurredAt:    time.Now().UTC(),
				CorrelationId: correlationId,
				Data:          coreevents.UserSessionsRevokedData{},
			},
		},
	)
}

func (r *OutboxEventRepository) EnqueueNotificationRequested(
	tx *gorm.DB,
	correlationId string,
	data coreevents.NotificationRequestedData,
) error {
	if tx == nil || data.RecipientUserPublicId == uuid.Nil ||
		data.UserProjection.PublicId != data.RecipientUserPublicId || data.Type == "" ||
		data.TemplateKey == "" || data.TemplateVersion <= 0 || data.DedupeKey == "" {
		return errors.New("notification request is incomplete")
	}

	envelope := cevent.EventEnvelope[coreevents.NotificationRequestedData]{
		SchemaVersion: cevent.Version,
		EventId:       uuid.New(),
		EventType:     coreevents.EventType_NotificationRequested,
		AggregateType: coreevents.AggregateType_Notification,
		AggregateId:   data.RecipientUserPublicId,
		KafkaKey:      data.RecipientUserPublicId.String(),
		OccurredAt:    time.Now().UTC(),
		CorrelationId: correlationId,
		Data:          data,
	}

	return EnqueueOutboxEvents(
		tx,
		coreevents.CoreNotificationTopic,
		[]cevent.EventEnvelope[coreevents.NotificationRequestedData]{envelope},
	)
}

func (r *OutboxEventRepository) EnqueueYjsMaintenanceHint(
	tx *gorm.DB,
	correlationId string,
	blockPackId uuid.UUID,
	reason string,
) error {
	if tx == nil || blockPackId == uuid.Nil {
		return errors.New("Yjs maintenance hint requires a transaction and BlockPack ID")
	}

	return r.EnqueueManyYjsMaintenanceHints(tx, correlationId, []uuid.UUID{blockPackId}, reason)
}

func (r *OutboxEventRepository) EnqueueManyYjsMaintenanceHints(
	tx *gorm.DB,
	correlationId string,
	blockPackIds []uuid.UUID,
	reason string,
) error {
	if tx == nil {
		return errors.New("Yjs maintenance hints require a transaction")
	}
	if len(blockPackIds) == 0 {
		return nil
	}

	var documents []schemas.BlockPackYjsDocument
	if err := tx.Model(&schemas.BlockPackYjsDocument{}).
		Where("block_pack_id IN ? AND deleted_at IS NULL", blockPackIds).
		Find(&documents).Error; err != nil {
		return err
	}
	if len(documents) != len(blockPackIds) {
		return errors.New("Yjs maintenance hints require documents for every BlockPack ID")
	}

	occurredAt := time.Now().UTC()
	events := make([]cevent.EventEnvelope[coreevents.YjsMaintenanceHintData], 0, len(documents))
	for _, document := range documents {
		events = append(events, cevent.EventEnvelope[coreevents.YjsMaintenanceHintData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_YjsMaintenanceHint,
			AggregateType: coreevents.AggregateType_BlockPack,
			AggregateId:   document.BlockPackId,
			KafkaKey:      document.BlockPackId.String(),
			OccurredAt:    occurredAt,
			CorrelationId: correlationId,
			Data: coreevents.YjsMaintenanceHintData{
				BlockPackId:            document.BlockPackId,
				DocumentId:             document.Id,
				LatestUpdateSequence:   document.LastUpdateSequence,
				CompactedUntilSequence: document.CompactedUntilSequence,
				ProjectedUntilSequence: document.ProjectedUntilSequence,
				LastCompactedAt:        document.LastCompactedAt,
				UncompactedUpdateCount: document.LastUpdateSequence - document.CompactedUntilSequence,
				SnapshotBytes:          len(document.Snapshot),
				StateVectorBytes:       len(document.StateVector),
				Reason:                 reason,
			},
		})
	}

	return EnqueueOutboxEvents(tx, coreevents.CoreYjsMaintenanceHintTopic, events)
}

func (r *OutboxEventRepository) ClaimAvailable(
	ctx context.Context,
	workerId string,
	batchSize int,
	claimTimeout time.Duration,
	opts ...RepositoryOptions,
) ([]schemas.OutboxEvent, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	return r.contractRepository.ClaimAvailable(ctx, workerId, batchSize, claimTimeout, parsedOptions)
}

func (r *OutboxEventRepository) MarkPublishedMany(
	ctx context.Context,
	eventIds []uuid.UUID,
	workerId string,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	return r.contractRepository.MarkPublishedMany(ctx, eventIds, workerId, parsedOptions)
}

func (r *OutboxEventRepository) MarkFailedMany(
	ctx context.Context,
	failureInputs []inputs.FailedOutboxEventInput,
	workerId string,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	return r.contractRepository.MarkFailedMany(ctx, failureInputs, workerId, parsedOptions)
}

func (r *OutboxEventRepository) DeletePublishedBefore(
	ctx context.Context,
	publishedBefore time.Time,
	opts ...RepositoryOptions,
) (int64, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	return r.contractRepository.DeletePublishedBefore(ctx, publishedBefore, parsedOptions)
}
