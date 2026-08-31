package execution

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"

	handlers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/handlers"
	matchers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/matchers"
	resolvers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/resolvers"
)

type RoutineTaskExecutionServiceInterface interface {
	ApplyPreparedRoutineTasks(
		ctx context.Context,
		eventId uuid.UUID,
		request *cdurablejob.MarkCompletedRoutineTasksRequestDto,
	) *cexceptions.Exception
	ApplyFailedRoutineTasks(
		ctx context.Context,
		eventId uuid.UUID,
		request *cdurablejob.MarkFailedRoutineTasksRequestDto,
	) *cexceptions.Exception
}

type RoutineTaskExecutionService struct {
	validator                   *validator.Validate
	db                          *gorm.DB
	routineTaskRecordRepository srepositories.RoutineTaskRecordRepositoryInterface
	subShelfHandler             handlers.SubShelfHandlerInterface
	blockPackHandler            handlers.BlockPackHandlerInterface
	routineHandler              handlers.RoutineHandlerInterface
	materialHandler             handlers.MaterialHandlerInterface
}

func NewRoutineTaskExecutionService(
	validatorInstance *validator.Validate,
	db *gorm.DB,
	yjsDocumentInitializer handlers.YjsDocumentInitializer,
	yjsBlockPackUpdater handlers.YjsBlockPackUpdater,
) RoutineTaskExecutionServiceInterface {
	if validatorInstance == nil {
		validatorInstance = validator.New()
	}

	patternResolver := resolvers.NewRoutineTaskPatternResolver(db)
	templateBlockMatcher := matchers.NewRoutineTaskTemplateMatcher()
	service := &RoutineTaskExecutionService{
		validator: validatorInstance,
		db:        db,
		routineTaskRecordRepository: srepositories.NewRoutineTaskRecordRepository(
			db,
			sscopes.NewRoutineTaskRecordScope(),
		),
		subShelfHandler: handlers.NewSubShelfHandler(
			db,
			validatorInstance,
			patternResolver,
			templateBlockMatcher,
		),
		blockPackHandler: handlers.NewBlockPackHandler(
			db,
			validatorInstance,
			patternResolver,
			templateBlockMatcher,
			yjsDocumentInitializer,
			yjsBlockPackUpdater,
		),
		routineHandler: handlers.NewRoutineHandler(
			db,
			validatorInstance,
			patternResolver,
			templateBlockMatcher,
		),
		materialHandler: handlers.NewMaterialHandler(
			db,
			validatorInstance,
			patternResolver,
			templateBlockMatcher,
		),
	}
	return service
}

