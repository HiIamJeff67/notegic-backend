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

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	durablejobconfig "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/configs"
	routinetaskservice "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask"
)

type Engine struct {
	ticker             *time.Ticker
	workerId           uuid.UUID
	batchSize          int
	isHealthy          atomic.Bool
	isManagingWork     atomic.Bool
	claimer            *Claimer
	recoveryService    routinetaskservice.RoutineTaskRecoveryServiceInterface
	routineTaskManager Manager
}

func NewEngine(
	_ durablejobconfig.Config,
	claimer *Claimer,
	maxWorkers ...int,
) *Engine {
	initialMaxWorkers := sconstants.RoutineTaskEngineMaxWorkers
	if len(maxWorkers) > 0 {
		initialMaxWorkers = min(initialMaxWorkers, maxWorkers[0])
	}

	engine := &Engine{
		ticker:    time.NewTicker(sconstants.RoutineTaskEngineTickerDuration),
		workerId:  uuid.New(),
		batchSize: initialMaxWorkers,
		claimer:   claimer,
	}
	engine.routineTaskManager = NewManager(initialMaxWorkers, engine.workerId)
	engine.isHealthy.Store(true)

	return engine
}

func (e *Engine) SetRoutineTaskRecoveryService(
	service routinetaskservice.RoutineTaskRecoveryServiceInterface,
) {
	e.recoveryService = service
}

func (e *Engine) SetResultWriter(writer ResultWriteFunc) {
	e.routineTaskManager.SetResultWriter(writer)
}

func (e *Engine) SetRoutineTaskRunningPublisher(
	publisher RoutineTaskRunningPublisher,
) {
	e.routineTaskManager.SetRoutineTaskRunningPublisher(publisher)
}

func (e *Engine) GetClaimRoutineTasksRequest() (cdurablejob.ClaimRoutineTasksRequestDto, bool) {
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

	if err := e.routineTaskManager.Manage(ctx, assignments); err != nil {
		e.isHealthy.Store(false)
		return err
	}

	e.isHealthy.Store(true)
	return nil
}

func (e *Engine) Start(
	ctx context.Context,
) func() {
	workerCtx, cancel := context.WithCancel(ctx)
	done := make(chan struct{})
	var shutdownOnce sync.Once

	go func() {
		defer close(done)
		defer e.Stop()

		var recoveryErr error
		if e.recoveryService != nil {
			_, recoveryErr = e.recoveryService.RecoverStaleRoutineTaskRecords(
				workerCtx,
				time.Now().UTC().Add(-sconstants.RoutineTaskExecutionLeaseTTL),
			)
		}
		if recoveryErr != nil {
			e.isHealthy.Store(false)
			if slogs.NotegicLogger != nil {
				slogs.NotegicLogger.Error(
					workerCtx,
					recoveryErr,
					"Failed to recover stale routine task records",
				)
			}
		}
		if recoveryErr == nil {
			request, shouldRequest := e.GetClaimRoutineTasksRequest()
			if shouldRequest {
				if e.claimer == nil {
					e.isHealthy.Store(false)
					if slogs.NotegicLogger != nil {
						slogs.NotegicLogger.Error(
							workerCtx,
							errors.New("DurableJob routine task claimer is not configured"),
							"Failed to claim and handle routine task assignments",
						)
					}
				} else if response, exception := e.claimer.ClaimRoutineTasks(workerCtx, request); exception != nil {
					e.isHealthy.Store(false)
					if slogs.NotegicLogger != nil {
						slogs.NotegicLogger.Error(
							workerCtx,
							exception,
							"Failed to claim and handle routine task assignments",
						)
					}
				} else if response == nil {
					e.isHealthy.Store(false)
					if slogs.NotegicLogger != nil {
						slogs.NotegicLogger.Error(
							workerCtx,
							errors.New("DurableJob routine task claimer returned an empty response"),
							"Failed to claim and handle routine task assignments",
						)
					}
				} else if err := e.HandleRoutineTaskAssignments(workerCtx, response.Assignments); err != nil {
					e.isHealthy.Store(false)
					if slogs.NotegicLogger != nil {
						slogs.NotegicLogger.Error(
							workerCtx,
							err,
							"Failed to claim and handle routine task assignments",
						)
					}
				} else {
					e.isHealthy.Store(true)
				}
			}
		}
		for {
			select {
			case <-workerCtx.Done():
				return
			case <-e.ticker.C:
				if e.recoveryService != nil {
					_, err := e.recoveryService.RecoverStaleRoutineTaskRecords(
						workerCtx,
						time.Now().UTC().Add(-sconstants.RoutineTaskExecutionLeaseTTL),
					)
					if err != nil {
						e.isHealthy.Store(false)
						if slogs.NotegicLogger != nil {
							slogs.NotegicLogger.Error(
								workerCtx,
								err,
								"Failed to recover stale routine task records",
							)
						}
						continue
					}
				}
				request, shouldRequest := e.GetClaimRoutineTasksRequest()
				if !shouldRequest {
					continue
				}
				if e.claimer == nil {
					e.isHealthy.Store(false)
					if slogs.NotegicLogger != nil {
						slogs.NotegicLogger.Error(
							workerCtx,
							errors.New("DurableJob routine task claimer is not configured"),
							"Failed to claim and handle routine task assignments",
						)
					}
					continue
				}
				response, exception := e.claimer.ClaimRoutineTasks(workerCtx, request)
				if exception != nil {
					e.isHealthy.Store(false)
					if slogs.NotegicLogger != nil {
						slogs.NotegicLogger.Error(
							workerCtx,
							exception,
							"Failed to claim and handle routine task assignments",
						)
					}
					continue
				}
				if response == nil {
					e.isHealthy.Store(false)
					if slogs.NotegicLogger != nil {
						slogs.NotegicLogger.Error(
							workerCtx,
							errors.New("DurableJob routine task claimer returned an empty response"),
							"Failed to claim and handle routine task assignments",
						)
					}
					continue
				}
				if err := e.HandleRoutineTaskAssignments(workerCtx, response.Assignments); err != nil {
					e.isHealthy.Store(false)
					if slogs.NotegicLogger != nil {
						slogs.NotegicLogger.Error(
							workerCtx,
							err,
							"Failed to claim and handle routine task assignments",
						)
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
