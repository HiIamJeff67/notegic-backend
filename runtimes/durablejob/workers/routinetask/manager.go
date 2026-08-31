package routinetask

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	cdurablejobroutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	routinetaskservice "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask"
	validation "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/validations"
)

type Manager struct {
	maxWorkers         int
	activeWorkers      atomic.Int32
	workerPool         sync.WaitGroup
	sem                chan struct{}
	workerId           uuid.UUID
	failed             []failedRoutineTask
	failedMutex        sync.Mutex
	success            []preparedRoutineTask
	successMutex       sync.Mutex
	preparationService *routinetaskservice.RoutineTaskPreparationService
	resultWriter       ResultWriteFunc
	runningPublisher   RoutineTaskRunningPublisher
}

type RoutineTaskRunningPublisher func(
	context.Context,
	cdurablejobroutinetasktypes.RoutineTaskAssignment,
) error

type preparedRoutineTask struct {
	preparedTask cdurablejobroutinetasktypes.PreparedRoutineTask
	completedAt  time.Time
}

type failedRoutineTask struct {
	assignment  cdurablejobroutinetasktypes.RoutineTaskAssignment
	failedAt    time.Time
	errorCode   cenums.RoutineTaskRecordErrorCode
	errorReason string
}

func NewManager(
	maxWorkers int,
	workerIds ...uuid.UUID,
) Manager {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}

	workerId := uuid.New()
	if len(workerIds) > 0 && workerIds[0] != uuid.Nil {
		workerId = workerIds[0]
	}

	return Manager{
		maxWorkers:         maxWorkers,
		sem:                make(chan struct{}, maxWorkers),
		workerId:           workerId,
		preparationService: routinetaskservice.NewRoutineTaskPreparationService(validation.New()),
	}
}

func (hm *Manager) SetResultWriter(writer ResultWriteFunc) {
	hm.resultWriter = writer
}

func (hm *Manager) SetRoutineTaskRunningPublisher(
	publisher RoutineTaskRunningPublisher,
) {
	hm.runningPublisher = publisher
}

func (hm *Manager) Manage(
	ctx context.Context,
	assignments []cdurablejobroutinetasktypes.RoutineTaskAssignment,
) error {
	if len(assignments) == 0 {
		return nil
	}

	hm.resetResults(len(assignments))
	for _, assignment := range assignments {
		assignment := assignment
		hm.sem <- struct{}{}
		hm.workerPool.Add(1)
		hm.activeWorkers.Add(1)
		go func() {
			defer func() {
				<-hm.sem
				hm.activeWorkers.Add(-1)
				hm.workerPool.Done()
			}()

			if hm.runningPublisher != nil {
				if err := hm.runningPublisher(ctx, assignment); err != nil && slogs.NotegicLogger != nil {
					slogs.NotegicLogger.Error(
						ctx,
						err,
						"Failed to publish RoutineTask running lifecycle event",
					)
				}
			}

			preparedTask, err := hm.preparationService.Prepare(ctx, assignment)
			if err != nil || preparedTask == nil {
				errorCode := cenums.RoutineTaskRecordErrorCode_HandlerFailed
				errorReason := "routine task preparation failed"
				if err != nil {
					var durableJobError *cexceptions.Exception
					if errors.As(err, &durableJobError) {
						switch durableJobError.Reason {
						case "Canceled":
							errorCode = cenums.RoutineTaskRecordErrorCode_Canceled
						case "Timeout":
							errorCode = cenums.RoutineTaskRecordErrorCode_Timeout
						case "InvalidRoutineTaskPayload":
							errorCode = cenums.RoutineTaskRecordErrorCode_PayloadInvalid
						case "TargetNotFound":
							errorCode = cenums.RoutineTaskRecordErrorCode_TargetNotFound
						case "PermissionDenied":
							errorCode = cenums.RoutineTaskRecordErrorCode_PermissionDenied
						}
						if durableJobError.Reason != "" {
							errorReason = durableJobError.Reason
						}
					} else if errors.Is(err, context.Canceled) {
						errorCode = cenums.RoutineTaskRecordErrorCode_Canceled
					} else if errors.Is(err, context.DeadlineExceeded) {
						errorCode = cenums.RoutineTaskRecordErrorCode_Timeout
					} else {
						errorReason = err.Error()
					}
					if len(errorReason) > 256 {
						errorReason = errorReason[:256]
					}
				}
				hm.appendFailure(failedRoutineTask{
					assignment:  assignment,
					failedAt:    time.Now().UTC(),
					errorCode:   errorCode,
					errorReason: errorReason,
				})
				return
			}

			hm.appendSuccess(preparedRoutineTask{
				preparedTask: *preparedTask,
				completedAt:  time.Now().UTC(),
			})
		}()
	}

	hm.workerPool.Wait()
	return hm.publishResults(ctx)
}

