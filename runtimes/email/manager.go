package email

import (
	"container/heap"
	"context"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	emailexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/email/exceptions"
	emailtypes "github.com/HiIamJeff67/notegic-backend/runtimes/email/types"
)

type EmailWorkerManager struct {
	maxWorkers    int
	activeWorkers int32
	workerPool    sync.WaitGroup
	ctx           context.Context
	cancel        context.CancelFunc
	emailSender   EmailSenderInterface

	buffer      *emailtypes.EmailBuffer
	bufferMutex sync.RWMutex

	monitorTicker *time.Ticker
	isMonitoring  int32
}

func NewEmailWorkerManager(maxWorkers int, sender EmailSenderInterface) *EmailWorkerManager {
	ctx, cancel := context.WithCancel(context.Background())
	return &EmailWorkerManager{
		maxWorkers:  maxWorkers,
		ctx:         ctx,
		cancel:      cancel,
		emailSender: sender,
		buffer:      emailtypes.NewEmailBuffer(),
	}
}

func (ewm *EmailWorkerManager) enqueueResponse(
	emailObject emailtypes.EmailObject,
	emailTaskType emailtypes.EmailTaskType,
	maxRetries int,
	priority int,
) error {
	task := &emailtypes.EmailTask{
		ID:         ewm.generateTaskID(),
		Type:       emailTaskType,
		Object:     emailObject,
		CreatedAt:  time.Now(),
		MaxRetries: maxRetries,
		Priority:   priority,
	}
	if err := ewm.enqueueTask(task); err != nil {
		return emailexceptions.
			NewDeliveryException("Email").
			EnqueueFailed(err)
	}
	return nil
}

/* ============================== Auxiliary Functions ============================== */

func (ewm *EmailWorkerManager) generateTaskID() string {
	return fmt.Sprintf("email_task_%s", uuid.New().String())
}

/* ============================== Private Methods ============================== */

func (ewm *EmailWorkerManager) processTask(task *emailtypes.EmailTask, workerID int) {
	err := ewm.emailSender.Send(context.Background(), task.Object)
	if err != nil {
		if slogs.NotegicLogger != nil {
			slogs.NotegicLogger.Error(
				context.Background(),
				err,
				fmt.Sprintf(
					"Worker %d failed to send email (attempt %d/%d)",
					workerID,
					task.Retries+1,
					task.MaxRetries,
				),
			)
		}

		task.Retries++
		if task.Retries < task.MaxRetries {
			task.Priority = max(0, task.Priority-1)
			go func() {
				retryDelay := time.Duration(task.Retries) * 30 * time.Second
				time.Sleep(retryDelay)
				_ = ewm.enqueueTask(task)
			}()
		}
		return
	}

	if slogs.NotegicLogger != nil {
		slogs.NotegicLogger.Debug(
			context.Background(),
			fmt.Sprintf("Worker %d successfully sent email task Id is %s", workerID, task.ID),
		)
	}
}

func (ewm *EmailWorkerManager) createWorker(task *emailtypes.EmailTask) {
	current := atomic.AddInt32(&ewm.activeWorkers, 1)
	workerID := int(current)
	ewm.workerPool.Add(1)

	go func() {
		defer func() {
			atomic.AddInt32(&ewm.activeWorkers, -1)
			ewm.workerPool.Done()
		}()
		ewm.processTask(task, workerID)
	}()
}

func (ewm *EmailWorkerManager) dispatchTasks() {
	ewm.bufferMutex.Lock()
	defer ewm.bufferMutex.Unlock()

	activeWorkers := ewm.GetActiveWorkerCount()
	workersNeeded := min(ewm.buffer.Len(), ewm.maxWorkers-activeWorkers)
	for index := 0; index < workersNeeded; index++ {
		if ewm.buffer.IsEmpty() {
			return
		}
		ewm.createWorker(heap.Pop(ewm.buffer).(*emailtypes.EmailTask))
	}
}

