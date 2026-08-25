package services

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cnotifications "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/api"
	cnotificationtypes "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/types"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	searchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"

	repositories "github.com/HiIamJeff67/notegic-backend/internal/notification/data/postgres/repositories"
	notificationexceptions "github.com/HiIamJeff67/notegic-backend/internal/notification/exceptions"
)

type NotificationServiceInterface interface {
	ConsumeNotificationRequested(
		ctx context.Context,
		event cevent.EventEnvelope[coreevents.NotificationRequestedData],
	) error
	SearchPrivateNotifications(
		ctx context.Context,
		request *cnotifications.SearchPrivateNotificationsRequestDto,
	) (*cnotifications.SearchPrivateNotificationsResponseDto, error)
	CountMyUnreadNotifications(
		ctx context.Context,
		request *cnotifications.CountUnreadNotificationsRequestDto,
	) (*cnotifications.CountUnreadNotificationsResponseDto, error)
	MarkMyNotificationsRead(
		ctx context.Context,
		request *cnotifications.MarkNotificationsReadRequestDto,
	) (*cnotifications.MarkNotificationsReadResponseDto, error)
	SoftDeleteMyNotifications(
		ctx context.Context,
		request *cnotifications.DeleteNotificationsRequestDto,
	) (*cnotifications.DeleteNotificationsResponseDto, error)
	HardDeleteExpiredNotifications(ctx context.Context, now time.Time, retention time.Duration) (int64, error)
	DeleteAllNotificationsForUser(ctx context.Context, userPublicId uuid.UUID) error
}

type NotificationService struct {
	repository repositories.NotificationRepository
	validator  *validator.Validate
}

func NewNotificationService(
	repository repositories.NotificationRepository,
	notificationValidator *validator.Validate,
) NotificationServiceInterface {
	return &NotificationService{
		repository: repository,
		validator:  notificationValidator,
	}
}

/* ============================== Service Methods for Notification ============================== */

func (s *NotificationService) ConsumeNotificationRequested(
	ctx context.Context,
	event cevent.EventEnvelope[coreevents.NotificationRequestedData],
) error {
	if event.EventType != coreevents.EventType_NotificationRequested {
		return notificationexceptions.NewEventException("Notification").UnsupportedEventType()
	}
	if event.AggregateId != event.Data.RecipientUserPublicId {
		return notificationexceptions.NewEventException("Notification").AggregateRecipientMismatch()
	}
	if err := s.validator.Struct(cnotificationtypes.NotificationMetadata{
		Type:            string(event.Data.Type),
		Priority:        string(event.Data.Priority),
		TemplateVersion: event.Data.TemplateVersion,
	}); err != nil {
		return notificationexceptions.NewEventException("Notification").InvalidMetadata(err)
	}
	if event.Data.TemplateVersion != 1 {
		return notificationexceptions.NewEventException("Notification").UnsupportedTemplateVersion(
			fmt.Errorf("version: %d", event.Data.TemplateVersion),
		)
	}
	switch event.Data.Type {
	case coreevents.NotificationType_News:
		if event.Data.TemplateKey != cnotificationtypes.TemplateKey_News {
			return notificationexceptions.NewEventException("Notification").InvalidNewsTemplateKey()
		}
		var payload cnotificationtypes.NewsPayload
		if err := json.Unmarshal(event.Data.Payload, &payload); err != nil {
			return notificationexceptions.NewPayloadException("Notification").PayloadDecodeFailed(err)
		}
		if err := s.validator.Struct(payload); err != nil {
			return notificationexceptions.NewPayloadException("Notification").InvalidNewsPayload(err)
		}
	case coreevents.NotificationType_Warning:
		if event.Data.TemplateKey != cnotificationtypes.TemplateKey_Warning {
			return notificationexceptions.NewEventException("Notification").InvalidWarningTemplateKey()
		}
		var payload cnotificationtypes.WarningPayload
		if err := json.Unmarshal(event.Data.Payload, &payload); err != nil {
			return notificationexceptions.NewPayloadException("Notification").PayloadDecodeFailed(err)
		}
		if err := s.validator.Struct(payload); err != nil {
			return notificationexceptions.NewPayloadException("Notification").InvalidWarningPayload(err)
		}
	case coreevents.NotificationType_Important:
		if event.Data.TemplateKey != cnotificationtypes.TemplateKey_Important {
			return notificationexceptions.NewEventException("Notification").InvalidImportantTemplateKey()
		}
		var payload cnotificationtypes.ImportantPayload
		if err := json.Unmarshal(event.Data.Payload, &payload); err != nil {
			return notificationexceptions.NewPayloadException("Notification").PayloadDecodeFailed(err)
		}
		if err := s.validator.Struct(payload); err != nil {
			return notificationexceptions.NewPayloadException("Notification").InvalidImportantPayload(err)
		}
	default:
		return notificationexceptions.NewEventException("Notification").UnsupportedType(
			fmt.Errorf("type: %q", event.Data.Type),
		)
	}

	if err := s.repository.CreateFromRequest(ctx, event); err != nil {
		return notificationexceptions.NewOperationException("Notification").CreateFailed(err)
	}
	return nil
}

