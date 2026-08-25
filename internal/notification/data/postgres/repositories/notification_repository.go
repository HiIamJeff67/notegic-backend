package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cnotificationevents "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	cinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/inputs"
	crepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	platformschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type NotificationRepository interface {
	CreateFromRequest(ctx context.Context, event cevent.EventEnvelope[coreevents.NotificationRequestedData]) error
	List(
		ctx context.Context,
		userPublicId uuid.UUID,
		beforeCreatedAt *time.Time,
		beforeId *uuid.UUID,
		limit int,
	) ([]schemas.Notification, error)
	CountUnread(ctx context.Context, userPublicId uuid.UUID) (int64, error)
	MarkRead(ctx context.Context, userPublicId uuid.UUID, notificationIds []uuid.UUID) (int64, error)
	SoftDelete(ctx context.Context, userPublicId uuid.UUID, notificationIds []uuid.UUID) (int64, error)
	DeleteForUser(ctx context.Context, userPublicId uuid.UUID) (int64, error)
	DeleteExpired(ctx context.Context, now time.Time, retention time.Duration) (int64, error)
	ClaimOutbox(ctx context.Context, workerId string, batchSize int, claimTimeout time.Duration) ([]platformschemas.OutboxEvent, error)
	MarkOutboxPublished(ctx context.Context, workerId string, eventIds []uuid.UUID) error
	MarkOutboxFailed(ctx context.Context, workerId string, eventIds []uuid.UUID, message string, availableAt time.Time) error
	DeletePublishedOutbox(ctx context.Context, publishedBefore time.Time) (int64, error)
}

type NotificationRepositoryImpl struct {
	db *gorm.DB
}

func NewNotificationRepository(db *gorm.DB) NotificationRepository {
	return &NotificationRepositoryImpl{db: db}
}

func (r *NotificationRepositoryImpl) CreateFromRequest(
	ctx context.Context,
	event cevent.EventEnvelope[coreevents.NotificationRequestedData],
) error {
	if r == nil || r.db == nil {
		return errors.New("notification repository database is required")
	}
	if event.EventId == uuid.Nil || event.Data.RecipientUserPublicId == uuid.Nil || event.Data.DedupeKey == "" {
		return errors.New("notification request is incomplete")
	}

	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		isNewInboxEvent, inboxException := crepositories.NewInboxEventRepository().CreateOne(
			cinputs.CreateInboxEventInput{EventId: event.EventId},
			crepositories.RepositoryOptionFields{DB: tx, IsTransactionStarted: true},
		)
		if inboxException != nil {
			return inboxException
		}
		if !isNewInboxEvent {
			return nil
		}

		_, userViewException := crepositories.NewUserViewRepository().GetOneByPublicId(
			cinputs.GetUserViewByPublicIdInput{PublicId: event.Data.RecipientUserPublicId},
			crepositories.RepositoryOptionFields{DB: tx, IsTransactionStarted: true},
		)
		if userViewException != nil {
			if userViewException.Reason == "NotFound" {
				return nil
			}
			return userViewException
		}

		notification := schemas.Notification{
			Id:                    uuid.New(),
			RecipientUserPublicId: event.Data.RecipientUserPublicId,
			Type:                  string(event.Data.Type),
			Priority:              string(event.Data.Priority),
			TemplateKey:           event.Data.TemplateKey,
			TemplateVersion:       event.Data.TemplateVersion,
			Payload:               datatypes.JSON(event.Data.Payload),
			DedupeKey:             event.Data.DedupeKey,
			CreatedAt:             event.OccurredAt,
			ExpiresAt:             event.Data.ExpiresAt,
		}
		result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&notification)
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return nil
		}

		createdData := cnotificationevents.NotificationCreatedData{
			NotificationId:        notification.Id,
			RecipientUserPublicId: notification.RecipientUserPublicId,
			Type:                  notification.Type,
			Priority:              notification.Priority,
			TemplateKey:           notification.TemplateKey,
			TemplateVersion:       notification.TemplateVersion,
			Payload:               json.RawMessage(notification.Payload),
			CreatedAt:             notification.CreatedAt,
			ExpiresAt:             notification.ExpiresAt,
		}
		createdEvent := cevent.EventEnvelope[cnotificationevents.NotificationCreatedData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     cnotificationevents.EventType_NotificationCreated,
			AggregateType: cnotificationevents.AggregateType_Notification,
			AggregateId:   notification.Id,
			KafkaKey:      notification.Id.String(),
			OccurredAt:    time.Now().UTC(),
			CorrelationId: event.CorrelationId,
			CausationId:   &event.EventId,
			Trace:         event.Trace,
			Data:          createdData,
		}
		payload, err := json.Marshal(createdEvent)
		if err != nil {
			return err
		}
		metadata, err := json.Marshal(map[string]any{
			"schemaVersion": cevent.Version,
			"correlationId": event.CorrelationId,
			"causationId":   event.EventId,
			"occurredAt":    createdEvent.OccurredAt,
			"trace":         event.Trace,
		})
		if err != nil {
			return err
		}

		return tx.Create(&platformschemas.OutboxEvent{
			Id:            createdEvent.EventId,
			AggregateType: createdEvent.AggregateType,
			AggregateId:   createdEvent.AggregateId,
			EventType:     createdEvent.EventType,
			Topic:         cnotificationevents.NotificationTopic,
			KafkaKey:      createdEvent.KafkaKey,
			Payload:       datatypes.JSON(payload),
			Metadata:      datatypes.JSON(metadata),
			AvailableAt:   time.Now().UTC(),
		}).Error
	})
}

