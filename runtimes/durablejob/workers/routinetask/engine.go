package routinetask

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	cdurablejobroutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	durablejobconfig "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/configs"
	routineexecution "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask"
	routinetaskrecoverers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/recovery/recoverers"
	realtimegatewayproducers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/transports/realtimegateway/producers"
)

type Engine struct {
	ticker             *time.Ticker
	workerId           uuid.UUID
	batchSize          int
	isHealthy          atomic.Bool
	isManagingWork     atomic.Bool
	claimer            *Claimer
	recoverer          routinetaskrecoverers.StaleRecordRecovererInterface
	routineTaskManager Manager
}

func NewEngine(
	_ durablejobconfig.Config,
	claimer *Claimer,
	planService *routineexecution.PlanService,
	executionService routineexecution.RoutineTaskExecutionServiceInterface,
	runningPublisher *realtimegatewayproducers.RoutineTaskLifecycleProducer,
	completionPublisher *realtimegatewayproducers.RoutineTaskCompletionProducer,
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
	var managerDB *gorm.DB
	if claimer != nil {
		managerDB = claimer.db
	}
	engine.routineTaskManager = NewManager(
		managerDB,
		planService,
		executionService,
		runningPublisher,
		completionPublisher,
		engine.workerId,
	)
	engine.isHealthy.Store(true)

	return engine
}

func (e *Engine) SetRoutineTaskRecoverer(
	recoverer routinetaskrecoverers.StaleRecordRecovererInterface,
) {
	e.recoverer = recoverer
}

func (e *Engine) GetClaimRoutinesRequest() (cdurablejob.ClaimRoutinesRequestDto, bool) {
	if e.isManagingWork.Load() {
		return cdurablejob.ClaimRoutinesRequestDto{}, false
	}

	return cdurablejob.ClaimRoutinesRequestDto{
		RequestId: uuid.New(),
		WorkerId:  e.workerId,
		BatchSize: e.batchSize,
	}, true
}

func (e *Engine) HandleRoutineAssignments(
	ctx context.Context,
	routines []cdurablejobroutinetasktypes.RoutineAssignment,
) error {
	if len(routines) == 0 {
		return nil
	}
	if !e.isManagingWork.CompareAndSwap(false, true) {
		return errors.New("DurableJob routine task worker is at capacity")
	}
	defer e.isManagingWork.Store(false)

	if err := e.routineTaskManager.Manage(ctx, routines); err != nil {
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
		if e.recoverer != nil {
			_, recoveryErr = e.recoverer.RecoverStaleRoutineTaskRecords(
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
			request, shouldRequest := e.GetClaimRoutinesRequest()
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
				} else if response, exception := e.claimer.ClaimRoutines(workerCtx, request); exception != nil {
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
				} else if err := e.HandleRoutineAssignments(workerCtx, response.RoutineAssignments); err != nil {
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
				if e.recoverer != nil {
					_, err := e.recoverer.RecoverStaleRoutineTaskRecords(
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
				request, shouldRequest := e.GetClaimRoutinesRequest()
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
				response, exception := e.claimer.ClaimRoutines(workerCtx, request)
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
				if err := e.HandleRoutineAssignments(workerCtx, response.RoutineAssignments); err != nil {
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
