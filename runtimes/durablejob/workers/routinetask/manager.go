package routinetask

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	cdurablejobroutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	routinetasksql "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/sqls/routinetask"
	routinetaskdependencies "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/dependencies"
	routinetaskbuilders "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/plan/builders"
	routinetaskpreparers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/preparation/preparers"
	validation "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/validations"
)

type Manager struct {
	maxWorkers       int
	activeWorkers    atomic.Int32
	workerPool       sync.WaitGroup
	sem              chan struct{}
	workerId         uuid.UUID
	db               *gorm.DB
	planBuilder      *routinetaskbuilders.DeterministicPlanBuilder
	failed           []failedRoutineTask
	failedMutex      sync.Mutex
	success          []preparedRoutineTask
	successMutex     sync.Mutex
	preparer         *routinetaskpreparers.AssignmentPreparer
	resultWriter     ResultWriteFunc
	runningPublisher RoutineTaskRunningPublisher
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
		maxWorkers:  maxWorkers,
		sem:         make(chan struct{}, maxWorkers),
		workerId:    workerId,
		planBuilder: &routinetaskbuilders.DeterministicPlanBuilder{},
		preparer:    routinetaskpreparers.NewAssignmentPreparer(validation.New()),
	}
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

func (hm *Manager) transitionRoutinePhase(
	ctx context.Context,
	routineIds []uuid.UUID,
	phase cenums.RoutinePhase,
) error {
	if hm.db == nil {
		return nil
	}
	return hm.setRoutinePhase(hm.db.WithContext(ctx), routineIds, phase)
}

func (hm *Manager) validateRoutineDependencies(
	db *gorm.DB,
	assignments []cdurablejobroutinetasktypes.RoutineTaskAssignment,
) (map[uuid.UUID]string, error) {
	invalidReasons := make(map[uuid.UUID]string)
	if db == nil || len(assignments) == 0 {
		return invalidReasons, nil
	}

	routineRecordIds := make([]uuid.UUID, 0, len(assignments))
	seenRoutineRecordIds := make(map[uuid.UUID]struct{}, len(assignments))
	for _, assignment := range assignments {
		if _, exists := seenRoutineRecordIds[assignment.RoutineRecordId]; exists {
			continue
		}
		seenRoutineRecordIds[assignment.RoutineRecordId] = struct{}{}
		routineRecordIds = append(routineRecordIds, assignment.RoutineRecordId)
	}

	var records []sschemas.RoutineRecord
	if result := db.
		Where("id IN ?", routineRecordIds).
		Find(&records); result.Error != nil {
		return nil, fmt.Errorf("find routine records for dependency validation: %w", result.Error)
	}

	recordById := make(map[uuid.UUID]sschemas.RoutineRecord, len(records))
	for _, record := range records {
		recordById[record.Id] = record
	}
	validatedRoutineRecordIds := make(map[uuid.UUID]struct{}, len(routineRecordIds))
	for _, assignment := range assignments {
		if _, validated := validatedRoutineRecordIds[assignment.RoutineRecordId]; validated {
			continue
		}
		validatedRoutineRecordIds[assignment.RoutineRecordId] = struct{}{}
		record, exists := recordById[assignment.RoutineRecordId]
		if !exists {
			invalidReasons[assignment.RoutineRecordId] = "routine record for dependency validation was not found"
			continue
		}

		var snapshot struct {
			RoutineTasks []struct {
				Id                     uuid.UUID                 `json:"id"`
				Purpose                cenums.RoutineTaskPurpose `json:"purpose"`
				PreviousRoutineTaskIds []uuid.UUID               `json:"previousRoutineTaskIds"`
			} `json:"routineTasks"`
		}
		if len(record.Snapshot) > 0 && string(record.Snapshot) != "{}" {
			if err := json.Unmarshal(record.Snapshot, &snapshot); err != nil {
				invalidReasons[record.Id] = fmt.Sprintf("decode routine task dependency snapshot: %v", err)
				continue
			}
		}
		taskIds := make([]uuid.UUID, len(snapshot.RoutineTasks))
		dependencies := make([]routinetaskdependencies.Edge, 0)
		for index, task := range snapshot.RoutineTasks {
			taskIds[index] = task.Id
			for _, previousTaskId := range task.PreviousRoutineTaskIds {
				dependencies = append(dependencies, routinetaskdependencies.Edge{
					TaskId:         task.Id,
					PreviousTaskId: previousTaskId,
				})
			}
		}
		if err := routinetaskdependencies.Validate(taskIds, dependencies); err != nil {
			invalidReasons[record.Id] = err.Error()
		}
	}

	return invalidReasons, nil
}

