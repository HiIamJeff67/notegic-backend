package workers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	smetrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"
	general "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/general"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type YjsMaintenanceReconciliationWorkerInterface interface {
	Start(ctx context.Context) func()
	Reconcile(ctx context.Context) error
}

type YjsMaintenanceReconciliationWorker struct {
	db                    *gorm.DB
	outboxEventRepository general.OutboxEventRepositoryInterface[coreevents.YjsMaintenanceHintData]
}

func NewYjsMaintenanceReconciliationWorker(
	db *gorm.DB,
	outboxEventRepository general.OutboxEventRepositoryInterface[coreevents.YjsMaintenanceHintData],
) YjsMaintenanceReconciliationWorkerInterface {
	return &YjsMaintenanceReconciliationWorker{
		db:                    db,
		outboxEventRepository: outboxEventRepository,
	}
}

/* ============================== Constants ============================== */

const (
	yjsMaintenanceReconciliationBatchSize = 256
	yjsMaintenanceReconciliationInterval  = time.Hour
)

/* ============================== Worker Methods ============================== */

func (w *YjsMaintenanceReconciliationWorker) Start(ctx context.Context) func() {
	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		defer close(done)
		if err := w.Reconcile(workerCtx); err != nil && workerCtx.Err() == nil && slogs.NotegicLogger != nil {
			slogs.NotegicLogger.Error(workerCtx, err, "Yjs maintenance reconciliation failed")
		}

		ticker := time.NewTicker(yjsMaintenanceReconciliationInterval)
		defer ticker.Stop()
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				if err := w.Reconcile(workerCtx); err != nil && workerCtx.Err() == nil && slogs.NotegicLogger != nil {
					slogs.NotegicLogger.Error(workerCtx, err, "Yjs maintenance reconciliation failed")
				}
			}
		}
	}()

	return func() {
		cancel()
		<-done
	}
}

func (w *YjsMaintenanceReconciliationWorker) Reconcile(ctx context.Context) error {
	if w == nil || w.db == nil || w.outboxEventRepository == nil {
		return errors.New("Yjs maintenance reconciliation dependencies are required")
	}

	var documents []sschemas.BlockPackYjsDocument
	result := w.db.WithContext(ctx).
		Select("id, block_pack_id, last_update_sequence, compacted_until_sequence, projected_until_sequence, last_compacted_at, snapshot, state_vector").
		Where("deleted_at IS NULL").
		Where("last_update_sequence > 0").
		Where("last_update_sequence > compacted_until_sequence OR last_update_sequence > projected_until_sequence").
		Order("updated_at ASC").
		Limit(yjsMaintenanceReconciliationBatchSize).
		Find(&documents)
	if result.Error != nil {
		return fmt.Errorf("load stale Yjs documents: %w", result.Error)
	}
	if len(documents) == 0 {
		if smetrics.NotegicMeter != nil {
			smetrics.NotegicMeter.Value(ctx, "yjs.maintenance.reconciliation.stale_documents", 0)
		}
		return nil
	}
	if smetrics.NotegicMeter != nil {
		smetrics.NotegicMeter.Value(ctx, "yjs.maintenance.reconciliation.stale_documents", int64(len(documents)))
	}

	tx := w.db.WithContext(ctx).Begin()
	events := make([]cevent.EventEnvelope[coreevents.YjsMaintenanceHintData], 0, len(documents))
	correlationId := uuid.NewString()
	for _, document := range documents {
		events = append(events, cevent.EventEnvelope[coreevents.YjsMaintenanceHintData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_YjsMaintenanceHint,
			AggregateType: coreevents.AggregateType_BlockPack,
			AggregateId:   document.BlockPackId,
			KafkaKey:      document.BlockPackId.String(),
			OccurredAt:    time.Now().UTC(),
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
				Reason:                 "reconciliation",
			},
		})
	}
	if err := w.outboxEventRepository.EnqueueOutboxEvents(
		tx,
		coreevents.CoreYjsMaintenanceHintTopic,
		events,
	); err != nil {
		tx.Rollback()
		return fmt.Errorf("enqueue Yjs maintenance hints: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit Yjs maintenance reconciliation transaction: %w", err)
	}
	if smetrics.NotegicMeter != nil {
		smetrics.NotegicMeter.Count(ctx, "yjs.maintenance.reconciliation.hints_enqueued", int64(len(documents)))
	}

	return nil
}
