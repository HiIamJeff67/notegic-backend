package workers

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	coreconfig "github.com/HiIamJeff67/notegic-backend/internal/core/configs"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
)

type QuotaCycleWorkerInterface interface {
	Start(ctx context.Context) func()
	Reconcile(ctx context.Context) error
}

type QuotaCycleWorker struct {
	db                  *gorm.DB
	config              coreconfig.QuotaCycleWorkerConfig
	userQuotaRepository repositories.UserQuotaRepositoryInterface
}

func NewQuotaCycleWorker(
	db *gorm.DB,
	config coreconfig.QuotaCycleWorkerConfig,
	userQuotaRepository repositories.UserQuotaRepositoryInterface,
) QuotaCycleWorkerInterface {
	return &QuotaCycleWorker{
		db:                  db,
		config:              config,
		userQuotaRepository: userQuotaRepository,
	}
}

func (w *QuotaCycleWorker) Start(ctx context.Context) func() {
	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})

	go func() {
		defer close(done)
		if err := w.Reconcile(workerCtx); err != nil && workerCtx.Err() == nil && logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(workerCtx, err, "User quota cycle reconciliation failed")
		}

		ticker := time.NewTicker(w.config.Interval)
		defer ticker.Stop()
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
				if err := w.Reconcile(workerCtx); err != nil && workerCtx.Err() == nil && logs.NotegicLogger != nil {
					logs.NotegicLogger.Error(workerCtx, err, "User quota cycle reconciliation failed")
				}
			}
		}
	}()

	return func() {
		cancel()
		<-done
	}
}

func (w *QuotaCycleWorker) Reconcile(ctx context.Context) error {
	if w == nil || w.db == nil || w.userQuotaRepository == nil || w.config.Interval <= 0 {
		return errors.New("user quota cycle reconciliation dependencies are required")
	}

	now := time.Now().UTC()
	tx := w.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin user quota cycle reconciliation transaction: %w", tx.Error)
	}

	if exception := w.userQuotaRepository.InitializeMissing(
		ctx,
		now,
		options.WithTransactionDB(tx),
	); exception != nil {
		tx.Rollback()
		return fmt.Errorf("initialize missing user quotas: %w", exception)
	}

	if _, exception := w.userQuotaRepository.ResetDue(
		ctx,
		now,
		options.WithTransactionDB(tx),
	); exception != nil {
		tx.Rollback()
		return fmt.Errorf("reset due user quotas: %w", exception)
	}

	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit user quota cycle reconciliation transaction: %w", err)
	}

	return nil
}
