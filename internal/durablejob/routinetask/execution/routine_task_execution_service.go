package routines

import (
	"context"
	"net/http"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	crepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	platformschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	durablejobeventbuilders "github.com/HiIamJeff67/notegic-backend/internal/durablejob/routinetask/execution/eventbuilders"
	handlers "github.com/HiIamJeff67/notegic-backend/internal/durablejob/routinetask/execution/handlers"
	matchers "github.com/HiIamJeff67/notegic-backend/internal/durablejob/routinetask/execution/matchers"
	parsers "github.com/HiIamJeff67/notegic-backend/internal/durablejob/routinetask/execution/parsers"
	resolvers "github.com/HiIamJeff67/notegic-backend/internal/durablejob/routinetask/execution/resolvers"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type RoutineTaskExecutionServiceInterface interface {
	ValidateRoutineTaskPayload(
		purpose enums.RoutineTaskPurpose,
		payload datatypes.JSON,
	) *cexceptions.Exception
	ResolveRoutineTaskPatterns(
		ctx context.Context,
		tasks []schemas.RoutineTask,
		actorUserIds []uuid.UUID,
		patterns []croutinetasktypes.RoutineTaskPattern,
		allowedPermissions []enums.AccessControlPermission,
	) ([]map[string]string, []bool, *cexceptions.Exception)
	ApplyPreparedRoutineTasks(
		ctx context.Context,
		eventId uuid.UUID,
		request *cdurablejob.MarkCompletedRoutineTasksRequestDto,
	) *cexceptions.Exception
}

type RoutineTaskExecutionService struct {
	validator          *validator.Validate
	db                 *gorm.DB
	patternResolver    resolvers.RoutineTaskPatternResolverInterface
	routineTaskHandler handlers.RoutineTaskHandlerInterface
	rootShelfHandler   handlers.RootShelfHandlerInterface
	subShelfHandler    handlers.SubShelfHandlerInterface
	blockPackHandler   handlers.BlockPackHandlerInterface
	routineHandler     handlers.RoutineHandlerInterface
}

func NewRoutineTaskExecutionService(
	validatorInstance *validator.Validate,
	db *gorm.DB,
	yjsDocumentInitializer handlers.YjsDocumentInitializer,
) RoutineTaskExecutionServiceInterface {
	if validatorInstance == nil {
		validatorInstance = validator.New()
	}

	patternResolver := resolvers.NewRoutineTaskPatternResolver(db)
	templateBlockMatcher := matchers.NewRoutineTaskTemplateMatcher()
	service := &RoutineTaskExecutionService{
		validator:       validatorInstance,
		db:              db,
		patternResolver: patternResolver,
		routineTaskHandler: handlers.NewRoutineTaskHandler(
			parsers.NewRoutineTaskPayloadParser(validatorInstance),
		),
		rootShelfHandler: handlers.NewRootShelfHandler(
			db,
			validatorInstance,
			patternResolver,
			templateBlockMatcher,
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
			yjsDocumentInitializer,
			patternResolver,
			templateBlockMatcher,
		),
		routineHandler: handlers.NewRoutineHandler(
			db,
			validatorInstance,
			patternResolver,
			templateBlockMatcher,
		),
	}
	return service
}

/* ============================== Service Methods for RoutineTaskExecution ============================== */

func (s *RoutineTaskExecutionService) ValidateRoutineTaskPayload(
	purpose enums.RoutineTaskPurpose,
	payload datatypes.JSON,
) *cexceptions.Exception {
	return s.routineTaskHandler.HandleValidateRoutineTaskPayload(purpose, payload)
}

func (s *RoutineTaskExecutionService) ResolveRoutineTaskPatterns(
	ctx context.Context,
	tasks []schemas.RoutineTask,
	actorUserIds []uuid.UUID,
	patterns []croutinetasktypes.RoutineTaskPattern,
	allowedPermissions []enums.AccessControlPermission,
) ([]map[string]string, []bool, *cexceptions.Exception) {
	return s.patternResolver.ResolveMany(ctx, s.db, tasks, actorUserIds, patterns, allowedPermissions)
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

	if exception := s.applyPreparedRoutineTasks(ctx, tx, request); exception != nil {
		tx.Rollback()
		return exception
	}
	if exception := finalizeCompletedRoutineTasks(tx, request); exception != nil {
		tx.Rollback()
		return exception
	}
	completionEvents := make([]cevent.EventEnvelope[coreevents.RoutineTaskCompletedData], len(request.Tasks))
	completionEventBuilder := durablejobeventbuilders.NewRoutineTaskCompletionEventBuilder()
	for index, completedTask := range request.Tasks {
		completionEvents[index] = completionEventBuilder.Build(
			completedTask,
			request.WorkerId,
			time.Now().UTC(),
		)
	}
	if err := crepositories.EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		completionEvents,
	); err != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToEnqueueCompletionEvent",
			"RoutineTask",
			"ApplyPreparedRoutineTasks",
			"Failed to enqueue routine task completion events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
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

func (s *RoutineTaskExecutionService) applyPreparedRoutineTasks(
	ctx context.Context,
	db *gorm.DB,
	request *cdurablejob.MarkCompletedRoutineTasksRequestDto,
) *cexceptions.Exception {
	taskIds := make([]uuid.UUID, len(request.Tasks))
	for index, task := range request.Tasks {
		if task.PreparedTask == nil {
			return cexceptions.New(
				"InvalidDto",
				"RoutineTask",
				"ApplyPreparedRoutineTasks",
				"Every completed task must contain a prepared payload",
				http.StatusBadRequest,
			)
		}
		taskIds[index] = task.RoutineTaskId
	}

	var storedTasks []schemas.RoutineTask
	if err := db.WithContext(ctx).Where("id IN ?", taskIds).Find(&storedTasks).Error; err != nil {
		return cexceptions.New(
			"FailedToRead",
			"RoutineTask",
			"ApplyPreparedRoutineTasks",
			"Failed to read routine tasks for execution",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	storedTaskById := make(map[uuid.UUID]schemas.RoutineTask, len(storedTasks))
	for _, task := range storedTasks {
		storedTaskById[task.Id] = task
	}
	groupedTasks := make(map[enums.RoutineTaskPurpose][]schemas.RoutineTask)
	actorsByTaskId := make(map[enums.RoutineTaskPurpose]map[uuid.UUID]uuid.UUID)
	for _, completedTask := range request.Tasks {
		preparedTask := completedTask.PreparedTask
		storedTask, exists := storedTaskById[completedTask.RoutineTaskId]
		if !exists || storedTask.ActorUserId != preparedTask.ActorUserId || storedTask.Attempts != preparedTask.Attempt {
			return cexceptions.New(
				"ResultStateMismatch",
				"RoutineTask",
				"ApplyPreparedRoutineTasks",
				"The prepared routine task does not match the stored task",
				http.StatusConflict,
				true,
			)
		}

		storedTask.Payload = datatypes.JSON(preparedTask.Payload)
		storedTask.RecordId = completedTask.RoutineTaskRecordId
		storedTask.RecordScheduledAt = storedTask.ScheduledAt
		storedTask.ActualStartedAt = &completedTask.CompletedAt
		purpose := enums.RoutineTaskPurpose(preparedTask.Purpose)
		groupedTasks[purpose] = append(groupedTasks[purpose], storedTask)
		if actorsByTaskId[purpose] == nil {
			actorsByTaskId[purpose] = make(map[uuid.UUID]uuid.UUID)
		}
		actorsByTaskId[purpose][storedTask.Id] = preparedTask.ActorUserId
	}

	for purpose, tasks := range groupedTasks {
		var (
			successes          []bool
			exception          *cexceptions.Exception
			allowedPermissions []enums.AccessControlPermission
		)
		switch purpose {
		case enums.RoutineTaskPurpose_CreateRootShelf:
			allowedPermissions = []enums.AccessControlPermission{
				enums.AccessControlPermission_Owner,
				enums.AccessControlPermission_Admin,
				enums.AccessControlPermission_Write,
				enums.AccessControlPermission_Read,
			}
			successes, exception = s.rootShelfHandler.HandleCreateRootShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case enums.RoutineTaskPurpose_UpdateRootShelf:
			allowedPermissions = []enums.AccessControlPermission{
				enums.AccessControlPermission_Owner,
				enums.AccessControlPermission_Admin,
			}
			successes, exception = s.rootShelfHandler.HandleUpdateRootShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case enums.RoutineTaskPurpose_ResetRootShelf:
			allowedPermissions = []enums.AccessControlPermission{
				enums.AccessControlPermission_Owner,
				enums.AccessControlPermission_Admin,
			}
			successes, exception = s.rootShelfHandler.HandleResetRootShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case enums.RoutineTaskPurpose_CreateSubShelf:
			allowedPermissions = []enums.AccessControlPermission{
				enums.AccessControlPermission_Owner,
				enums.AccessControlPermission_Admin,
			}
			successes, exception = s.subShelfHandler.HandleCreateSubShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case enums.RoutineTaskPurpose_UpdateSubShelf:
			allowedPermissions = []enums.AccessControlPermission{
				enums.AccessControlPermission_Owner,
				enums.AccessControlPermission_Admin,
			}
			successes, exception = s.subShelfHandler.HandleUpdateSubShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case enums.RoutineTaskPurpose_ResetSubShelf:
			allowedPermissions = []enums.AccessControlPermission{
				enums.AccessControlPermission_Owner,
				enums.AccessControlPermission_Admin,
			}
			successes, exception = s.subShelfHandler.HandleResetSubShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case enums.RoutineTaskPurpose_CreateBlockPack:
			allowedPermissions = []enums.AccessControlPermission{
				enums.AccessControlPermission_Owner,
				enums.AccessControlPermission_Admin,
				enums.AccessControlPermission_Write,
			}
			successes, exception = s.blockPackHandler.HandleCreateBlockPack(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case enums.RoutineTaskPurpose_UpdateBlockPack:
			allowedPermissions = []enums.AccessControlPermission{
				enums.AccessControlPermission_Owner,
				enums.AccessControlPermission_Admin,
				enums.AccessControlPermission_Write,
			}
			successes, exception = s.blockPackHandler.HandleUpdateBlockPack(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case enums.RoutineTaskPurpose_ResetBlockPack:
			allowedPermissions = []enums.AccessControlPermission{
				enums.AccessControlPermission_Owner,
				enums.AccessControlPermission_Admin,
				enums.AccessControlPermission_Write,
			}
			successes, exception = s.blockPackHandler.HandleResetBlockPack(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case enums.RoutineTaskPurpose_CreateRoutine:
			allowedPermissions = []enums.AccessControlPermission{
				enums.AccessControlPermission_Owner,
				enums.AccessControlPermission_Admin,
				enums.AccessControlPermission_Write,
			}
			successes, exception = s.routineHandler.HandleCreateRoutine(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case enums.RoutineTaskPurpose_UpdateRoutine:
			allowedPermissions = []enums.AccessControlPermission{
				enums.AccessControlPermission_Owner,
				enums.AccessControlPermission_Admin,
				enums.AccessControlPermission_Write,
			}
			successes, exception = s.routineHandler.HandleUpdateRoutine(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		default:
			return cexceptions.New(
				"ExecutionOperationNotFound",
				"RoutineTask",
				"ApplyPreparedRoutineTasks",
				"No Core execution operation is registered for the routine task purpose",
				http.StatusInternalServerError,
			)
		}
		if exception != nil {
			return exception
		}
		for _, success := range successes {
			if !success {
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

	return nil
}

func finalizeCompletedRoutineTasks(
	tx *gorm.DB,
	request *cdurablejob.MarkCompletedRoutineTasksRequestDto,
) *cexceptions.Exception {
	now := time.Now().UTC()
	taskIds := make([]uuid.UUID, len(request.Tasks))
	recordIds := make([]uuid.UUID, len(request.Tasks))
	for index, task := range request.Tasks {
		taskIds[index] = task.RoutineTaskId
		recordIds[index] = task.RoutineTaskRecordId
	}

	result := tx.Model(&schemas.RoutineTask{}).
		Where("id IN ? AND status = ?", taskIds, enums.RoutineTaskStatus_Running).
		Updates(map[string]any{
			"status":          enums.RoutineTaskStatus_Idle,
			"attempts":        0,
			"actual_ended_at": now,
			"updated_at":      now,
		})
	if result.Error != nil {
		return cexceptions.New(
			"FailedToUpdate",
			"RoutineTask",
			"MarkCompletedRoutineTasks",
			"Failed to finalize routine tasks",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	if result.RowsAffected != int64(len(taskIds)) {
		var finalizedTaskCount int64
		tx.Model(&schemas.RoutineTask{}).
			Where("id IN ? AND status = ?", taskIds, enums.RoutineTaskStatus_Idle).
			Count(&finalizedTaskCount)
		if finalizedTaskCount != int64(len(taskIds)) {
			return cexceptions.New(
				"ResultStateMismatch",
				"RoutineTask",
				"MarkCompletedRoutineTasks",
				"Routine task completion count does not match the claimed batch",
				http.StatusConflict,
				true,
			)
		}
	}

	result = tx.Model(&platformschemas.RoutineTaskRecord{}).
		Where("id IN ? AND status = ?", recordIds, enums.RoutineTaskRecordStatus_Running).
		Updates(map[string]any{
			"status":          enums.RoutineTaskRecordStatus_Success,
			"actual_ended_at": now,
			"error_code":      nil,
			"error_reason":    nil,
			"updated_at":      now,
		})
	if result.Error != nil {
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
		tx.Model(&platformschemas.RoutineTaskRecord{}).
			Where("id IN ? AND status = ?", recordIds, enums.RoutineTaskRecordStatus_Success).
			Count(&finalizedRecordCount)
		if finalizedRecordCount != int64(len(recordIds)) {
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

	return nil
}