func (s *RoutineTaskExecutionService) ApplyPreparedRoutineTasks(
	ctx context.Context,
	eventId uuid.UUID,
	request *cdurablejob.MarkCompletedRoutineTasksRequestDto,
) *cexceptions.Exception {
	if eventId == uuid.Nil || request == nil || len(request.Tasks) == 0 || s.db == nil {
		return cexceptions.New(
			"InvalidDto",
			"RoutineTask",
			"ApplyPreparedRoutineTasks",
			"The prepared routine task result is invalid",
			http.StatusBadRequest,
		)
	}
	if err := s.validator.Struct(request); err != nil {
		return cexceptions.New(
			"InvalidDto",
			"RoutineTask",
			"ApplyPreparedRoutineTasks",
			"The prepared routine task result is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return cexceptions.New(
			"FailedToBeginTransaction",
			"RoutineTask",
			"ApplyPreparedRoutineTasks",
			"Failed to start the routine task completion transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}

	taskIds := make([]uuid.UUID, len(request.Tasks))
	recordIds := make([]uuid.UUID, len(request.Tasks))
	routineRecordIds := make([]uuid.UUID, len(request.Tasks))
	completedTaskIndexById := make(map[uuid.UUID]int, len(request.Tasks))
	seenTaskIds := make(map[uuid.UUID]struct{}, len(request.Tasks))
	seenRecordIds := make(map[uuid.UUID]struct{}, len(request.Tasks))
	for index, task := range request.Tasks {
		if task.PreparedTask == nil {
			tx.Rollback()
			return cexceptions.New(
				"InvalidDto",
				"RoutineTask",
				"ApplyPreparedRoutineTasks",
				"Every completed task must contain a prepared payload",
				http.StatusBadRequest,
			)
		}
		if _, exists := seenTaskIds[task.RoutineTaskId]; exists {
			tx.Rollback()
			return cexceptions.New(
				"InvalidDto",
				"RoutineTask",
				"ApplyPreparedRoutineTasks",
				"The prepared routine task result contains duplicate routine task ids",
				http.StatusBadRequest,
			)
		}
		if _, exists := seenRecordIds[task.RoutineTaskRecordId]; exists {
			tx.Rollback()
			return cexceptions.New(
				"InvalidDto",
				"RoutineTaskRecord",
				"ApplyPreparedRoutineTasks",
				"The prepared routine task result contains duplicate routine task record ids",
				http.StatusBadRequest,
			)
		}
		seenTaskIds[task.RoutineTaskId] = struct{}{}
		seenRecordIds[task.RoutineTaskRecordId] = struct{}{}
		taskIds[index] = task.RoutineTaskId
		recordIds[index] = task.RoutineTaskRecordId
		routineRecordIds[index] = task.RoutineRecordId
		completedTaskIndexById[task.RoutineTaskId] = index
	}

	var storedTasks []sschemas.RoutineTask
	if err := tx.WithContext(ctx).Where("id IN ?", taskIds).Find(&storedTasks).Error; err != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToRead",
			"RoutineTask",
			"ApplyPreparedRoutineTasks",
			"Failed to read routine tasks for execution",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	storedTaskById := make(map[uuid.UUID]sschemas.RoutineTask, len(storedTasks))
	for _, task := range storedTasks {
		storedTaskById[task.Id] = task
	}
	var storedRecords []sschemas.RoutineTaskRecord
	if err := tx.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id IN ?", recordIds).
		Find(&storedRecords).Error; err != nil {
		tx.Rollback()
		return cexceptions.New("FailedToRead", "RoutineTaskRecord", "ApplyPreparedRoutineTasks", "Failed to read routine task records for execution", http.StatusInternalServerError, true).WithOrigin(err)
	}
	recordById := make(map[uuid.UUID]sschemas.RoutineTaskRecord, len(storedRecords))
	for _, record := range storedRecords {
		recordById[record.Id] = record
	}
	var routineRecords []sschemas.RoutineRecord
	if err := tx.WithContext(ctx).Where("id IN ?", routineRecordIds).Find(&routineRecords).Error; err != nil {
		tx.Rollback()
		return cexceptions.New("FailedToRead", "RoutineRecord", "ApplyPreparedRoutineTasks", "Failed to read routine records for execution", http.StatusInternalServerError, true).WithOrigin(err)
	}
	routineRecordById := make(map[uuid.UUID]sschemas.RoutineRecord, len(routineRecords))
	for _, record := range routineRecords {
		routineRecordById[record.Id] = record
	}
	groupedTasks := make(map[cenums.RoutineTaskPurpose][]sschemas.RoutineTask)
	actorsByTaskId := make(map[cenums.RoutineTaskPurpose]map[uuid.UUID]uuid.UUID)
	alreadyCompletedRecordIds := make(map[uuid.UUID]struct{})
	for _, completedTask := range request.Tasks {
		preparedTask := completedTask.PreparedTask
		storedTask, exists := storedTaskById[completedTask.RoutineTaskId]
		storedRecord, recordExists := recordById[completedTask.RoutineTaskRecordId]
		routineRecord, routineRecordExists := routineRecordById[completedTask.RoutineRecordId]
		if !exists || !recordExists || !routineRecordExists ||
			storedTask.ActorUserId != preparedTask.ActorUserId ||
			storedTask.RoutineId != routineRecord.RoutineId ||
			storedRecord.RoutineTaskId != completedTask.RoutineTaskId ||
			storedRecord.RoutineRecordId != completedTask.RoutineRecordId ||
			(storedRecord.Status != cenums.RoutineTaskRecordStatus_Running &&
				storedRecord.Status != cenums.RoutineTaskRecordStatus_Success) ||
			storedRecord.Attempts != preparedTask.Attempt {
			tx.Rollback()
			return cexceptions.New(
				"ResultStateMismatch",
				"RoutineTask",
				"ApplyPreparedRoutineTasks",
				"The prepared routine task does not match the stored task",
				http.StatusConflict,
				true,
			)
		}
		if storedRecord.Status == cenums.RoutineTaskRecordStatus_Success {
			alreadyCompletedRecordIds[storedRecord.Id] = struct{}{}
			continue
		}

		storedTask.Payload = datatypes.JSON(preparedTask.Payload)
		storedTask.RecordId = completedTask.RoutineTaskRecordId
		storedTask.RecordScheduledAt = routineRecord.ScheduledAt
		purpose := cenums.RoutineTaskPurpose(preparedTask.Purpose)
		groupedTasks[purpose] = append(groupedTasks[purpose], storedTask)
		if actorsByTaskId[purpose] == nil {
			actorsByTaskId[purpose] = make(map[uuid.UUID]uuid.UUID)
		}
		actorsByTaskId[purpose][storedTask.Id] = preparedTask.ActorUserId
	}

	for purpose, tasks := range groupedTasks {
		var (
			successes          []bool
			executionResults   map[uuid.UUID]croutinetasktypes.ExecutionResult
			exception          *cexceptions.Exception
			allowedPermissions []cenums.AccessControlPermission
		)
		switch purpose {
		case cenums.RoutineTaskPurpose_GetSubShelf:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
				cenums.AccessControlPermission_Read,
			}
			if detailedHandler, ok := s.subShelfHandler.(handlers.SubShelfDetailedExecutionHandlerInterface); ok {
				successes, executionResults, exception = detailedHandler.HandleGetSubShelfWithResults(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
			} else {
				successes, exception = s.subShelfHandler.HandleGetSubShelf(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
			}
		case cenums.RoutineTaskPurpose_CreateSubShelf:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
			}
			successes, exception = s.subShelfHandler.HandleCreateSubShelf(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_UpdateSubShelf:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
			}
			successes, exception = s.subShelfHandler.HandleUpdateSubShelf(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_DeleteSubShelf:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
			}
			successes, exception = s.subShelfHandler.HandleDeleteSubShelf(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_GetBlockPack:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
				cenums.AccessControlPermission_Read,
			}
			if detailedHandler, ok := s.blockPackHandler.(handlers.BlockPackGetDetailedExecutionHandlerInterface); ok {
				successes, executionResults, exception = detailedHandler.HandleGetBlockPackWithResults(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
			} else {
				successes, exception = s.blockPackHandler.HandleGetBlockPack(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
			}
		case cenums.RoutineTaskPurpose_CreateBlockPack:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
			}
			successes, exception = s.blockPackHandler.HandleCreateBlockPack(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_UpdateBlockPack:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
			}
			if detailedHandler, ok := s.blockPackHandler.(handlers.BlockPackDetailedExecutionHandlerInterface); ok {
				successes, executionResults, exception = detailedHandler.HandleUpdateBlockPackWithResults(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
			} else {
				successes, exception = s.blockPackHandler.HandleUpdateBlockPack(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
			}
		case cenums.RoutineTaskPurpose_DeleteBlockPack:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
			}
			successes, exception = s.blockPackHandler.HandleDeleteBlockPack(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_GetRoutine:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
				cenums.AccessControlPermission_Read,
			}
			if detailedHandler, ok := s.routineHandler.(handlers.RoutineDetailedExecutionHandlerInterface); ok {
				successes, executionResults, exception = detailedHandler.HandleGetRoutineWithResults(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
			} else {
				successes, exception = s.routineHandler.HandleGetRoutine(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
			}
		case cenums.RoutineTaskPurpose_CreateRoutine:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
			}
			successes, exception = s.routineHandler.HandleCreateRoutine(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_UpdateRoutine:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
			}
			successes, exception = s.routineHandler.HandleUpdateRoutine(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_DeleteRoutine:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
			}
			successes, exception = s.routineHandler.HandleDeleteRoutine(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_GetMaterial:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
				cenums.AccessControlPermission_Read,
			}
			if detailedHandler, ok := s.materialHandler.(handlers.MaterialDetailedExecutionHandlerInterface); ok {
				successes, executionResults, exception = detailedHandler.HandleGetMaterialWithResults(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
			} else {
				successes, exception = s.materialHandler.HandleGetMaterial(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
			}
		case cenums.RoutineTaskPurpose_CreateMaterial:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
			}
			successes, exception = s.materialHandler.HandleCreateMaterial(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_UpdateMaterial:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
			}
			successes, exception = s.materialHandler.HandleUpdateMaterial(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_DeleteMaterial:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
			}
			successes, exception = s.materialHandler.HandleDeleteMaterial(ctx, tx, tasks, actorsByTaskId[purpose], allowedPermissions)
		default:
			tx.Rollback()
			return cexceptions.New(
				"ExecutionOperationNotFound",
				"RoutineTask",
				"ApplyPreparedRoutineTasks",
				"No DurableJob execution operation is registered for the routine task purpose",
				http.StatusInternalServerError,
			)
		}
		if exception != nil {
			tx.Rollback()
			return exception
		}
		for taskId, result := range executionResults {
			completedTaskIndex, exists := completedTaskIndexById[taskId]
			if !exists {
				continue
			}
			resultCopy := result
			request.Tasks[completedTaskIndex].ExecutionResult = &resultCopy
		}
		for _, success := range successes {
			if !success {
				tx.Rollback()
				return cexceptions.New(
					"ExecutionFailed",
					"RoutineTask",
					"ApplyPreparedRoutineTasks",
					"A prepared routine task could not be applied",
					http.StatusConflict,
					true,
				)
			}
		}
	}

	now := time.Now().UTC()
	recordIds = make([]uuid.UUID, len(request.Tasks))
	for index, task := range request.Tasks {
		recordIds[index] = task.RoutineTaskRecordId
	}

	result := tx.Model(&sschemas.RoutineTaskRecord{}).
		Where("id IN ? AND status = ?", recordIds, cenums.RoutineTaskRecordStatus_Running).
		Updates(map[string]any{
			"status":          cenums.RoutineTaskRecordStatus_Success,
			"actual_ended_at": now,
			"error_code":      nil,
			"error_reason":    nil,
			"updated_at":      now,
		})
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToUpdate",
			"RoutineTaskRecord",
			"MarkCompletedRoutineTasks",
			"Failed to finalize routine task records",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	if result.RowsAffected != int64(len(recordIds)) {
		var finalizedRecordCount int64
		tx.Model(&sschemas.RoutineTaskRecord{}).
			Where("id IN ? AND status = ?", recordIds, cenums.RoutineTaskRecordStatus_Success).
			Count(&finalizedRecordCount)
		if finalizedRecordCount != int64(len(recordIds)) {
			tx.Rollback()
			return cexceptions.New(
				"ResultStateMismatch",
				"RoutineTaskRecord",
				"MarkCompletedRoutineTasks",
				"Routine task record completion count does not match the claimed batch",
				http.StatusConflict,
				true,
			)
		}
	}
	resultSnapshotPlaceholders := make([]string, 0, len(request.Tasks))
	resultSnapshotArgs := make([]any, 0, len(request.Tasks)*2)
	for _, task := range request.Tasks {
		if _, exists := alreadyCompletedRecordIds[task.RoutineTaskRecordId]; exists {
			continue
		}
		snapshot := []byte("{}")
		if task.ExecutionResult != nil {
			encodedSnapshot, err := json.Marshal(task.ExecutionResult)
			if err != nil {
				tx.Rollback()
				return cexceptions.New("FailedToEncode", "RoutineTaskRecord", "MarkCompletedRoutineTasks", "Failed to encode the routine task execution result", http.StatusInternalServerError, true).WithOrigin(err)
			}
			snapshot = encodedSnapshot
		}
		resultSnapshotPlaceholders = append(resultSnapshotPlaceholders, "(?::uuid, ?::jsonb)")
		resultSnapshotArgs = append(resultSnapshotArgs, task.RoutineTaskRecordId, snapshot)
	}
	if len(resultSnapshotPlaceholders) > 0 {
		result = tx.Exec(fmt.Sprintf(`UPDATE "RoutineTaskRecordTable" AS routine_task_record
			SET result_snapshot = results.result_snapshot, updated_at = ?
			FROM (VALUES %s) AS results(id, result_snapshot)
			WHERE routine_task_record.id = results.id`, strings.Join(resultSnapshotPlaceholders, ",")), append([]any{now}, resultSnapshotArgs...)...)
		if result.Error != nil {
			tx.Rollback()
			return cexceptions.New("FailedToUpdate", "RoutineTaskRecord", "MarkCompletedRoutineTasks", "Failed to store routine task execution results", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
	}
	result = tx.Exec(`UPDATE "RoutineTaskRecordTable" AS child
		SET status = ?, updated_at = ?
		WHERE child.status = ? AND child.routine_record_id IN ?
		AND NOT EXISTS (
			SELECT 1 FROM "RoutineDependencyTable" dependency
			INNER JOIN "RoutineTaskRecordTable" previous
				ON previous.routine_task_id = dependency.previous_routine_task_id
				AND previous.routine_record_id = child.routine_record_id
			WHERE dependency.routine_task_id = child.routine_task_id
				AND previous.status <> ?
		)`, cenums.RoutineTaskRecordStatus_Ready, now, cenums.RoutineTaskRecordStatus_Waiting, routineRecordIds, cenums.RoutineTaskRecordStatus_Success)
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New("FailedToUpdate", "RoutineTaskRecord", "MarkCompletedRoutineTasks", "Failed to release dependent routine tasks", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	result = tx.Exec(`UPDATE "RoutineRecordTable" AS routine_record
		SET success_task_count = counts.success_task_count, failed_task_count = counts.failed_task_count,
			blocked_task_count = counts.blocked_task_count, running_task_count = counts.running_task_count,
			waiting_task_count = counts.waiting_task_count,
			status = CASE
				WHEN counts.running_task_count > 0 OR counts.waiting_task_count > 0 THEN ?::"RoutineRecordStatus"
				WHEN counts.failed_task_count > 0 OR counts.blocked_task_count > 0 THEN ?::"RoutineRecordStatus"
				ELSE ?::"RoutineRecordStatus"
			END,
			actual_ended_at = CASE WHEN counts.running_task_count = 0 AND counts.waiting_task_count = 0 THEN ? ELSE routine_record.actual_ended_at END,
			updated_at = ?
		FROM (
			SELECT routine_record_id,
				COUNT(*) FILTER (WHERE status = 'Success')::integer AS success_task_count,
				COUNT(*) FILTER (WHERE status = 'Failed')::integer AS failed_task_count,
				COUNT(*) FILTER (WHERE status = 'Blocked')::integer AS blocked_task_count,
				COUNT(*) FILTER (WHERE status = 'Running')::integer AS running_task_count,
				COUNT(*) FILTER (WHERE status IN ('Waiting', 'Ready'))::integer AS waiting_task_count
			FROM "RoutineTaskRecordTable" WHERE routine_record_id IN ? GROUP BY routine_record_id
		) counts WHERE routine_record.id = counts.routine_record_id`, cenums.RoutineRecordStatus_Running, cenums.RoutineRecordStatus_Failed, cenums.RoutineRecordStatus_Success, now, now, routineRecordIds)
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New("FailedToUpdate", "RoutineRecord", "MarkCompletedRoutineTasks", "Failed to update routine record aggregates", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	result = tx.Exec(`UPDATE "RoutineTable" AS routine
		SET status = CASE WHEN routine.period IS NULL THEN ? ELSE ? END, updated_at = ?
		WHERE routine.id IN (
			SELECT routine_record.routine_id FROM "RoutineRecordTable" routine_record
			WHERE routine_record.id IN ? AND routine_record.status IN (?, ?, ?)
		)`, cenums.RoutineStatus_Completed, cenums.RoutineStatus_Scheduled, now, routineRecordIds, cenums.RoutineRecordStatus_Success, cenums.RoutineRecordStatus_Failed, cenums.RoutineRecordStatus_Blocked)
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New("FailedToUpdate", "Routine", "MarkCompletedRoutineTasks", "Failed to finalize routine schedule", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	if err := tx.Commit().Error; err != nil {
		return cexceptions.New(
			"FailedToCommitTransaction",
			"RoutineTask",
			"ApplyPreparedRoutineTasks",
			"Failed to commit the routine task completion transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return nil
}

func (s *RoutineTaskExecutionService) ApplyFailedRoutineTasks(
	ctx context.Context,
	eventId uuid.UUID,
	request *cdurablejob.MarkFailedRoutineTasksRequestDto,
) *cexceptions.Exception {
	if eventId == uuid.Nil || request == nil || len(request.Tasks) == 0 || s.db == nil {
		return cexceptions.New(
			"InvalidDto",
			"RoutineTask",
			"ApplyFailedRoutineTasks",
			"The failed routine task result is invalid",
			http.StatusBadRequest,
		)
	}
	if err := s.validator.Struct(request); err != nil {
		return cexceptions.New(
			"InvalidDto",
			"RoutineTask",
			"ApplyFailedRoutineTasks",
			"The failed routine task result is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	recordInputs := make([]sinputs.UpdateRoutineTaskRecordFailureInput, 0, len(request.Tasks))
	recordIds := make([]uuid.UUID, 0, len(request.Tasks))
	routineRecordIds := make([]uuid.UUID, 0, len(request.Tasks))
	seenTaskIds := make(map[uuid.UUID]struct{}, len(request.Tasks))
	seenRecordIds := make(map[uuid.UUID]struct{}, len(request.Tasks))
	for _, task := range request.Tasks {
		if _, exists := seenTaskIds[task.RoutineTaskId]; exists {
			return cexceptions.New("InvalidDto", "RoutineTask", "ApplyFailedRoutineTasks", "The failed routine task result contains duplicate routine task ids", http.StatusBadRequest)
		}
		if _, exists := seenRecordIds[task.RoutineTaskRecordId]; exists {
			return cexceptions.New("InvalidDto", "RoutineTaskRecord", "ApplyFailedRoutineTasks", "The failed routine task result contains duplicate routine task record ids", http.StatusBadRequest)
		}
		seenTaskIds[task.RoutineTaskId] = struct{}{}
		seenRecordIds[task.RoutineTaskRecordId] = struct{}{}
		recordInputs = append(recordInputs, sinputs.UpdateRoutineTaskRecordFailureInput{
			Id:          task.RoutineTaskRecordId,
			ErrorCode:   task.ErrorCode,
			ErrorReason: task.ErrorReason,
			FailedAt:    task.FailedAt,
		})
		recordIds = append(recordIds, task.RoutineTaskRecordId)
		routineRecordIds = append(routineRecordIds, task.RoutineRecordId)
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return cexceptions.New(
			"FailedToBeginTransaction",
			"RoutineTask",
			"ApplyFailedRoutineTasks",
			"Failed to start the routine task failure transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	var runningRecords []sschemas.RoutineTaskRecord
	if result := tx.Model(&sschemas.RoutineTaskRecord{}).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id IN ?", recordIds).
		Find(&runningRecords); result.Error != nil {
		tx.Rollback()
		return cexceptions.New("FailedToRead", "RoutineTaskRecord", "ApplyFailedRoutineTasks", "Failed to read routine task records for failure finalization", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	runningRecordById := make(map[uuid.UUID]sschemas.RoutineTaskRecord, len(runningRecords))
	for _, record := range runningRecords {
		runningRecordById[record.Id] = record
	}
	for _, task := range request.Tasks {
		record, exists := runningRecordById[task.RoutineTaskRecordId]
		if !exists || record.RoutineTaskId != task.RoutineTaskId || record.RoutineRecordId != task.RoutineRecordId ||
			(record.Status != cenums.RoutineTaskRecordStatus_Running && record.Status != cenums.RoutineTaskRecordStatus_Failed) {
			tx.Rollback()
			return cexceptions.New("ResultStateMismatch", "RoutineTaskRecord", "ApplyFailedRoutineTasks", "The failed routine task does not match the running task record", http.StatusConflict, true)
		}
	}

	updatedRecordCount, exception := s.routineTaskRecordRepository.UpdateManyAsFailed(
		recordInputs,
		srepositories.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return exception
	}
	if updatedRecordCount != int64(len(recordInputs)) {
		var failedRecordCount int64
		result := tx.Model(&sschemas.RoutineTaskRecord{}).
			Where("id IN ? AND status = ?", recordIds, cenums.RoutineTaskRecordStatus_Failed).
			Count(&failedRecordCount)
		if result.Error != nil || failedRecordCount != int64(len(recordIds)) {
			tx.Rollback()
			if result.Error != nil {
				return cexceptions.New(
					"FailedToRead",
					"RoutineTaskRecord",
					"ApplyFailedRoutineTasks",
					"Failed to verify routine task record failure finalization",
					http.StatusInternalServerError,
					true,
				).WithOrigin(result.Error)
			}
			return cexceptions.New(
				"ResultStateMismatch",
				"RoutineTaskRecord",
				"ApplyFailedRoutineTasks",
				"Routine task record failure count does not match the claimed batch",
				http.StatusConflict,
				true,
			)
		}
	}
	now := time.Now().UTC()
	result := tx.Exec(`WITH RECURSIVE blocked_tasks(routine_record_id, routine_task_id) AS (
		SELECT routine_record_id, routine_task_id
		FROM "RoutineTaskRecordTable"
		WHERE id IN ?
		UNION
		SELECT blocked_tasks.routine_record_id, dependency.routine_task_id
		FROM blocked_tasks
		INNER JOIN "RoutineDependencyTable" dependency
			ON dependency.previous_routine_task_id = blocked_tasks.routine_task_id
	)
	UPDATE "RoutineTaskRecordTable" AS routine_task_record
	SET status = ?, updated_at = ?
	FROM blocked_tasks
	WHERE routine_task_record.routine_record_id = blocked_tasks.routine_record_id
		AND routine_task_record.routine_task_id = blocked_tasks.routine_task_id
		AND routine_task_record.status IN (?, ?)`, recordIds, cenums.RoutineTaskRecordStatus_Blocked, now, cenums.RoutineTaskRecordStatus_Waiting, cenums.RoutineTaskRecordStatus_Ready)
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New("FailedToUpdate", "RoutineTaskRecord", "ApplyFailedRoutineTasks", "Failed to block dependent routine tasks", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	result = tx.Exec(`UPDATE "RoutineRecordTable" AS routine_record
		SET success_task_count = counts.success_task_count, failed_task_count = counts.failed_task_count,
			blocked_task_count = counts.blocked_task_count, running_task_count = counts.running_task_count,
			waiting_task_count = counts.waiting_task_count,
			status = CASE
				WHEN counts.running_task_count > 0 OR counts.waiting_task_count > 0 THEN ?::"RoutineRecordStatus"
				WHEN counts.failed_task_count > 0 OR counts.blocked_task_count > 0 THEN ?::"RoutineRecordStatus"
				ELSE ?::"RoutineRecordStatus"
			END,
			actual_ended_at = CASE WHEN counts.running_task_count = 0 AND counts.waiting_task_count = 0 THEN ? ELSE routine_record.actual_ended_at END,
			updated_at = ?
		FROM (
			SELECT routine_record_id,
				COUNT(*) FILTER (WHERE status = 'Success')::integer AS success_task_count,
				COUNT(*) FILTER (WHERE status = 'Failed')::integer AS failed_task_count,
				COUNT(*) FILTER (WHERE status = 'Blocked')::integer AS blocked_task_count,
				COUNT(*) FILTER (WHERE status = 'Running')::integer AS running_task_count,
				COUNT(*) FILTER (WHERE status IN ('Waiting', 'Ready'))::integer AS waiting_task_count
			FROM "RoutineTaskRecordTable" WHERE routine_record_id IN ? GROUP BY routine_record_id
		) counts WHERE routine_record.id = counts.routine_record_id`, cenums.RoutineRecordStatus_Running, cenums.RoutineRecordStatus_Failed, cenums.RoutineRecordStatus_Success, now, now, routineRecordIds)
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New("FailedToUpdate", "RoutineRecord", "ApplyFailedRoutineTasks", "Failed to update routine record aggregates", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	result = tx.Exec(`UPDATE "RoutineTable" AS routine
		SET status = CASE WHEN routine.period IS NULL THEN ? ELSE ? END, updated_at = ?
		WHERE routine.id IN (
			SELECT routine_record.routine_id FROM "RoutineRecordTable" routine_record
			WHERE routine_record.id IN ? AND routine_record.status IN (?, ?, ?)
		)`, cenums.RoutineStatus_Completed, cenums.RoutineStatus_Scheduled, now, routineRecordIds, cenums.RoutineRecordStatus_Success, cenums.RoutineRecordStatus_Failed, cenums.RoutineRecordStatus_Blocked)
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New("FailedToUpdate", "Routine", "ApplyFailedRoutineTasks", "Failed to finalize routine schedule", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}

	if err := tx.Commit().Error; err != nil {
		return cexceptions.New(
			"FailedToCommitTransaction",
			"RoutineTask",
			"ApplyFailedRoutineTasks",
			"Failed to commit the routine task failure transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return nil
}
