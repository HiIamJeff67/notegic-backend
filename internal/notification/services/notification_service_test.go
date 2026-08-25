package services

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cnotifications "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/api"
	cnotificationtypes "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/types"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	platformschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	sharedvalidations "github.com/HiIamJeff67/notegic-backend/shared/validations"

	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	notificationvalidations "github.com/HiIamJeff67/notegic-backend/internal/notification/validations"
)

type notificationRepositoryStub struct {
	createCalls        int
	createErr          error
	deleteForUserCalls int
	deleteForUserErr   error
	notifications      []schemas.Notification
}

func (r *notificationRepositoryStub) CreateFromRequest(
	context.Context,
	cevent.EventEnvelope[coreevents.NotificationRequestedData],
) error {
	r.createCalls++
	return r.createErr
}

func (r *notificationRepositoryStub) List(context.Context, uuid.UUID, *time.Time, *uuid.UUID, int) ([]schemas.Notification, error) {
	return r.notifications, nil
}

func (r *notificationRepositoryStub) CountUnread(context.Context, uuid.UUID) (int64, error) {
	return 0, nil
}

func (r *notificationRepositoryStub) MarkRead(context.Context, uuid.UUID, []uuid.UUID) (int64, error) {
	return 0, nil
}

func (r *notificationRepositoryStub) SoftDelete(context.Context, uuid.UUID, []uuid.UUID) (int64, error) {
	return 0, nil
}

func (r *notificationRepositoryStub) DeleteForUser(context.Context, uuid.UUID) (int64, error) {
	r.deleteForUserCalls++
	return 0, r.deleteForUserErr
}

func (r *notificationRepositoryStub) DeleteExpired(context.Context, time.Time, time.Duration) (int64, error) {
	return 0, nil
}

func (r *notificationRepositoryStub) ClaimOutbox(context.Context, string, int, time.Duration) ([]platformschemas.OutboxEvent, error) {
	return nil, nil
}

func (r *notificationRepositoryStub) MarkOutboxPublished(context.Context, string, []uuid.UUID) error {
	return nil
}

func (r *notificationRepositoryStub) MarkOutboxFailed(context.Context, string, []uuid.UUID, string, time.Time) error {
	return nil
}

func (r *notificationRepositoryStub) DeletePublishedOutbox(context.Context, time.Time) (int64, error) {
	return 0, nil
}

func newNotificationServiceForTest(repository *notificationRepositoryStub) NotificationServiceInterface {
	validate := validator.New()
	sharedvalidations.RegisterStringsValidation(validate)
	sharedvalidations.RegisterTimesValidation(validate)
	notificationvalidations.RegisterNotificationValidation(validate)
	notificationvalidations.RegisterNewsValidation(validate)
	notificationvalidations.RegisterWarningValidation(validate)
	notificationvalidations.RegisterImportantValidation(validate)

	return NewNotificationService(repository, validate)
}

func TestConsumeRequestedValidatesPayloadBeforePersisting(t *testing.T) {
	repository := &notificationRepositoryStub{}
	service := newNotificationServiceForTest(repository)
	recipientUserPublicId := uuid.New()
	payload, err := json.Marshal(cnotificationtypes.NewsPayload{
		Title:   "Release update",
		Summary: "A new release is available.",
		Body:    "Read the release notes for more details.",
	})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	event := cevent.EventEnvelope[coreevents.NotificationRequestedData]{
		SchemaVersion: cevent.Version,
		EventId:       uuid.New(),
		EventType:     coreevents.EventType_NotificationRequested,
		AggregateType: coreevents.AggregateType_Notification,
		AggregateId:   recipientUserPublicId,
		KafkaKey:      recipientUserPublicId.String(),
		OccurredAt:    time.Now().UTC(),
		CorrelationId: "request-id",
		Data: coreevents.NotificationRequestedData{
			RecipientUserPublicId: recipientUserPublicId,
			Type:                  coreevents.NotificationType_News,
			Priority:              coreevents.NotificationPriority_Normal,
			TemplateKey:           cnotificationtypes.TemplateKey_News,
			TemplateVersion:       1,
			Payload:               payload,
			DedupeKey:             "welcome:" + recipientUserPublicId.String(),
		},
	}

	if err := service.ConsumeNotificationRequested(context.Background(), event); err != nil {
		t.Fatalf("expected valid notification request, got %v", err)
	}
	if repository.createCalls != 1 {
		t.Fatalf("expected one repository call, got %d", repository.createCalls)
	}
}

