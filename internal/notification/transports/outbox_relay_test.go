package transports

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	cmodels "github.com/HiIamJeff67/notegic-backend/contracts/types/models"

	repositories "github.com/HiIamJeff67/notegic-backend/internal/notification/data/postgres/repositories"
)

type notificationRepositoryStub struct {
	repositories.NotificationRepository
	events            []cmodels.OutboxEvent
	claimCalls        int
	failedEventIds    []uuid.UUID
	failedAvailableAt time.Time
	failedMessage     string
	publishedEventIds []uuid.UUID
}

func (r *notificationRepositoryStub) ClaimOutbox(
	context.Context,
	string,
	int,
	time.Duration,
) ([]cmodels.OutboxEvent, error) {
	r.claimCalls++
	return r.events, nil
}

func (r *notificationRepositoryStub) MarkOutboxPublished(
	_ context.Context,
	_ string,
	eventIds []uuid.UUID,
) error {
	r.publishedEventIds = append([]uuid.UUID(nil), eventIds...)
	return nil
}

func (r *notificationRepositoryStub) MarkOutboxFailed(
	_ context.Context,
	_ string,
	eventIds []uuid.UUID,
	message string,
	availableAt time.Time,
) error {
	r.failedEventIds = append([]uuid.UUID(nil), eventIds...)
	r.failedMessage = message
	r.failedAvailableAt = availableAt
	return nil
}

func TestOutboxRelaySchedulesRetryWhenProducerIsUnavailable(t *testing.T) {
	eventId := uuid.New()
	startedAt := time.Now().UTC()
	repository := &notificationRepositoryStub{
		events: []cmodels.OutboxEvent{{
			Id:           eventId,
			PublishCount: 1,
		}},
	}
	relay := NewOutboxRelay(
		repository,
		nil,
		time.Second,
		time.Second,
		time.Second,
		30*time.Second,
		10,
		time.Hour,
		24*time.Hour,
	)

	relay.relay(context.Background())

	if repository.claimCalls != 1 {
		t.Fatalf("claim calls = %d, want 1", repository.claimCalls)
	}
	if len(repository.publishedEventIds) != 0 {
		t.Fatalf("published event IDs = %v, want none", repository.publishedEventIds)
	}
	if len(repository.failedEventIds) != 1 || repository.failedEventIds[0] != eventId {
		t.Fatalf("failed event IDs = %v, want [%s]", repository.failedEventIds, eventId)
	}
	if repository.failedMessage != "Kafka publish failed" {
		t.Fatalf("failure message = %q, want Kafka publish failed", repository.failedMessage)
	}
	if !repository.failedAvailableAt.After(startedAt) {
		t.Fatalf("retry available at %s, want after %s", repository.failedAvailableAt, startedAt)
	}
}
