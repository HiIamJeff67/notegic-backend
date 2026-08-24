package workers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	metrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"

	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
)

type YjsMaintenanceReconciliationWorkerInterface interface {
	Start(ctx context.Context) func()
	Reconcile(ctx context.Context) error
}

type YjsMaintenanceReconciliationWorker struct {
	db                    *gorm.DB
	outboxEventRepository repositories.OutboxEventRepositoryInterface
}

func NewYjsMaintenanceReconciliationWorker(
	db *gorm.DB,
	outboxEventRepository repositories.OutboxEventRepositoryInterface,
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

/* ============================== Auxiliary Functions ============================== */

func (w *YjsMaintenanceReconciliationWorker) reconcile(ctx context.Context) {
	if err := w.Reconcile(ctx); err != nil && ctx.Err() == nil {
		// Reconciliation is a safety net. The normal outbox path remains available
		// when a scan fails, so the next interval can retry without stopping Core.
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(ctx, err, "Yjs maintenance reconciliation failed")
		}
		return
	}
}

/* ============================== Worker Methods ============================== */

func (w *YjsMaintenanceReconciliationWorker) Start(ctx context.Context) func() {
	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		defer close(done)
		w.reconcile(workerCtx)

		ticker := time.NewTicker(yjsMaintenanceReconciliationInterval)
		defer ticker.Stop()
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				w.reconcile(workerCtx)
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

	var documents []schemas.BlockPackYjsDocument
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
		if metrics.NotegicMeter != nil {
			metrics.NotegicMeter.Value(ctx, "yjs.maintenance.reconciliation.stale_documents", 0)
		}
		return nil
	}
	if metrics.NotegicMeter != nil {
		metrics.NotegicMeter.Value(ctx, "yjs.maintenance.reconciliation.stale_documents", int64(len(documents)))
	}

	tx := w.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin Yjs maintenance reconciliation transaction: %w", tx.Error)
	}
	blockPackIds := make([]uuid.UUID, len(documents))
	for index, document := range documents {
		blockPackIds[index] = document.BlockPackId
	}
	if err := w.outboxEventRepository.EnqueueManyYjsMaintenanceHints(
		tx,
		uuid.NewString(),
		blockPackIds,
		"reconciliation",
	); err != nil {
		tx.Rollback()
		return fmt.Errorf("enqueue Yjs maintenance hints: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit Yjs maintenance reconciliation transaction: %w", err)
	}
	if metrics.NotegicMeter != nil {
		metrics.NotegicMeter.Count(ctx, "yjs.maintenance.reconciliation.hints_enqueued", int64(len(documents)))
	}

	return nil
}