func (ewm *EmailWorkerManager) tryStartMonitoring() {
	if !atomic.CompareAndSwapInt32(&ewm.isMonitoring, 0, 1) {
		return
	}

	ewm.monitorTicker = time.NewTicker(sconstants.EmailWorkerManagerTickerDuration)
	go func() {
		defer atomic.StoreInt32(&ewm.isMonitoring, 0)
		for {
			select {
			case <-ewm.ctx.Done():
				return
			case <-ewm.monitorTicker.C:
				ewm.dispatchTasks()
				ewm.bufferMutex.RLock()
				isIdle := ewm.buffer.IsEmpty() && ewm.GetActiveWorkerCount() == 0
				ewm.bufferMutex.RUnlock()
				if isIdle {
					ewm.monitorTicker.Stop()
					return
				}
			}
		}
	}()
}

func (ewm *EmailWorkerManager) enqueueTask(task *emailtypes.EmailTask) error {
	ewm.bufferMutex.Lock()
	ewm.buffer.EnqueueTask(task)
	bufferSize := ewm.buffer.Len()
	ewm.bufferMutex.Unlock()

	if slogs.NotegicLogger != nil {
		slogs.NotegicLogger.Debug(
			context.Background(),
			fmt.Sprintf("Enqueued email task: ID=%s, Type=%s, Priority=%d, Queue size: %d", task.ID, task.Type, task.Priority, bufferSize),
		)
	}
	ewm.tryStartMonitoring()
	return nil
}

/* ============================== Public Methods ============================== */

func (ewm *EmailWorkerManager) GetActiveWorkerCount() int {
	return int(atomic.LoadInt32(&ewm.activeWorkers))
}

func (ewm *EmailWorkerManager) Shutdown() {
	ewm.cancel()
	if ewm.monitorTicker != nil {
		ewm.monitorTicker.Stop()
	}
	ewm.workerPool.Wait()
}

func (ewm *EmailWorkerManager) GetStatus() map[string]interface{} {
	ewm.bufferMutex.RLock()
	bufferSize := ewm.buffer.Len()
	ewm.bufferMutex.RUnlock()

	return map[string]interface{}{
		"bufferSize":    bufferSize,
		"activeWorkers": ewm.GetActiveWorkerCount(),
		"maxWorkers":    ewm.maxWorkers,
		"isMonitoring":  atomic.LoadInt32(&ewm.isMonitoring) == 1,
	}
}

func (ewm *EmailWorkerManager) EnqueueWelcomeEmail(
	response *cemailevents.SendWelcomeEmailResponseDto,
) error {
	return ewm.enqueueResponse(
		emailtypes.EmailObject{
			To:               response.To,
			Subject:          response.Subject,
			Body:             response.Body,
			EmailContentType: response.EmailContentType,
		},
		emailtypes.EmailTaskType_Welcome,
		response.MaxRetries,
		response.Priority,
	)
}

func (ewm *EmailWorkerManager) EnqueueValidationEmail(
	response *cemailevents.SendValidationEmailResponseDto,
) error {
	return ewm.enqueueResponse(
		emailtypes.EmailObject{
			To:               response.To,
			Subject:          response.Subject,
			Body:             response.Body,
			EmailContentType: response.EmailContentType,
		},
		emailtypes.EmailTaskType_Validation,
		response.MaxRetries,
		response.Priority,
	)
}

func (ewm *EmailWorkerManager) EnqueueSecurityAlertEmail(
	response *cemailevents.SendSecurityAlertEmailResponseDto,
) error {
	return ewm.enqueueResponse(
		emailtypes.EmailObject{
			To:               response.To,
			Subject:          response.Subject,
			Body:             response.Body,
			EmailContentType: response.EmailContentType,
		},
		emailtypes.EmailTaskType_Security,
		response.MaxRetries,
		response.Priority,
	)
}
