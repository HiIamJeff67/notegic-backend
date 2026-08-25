package repositories

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cnotificationtypes "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/types"
	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	platformschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

func TestNotificationRepositorySuppressesRequestsForDeletedUsers(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:notification-repository-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := createNotificationRepositoryTestTables(db); err != nil {
		t.Fatalf("create test database tables: %v", err)
	}

	repository := NewNotificationRepository(db)
	userPublicId := uuid.New()
	if err := db.Create(&platformschemas.UserView{PublicId: userPublicId, Status: enums.UserStatus_Online}).Error; err != nil {
		t.Fatalf("create user view: %v", err)
	}
	firstEvent := newNotificationRequestEvent(t, userPublicId, "first")
	if err := repository.CreateFromRequest(context.Background(), firstEvent); err != nil {
		t.Fatalf("persist first notification: %v", err)
	}

	deletedCount, err := repository.DeleteForUser(context.Background(), userPublicId)
	if err != nil {
		t.Fatalf("delete notifications for user: %v", err)
	}
	if deletedCount != 1 {
		t.Fatalf("deleted notification count = %d, want 1", deletedCount)
	}
	if err := db.Delete(&platformschemas.UserView{}, "public_id = ?", userPublicId).Error; err != nil {
		t.Fatalf("delete user view: %v", err)
	}

	if err := repository.CreateFromRequest(context.Background(), newNotificationRequestEvent(t, userPublicId, "delayed")); err != nil {
		t.Fatalf("process delayed notification request: %v", err)
	}

	notifications, err := repository.List(context.Background(), userPublicId, nil, nil, 100)
	if err != nil {
		t.Fatalf("list notifications after deletion: %v", err)
	}
	if len(notifications) != 0 {
		t.Fatalf("notifications after user deletion = %d, want 0", len(notifications))
	}
}

func TestNotificationRepositoryListUsesCompositeCursor(t *testing.T) {
	db, err := gorm.Open(sqlite.Open("file:notification-repository-cursor-test?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	if err := createNotificationRepositoryTestTables(db); err != nil {
		t.Fatalf("create test database tables: %v", err)
	}

	repository := NewNotificationRepository(db)
	userPublicId := uuid.New()
	createdAt := time.Now().UTC().Truncate(time.Microsecond)
	for index := 0; index < 3; index++ {
		if err := db.Create(&schemas.Notification{
			Id:                    uuid.New(),
			RecipientUserPublicId: userPublicId,
			Type:                  "news",
			Priority:              "normal",
			TemplateKey:           "news",
			TemplateVersion:       1,
			Payload:               datatypes.JSON(`{"title":"Title","summary":"Summary","body":"Body"}`),
			DedupeKey:             "cursor-test:" + uuid.NewString(),
			CreatedAt:             createdAt,
		}).Error; err != nil {
			t.Fatalf("create notification %d: %v", index, err)
		}
	}

	firstPage, err := repository.List(context.Background(), userPublicId, nil, nil, 1)
	if err != nil {
		t.Fatalf("load first notification page: %v", err)
	}
	if len(firstPage) != 1 {
		t.Fatalf("first page count = %d, want 1", len(firstPage))
	}

	secondPage, err := repository.List(
		context.Background(),
		userPublicId,
		&firstPage[0].CreatedAt,
		&firstPage[0].Id,
		10,
	)
	if err != nil {
		t.Fatalf("load second notification page: %v", err)
	}
	if len(secondPage) != 2 {
		t.Fatalf("second page count = %d, want 2", len(secondPage))
	}
	for _, notification := range secondPage {
		if notification.Id == firstPage[0].Id {
			t.Fatal("composite cursor returned the previous notification again")
		}
	}
}

func createNotificationRepositoryTestTables(db *gorm.DB) error {
	for _, statement := range []string{
		`CREATE TABLE NotificationTable (id BLOB PRIMARY KEY, recipient_user_public_id BLOB NOT NULL, type TEXT NOT NULL, priority TEXT NOT NULL, template_key TEXT NOT NULL, template_version INTEGER NOT NULL, payload BLOB NOT NULL, dedupe_key TEXT NOT NULL, created_at DATETIME NOT NULL, read_at DATETIME, deleted_at DATETIME, expires_at DATETIME)`,
		`CREATE UNIQUE INDEX notification_dedupe_key_index ON NotificationTable (dedupe_key)`,
		`CREATE TABLE InboxEventTable (event_id BLOB PRIMARY KEY, consumed_at DATETIME NOT NULL)`,
		`CREATE TABLE OutboxEventTable (id BLOB PRIMARY KEY, aggregate_type TEXT NOT NULL, aggregate_id BLOB NOT NULL, event_type TEXT NOT NULL, topic TEXT NOT NULL, kafka_key TEXT NOT NULL, payload BLOB NOT NULL, metadata BLOB NOT NULL, available_at DATETIME NOT NULL, published_at DATETIME, publish_count INTEGER NOT NULL DEFAULT 0, last_error TEXT, claimed_by TEXT, claimed_at DATETIME, created_at DATETIME NOT NULL)`,
		`CREATE TABLE UserView (id BLOB PRIMARY KEY, public_id BLOB NOT NULL UNIQUE, plan TEXT NOT NULL, status TEXT NOT NULL, created_at DATETIME NOT NULL)`,
	} {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}

	return nil
}

func newNotificationRequestEvent(
	t *testing.T,
	userPublicId uuid.UUID,
	dedupeSuffix string,
) cevent.EventEnvelope[coreevents.NotificationRequestedData] {
	t.Helper()

	payload, err := json.Marshal(cnotificationtypes.NewsPayload{
		Title:   "Release update",
		Summary: "A new release is available.",
		Body:    "Read the release notes.",
	})
	if err != nil {
		t.Fatalf("marshal notification payload: %v", err)
	}

	return cevent.EventEnvelope[coreevents.NotificationRequestedData]{
		SchemaVersion: cevent.Version,
		EventId:       uuid.New(),
		EventType:     coreevents.EventType_NotificationRequested,
		AggregateType: coreevents.AggregateType_Notification,
		AggregateId:   userPublicId,
		KafkaKey:      userPublicId.String(),
		OccurredAt:    time.Now().UTC(),
		CorrelationId: "notification-repository-test",
		Data: coreevents.NotificationRequestedData{
			RecipientUserPublicId: userPublicId,
			Type:                  coreevents.NotificationType_News,
			Priority:              coreevents.NotificationPriority_Normal,
			TemplateKey:           cnotificationtypes.TemplateKey_News,
			TemplateVersion:       1,
			Payload:               payload,
			DedupeKey:             "repository-test:" + dedupeSuffix,
		},
	}
}