func (s *NotificationService) SearchPrivateNotifications(
	ctx context.Context,
	request *cnotifications.SearchPrivateNotificationsRequestDto,
) (*cnotifications.SearchPrivateNotificationsResponseDto, error) {
	startTime := time.Now()

	if request == nil || request.RecipientUserPublicId == uuid.Nil {
		return nil, notificationexceptions.NewRequestException("Notification").RecipientRequired()
	}
	if err := s.validator.Struct(request); err != nil {
		return nil, notificationexceptions.NewRequestException("Notification").InvalidSearchRequest(err)
	}

	limit := request.First
	if limit <= 0 || limit > 100 {
		limit = 50
	}

	var cursor *searchcursor.SearchCursor[cnotifications.SearchNotificationCursorFields]
	if request.After != nil && strings.TrimSpace(*request.After) != "" {
		decodedCursor, err := searchcursor.Decode[cnotifications.SearchNotificationCursorFields](*request.After)
		if err != nil {
			return nil, notificationexceptions.NewRequestException("Notification").InvalidSearchRequest(err)
		}
		if decodedCursor.Fields.CreatedAt.IsZero() || decodedCursor.Fields.Id == uuid.Nil {
			return nil, notificationexceptions.NewRequestException("Notification").InvalidSearchRequest(
				fmt.Errorf("notification search cursor is incomplete"),
			)
		}
		cursor = decodedCursor
	}

	var beforeCreatedAt *time.Time
	var beforeId *uuid.UUID
	if cursor != nil {
		beforeCreatedAt = &cursor.Fields.CreatedAt
		beforeId = &cursor.Fields.Id
	}

	notifications, err := s.repository.List(
		ctx,
		request.RecipientUserPublicId,
		beforeCreatedAt,
		beforeId,
		limit+1,
	)
	if err != nil {
		return nil, notificationexceptions.NewOperationException("Notification").SearchFailed(err)
	}

	hasNextPage := len(notifications) > limit
	if hasNextPage {
		notifications = notifications[:limit]
	}

	response := &cnotifications.SearchPrivateNotificationsResponseDto{
		SearchEdges: make([]cnotifications.SearchPrivateNotificationEdge, len(notifications)),
		SearchPageInfo: cnotifications.SearchNotificationPageInfo{
			HasNextPage:     hasNextPage,
			HasPreviousPage: cursor != nil,
		},
	}
	for index, notification := range notifications {
		payload := map[string]any{}
		if len(notification.Payload) > 0 {
			if err := json.Unmarshal(notification.Payload, &payload); err != nil {
				return nil, notificationexceptions.NewPayloadException("Notification").ResponsePayloadDecodeFailed(err)
			}
		}
		notificationResponse := cnotifications.NotificationResponseDto{
			Id:                    notification.Id,
			RecipientUserPublicId: notification.RecipientUserPublicId,
			Type:                  notification.Type,
			Priority:              notification.Priority,
			TemplateKey:           notification.TemplateKey,
			TemplateVersion:       notification.TemplateVersion,
			Payload:               payload,
			CreatedAt:             notification.CreatedAt,
			ReadAt:                notification.ReadAt,
			DeletedAt:             notification.DeletedAt,
			ExpiresAt:             notification.ExpiresAt,
		}
		encodedCursor, err := searchcursor.EncodeFromData(cnotifications.SearchNotificationCursorFields{
			CreatedAt: notification.CreatedAt,
			Id:        notification.Id,
		})
		if err != nil || encodedCursor == nil {
			if err == nil {
				err = fmt.Errorf("encoded notification cursor is nil")
			}
			return nil, notificationexceptions.NewOperationException("Notification").SearchFailed(err)
		}
		response.SearchEdges[index] = cnotifications.SearchPrivateNotificationEdge{
			EncodedSearchCursor: *encodedCursor,
			Node:                notificationResponse,
		}
	}
	if len(response.SearchEdges) > 0 {
		response.SearchPageInfo.StartEncodedSearchCursor = &response.SearchEdges[0].EncodedSearchCursor
		response.SearchPageInfo.EndEncodedSearchCursor = &response.SearchEdges[len(response.SearchEdges)-1].EncodedSearchCursor
	}
	response.TotalCount = int32(len(response.SearchEdges))
	response.SearchTime = float64(time.Since(startTime).Nanoseconds()) / 1e6

	return response, nil
}

