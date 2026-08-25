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

	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	metrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"
	traces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"
	cinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/inputs"
	crepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"

	coreconfig "github.com/HiIamJeff67/notegic-backend/internal/core/configs"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
)

type OutboxRelay struct {
	db                    *gorm.DB
	outboxEventRepository repositories.OutboxEventRepositoryInterface
	producer              *platformkafka.Producer
	config                coreconfig.OutboxRelayConfig
	workerId              string
}

func NewOutboxRelay(
	db *gorm.DB,
	outboxEventRepository repositories.OutboxEventRepositoryInterface,
	producer *platformkafka.Producer,
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
	if traces.NotegicTracer != nil {
		relayCtx, span := traces.NotegicTracer.Start(ctx, "outbox.relay")
		defer traces.NotegicTracer.End(span, nil)
		ctx = relayCtx
	}

	events, exception := r.outboxEventRepository.ClaimAvailable(
		ctx,
		r.workerId,
		r.config.BatchSize,
		r.config.ClaimTimeout,
		options.WithDB(r.db),
	)
	if exception != nil {
		if traces.NotegicTracer != nil {
			traces.NotegicTracer.RecordError(ctx, exception)
		}
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(ctx, exception, "Failed to claim outbox events")
		}
		return
	}
	if len(events) == 0 {
		return
	}
	if metrics.NotegicMeter != nil {
		metrics.NotegicMeter.Count(ctx, "outbox.relay.claimed.count", int64(len(events)))
	}

	publishedEventIds := make([]uuid.UUID, 0, len(events))
	failureInputs := make([]cinputs.FailedOutboxEventInput, 0)
	for _, event := range events {
		payload, err := crepositories.SerializeOutboxEvent(event)
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
		failureInputs = append(failureInputs, cinputs.FailedOutboxEventInput{
			Id:          event.Id,
			LastError:   err.Error(),
			AvailableAt: time.Now().Add(backoff),
		})
		if metrics.NotegicMeter != nil {
			metrics.NotegicMeter.Count(ctx, "outbox.relay.failure.count", 1)
		}
		if traces.NotegicTracer != nil {
			traces.NotegicTracer.RecordError(ctx, err)
		}
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(ctx, err, "Failed to publish outbox event")
		}
	}

	if exception := r.outboxEventRepository.MarkPublishedMany(
		ctx,
		publishedEventIds,
		r.workerId,
		options.WithDB(r.db),
	); exception != nil {
		if traces.NotegicTracer != nil {
			traces.NotegicTracer.RecordError(ctx, exception)
		}
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(ctx, exception, "Failed to mark outbox events as published")
		}
	} else if metrics.NotegicMeter != nil && len(publishedEventIds) > 0 {
		metrics.NotegicMeter.Count(ctx, "outbox.relay.published.count", int64(len(publishedEventIds)))
	}

	if exception := r.outboxEventRepository.MarkFailedMany(
		ctx,
		failureInputs,
		r.workerId,
		options.WithDB(r.db),
	); exception != nil {
		if traces.NotegicTracer != nil {
			traces.NotegicTracer.RecordError(ctx, exception)
		}
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(ctx, exception, "Failed to schedule outbox event retries")
		}
	} else if metrics.NotegicMeter != nil && len(failureInputs) > 0 {
		metrics.NotegicMeter.Count(ctx, "outbox.relay.retry.count", int64(len(failureInputs)))
	}
}

func (r *OutboxRelay) cleanup(ctx context.Context) {
	deletedCount, exception := r.outboxEventRepository.DeletePublishedBefore(
		ctx,
		time.Now().Add(-r.config.Retention),
		options.WithDB(r.db),
	)
	if exception != nil {
		if traces.NotegicTracer != nil {
			traces.NotegicTracer.RecordError(ctx, exception)
		}
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(ctx, exception, "Failed to clean published outbox events")
		}
		return
	}
	if metrics.NotegicMeter != nil && deletedCount > 0 {
		metrics.NotegicMeter.Count(ctx, "outbox.cleanup.deleted.count", deletedCount)
	}
}
