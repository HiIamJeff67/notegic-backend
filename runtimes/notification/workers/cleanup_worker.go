package workers

import (
	"context"
	"sync"
	"time"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	services "github.com/HiIamJeff67/notegic-backend/runtimes/notification/services"
)

type CleanupWorker struct {
	service   services.NotificationServiceInterface
	interval  time.Duration
	retention time.Duration
}

func NewCleanupWorker(
	service services.NotificationServiceInterface,
	interval time.Duration,
	retention time.Duration,
) *CleanupWorker {
	return &CleanupWorker{
		service:   service,
		interval:  interval,
		retention: retention,
	}
}

func (w *CleanupWorker) Start(ctx context.Context) func() {
	workerCtx, cancel := context.WithCancel(ctx)
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		ticker := time.NewTicker(w.interval)
		defer ticker.Stop()
		for {
			if _, err := w.service.HardDeleteExpiredNotifications(workerCtx, time.Now().UTC(), w.retention); err != nil && slogs.NotegicLogger != nil {
				slogs.NotegicLogger.Error(workerCtx, err, "Failed to clean Notification records")
			}
			select {
			case <-workerCtx.Done():
				return
			case <-ticker.C:
			}
		}
	}()

	return func() {
		cancel()
		waitGroup.Wait()
	}
}