func (s *NotificationService) CountMyUnreadNotifications(
	ctx context.Context,
	request *cnotifications.CountUnreadNotificationsRequestDto,
) (*cnotifications.CountUnreadNotificationsResponseDto, error) {
	if request == nil || request.RecipientUserPublicId == uuid.Nil {
		return nil, notificationexceptions.NewRequestException("Notification").RecipientRequired()
	}
	if err := s.validator.Struct(request); err != nil {
		return nil, notificationexceptions.NewRequestException("Notification").InvalidCountRequest(err)
	}
	count, err := s.repository.CountUnread(ctx, request.RecipientUserPublicId)
	if err != nil {
		return nil, notificationexceptions.NewOperationException("Notification").CountUnreadFailed(err)
	}

	return &cnotifications.CountUnreadNotificationsResponseDto{Count: count}, nil
}

func (s *NotificationService) MarkMyNotificationsRead(
	ctx context.Context,
	request *cnotifications.MarkNotificationsReadRequestDto,
) (*cnotifications.MarkNotificationsReadResponseDto, error) {
	if request == nil || request.RecipientUserPublicId == uuid.Nil {
		return nil, notificationexceptions.NewRequestException("Notification").RecipientRequired()
	}
	if err := s.validator.Struct(request); err != nil {
		return nil, notificationexceptions.NewRequestException("Notification").InvalidMarkReadRequest(err)
	}
	count, err := s.repository.MarkRead(ctx, request.RecipientUserPublicId, request.NotificationIds)
	if err != nil {
		return nil, notificationexceptions.NewOperationException("Notification").MarkReadFailed(err)
	}

	return &cnotifications.MarkNotificationsReadResponseDto{UpdatedCount: count}, nil
}

func (s *NotificationService) SoftDeleteMyNotifications(
	ctx context.Context,
	request *cnotifications.DeleteNotificationsRequestDto,
) (*cnotifications.DeleteNotificationsResponseDto, error) {
	if request == nil || request.RecipientUserPublicId == uuid.Nil {
		return nil, notificationexceptions.NewRequestException("Notification").RecipientRequired()
	}
	if err := s.validator.Struct(request); err != nil {
		return nil, notificationexceptions.NewRequestException("Notification").InvalidDeleteRequest(err)
	}
	count, err := s.repository.SoftDelete(ctx, request.RecipientUserPublicId, request.NotificationIds)
	if err != nil {
		return nil, notificationexceptions.NewOperationException("Notification").DeleteFailed(err)
	}

	return &cnotifications.DeleteNotificationsResponseDto{DeletedCount: count}, nil
}

func (s *NotificationService) HardDeleteExpiredNotifications(
	ctx context.Context,
	now time.Time,
	retention time.Duration,
) (int64, error) {
	count, err := s.repository.DeleteExpired(ctx, now, retention)
	if err != nil {
		return 0, notificationexceptions.NewOperationException("Notification").HardDeleteFailed(err)
	}
	return count, nil
}

func (s *NotificationService) DeleteAllNotificationsForUser(
	ctx context.Context,
	userPublicId uuid.UUID,
) error {
	if userPublicId == uuid.Nil {
		return notificationexceptions.NewRequestException("Notification").UserRequired()
	}

	_, err := s.repository.DeleteForUser(ctx, userPublicId)
	if err != nil {
		return notificationexceptions.NewOperationException("Notification").DeleteAllForUserFailed(err)
	}
	return nil
}
