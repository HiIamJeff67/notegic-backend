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
	planService         *routineexecution.PlanService
	executionService    routineexecution.RoutineTaskExecutionServiceInterface
	runningPublisher    *realtimegatewayproducers.RoutineTaskLifecycleProducer
	completionPublisher *realtimegatewayproducers.RoutineTaskCompletionProducer
}

func NewManager(
	db *gorm.DB,
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
		db:                  db,
		planService:         planService,
		executionService:    executionService,
		runningPublisher:    runningPublisher,
		completionPublisher: completionPublisher,
	}
}

func (hm *Manager) setPhase(
	db *gorm.DB,
	routineIds []uuid.UUID,
	routineTaskIds []uuid.UUID,
	phase cenums.RoutinePhase,
) error {
	if len(routineIds) == 0 && len(routineTaskIds) == 0 {
		return nil
	}
	if db == nil {
		return fmt.Errorf("DurableJob routine phase database is not configured")
	}

	tx := db.Begin()
	if len(routineIds) > 0 {
		if result := tx.
			Model(&sschemas.Routine{}).
			Where("id IN ?", routineIds).
			Update("phase", phase); result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	if len(routineTaskIds) > 0 {
		if result := tx.
			Model(&sschemas.RoutineTask{}).
			Where("id IN ?", routineTaskIds).
			Update("phase", phase); result.Error != nil {
			tx.Rollback()
			return result.Error
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
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
	routineTaskIds := make([]uuid.UUID, 0, len(assignments))
	seenRoutineTaskIds := make(map[uuid.UUID]struct{}, len(assignments))
	for _, assignment := range assignments {
		if _, exists := seenRoutineIds[assignment.RoutineId]; !exists {
			seenRoutineIds[assignment.RoutineId] = struct{}{}
			routineIds = append(routineIds, assignment.RoutineId)
		}
		if _, exists := seenRoutineTaskIds[assignment.RoutineTaskId]; !exists {
			seenRoutineTaskIds[assignment.RoutineTaskId] = struct{}{}
			routineTaskIds = append(routineTaskIds, assignment.RoutineTaskId)
		}
	}
	// Build and persist the deterministic plan, including assignment planning.
	success, failed, invalidReasons, err := hm.planService.BuildRoutineTaskPlans(ctx, assignments)
	if err != nil {
		return err
	}
	if len(invalidReasons) > 0 {
		filteredSuccesses := success[:0]
		for _, completedTask := range success {
			if _, invalid := invalidReasons[completedTask.RoutineRecordId]; !invalid {
				filteredSuccesses = append(filteredSuccesses, completedTask)
			}
		}
		success = filteredSuccesses
		filteredFailures := failed[:0]
		for _, failedTask := range failed {
			if _, invalid := invalidReasons[failedTask.RoutineRecordId]; !invalid {
				filteredFailures = append(filteredFailures, failedTask)
			}
		}
		failed = filteredFailures
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
	validRoutineTaskIds := make([]uuid.UUID, 0, len(routineTaskIds))
	for _, assignment := range assignments {
		if _, invalid := invalidReasons[assignment.RoutineRecordId]; invalid {
			continue
		}
		if _, exists := seenRoutineTaskIds[assignment.RoutineTaskId]; exists {
			validRoutineTaskIds = append(validRoutineTaskIds, assignment.RoutineTaskId)
			delete(seenRoutineTaskIds, assignment.RoutineTaskId)
		}
	}
	if err := hm.setPhase(phaseDB, validRoutineIds, validRoutineTaskIds, cenums.RoutinePhase_Execution); err != nil {
		return err
	}
	if len(success) > 0 || len(failed) > 0 {
		if hm.executionService == nil {
			return fmt.Errorf("DurableJob routine task execution service is not configured")
		}
	}
	if len(success) > 0 {
		result := cdurablejobroutinetasktypes.Result{
			Kind:          cdurablejobroutinetasktypes.ResultKind_Completed,
			WorkerId:      hm.workerId,
			CorrelationId: uuid.New().String(),
			Data: cdurablejob.MarkCompletedRoutineTasksRequestDto{
				WorkerId: hm.workerId,
				Tasks:    success,
			},
		}
		if exception := hm.executionService.ApplyResult(ctx, uuid.New(), result); exception != nil {
			_ = hm.setPhase(phaseDB, validRoutineIds, validRoutineTaskIds, cenums.RoutinePhase_Recovery)
			return exception
		}
		if hm.completionPublisher != nil {
			err := hm.completionPublisher.ProduceRoutineTaskCompleted(
				ctx,
				success,
				hm.workerId,
			)
			if err != nil {
				_ = hm.setPhase(phaseDB, validRoutineIds, validRoutineTaskIds, cenums.RoutinePhase_Recovery)
				return err
			}
		}
	}
	if len(failed) > 0 {
		failedRoutineTaskIds := make([]uuid.UUID, 0, len(failed))
		for _, failedTask := range failed {
			failedRoutineTaskIds = append(failedRoutineTaskIds, failedTask.RoutineTaskId)
		}
		if err := hm.setPhase(phaseDB, validRoutineIds, failedRoutineTaskIds, cenums.RoutinePhase_Recovery); err != nil {
			return err
		}
		result := cdurablejobroutinetasktypes.Result{
			Kind:          cdurablejobroutinetasktypes.ResultKind_Failed,
			WorkerId:      hm.workerId,
			CorrelationId: uuid.New().String(),
			Data: cdurablejob.MarkFailedRoutineTasksRequestDto{
				WorkerId: hm.workerId,
				Tasks:    failed,
			},
		}
		if exception := hm.executionService.ApplyResult(ctx, uuid.New(), result); exception != nil {
			return exception
		}
	}
	return hm.setPhase(phaseDB, validRoutineIds, validRoutineTaskIds, cenums.RoutinePhase_Analysis)
}
