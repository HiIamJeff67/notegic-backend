package routinetask

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	durablejobcontract "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	durablejobroutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

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
	registries       map[enums.RoutineTaskPurpose]handlers.PurposeHandler
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
	durablejobroutinetasktypes.RoutineTaskAssignment,
) error

type preparedRoutineTask struct {
	preparedTask durablejobroutinetasktypes.PreparedRoutineTask
	completedAt  time.Time
}

type failedRoutineTask struct {
	assignment  durablejobroutinetasktypes.RoutineTaskAssignment
	failedAt    time.Time
	errorCode   enums.RoutineTaskRecordErrorCode
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

	registries := make(map[enums.RoutineTaskPurpose]handlers.PurposeHandler, 14)
	for _, purpose := range []enums.RoutineTaskPurpose{
		enums.RoutineTaskPurpose_CreateRootShelf,
		enums.RoutineTaskPurpose_UpdateRootShelf,
		enums.RoutineTaskPurpose_ResetRootShelf,
		enums.RoutineTaskPurpose_CreateSubShelf,
		enums.RoutineTaskPurpose_UpdateSubShelf,
		enums.RoutineTaskPurpose_ResetSubShelf,
		enums.RoutineTaskPurpose_CreateBlockPack,
		enums.RoutineTaskPurpose_UpdateBlockPack,
		enums.RoutineTaskPurpose_ResetBlockPack,
		enums.RoutineTaskPurpose_AppendBlock,
		enums.RoutineTaskPurpose_UpdateBlock,
		enums.RoutineTaskPurpose_ResetBlock,
		enums.RoutineTaskPurpose_CreateRoutine,
		enums.RoutineTaskPurpose_UpdateRoutine,
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
	assignments []durablejobroutinetasktypes.RoutineTaskAssignment,
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
				errorCode:   enums.RoutineTaskRecordErrorCode_HandlerFailed,
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
				if err := hm.runningPublisher(ctx, assignment); err != nil && logs.NotegicLogger != nil {
					logs.NotegicLogger.Error(
						ctx,
						err,
						"Failed to publish RoutineTask running lifecycle event",
					)
				}
			}

			preparedTask, err := registry.HandlerFunc(ctx, assignment)
			if err != nil || preparedTask == nil {
				errorCode := enums.RoutineTaskRecordErrorCode_HandlerFailed
				errorReason := "routine task preparation failed"
				if err != nil {
					var durableJobError *exceptions.Exception
					if errors.As(err, &durableJobError) {
						switch durableJobError.Reason {
						case "Canceled":
							errorCode = enums.RoutineTaskRecordErrorCode_Canceled
						case "Timeout":
							errorCode = enums.RoutineTaskRecordErrorCode_Timeout
						case "InvalidRoutineTaskPayload":
							errorCode = enums.RoutineTaskRecordErrorCode_PayloadInvalid
						case "TargetNotFound":
							errorCode = enums.RoutineTaskRecordErrorCode_TargetNotFound
						case "PermissionDenied":
							errorCode = enums.RoutineTaskRecordErrorCode_PermissionDenied
						}
						if durableJobError.Reason != "" {
							errorReason = durableJobError.Reason
						}
					} else if errors.Is(err, context.Canceled) {
						errorCode = enums.RoutineTaskRecordErrorCode_Canceled
					} else if errors.Is(err, context.DeadlineExceeded) {
						errorCode = enums.RoutineTaskRecordErrorCode_Timeout
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
		request := durablejobcontract.MarkCompletedRoutineTasksRequestDto{
			WorkerId: hm.workerId,
			Tasks:    make([]durablejobroutinetasktypes.CompletedRoutineTask, len(successes)),
		}
		for index, result := range successes {
			request.Tasks[index] = durablejobroutinetasktypes.CompletedRoutineTask{
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
		request := durablejobcontract.MarkFailedRoutineTasksRequestDto{
			WorkerId: hm.workerId,
			Tasks:    make([]durablejobroutinetasktypes.FailedRoutineTask, len(failures)),
		}
		for index, failure := range failures {
			request.Tasks[index] = durablejobroutinetasktypes.FailedRoutineTask{
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
