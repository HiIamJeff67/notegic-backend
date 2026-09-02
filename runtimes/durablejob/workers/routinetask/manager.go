package routinetask

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	cdurablejobroutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	routineexecution "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask"
	realtimegatewayproducers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/transports/realtimegateway/producers"
)

type Manager struct {
	workerId            uuid.UUID
	db                  *gorm.DB
	failed              []cdurablejobroutinetasktypes.FailedRoutineTask
	success             []cdurablejobroutinetasktypes.CompletedRoutineTask
	planService         *routineexecution.PlanService
	executionService    routineexecution.RoutineTaskExecutionServiceInterface
	runningPublisher    *realtimegatewayproducers.RoutineTaskLifecycleProducer
	completionPublisher *realtimegatewayproducers.RoutineTaskCompletionProducer
}

func NewManager(
	planService *routineexecution.PlanService,
	executionService routineexecution.RoutineTaskExecutionServiceInterface,
	runningPublisher *realtimegatewayproducers.RoutineTaskLifecycleProducer,
	completionPublisher *realtimegatewayproducers.RoutineTaskCompletionProducer,
	workerIds ...uuid.UUID,
) Manager {
	workerId := uuid.New()
	if len(workerIds) > 0 && workerIds[0] != uuid.Nil {
		workerId = workerIds[0]
	}

	return Manager{
		workerId:            workerId,
		planService:         planService,
		executionService:    executionService,
		runningPublisher:    runningPublisher,
		completionPublisher: completionPublisher,
	}
}

func (hm *Manager) setRoutinePhase(
	db *gorm.DB,
	routineIds []uuid.UUID,
	phase cenums.RoutinePhase,
) error {
	if db == nil || len(routineIds) == 0 {
		return nil
	}
	return db.
		Model(&sschemas.Routine{}).
		Where("id IN ?", routineIds).
		Update("phase", phase).Error
}

func (hm *Manager) Manage(
	ctx context.Context,
	routines []cdurablejobroutinetasktypes.RoutineAssignment,
) error {
	if len(routines) == 0 {
		return nil
	}
	assignments := make([]cdurablejobroutinetasktypes.RoutineTaskAssignment, 0)
	for _, routine := range routines {
		assignments = append(assignments, routine.RoutineTasks...)
	}
	if len(assignments) == 0 {
		return nil
	}
	if hm.planService == nil {
		return fmt.Errorf("DurableJob routine task plan service is not configured")
	}

	routineIds := make([]uuid.UUID, 0, len(assignments))
	seenRoutineIds := make(map[uuid.UUID]struct{}, len(assignments))
	for _, assignment := range assignments {
		if _, exists := seenRoutineIds[assignment.RoutineId]; exists {
			continue
		}
		seenRoutineIds[assignment.RoutineId] = struct{}{}
		routineIds = append(routineIds, assignment.RoutineId)
	}
	// Build and persist the deterministic plan, including assignment planning.
	var invalidReasons map[uuid.UUID]string
	var err error
	hm.success, hm.failed, invalidReasons, err = hm.planService.BuildRoutineTaskPlans(ctx, assignments)
	if err != nil {
		return err
	}
	if len(invalidReasons) > 0 {
		filteredSuccesses := hm.success[:0]
		for _, success := range hm.success {
			if _, invalid := invalidReasons[success.RoutineRecordId]; !invalid {
				filteredSuccesses = append(filteredSuccesses, success)
			}
		}
		hm.success = filteredSuccesses
		filteredFailures := hm.failed[:0]
		for _, failure := range hm.failed {
			if _, invalid := invalidReasons[failure.RoutineRecordId]; !invalid {
				filteredFailures = append(filteredFailures, failure)
			}
		}
		hm.failed = filteredFailures
	}
	invalidRoutineIds := make(map[uuid.UUID]struct{})
	for _, assignment := range assignments {
		if _, invalid := invalidReasons[assignment.RoutineRecordId]; invalid {
			invalidRoutineIds[assignment.RoutineId] = struct{}{}
		}
	}
	if hm.runningPublisher != nil {
		for _, assignment := range assignments {
			if _, invalid := invalidReasons[assignment.RoutineRecordId]; invalid {
				continue
			}
			err := hm.runningPublisher.ProduceRoutineTaskRunning(ctx, assignment)
			if err != nil && slogs.NotegicLogger != nil {
				slogs.NotegicLogger.Error(
					ctx,
					err,
					"Failed to publish RoutineTask running lifecycle event",
				)
			}
		}
	}

	// Apply execution results and advance the routine phase.
	validRoutineIds := make([]uuid.UUID, 0, len(routineIds))
	for _, routineId := range routineIds {
		if _, invalid := invalidRoutineIds[routineId]; !invalid {
			validRoutineIds = append(validRoutineIds, routineId)
		}
	}
	phaseDB := hm.db
	if phaseDB != nil {
		phaseDB = phaseDB.WithContext(ctx)
	}
	if err := hm.setRoutinePhase(phaseDB, validRoutineIds, cenums.RoutinePhase_Execution); err != nil {
		return err
	}
	if len(hm.success) > 0 || len(hm.failed) > 0 {
		if hm.executionService == nil {
			return fmt.Errorf("DurableJob routine task execution service is not configured")
		}
	}
	if len(hm.success) > 0 {
		result := cdurablejobroutinetasktypes.Result{
			Kind:          cdurablejobroutinetasktypes.ResultKind_Completed,
			WorkerId:      hm.workerId,
			CorrelationId: uuid.New().String(),
			Data: cdurablejob.MarkCompletedRoutineTasksRequestDto{
				WorkerId: hm.workerId,
				Tasks:    hm.success,
			},
		}
		if exception := hm.executionService.ApplyResult(ctx, uuid.New(), result); exception != nil {
			_ = hm.setRoutinePhase(phaseDB, validRoutineIds, cenums.RoutinePhase_Recovery)
			return exception
		}
		if hm.completionPublisher != nil {
			err := hm.completionPublisher.ProduceRoutineTaskCompleted(
				ctx,
				hm.success,
				hm.workerId,
			)
			if err != nil {
				_ = hm.setRoutinePhase(phaseDB, validRoutineIds, cenums.RoutinePhase_Recovery)
				return err
			}
		}
	}
	if len(hm.failed) > 0 {
		if err := hm.setRoutinePhase(phaseDB, validRoutineIds, cenums.RoutinePhase_Recovery); err != nil {
			return err
		}
		result := cdurablejobroutinetasktypes.Result{
			Kind:          cdurablejobroutinetasktypes.ResultKind_Failed,
			WorkerId:      hm.workerId,
			CorrelationId: uuid.New().String(),
			Data: cdurablejob.MarkFailedRoutineTasksRequestDto{
				WorkerId: hm.workerId,
				Tasks:    hm.failed,
			},
		}
		if exception := hm.executionService.ApplyResult(ctx, uuid.New(), result); exception != nil {
			return exception
		}
	}
	return hm.setRoutinePhase(phaseDB, validRoutineIds, cenums.RoutinePhase_Analysis)
}
