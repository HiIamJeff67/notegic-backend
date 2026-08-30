package routines

import (
	"context"
	"net/http"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	durablejobeventbuilders "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/routinetask/execution/eventbuilders"
	handlers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/routinetask/execution/handlers"
	matchers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/routinetask/execution/matchers"
	parsers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/routinetask/execution/parsers"
	resolvers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/routinetask/execution/resolvers"
)

type RoutineTaskExecutionServiceInterface interface {
	ValidateRoutineTaskPayload(
		purpose cenums.RoutineTaskPurpose,
		payload datatypes.JSON,
	) *cexceptions.Exception
	ResolveRoutineTaskPatterns(
		ctx context.Context,
		tasks []sschemas.RoutineTask,
		actorUserIds []uuid.UUID,
		patterns []croutinetasktypes.RoutineTaskPattern,
		allowedPermissions []cenums.AccessControlPermission,
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
	purpose cenums.RoutineTaskPurpose,
	payload datatypes.JSON,
) *cexceptions.Exception {
	return s.routineTaskHandler.HandleValidateRoutineTaskPayload(purpose, payload)
}

func (s *RoutineTaskExecutionService) ResolveRoutineTaskPatterns(
	ctx context.Context,
	tasks []sschemas.RoutineTask,
	actorUserIds []uuid.UUID,
	patterns []croutinetasktypes.RoutineTaskPattern,
	allowedPermissions []cenums.AccessControlPermission,
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
	if err := srepositories.EnqueueOutboxEvents(
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

	var storedTasks []sschemas.RoutineTask
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

	storedTaskById := make(map[uuid.UUID]sschemas.RoutineTask, len(storedTasks))
	for _, task := range storedTasks {
		storedTaskById[task.Id] = task
	}
	groupedTasks := make(map[cenums.RoutineTaskPurpose][]sschemas.RoutineTask)
	actorsByTaskId := make(map[cenums.RoutineTaskPurpose]map[uuid.UUID]uuid.UUID)
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
			exception          *cexceptions.Exception
			allowedPermissions []cenums.AccessControlPermission
		)
		switch purpose {
		case cenums.RoutineTaskPurpose_CreateRootShelf:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
				cenums.AccessControlPermission_Read,
			}
			successes, exception = s.rootShelfHandler.HandleCreateRootShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_UpdateRootShelf:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
			}
			successes, exception = s.rootShelfHandler.HandleUpdateRootShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_ResetRootShelf:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
			}
			successes, exception = s.rootShelfHandler.HandleResetRootShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_CreateSubShelf:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
			}
			successes, exception = s.subShelfHandler.HandleCreateSubShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_UpdateSubShelf:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
			}
			successes, exception = s.subShelfHandler.HandleUpdateSubShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_ResetSubShelf:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
			}
			successes, exception = s.subShelfHandler.HandleResetSubShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_CreateBlockPack:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
			}
			successes, exception = s.blockPackHandler.HandleCreateBlockPack(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_UpdateBlockPack:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
			}
			successes, exception = s.blockPackHandler.HandleUpdateBlockPack(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_ResetBlockPack:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
			}
			successes, exception = s.blockPackHandler.HandleResetBlockPack(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_CreateRoutine:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
			}
			successes, exception = s.routineHandler.HandleCreateRoutine(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case cenums.RoutineTaskPurpose_UpdateRoutine:
			allowedPermissions = []cenums.AccessControlPermission{
				cenums.AccessControlPermission_Owner,
				cenums.AccessControlPermission_Admin,
				cenums.AccessControlPermission_Write,
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

	result := tx.Model(&sschemas.RoutineTask{}).
		Where("id IN ? AND status = ?", taskIds, cenums.RoutineTaskStatus_Running).
		Updates(map[string]any{
			"status":          cenums.RoutineTaskStatus_Idle,
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
		tx.Model(&sschemas.RoutineTask{}).
			Where("id IN ? AND status = ?", taskIds, cenums.RoutineTaskStatus_Idle).
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

	result = tx.Model(&sschemas.RoutineTaskRecord{}).
		Where("id IN ? AND status = ?", recordIds, cenums.RoutineTaskRecordStatus_Running).
		Updates(map[string]any{
			"status":          cenums.RoutineTaskRecordStatus_Success,
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
		tx.Model(&sschemas.RoutineTaskRecord{}).
			Where("id IN ? AND status = ?", recordIds, cenums.RoutineTaskRecordStatus_Success).
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