func (hm *Manager) resetResults(capacity int) {
	hm.failedMutex.Lock()
	hm.failed = make([]failedRoutineTask, 0, capacity)
	hm.failedMutex.Unlock()

	hm.successMutex.Lock()
	hm.success = make([]preparedRoutineTask, 0, capacity)
	hm.successMutex.Unlock()
}

func (hm *Manager) appendSuccess(result preparedRoutineTask) {
	hm.successMutex.Lock()
	hm.success = append(hm.success, result)
	hm.successMutex.Unlock()
}

func (hm *Manager) appendFailure(result failedRoutineTask) {
	hm.failedMutex.Lock()
	hm.failed = append(hm.failed, result)
	hm.failedMutex.Unlock()
}

func (hm *Manager) publishResults(ctx context.Context) error {
	hm.successMutex.Lock()
	successes := append([]preparedRoutineTask(nil), hm.success...)
	hm.successMutex.Unlock()

	hm.failedMutex.Lock()
	failures := append([]failedRoutineTask(nil), hm.failed...)
	hm.failedMutex.Unlock()

	if len(successes)+len(failures) == 0 {
		return nil
	}
	if hm.resultWriter == nil {
		return errors.New("DurableJob routine task result writer is not configured")
	}

	correlationId := uuid.New().String()
	if len(successes) > 0 {
		request := cdurablejob.MarkCompletedRoutineTasksRequestDto{
			WorkerId: hm.workerId,
			Tasks:    make([]cdurablejobroutinetasktypes.CompletedRoutineTask, len(successes)),
		}
		for index, result := range successes {
			request.Tasks[index] = cdurablejobroutinetasktypes.CompletedRoutineTask{
				RoutineTaskId:       result.preparedTask.RoutineTaskId,
				RoutineTaskRecordId: result.preparedTask.RoutineTaskRecordId,
				RoutineRecordId:     result.preparedTask.RoutineRecordId,
				CompletedAt:         result.completedAt,
				PreparedTask:        &result.preparedTask,
			}
		}
		if err := hm.resultWriter(ctx, Result{
			Kind:          ResultKind_Completed,
			WorkerId:      hm.workerId,
			CorrelationId: correlationId,
			Data:          request,
		}); err != nil {
			return err
		}
	}

	if len(failures) > 0 {
		request := cdurablejob.MarkFailedRoutineTasksRequestDto{
			WorkerId: hm.workerId,
			Tasks:    make([]cdurablejobroutinetasktypes.FailedRoutineTask, len(failures)),
		}
		for index, failure := range failures {
			request.Tasks[index] = cdurablejobroutinetasktypes.FailedRoutineTask{
				RoutineTaskId:       failure.assignment.RoutineTaskId,
				RoutineTaskRecordId: failure.assignment.RoutineTaskRecordId,
				RoutineRecordId:     failure.assignment.RoutineRecordId,
				FailedAt:            failure.failedAt,
				ErrorCode:           failure.errorCode,
				ErrorReason:         failure.errorReason,
			}
		}
		if err := hm.resultWriter(ctx, Result{
			Kind:          ResultKind_Failed,
			WorkerId:      hm.workerId,
			CorrelationId: correlationId,
			Data:          request,
		}); err != nil {
			return err
		}
	}

	return nil
}
