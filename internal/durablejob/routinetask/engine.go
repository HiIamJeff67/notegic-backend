package routinetask

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"

	constants "github.com/HiIamJeff67/notegic-backend/shared/constants"

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	cdurablejobroutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	durablejobconfig "github.com/HiIamJeff67/notegic-backend/internal/durablejob/configs"
)

type Engine struct {
	ticker         *time.Ticker
	workerId       uuid.UUID
	batchSize      int
	isHealthy      atomic.Bool
	isManagingWork atomic.Bool
	handlerManager HandlerManager
}

func NewEngine(
	_ durablejobconfig.Config,
	maxWorkers ...int,
) *Engine {
	initialMaxWorkers := constants.RoutineTaskEngineMaxWorkers
	if len(maxWorkers) > 0 {
		initialMaxWorkers = min(initialMaxWorkers, maxWorkers[0])
	}

	engine := &Engine{
		ticker:    time.NewTicker(constants.RoutineTaskEngineTickerDuration),
		workerId:  uuid.New(),
		batchSize: initialMaxWorkers,
	}
	engine.handlerManager = NewHandlerManager(initialMaxWorkers, engine.workerId)
	engine.isHealthy.Store(true)

	return engine
}

func (e *Engine) SetResultPublisher(publisher ResultPublisher) {
	e.handlerManager.SetResultPublisher(publisher)
}

func (e *Engine) SetRoutineTaskRunningPublisher(
	publisher RoutineTaskRunningPublisher,
) {
	e.handlerManager.SetRoutineTaskRunningPublisher(publisher)
}

func (e *Engine) NewClaimRoutineTasksRequest() (cdurablejob.ClaimRoutineTasksRequestDto, bool) {
	if e.isManagingWork.Load() {
		return cdurablejob.ClaimRoutineTasksRequestDto{}, false
	}

	return cdurablejob.ClaimRoutineTasksRequestDto{
		RequestId: uuid.New(),
		WorkerId:  e.workerId,
		BatchSize: e.batchSize,
	}, true
}

func (e *Engine) HandleRoutineTaskAssignments(
	ctx context.Context,
	assignments []cdurablejobroutinetasktypes.RoutineTaskAssignment,
) error {
	if len(assignments) == 0 {
		return nil
	}
	if !e.isManagingWork.CompareAndSwap(false, true) {
		return errors.New("DurableJob routine task worker is at capacity")
	}
	defer e.isManagingWork.Store(false)

	if err := e.handlerManager.Manage(ctx, assignments); err != nil {
		e.isHealthy.Store(false)
		return err
	}

	e.isHealthy.Store(true)
	return nil
}

func (e *Engine) Start(
	ctx context.Context,
	requestRoutineTasks func(context.Context, cdurablejob.ClaimRoutineTasksRequestDto) error,
) func() {
	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	var shutdownOnce sync.Once

	go func() {
		defer close(done)
		defer e.Stop()

		request, shouldRequest := e.NewClaimRoutineTasksRequest()
		if shouldRequest {
			if err := requestRoutineTasks(workerCtx, request); err != nil {
				e.isHealthy.Store(false)
				if logs.NotegicLogger != nil {
					logs.NotegicLogger.Error(workerCtx, err, "Failed to publish routine task claim request")
				}
			}
		}
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-e.ticker.C:
				request, shouldRequest := e.NewClaimRoutineTasksRequest()
				if !shouldRequest {
					continue
				}
				if err := requestRoutineTasks(workerCtx, request); err != nil {
					e.isHealthy.Store(false)
					if logs.NotegicLogger != nil {
						logs.NotegicLogger.Error(workerCtx, err, "Failed to publish routine task claim request")
					}
					continue
				}
				e.isHealthy.Store(true)
			}
		}
	}()

	return func() {
		shutdownOnce.Do(func() {
			cancel()
			<-done
		})
	}
}

func (e *Engine) Stop() {
	if e.ticker != nil {
		e.ticker.Stop()
	}
	e.isHealthy.Store(false)
}

func (e *Engine) IsReady() bool {
	return e.isHealthy.Load()
}
