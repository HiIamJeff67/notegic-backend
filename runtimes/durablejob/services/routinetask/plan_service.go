package routinetask

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	cdurablejobroutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	routinetasksql "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/sqls/routinetask"
	durablejobexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/exceptions"
	routinetaskbuilders "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/plan/builders"
	routinetaskpreparers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/plan/preparers"
)

type PlanService struct {
	db          *gorm.DB
	planBuilder *routinetaskbuilders.DeterministicPlanBuilder
	preparer    *routinetaskpreparers.AssignmentPreparer
}

func NewPlanService(
	db *gorm.DB,
	validatorInstance *validator.Validate,
	routineTaskException durablejobexceptions.RoutineTaskException,
) *PlanService {
	return &PlanService{
		db:          db,
		planBuilder: &routinetaskbuilders.DeterministicPlanBuilder{},
		preparer:    routinetaskpreparers.NewAssignmentPreparer(validatorInstance, routineTaskException),
	}
}

func (s *PlanService) markInvalidRoutines(
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
	result := db.
		Model(&sschemas.RoutineTaskRecord{}).
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
	result = db.
		Model(&sschemas.RoutineRecord{}).
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
	routineIdsQuery := db.
		Model(&sschemas.RoutineRecord{}).
		Select("routine_id").
		Where("id IN ?", recordIds)
	if result = db.
		Model(&sschemas.Routine{}).
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

func (s *PlanService) BuildRoutineTaskPlans(
	ctx context.Context,
	assignments []cdurablejobroutinetasktypes.RoutineTaskAssignment,
) ([]cdurablejobroutinetasktypes.CompletedRoutineTask, []cdurablejobroutinetasktypes.FailedRoutineTask, map[uuid.UUID]string, error) {
	completedTasks := make([]cdurablejobroutinetasktypes.CompletedRoutineTask, 0, len(assignments))
	failedTasks := make([]cdurablejobroutinetasktypes.FailedRoutineTask, 0)
	for _, assignment := range assignments {
		preparedTask, err := s.preparer.Prepare(ctx, assignment)
		if err == nil && preparedTask != nil {
			completedTasks = append(completedTasks, cdurablejobroutinetasktypes.CompletedRoutineTask{
				RoutineTaskId:       preparedTask.RoutineTaskId,
				RoutineTaskRecordId: preparedTask.RoutineTaskRecordId,
				RoutineRecordId:     preparedTask.RoutineRecordId,
				CompletedAt:         time.Now().UTC(),
				PreparedTask:        preparedTask,
			})
			continue
		}

		errorCode := cenums.RoutineTaskRecordErrorCode_HandlerFailed
		errorReason := "routine task assignment planning failed"
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
		}
		if len(errorReason) > 256 {
			errorReason = errorReason[:256]
		}
		failedTasks = append(failedTasks, cdurablejobroutinetasktypes.FailedRoutineTask{
			RoutineTaskId:       assignment.RoutineTaskId,
			RoutineTaskRecordId: assignment.RoutineTaskRecordId,
			RoutineRecordId:     assignment.RoutineRecordId,
			FailedAt:            time.Now().UTC(),
			ErrorCode:           errorCode,
			ErrorReason:         errorReason,
		})
	}
	invalidReasons := make(map[uuid.UUID]string, len(failedTasks))
	for _, failedTask := range failedTasks {
		invalidReasons[failedTask.RoutineRecordId] = failedTask.ErrorReason
	}
	if s.db == nil || len(assignments) == 0 {
		return completedTasks, failedTasks, invalidReasons, nil
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

	tx := s.db.WithContext(ctx).Begin()
	var records []sschemas.RoutineRecord
	if result := tx.
		Model(&sschemas.RoutineRecord{}).
		Where("id IN ?", routineRecordIds).
		Find(&records); result.Error != nil {
		tx.Rollback()
		return nil, nil, nil, fmt.Errorf("find routine records for plan building: %w", result.Error)
	}

	routineIds := make([]uuid.UUID, 0, len(records))
	seenRoutineIds := make(map[uuid.UUID]struct{}, len(records))
	recordById := make(map[uuid.UUID]sschemas.RoutineRecord, len(records))
	for _, record := range records {
		recordById[record.Id] = record
		if _, exists := seenRoutineIds[record.RoutineId]; exists {
			continue
		}
		seenRoutineIds[record.RoutineId] = struct{}{}
		routineIds = append(routineIds, record.RoutineId)
	}
	routineTaskIds := make([]uuid.UUID, 0, len(assignments))
	seenRoutineTaskIds := make(map[uuid.UUID]struct{}, len(assignments))
	for _, assignment := range assignments {
		if _, exists := seenRoutineTaskIds[assignment.RoutineTaskId]; exists {
			continue
		}
		seenRoutineTaskIds[assignment.RoutineTaskId] = struct{}{}
		routineTaskIds = append(routineTaskIds, assignment.RoutineTaskId)
	}

	if result := tx.
		Model(&sschemas.Routine{}).
		Where("id IN ?", routineIds).
		Update("phase", cenums.RoutinePhase_Plan); result.Error != nil {
		tx.Rollback()
		return nil, nil, nil, fmt.Errorf("mark routines as planning: %w", result.Error)
	}
	if result := tx.
		Model(&sschemas.RoutineTask{}).
		Where("id IN ?", routineTaskIds).
		Update("phase", cenums.RoutinePhase_Plan); result.Error != nil {
		tx.Rollback()
		return nil, nil, nil, fmt.Errorf("mark routine tasks as planning: %w", result.Error)
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
		builtPlan, err := s.planBuilder.Build(
			record.RoutineId,
			tasks,
			dependencies,
			snapshot.RoutineTaskPlan,
		)
		if err != nil {
			invalidReasons[record.Id] = err.Error()
			continue
		}
		snapshot.RoutineTaskPlan = builtPlan
		snapshotData, err := json.Marshal(snapshot)
		if err != nil {
			invalidReasons[record.Id] = fmt.Sprintf("encode routine task plan snapshot: %v", err)
			continue
		}
		plannedSnapshots[record.Id] = snapshotData
	}

	if len(plannedSnapshots) > 0 {
		caseSQL := "CASE id"
		caseArgs := make([]any, 0, len(plannedSnapshots)*2)
		recordIds := make([]uuid.UUID, 0, len(plannedSnapshots))
		for recordId, snapshotData := range plannedSnapshots {
			caseSQL += " WHEN ? THEN ?::jsonb"
			caseArgs = append(caseArgs, recordId, snapshotData)
			recordIds = append(recordIds, recordId)
		}
		caseSQL += " ELSE snapshot END"
		if result := tx.
			Model(&sschemas.RoutineRecord{}).
			Where("id IN ?", recordIds).
			Updates(map[string]any{
				"snapshot":   gorm.Expr(caseSQL, caseArgs...),
				"updated_at": time.Now().UTC(),
			}); result.Error != nil {
			tx.Rollback()
			return nil, nil, nil, fmt.Errorf("persist routine task plans: %w", result.Error)
		}
	}
	if err := s.markInvalidRoutines(tx, invalidReasons, cenums.RoutinePhase_Plan); err != nil {
		tx.Rollback()
		return nil, nil, nil, fmt.Errorf("mark invalid routine plans: %w", err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, nil, nil, fmt.Errorf("commit routine task plan transaction: %w", err)
	}

	return completedTasks, failedTasks, invalidReasons, nil
}