func (r *NotificationRepositoryImpl) List(
	ctx context.Context,
	userPublicId uuid.UUID,
	beforeCreatedAt *time.Time,
	beforeId *uuid.UUID,
	limit int,
) ([]schemas.Notification, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	query := r.db.WithContext(ctx).
		Where("recipient_user_public_id = ?", userPublicId).
		Where("deleted_at IS NULL").
		Where("expires_at IS NULL OR expires_at > ?", time.Now().UTC()).
		Order("created_at DESC").
		Order("id DESC").
		Limit(limit)
	if beforeCreatedAt != nil && beforeId != nil {
		query = query.Where(
			"(created_at < ? OR (created_at = ? AND id < ?))",
			*beforeCreatedAt,
			*beforeCreatedAt,
			*beforeId,
		)
	}

	var notifications []schemas.Notification
	if err := query.Find(&notifications).Error; err != nil {
		return nil, err
	}

	return notifications, nil
}

func (r *NotificationRepositoryImpl) CountUnread(ctx context.Context, userPublicId uuid.UUID) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&schemas.Notification{}).
		Where("recipient_user_public_id = ?", userPublicId).
		Where("read_at IS NULL AND deleted_at IS NULL").
		Where("expires_at IS NULL OR expires_at > ?", time.Now().UTC()).
		Count(&count).Error

	return count, err
}

func (r *NotificationRepositoryImpl) MarkRead(
	ctx context.Context,
	userPublicId uuid.UUID,
	notificationIds []uuid.UUID,
) (int64, error) {
	if len(notificationIds) == 0 {
		return 0, nil
	}
	result := r.db.WithContext(ctx).
		Model(&schemas.Notification{}).
		Where("recipient_user_public_id = ? AND id IN ? AND deleted_at IS NULL", userPublicId, notificationIds).
		Where("read_at IS NULL").
		Updates(map[string]any{"read_at": time.Now().UTC()})

	return result.RowsAffected, result.Error
}

func (r *NotificationRepositoryImpl) SoftDelete(
	ctx context.Context,
	userPublicId uuid.UUID,
	notificationIds []uuid.UUID,
) (int64, error) {
	if len(notificationIds) == 0 {
		return 0, nil
	}
	result := r.db.WithContext(ctx).
		Model(&schemas.Notification{}).
		Where("recipient_user_public_id = ? AND id IN ? AND deleted_at IS NULL", userPublicId, notificationIds).
		Updates(map[string]any{"deleted_at": time.Now().UTC()})

	return result.RowsAffected, result.Error
}