func (hm *Manager) buildPlans(
	db *gorm.DB,
	assignments []cdurablejobroutinetasktypes.RoutineTaskAssignment,
) (map[uuid.UUID]string, error) {
	invalidReasons := make(map[uuid.UUID]string)
	if db == nil || len(assignments) == 0 {
		return invalidReasons, nil
	}

	routineRecordIds := make([]uuid.UUID, 0, len(assignments))
	seenRoutineRecordIds := make(map[uuid.UUID]struct{}, len(assignments))
	for _, assignment := range assignments {
		if _, exists := seenRoutineRecordIds[assignment.RoutineRecordId]; exists {
			continue
		}
		seenRoutineRecordIds[assignment.RoutineRecordId] = struct{}{}
		routineRecordIds = append(routineRecordIds, assignment.RoutineRecordId)
	}

	var records []sschemas.RoutineRecord
	if result := db.
		Where("id IN ?", routineRecordIds).
		Find(&records); result.Error != nil {
		return nil, fmt.Errorf("find routine records for plan building: %w", result.Error)
	}

	recordById := make(map[uuid.UUID]sschemas.RoutineRecord, len(records))
	for _, record := range records {
		recordById[record.Id] = record
	}

	plannedSnapshots := make(map[uuid.UUID][]byte, len(records))
	for _, assignment := range assignments {
		if _, exists := plannedSnapshots[assignment.RoutineRecordId]; exists {
			continue
		}
		record, exists := recordById[assignment.RoutineRecordId]
		if !exists {
			invalidReasons[assignment.RoutineRecordId] = "routine record for plan building was not found"
			continue
		}

		var snapshot struct {
			RoutineTaskPlan *cdurablejobroutinetasktypes.RoutineTaskPlan `json:"routineTaskPlan"`
			RoutineTasks    []struct {
				Id                     uuid.UUID                 `json:"id"`
				Purpose                cenums.RoutineTaskPurpose `json:"purpose"`
				Payload                json.RawMessage           `json:"payload"`
				PreviousRoutineTaskIds []uuid.UUID               `json:"previousRoutineTaskIds"`
			} `json:"routineTasks"`
		}
		if len(record.Snapshot) > 0 && string(record.Snapshot) != "{}" {
			if err := json.Unmarshal(record.Snapshot, &snapshot); err != nil {
				invalidReasons[record.Id] = fmt.Sprintf("decode routine task plan snapshot: %v", err)
				continue
			}
		}
		tasks := make([]sschemas.RoutineTask, len(snapshot.RoutineTasks))
		dependencies := make([]sschemas.RoutineTaskDependency, 0)
		for index, task := range snapshot.RoutineTasks {
			tasks[index] = sschemas.RoutineTask{
				Id:        task.Id,
				RoutineId: record.RoutineId,
				Purpose:   task.Purpose,
				Payload:   datatypes.JSON(task.Payload),
			}
			for _, previousTaskId := range task.PreviousRoutineTaskIds {
				dependencies = append(dependencies, sschemas.RoutineTaskDependency{
					RoutineTaskId:         task.Id,
					PreviousRoutineTaskId: previousTaskId,
				})
			}
		}
		plan, err := hm.planBuilder.Build(record.RoutineId, tasks, dependencies, snapshot.RoutineTaskPlan)
		if err != nil {
			invalidReasons[record.Id] = err.Error()
			continue
		}
		snapshot.RoutineTaskPlan = plan
		snapshotData, err := json.Marshal(snapshot)
		if err != nil {
			invalidReasons[record.Id] = fmt.Sprintf("encode routine task plan snapshot: %v", err)
			continue
		}
		plannedSnapshots[record.Id] = snapshotData
	}

	if len(plannedSnapshots) == 0 {
		return invalidReasons, nil
	}
	caseSQL := "CASE id"
	caseArgs := make([]any, 0, len(plannedSnapshots)*2)
	recordIds := make([]uuid.UUID, 0, len(plannedSnapshots))
	for recordId, snapshotData := range plannedSnapshots {
		caseSQL += " WHEN ? THEN ?::jsonb"
		caseArgs = append(caseArgs, recordId, snapshotData)
		recordIds = append(recordIds, recordId)
	}
	caseSQL += " ELSE snapshot END"
	result := db.
		Model(&sschemas.RoutineRecord{}).
		Where("id IN ?", recordIds).
		Updates(map[string]any{
			"snapshot":   gorm.Expr(caseSQL, caseArgs...),
			"updated_at": time.Now().UTC(),
		})
	if result.Error != nil {
		return nil, fmt.Errorf("persist routine task plans: %w", result.Error)
	}
	return invalidReasons, nil
}

