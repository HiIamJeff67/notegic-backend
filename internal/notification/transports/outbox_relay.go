package transports

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
)

type OutboxRelay struct {
	repository      srepositories.NotificationRepository
	producer        *skafka.Producer
	pollInterval    time.Duration
	claimTimeout    time.Duration
	initialBackoff  time.Duration
	maximumBackoff  time.Duration
	batchSize       int
	cleanupInterval time.Duration
	retention       time.Duration
	workerId        string
}

func NewOutboxRelay(
	repository srepositories.NotificationRepository,
	producer *skafka.Producer,
	pollInterval time.Duration,
	claimTimeout time.Duration,
	initialBackoff time.Duration,
	maximumBackoff time.Duration,
	batchSize int,
	cleanupInterval time.Duration,
	retention time.Duration,
) *OutboxRelay {
	return &OutboxRelay{
		repository:      repository,
		producer:        producer,
		pollInterval:    pollInterval,
		claimTimeout:    claimTimeout,
		initialBackoff:  initialBackoff,
		maximumBackoff:  maximumBackoff,
		batchSize:       batchSize,
		cleanupInterval: cleanupInterval,
		retention:       retention,
		workerId:        "notification-outbox-relay",
	}
}

func (r *OutboxRelay) Start(ctx context.Context) func() {
	workerCtx, cancel := context.WithCancel(ctx)
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		runTicker := time.NewTicker(r.pollInterval)
		defer runTicker.Stop()
		lastCleanup := time.Now().UTC()
		for {
			r.relay(workerCtx)
			if time.Since(lastCleanup) >= r.cleanupInterval {
				if _, err := r.repository.DeletePublishedOutbox(workerCtx, time.Now().UTC().Add(-r.retention)); err != nil && slogs.NotegicLogger != nil {
					slogs.NotegicLogger.Error(workerCtx, err, "Failed to clean Notification outbox events")
				}
				lastCleanup = time.Now().UTC()
			}
			select {
			case <-workerCtx.Done():
				return
			case <-runTicker.C:
			}
		}
	}()

	return func() {
		cancel()
		waitGroup.Wait()
	}
}

func (r *OutboxRelay) relay(ctx context.Context) {
	events, err := r.repository.ClaimOutbox(ctx, r.workerId, r.batchSize, r.claimTimeout)
	if err != nil {
		if slogs.NotegicLogger != nil {
			slogs.NotegicLogger.Error(ctx, err, "Failed to claim Notification outbox events")
		}
		return
	}
	if len(events) == 0 {
		return
	}
	publishedIds := make([]uuid.UUID, 0, len(events))
	failedIds := make([]uuid.UUID, 0)
	for _, event := range events {
		if r.producer == nil {
			failedIds = append(failedIds, event.Id)
			continue
		}
		if err := r.producer.Produce(ctx, event.Topic.String(), event.KafkaKey, event.Payload); err != nil {
			failedIds = append(failedIds, event.Id)
			if slogs.NotegicLogger != nil {
				slogs.NotegicLogger.Error(ctx, err, "Failed to publish Notification outbox event")
			}
			continue
		}
		publishedIds = append(publishedIds, event.Id)
	}
	if err := r.repository.MarkOutboxPublished(ctx, r.workerId, publishedIds); err != nil && slogs.NotegicLogger != nil {
		slogs.NotegicLogger.Error(ctx, err, "Failed to mark Notification outbox events published")
	}
	if len(failedIds) > 0 {
		failedSet := make(map[uuid.UUID]struct{}, len(failedIds))
		for _, id := range failedIds {
			failedSet[id] = struct{}{}
		}
		maximumPublishCount := 0
		for _, event := range events {
			if _, failed := failedSet[event.Id]; !failed {
				continue
			}
			if int(event.PublishCount) > maximumPublishCount {
				maximumPublishCount = int(event.PublishCount)
			}
		}
		retryBackoff := r.initialBackoff
		for index := 0; index < maximumPublishCount && retryBackoff < r.maximumBackoff; index++ {
			retryBackoff *= 2
		}
		if retryBackoff > r.maximumBackoff {
			retryBackoff = r.maximumBackoff
		}
		if err := r.repository.MarkOutboxFailed(ctx, r.workerId, failedIds, "Kafka publish failed", time.Now().UTC().Add(retryBackoff)); err != nil && slogs.NotegicLogger != nil {
			slogs.NotegicLogger.Error(ctx, err, "Failed to schedule Notification outbox retries")
		}
	}
}
