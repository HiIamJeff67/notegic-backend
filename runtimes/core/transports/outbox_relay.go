package transports

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	smetrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"
	straces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"

	coreconfig "github.com/HiIamJeff67/notegic-backend/runtimes/core/configs"
)

type OutboxRelay struct {
	db                    *gorm.DB
	outboxEventRepository srepositories.OutboxEventRepositoryInterface
	producer              *skafka.Producer
	config                coreconfig.OutboxRelayConfig
	workerId              string
}

func NewOutboxRelay(
	db *gorm.DB,
	outboxEventRepository srepositories.OutboxEventRepositoryInterface,
	producer *skafka.Producer,
	config coreconfig.OutboxRelayConfig,
) *OutboxRelay {
	hostname, err := os.Hostname()
	if err != nil || hostname == "" {
		hostname = "core"
	}

	return &OutboxRelay{
		db:                    db,
		outboxEventRepository: outboxEventRepository,
		producer:              producer,
		config:                config,
		workerId:              fmt.Sprintf("%s-%d", hostname, os.Getpid()),
	}
}

func (r *OutboxRelay) Start(ctx context.Context) func() {
	workerCtx, cancel := context.WithCancel(ctx)
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		r.run(workerCtx)
	}()

	return func() {
		cancel()
		waitGroup.Wait()
	}
}

func (r *OutboxRelay) run(ctx context.Context) {
	relayTicker := time.NewTicker(r.config.PollInterval)
	cleanupTicker := time.NewTicker(r.config.CleanupInterval)
	defer relayTicker.Stop()
	defer cleanupTicker.Stop()

	r.relay(ctx)
	for {
		select {
		case <-ctx.Done():
			return
		case <-relayTicker.C:
			r.relay(ctx)
		case <-cleanupTicker.C:
			r.cleanup(ctx)
		}
	}
}

func (r *OutboxRelay) relay(ctx context.Context) {
	if straces.NotegicTracer != nil {
		relayCtx, span := straces.NotegicTracer.Start(ctx, "outbox.relay")
		defer straces.NotegicTracer.End(span, nil)
		ctx = relayCtx
	}

	events, exception := r.outboxEventRepository.ClaimAvailable(
		ctx,
		r.workerId,
		r.config.BatchSize,
		r.config.ClaimTimeout,
		srepositories.WithDB(r.db),
	)
	if exception != nil {
		if straces.NotegicTracer != nil {
			straces.NotegicTracer.RecordError(ctx, exception)
		}
		if slogs.NotegicLogger != nil {
			slogs.NotegicLogger.Error(ctx, exception, "Failed to claim outbox events")
		}
		return
	}
	if len(events) == 0 {
		return
	}
	if smetrics.NotegicMeter != nil {
		smetrics.NotegicMeter.Count(ctx, "outbox.relay.claimed.count", int64(len(events)))
	}

	publishedEventIds := make([]uuid.UUID, 0, len(events))
	failureInputs := make([]sinputs.FailedOutboxEventInput, 0)
	for _, event := range events {
		payload, err := srepositories.SerializeOutboxEvent(event)
		if err == nil && r.producer == nil {
			err = errors.New("Kafka producer is unavailable")
		}
		if err == nil {
			err = r.producer.Produce(ctx, event.Topic.String(), event.KafkaKey, payload)
		}
		if err == nil {
			publishedEventIds = append(publishedEventIds, event.Id)
			continue
		}

		backoff := r.config.InitialBackoff
		for attempt := int32(0); attempt < event.PublishCount && backoff < r.config.MaximumBackoff; attempt++ {
			backoff *= 2
		}
		if backoff > r.config.MaximumBackoff {
			backoff = r.config.MaximumBackoff
		}
		failureInputs = append(failureInputs, sinputs.FailedOutboxEventInput{
			Id:          event.Id,
			LastError:   err.Error(),
			AvailableAt: time.Now().Add(backoff),
		})
		if smetrics.NotegicMeter != nil {
			smetrics.NotegicMeter.Count(ctx, "outbox.relay.failure.count", 1)
		}
		if straces.NotegicTracer != nil {
			straces.NotegicTracer.RecordError(ctx, err)
		}
		if slogs.NotegicLogger != nil {
			slogs.NotegicLogger.Error(ctx, err, "Failed to publish outbox event")
		}
	}

	if exception := r.outboxEventRepository.MarkPublishedMany(
		ctx,
		publishedEventIds,
		r.workerId,
		srepositories.WithDB(r.db),
	); exception != nil {
		if straces.NotegicTracer != nil {
			straces.NotegicTracer.RecordError(ctx, exception)
		}
		if slogs.NotegicLogger != nil {
			slogs.NotegicLogger.Error(ctx, exception, "Failed to mark outbox events as published")
		}
	} else if smetrics.NotegicMeter != nil && len(publishedEventIds) > 0 {
		smetrics.NotegicMeter.Count(ctx, "outbox.relay.published.count", int64(len(publishedEventIds)))
	}

	if exception := r.outboxEventRepository.MarkFailedMany(
		ctx,
		failureInputs,
		r.workerId,
		srepositories.WithDB(r.db),
	); exception != nil {
		if straces.NotegicTracer != nil {
			straces.NotegicTracer.RecordError(ctx, exception)
		}
		if slogs.NotegicLogger != nil {
			slogs.NotegicLogger.Error(ctx, exception, "Failed to schedule outbox event retries")
		}
	} else if smetrics.NotegicMeter != nil && len(failureInputs) > 0 {
		smetrics.NotegicMeter.Count(ctx, "outbox.relay.retry.count", int64(len(failureInputs)))
	}
}

func (r *OutboxRelay) cleanup(ctx context.Context) {
	deletedCount, exception := r.outboxEventRepository.DeletePublishedBefore(
		ctx,
		time.Now().Add(-r.config.Retention),
		srepositories.WithDB(r.db),
	)
	if exception != nil {
		if straces.NotegicTracer != nil {
			straces.NotegicTracer.RecordError(ctx, exception)
		}
		if slogs.NotegicLogger != nil {
			slogs.NotegicLogger.Error(ctx, exception, "Failed to clean published outbox events")
		}
		return
	}
	if smetrics.NotegicMeter != nil && deletedCount > 0 {
		smetrics.NotegicMeter.Count(ctx, "outbox.cleanup.deleted.count", deletedCount)
	}
}
