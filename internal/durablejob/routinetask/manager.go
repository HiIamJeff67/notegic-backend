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

	handlers "github.com/HiIamJeff67/notegic-backend/internal/durablejob/routinetask/handlers"
	validation "github.com/HiIamJeff67/notegic-backend/internal/durablejob/validations"
)

type HandlerManager struct {
	maxWorkers       int
	activeWorkers    atomic.Int32
	workerPool       sync.WaitGroup
	sem              chan struct{}
	workerId         uuid.UUID
	failed           []failedRoutineTask
	failedMutex      sync.Mutex
	success          []preparedRoutineTask
	successMutex     sync.Mutex
	registries       map[cenums.RoutineTaskPurpose]handlers.PurposeHandler
	resultPublisher  ResultPublisher
	runningPublisher RoutineTaskRunningPublisher
}

type RoutineTaskResultKind string

const (
	RoutineTaskResultKind_Completed RoutineTaskResultKind = "completed"
	RoutineTaskResultKind_Failed    RoutineTaskResultKind = "failed"
)

type RoutineTaskResult struct {
	Kind          RoutineTaskResultKind
	WorkerId      uuid.UUID
	CorrelationId string
	Data          any
}

type ResultPublisher func(context.Context, RoutineTaskResult) error

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

func NewHandlerManager(
	maxWorkers int,
	workerIds ...uuid.UUID,
) HandlerManager {
	if maxWorkers <= 0 {
		maxWorkers = 1
	}

	workerId := uuid.New()
	if len(workerIds) > 0 && workerIds[0] != uuid.Nil {
		workerId = workerIds[0]
	}

	validator := validation.New()
	prepareHandler := handlers.NewPurposeHandler(validator)

	registries := make(map[cenums.RoutineTaskPurpose]handlers.PurposeHandler, 14)
	for _, purpose := range []cenums.RoutineTaskPurpose{
		cenums.RoutineTaskPurpose_CreateRootShelf,
		cenums.RoutineTaskPurpose_UpdateRootShelf,
		cenums.RoutineTaskPurpose_ResetRootShelf,
		cenums.RoutineTaskPurpose_CreateSubShelf,
		cenums.RoutineTaskPurpose_UpdateSubShelf,
		cenums.RoutineTaskPurpose_ResetSubShelf,
		cenums.RoutineTaskPurpose_CreateBlockPack,
		cenums.RoutineTaskPurpose_UpdateBlockPack,
		cenums.RoutineTaskPurpose_ResetBlockPack,
		cenums.RoutineTaskPurpose_AppendBlock,
		cenums.RoutineTaskPurpose_UpdateBlock,
		cenums.RoutineTaskPurpose_ResetBlock,
		cenums.RoutineTaskPurpose_CreateRoutine,
		cenums.RoutineTaskPurpose_UpdateRoutine,
	} {
		registries[purpose] = prepareHandler
	}

	return HandlerManager{
		maxWorkers: maxWorkers,
		sem:        make(chan struct{}, maxWorkers),
		workerId:   workerId,
		registries: registries,
	}
}

func (hm *HandlerManager) SetResultPublisher(publisher ResultPublisher) {
	hm.resultPublisher = publisher
}

func (hm *HandlerManager) SetRoutineTaskRunningPublisher(
	publisher RoutineTaskRunningPublisher,
) {
	hm.runningPublisher = publisher
}

func (hm *HandlerManager) Manage(
	ctx context.Context,
	assignments []cdurablejobroutinetasktypes.RoutineTaskAssignment,
) error {
	if len(assignments) == 0 {
		return nil
	}

	hm.resetResults(len(assignments))
	for _, assignment := range assignments {
		registry, exists := hm.registries[assignment.Purpose]
		if !exists || registry.HandlerFunc == nil {
			hm.appendFailure(failedRoutineTask{
				assignment:  assignment,
				failedAt:    time.Now().UTC(),
				errorCode:   cenums.RoutineTaskRecordErrorCode_HandlerFailed,
				errorReason: "routine task purpose handler was not found",
			})
			continue
		}

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

			preparedTask, err := registry.HandlerFunc(ctx, assignment)
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

func (hm *HandlerManager) resetResults(capacity int) {
	hm.failedMutex.Lock()
	hm.failed = make([]failedRoutineTask, 0, capacity)
	hm.failedMutex.Unlock()

	hm.successMutex.Lock()
	hm.success = make([]preparedRoutineTask, 0, capacity)
	hm.successMutex.Unlock()
}

func (hm *HandlerManager) appendSuccess(result preparedRoutineTask) {
	hm.successMutex.Lock()
	hm.success = append(hm.success, result)
	hm.successMutex.Unlock()
}

func (hm *HandlerManager) appendFailure(result failedRoutineTask) {
	hm.failedMutex.Lock()
	hm.failed = append(hm.failed, result)
	hm.failedMutex.Unlock()
}

func (hm *HandlerManager) publishResults(ctx context.Context) error {
	hm.successMutex.Lock()
	successes := append([]preparedRoutineTask(nil), hm.success...)
	hm.successMutex.Unlock()

	hm.failedMutex.Lock()
	failures := append([]failedRoutineTask(nil), hm.failed...)
	hm.failedMutex.Unlock()

	if len(successes)+len(failures) == 0 {
		return nil
	}
	if hm.resultPublisher == nil {
		return errors.New("DurableJob routine task result publisher is not configured")
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
				CompletedAt:         result.completedAt,
				PreparedTask:        &result.preparedTask,
			}
		}
		if err := hm.resultPublisher(ctx, RoutineTaskResult{
			Kind:          RoutineTaskResultKind_Completed,
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
				FailedAt:            failure.failedAt,
				ErrorCode:           failure.errorCode,
				ErrorReason:         failure.errorReason,
			}
		}
		if err := hm.resultPublisher(ctx, RoutineTaskResult{
			Kind:          RoutineTaskResultKind_Failed,
			WorkerId:      hm.workerId,
			CorrelationId: correlationId,
			Data:          request,
		}); err != nil {
			return err
		}
	}

	return nil
}