func TestConsumeRequestedRejectsInvalidPayloadBeforePersisting(t *testing.T) {
	repository := &notificationRepositoryStub{}
	service := newNotificationServiceForTest(repository)
	recipientUserPublicId := uuid.New()
	payload, err := json.Marshal(cnotificationtypes.NewsPayload{Title: ""})
	if err != nil {
		t.Fatalf("failed to encode payload: %v", err)
	}

	event := cevent.EventEnvelope[coreevents.NotificationRequestedData]{
		EventId:       uuid.New(),
		EventType:     coreevents.EventType_NotificationRequested,
		AggregateType: coreevents.AggregateType_Notification,
		AggregateId:   recipientUserPublicId,
		KafkaKey:      recipientUserPublicId.String(),
		OccurredAt:    time.Now().UTC(),
		Data: coreevents.NotificationRequestedData{
			RecipientUserPublicId: recipientUserPublicId,
			Type:                  coreevents.NotificationType_News,
			Priority:              coreevents.NotificationPriority_Normal,
			TemplateKey:           cnotificationtypes.TemplateKey_News,
			TemplateVersion:       1,
			Payload:               payload,
			DedupeKey:             "invalid:" + recipientUserPublicId.String(),
		},
	}

	if err := service.ConsumeNotificationRequested(context.Background(), event); err == nil {
		t.Fatal("expected invalid notification payload to be rejected")
	}
	if repository.createCalls != 0 {
		t.Fatalf("expected no repository call for invalid payload, got %d", repository.createCalls)
	}
}

func TestDeleteForUserDelegatesToRepository(t *testing.T) {
	repository := &notificationRepositoryStub{}
	service := newNotificationServiceForTest(repository)

	if err := service.DeleteAllNotificationsForUser(context.Background(), uuid.New()); err != nil {
		t.Fatalf("expected user notification deletion to succeed, got %v", err)
	}
	if repository.deleteForUserCalls != 1 {
		t.Fatalf("expected one user notification deletion call, got %d", repository.deleteForUserCalls)
	}
}

func TestSearchPrivateNotificationsReturnsGraphQLStyleCursorPage(t *testing.T) {
	recipientUserPublicId := uuid.New()
	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	repository := &notificationRepositoryStub{
		notifications: []schemas.Notification{
			{
				Id:                    uuid.New(),
				RecipientUserPublicId: recipientUserPublicId,
				Type:                  "news",
				Priority:              "normal",
				TemplateKey:           "news",
				TemplateVersion:       1,
				Payload:               []byte(`{"title":"Release","summary":"Summary","body":"Body"}`),
				CreatedAt:             createdAt,
			},
			{
				Id:                    uuid.New(),
				RecipientUserPublicId: recipientUserPublicId,
				Type:                  "news",
				Priority:              "normal",
				TemplateKey:           "news",
				TemplateVersion:       1,
				Payload:               []byte(`{"title":"Older release","summary":"Summary","body":"Body"}`),
				CreatedAt:             createdAt.Add(-time.Minute),
			},
		},
	}
	service := newNotificationServiceForTest(repository)

	response, err := service.SearchPrivateNotifications(context.Background(), &cnotifications.SearchPrivateNotificationsRequestDto{
		RecipientUserPublicId: recipientUserPublicId,
		First:                 1,
	})
	if err != nil {
		t.Fatalf("search notifications: %v", err)
	}
	if len(response.SearchEdges) != 1 {
		t.Fatalf("search edge count = %d, want 1", len(response.SearchEdges))
	}
	if !response.SearchPageInfo.HasNextPage {
		t.Fatal("expected a next page")
	}
	if response.SearchPageInfo.EndEncodedSearchCursor == nil || *response.SearchPageInfo.EndEncodedSearchCursor == "" {
		t.Fatal("expected an opaque end cursor")
	}
	if response.SearchEdges[0].Node.Payload["title"] != "Release" {
		t.Fatalf("unexpected notification payload: %#v", response.SearchEdges[0].Node.Payload)
	}
}