func (r *NotificationRepositoryImpl) DeleteForUser(
	ctx context.Context,
	userPublicId uuid.UUID,
) (int64, error) {
	if userPublicId == uuid.Nil {
		return 0, errors.New("user public ID is required")
	}

	var deletedCount int64
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Where("recipient_user_public_id = ?", userPublicId).
			Delete(&schemas.Notification{})
		deletedCount = result.RowsAffected
		return result.Error
	})

	return deletedCount, err
}

func (r *NotificationRepositoryImpl) DeleteExpired(
	ctx context.Context,
	now time.Time,
	retention time.Duration,
) (int64, error) {
	cutoff := now.Add(-retention)
	result := r.db.WithContext(ctx).
		Where("(expires_at IS NOT NULL AND expires_at <= ?) OR (deleted_at IS NOT NULL AND deleted_at <= ?)", now, cutoff).
		Delete(&schemas.Notification{})

	return result.RowsAffected, result.Error
}

func (r *NotificationRepositoryImpl) ClaimOutbox(
	ctx context.Context,
	workerId string,
	batchSize int,
	claimTimeout time.Duration,
) ([]platformschemas.OutboxEvent, error) {
	if batchSize <= 0 {
		return nil, nil
	}

	var events []platformschemas.OutboxEvent
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		claimBefore := time.Now().UTC().Add(-claimTimeout)
		query := tx.Where("published_at IS NULL").
			Where("available_at <= ?", time.Now().UTC()).
			Where("claimed_at IS NULL OR claimed_at < ?", claimBefore).
			Order("available_at ASC").
			Limit(batchSize).
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"})
		if err := query.Find(&events).Error; err != nil {
			return err
		}
		if len(events) == 0 {
			return nil
		}

		ids := make([]uuid.UUID, len(events))
		for index, event := range events {
			ids[index] = event.Id
		}
		claimedAt := time.Now().UTC()
		result := tx.Model(&platformschemas.OutboxEvent{}).
			Where("id IN ?", ids).
			Updates(map[string]any{
				"claimed_by": workerId,
				"claimed_at": claimedAt,
			})
		return result.Error
	})
	if err != nil {
		return nil, err
	}

	return events, nil
}

func (r *NotificationRepositoryImpl) MarkOutboxPublished(
	ctx context.Context,
	workerId string,
	eventIds []uuid.UUID,
) error {
	if len(eventIds) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.WithContext(ctx).
		Model(&platformschemas.OutboxEvent{}).
		Where("id IN ? AND claimed_by = ? AND published_at IS NULL", eventIds, workerId).
		Updates(map[string]any{
			"published_at":  now,
			"publish_count": gorm.Expr("publish_count + 1"),
			"claimed_by":    nil,
			"claimed_at":    nil,
		}).Error
}

func (r *NotificationRepositoryImpl) MarkOutboxFailed(
	ctx context.Context,
	workerId string,
	eventIds []uuid.UUID,
	message string,
	availableAt time.Time,
) error {
	if len(eventIds) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Model(&platformschemas.OutboxEvent{}).
		Where("id IN ? AND claimed_by = ? AND published_at IS NULL", eventIds, workerId).
		Updates(map[string]any{
			"last_error":    message,
			"available_at":  availableAt,
			"publish_count": gorm.Expr("publish_count + 1"),
			"claimed_by":    nil,
			"claimed_at":    nil,
		}).Error
}

func (r *NotificationRepositoryImpl) DeletePublishedOutbox(
	ctx context.Context,
	publishedBefore time.Time,
) (int64, error) {
	result := r.db.WithContext(ctx).
		Where("published_at IS NOT NULL AND published_at < ?", publishedBefore).
		Delete(&platformschemas.OutboxEvent{})

	return result.RowsAffected, result.Error
}