func (hm *Manager) markInvalidRoutines(
	db *gorm.DB,
	invalidReasons map[uuid.UUID]string,
	phase cenums.RoutinePhase,
) error {
	if db == nil || len(invalidReasons) == 0 {
		return nil
	}
	recordIds := make([]uuid.UUID, 0, len(invalidReasons))
	for recordId := range invalidReasons {
		recordIds = append(recordIds, recordId)
	}
	now := time.Now().UTC()
	errorReasonSQL := "CASE routine_record_id"
	errorReasonArgs := make([]any, 0, len(invalidReasons)*2)
	for recordId, reason := range invalidReasons {
		if len(reason) > 256 {
			reason = reason[:256]
		}
		errorReasonSQL += " WHEN ? THEN ?"
		errorReasonArgs = append(errorReasonArgs, recordId, reason)
	}
	errorReasonSQL += " ELSE error_reason END"
	result := db.Model(&sschemas.RoutineTaskRecord{}).
		Where("routine_record_id IN ?", recordIds).
		Where("status IN ?", []cenums.RoutineTaskRecordStatus{
			cenums.RoutineTaskRecordStatus_Waiting,
			cenums.RoutineTaskRecordStatus_Ready,
			cenums.RoutineTaskRecordStatus_Running,
		}).
		Updates(map[string]any{
			"status":       cenums.RoutineTaskRecordStatus_Blocked,
			"error_code":   cenums.RoutineTaskRecordErrorCode_PayloadInvalid,
			"error_reason": gorm.Expr(errorReasonSQL, errorReasonArgs...),
			"updated_at":   now,
		})
	if result.Error != nil {
		return result.Error
	}
	result = db.Model(&sschemas.RoutineRecord{}).
		Where("id IN ?", recordIds).
		Updates(map[string]any{
			"status":          cenums.RoutineRecordStatus_Blocked,
			"actual_ended_at": now,
			"updated_at":      now,
		})
	if result.Error != nil {
		return result.Error
	}
	if result = db.Exec(
		routinetasksql.UpdateRoutineRecordAggregateSQL,
		cenums.RoutineRecordStatus_Running,
		cenums.RoutineRecordStatus_Blocked,
		cenums.RoutineRecordStatus_Failed,
		cenums.RoutineRecordStatus_Success,
		now,
		now,
		recordIds,
	); result.Error != nil {
		return result.Error
	}
	routineIdsQuery := db.Model(&sschemas.RoutineRecord{}).
		Select("routine_id").
		Where("id IN ?", recordIds)
	if result = db.Model(&sschemas.Routine{}).
		Where("id IN (?)", routineIdsQuery).
		Updates(map[string]any{
			"status":     cenums.RoutineStatus_OverDue,
			"phase":      phase,
			"updated_at": now,
		}); result.Error != nil {
		return result.Error
	}
	return nil
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
	routines []cdurablejobroutinetasktypes.RoutineAssignment,
) error {
	if len(routines) == 0 {
		return nil
	}
	assignments := make([]cdurablejobroutinetasktypes.RoutineTaskAssignment, 0)
	emptyRoutineIds := make([]uuid.UUID, 0)
	emptyRoutinePlaceholders := make([]cdurablejobroutinetasktypes.RoutineTaskAssignment, 0)
	for _, routine := range routines {
		if len(routine.RoutineTasks) > 0 {
			assignments = append(assignments, routine.RoutineTasks...)
			continue
		}
		emptyRoutineIds = append(emptyRoutineIds, routine.RoutineId)
		emptyRoutinePlaceholders = append(emptyRoutinePlaceholders, cdurablejobroutinetasktypes.RoutineTaskAssignment{
			RoutineId:       routine.RoutineId,
			RoutineRecordId: routine.RoutineRecordId,
		})
	}
	// Process routines that have no claimed tasks.
	if len(emptyRoutineIds) > 0 {
		var preparationInvalidReasons map[uuid.UUID]string
		if hm.db == nil {
			preparationInvalidReasons = make(map[uuid.UUID]string)
		} else {
			tx := hm.db.WithContext(ctx).Begin()
			if err := hm.setRoutinePhase(tx, emptyRoutineIds, cenums.RoutinePhase_Preparation); err != nil {
				tx.Rollback()
				return err
			}
			var err error
			preparationInvalidReasons, err = hm.validateRoutineDependencies(tx, emptyRoutinePlaceholders)
			if err != nil {
				tx.Rollback()
				return err
			}
			if err := hm.markInvalidRoutines(tx, preparationInvalidReasons, cenums.RoutinePhase_Preparation); err != nil {
				tx.Rollback()
				return err
			}
			if err := tx.Commit().Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("commit empty routine preparation transaction: %w", err)
			}
		}
		validEmptyRoutineIds := make([]uuid.UUID, 0, len(emptyRoutineIds))
		validEmptyPlaceholders := make([]cdurablejobroutinetasktypes.RoutineTaskAssignment, 0, len(emptyRoutinePlaceholders))
		for index, routineId := range emptyRoutineIds {
			if _, invalid := preparationInvalidReasons[emptyRoutinePlaceholders[index].RoutineRecordId]; invalid {
				continue
			}
			validEmptyRoutineIds = append(validEmptyRoutineIds, routineId)
			validEmptyPlaceholders = append(validEmptyPlaceholders, emptyRoutinePlaceholders[index])
		}
		emptyRoutineIds = validEmptyRoutineIds
		emptyRoutinePlaceholders = validEmptyPlaceholders
		if len(emptyRoutineIds) == 0 {
			return nil
		}
		var invalidReasons map[uuid.UUID]string
		if hm.db == nil {
			invalidReasons = make(map[uuid.UUID]string)
		} else {
			tx := hm.db.WithContext(ctx).Begin()
			if err := hm.setRoutinePhase(tx, emptyRoutineIds, cenums.RoutinePhase_Plan); err != nil {
				tx.Rollback()
				return err
			}
			var err error
			invalidReasons, err = hm.buildPlans(tx, emptyRoutinePlaceholders)
			if err != nil {
				tx.Rollback()
				return err
			}
			if err := hm.markInvalidRoutines(tx, invalidReasons, cenums.RoutinePhase_Plan); err != nil {
				tx.Rollback()
				return err
			}
			if err := tx.Commit().Error; err != nil {
				tx.Rollback()
				return fmt.Errorf("commit empty routine plan transaction: %w", err)
			}
		}
	}
	// Prepare and validate claimed task assignments.
	if len(assignments) == 0 {
		return nil
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
	var preparationInvalidReasons map[uuid.UUID]string
	if hm.db == nil {
		preparationInvalidReasons = make(map[uuid.UUID]string)
	} else {
		tx := hm.db.WithContext(ctx).Begin()
		if err := hm.setRoutinePhase(tx, routineIds, cenums.RoutinePhase_Preparation); err != nil {
			tx.Rollback()
			return err
		}
		var err error
		preparationInvalidReasons, err = hm.validateRoutineDependencies(tx, assignments)
		if err != nil {
			tx.Rollback()
			return err
		}
		if err := hm.markInvalidRoutines(tx, preparationInvalidReasons, cenums.RoutinePhase_Preparation); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark invalid routine dependencies: %w", err)
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("commit routine task preparation transaction: %w", err)
		}
	}
	if len(preparationInvalidReasons) > 0 {
		validAssignments := make([]cdurablejobroutinetasktypes.RoutineTaskAssignment, 0, len(assignments))
		for _, assignment := range assignments {
			if _, invalid := preparationInvalidReasons[assignment.RoutineRecordId]; !invalid {
				validAssignments = append(validAssignments, assignment)
			}
		}
		assignments = validAssignments
		if len(assignments) == 0 {
			return nil
		}
	}
	// Run task preparers concurrently.
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

			preparedTask, err := hm.preparer.Prepare(ctx, assignment)
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

	// Build and persist deterministic plans in one transaction.
	hm.workerPool.Wait()
	var invalidReasons map[uuid.UUID]string
	if hm.db == nil {
		invalidReasons = make(map[uuid.UUID]string)
	} else {
		tx := hm.db.WithContext(ctx).Begin()
		if err := hm.setRoutinePhase(tx, routineIds, cenums.RoutinePhase_Plan); err != nil {
			tx.Rollback()
			return err
		}
		var err error
		invalidReasons, err = hm.buildPlans(tx, assignments)
		if err != nil {
			tx.Rollback()
			return err
		}
		if err := hm.markInvalidRoutines(tx, invalidReasons, cenums.RoutinePhase_Plan); err != nil {
			tx.Rollback()
			return fmt.Errorf("mark invalid routine plans: %w", err)
		}
		if err := tx.Commit().Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("commit routine task plan transaction: %w", err)
		}
	}
	if len(invalidReasons) > 0 {
		hm.successMutex.Lock()
		filteredSuccesses := hm.success[:0]
		for _, success := range hm.success {
			if _, invalid := invalidReasons[success.preparedTask.RoutineRecordId]; !invalid {
				filteredSuccesses = append(filteredSuccesses, success)
			}
		}
		hm.success = filteredSuccesses
		hm.successMutex.Unlock()
		hm.failedMutex.Lock()
		filteredFailures := hm.failed[:0]
		for _, failure := range hm.failed {
			if _, invalid := invalidReasons[failure.assignment.RoutineRecordId]; !invalid {
				filteredFailures = append(filteredFailures, failure)
			}
		}
		hm.failed = filteredFailures
		hm.failedMutex.Unlock()
	}
	// Publish execution results and advance the routine phase.
	validRoutineIds := make([]uuid.UUID, 0, len(routineIds))
	for _, routineId := range routineIds {
		valid := true
		for _, assignment := range assignments {
			if assignment.RoutineId != routineId {
				continue
			}
			if _, invalid := invalidReasons[assignment.RoutineRecordId]; invalid {
				valid = false
			}
			break
		}
		if valid {
			validRoutineIds = append(validRoutineIds, routineId)
		}
	}
	if err := hm.transitionRoutinePhase(ctx, validRoutineIds, cenums.RoutinePhase_Execution); err != nil {
		return err
	}
	if err := hm.publishSuccesses(ctx); err != nil {
		_ = hm.transitionRoutinePhase(ctx, validRoutineIds, cenums.RoutinePhase_Recovery)
		return err
	}
	if hm.hasFailures() {
		if err := hm.transitionRoutinePhase(ctx, validRoutineIds, cenums.RoutinePhase_Recovery); err != nil {
			return err
		}
		if err := hm.publishFailures(ctx); err != nil {
			return err
		}
	}
	return hm.transitionRoutinePhase(ctx, validRoutineIds, cenums.RoutinePhase_Analysis)

}

func (hm *Manager) hasFailures() bool {
	hm.failedMutex.Lock()
	defer hm.failedMutex.Unlock()
	return len(hm.failed) > 0
}

func (hm *Manager) publishSuccesses(ctx context.Context) error {
	hm.successMutex.Lock()
	successes := append([]preparedRoutineTask(nil), hm.success...)
	hm.successMutex.Unlock()
	if len(successes) == 0 {
		return nil
	}
	if hm.resultWriter == nil {
		return errors.New("DurableJob routine task result writer is not configured")
	}

	correlationId := uuid.New().String()
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
	return hm.resultWriter(ctx, Result{
		Kind:          ResultKind_Completed,
		WorkerId:      hm.workerId,
		CorrelationId: correlationId,
		Data:          request,
	})
}

func (hm *Manager) publishFailures(ctx context.Context) error {
	hm.failedMutex.Lock()
	failures := append([]failedRoutineTask(nil), hm.failed...)
	hm.failedMutex.Unlock()
	if len(failures) == 0 {
		return nil
	}
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
	return hm.resultWriter(ctx, Result{
		Kind:          ResultKind_Failed,
		WorkerId:      hm.workerId,
		CorrelationId: uuid.New().String(),
		Data:          request,
	})
}
