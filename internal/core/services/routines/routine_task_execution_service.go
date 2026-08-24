package routines

import (
	"context"
	"net/http"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	coreeventscontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	durablejobcontract "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	routinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	eventcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	cinputs "github.com/HiIamJeff67/notegic-backend/contracts/types/models/inputs"
	crepositories "github.com/HiIamJeff67/notegic-backend/contracts/types/models/repositories"

	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	coreenums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	handlers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/handlers"
	matchers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/matchers"
	parsers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/parsers"
	resolvers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/resolvers"
	durablejobeventbuilders "github.com/HiIamJeff67/notegic-backend/internal/core/transports/durablejob/eventbuilders"
)

type RoutineTaskExecutionServiceInterface interface {
	ValidateRoutineTaskPayload(
		purpose coreenums.RoutineTaskPurpose,
		payload datatypes.JSON,
	) *exceptions.Exception
	ResolveRoutineTaskPatterns(
		ctx context.Context,
		tasks []schemas.RoutineTask,
		actorUserIds []uuid.UUID,
		patterns []routinetasktypes.RoutineTaskPattern,
		allowedPermissions []coreenums.AccessControlPermission,
	) ([]map[string]string, []bool, *exceptions.Exception)
	ApplyPreparedRoutineTasks(
		ctx context.Context,
		eventId uuid.UUID,
		request *durablejobcontract.MarkCompletedRoutineTasksRequestDto,
	) *exceptions.Exception
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
	purpose coreenums.RoutineTaskPurpose,
	payload datatypes.JSON,
) *exceptions.Exception {
	return s.routineTaskHandler.HandleValidateRoutineTaskPayload(purpose, payload)
}

func (s *RoutineTaskExecutionService) ResolveRoutineTaskPatterns(
	ctx context.Context,
	tasks []schemas.RoutineTask,
	actorUserIds []uuid.UUID,
	patterns []routinetasktypes.RoutineTaskPattern,
	allowedPermissions []coreenums.AccessControlPermission,
) ([]map[string]string, []bool, *exceptions.Exception) {
	return s.patternResolver.ResolveMany(ctx, s.db, tasks, actorUserIds, patterns, allowedPermissions)
}

func (s *RoutineTaskExecutionService) ApplyPreparedRoutineTasks(
	ctx context.Context,
	eventId uuid.UUID,
	request *durablejobcontract.MarkCompletedRoutineTasksRequestDto,
) *exceptions.Exception {
	if eventId == uuid.Nil || request == nil || len(request.Tasks) == 0 || s.db == nil {
		return exceptions.New(
			"InvalidDto",
			"RoutineTask",
			"ApplyPreparedRoutineTasks",
			"The prepared routine task result is invalid",
			http.StatusBadRequest,
		)
	}
	if err := s.validator.Struct(request); err != nil {
		return exceptions.New(
			"InvalidDto",
			"RoutineTask",
			"ApplyPreparedRoutineTasks",
			"The prepared routine task result is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return exceptions.New(
			"FailedToBeginTransaction",
			"RoutineTask",
			"ApplyPreparedRoutineTasks",
			"Failed to start the routine task completion transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}

	isNewInboxEvent, inboxException := crepositories.NewInboxEventRepository().CreateOne(
		cinputs.CreateInboxEventInput{EventId: eventId},
		crepositories.RepositoryOptionFields{DB: tx, IsTransactionStarted: true},
	)
	if inboxException != nil {
		tx.Rollback()
		return exceptions.New(
			"FailedToRecordInboxEvent",
			"RoutineTask",
			"ApplyPreparedRoutineTasks",
			"Failed to record the Kafka result event",
			http.StatusInternalServerError,
			true,
		).WithOrigin(inboxException)
	}
	if !isNewInboxEvent {
		if err := tx.Commit().Error; err != nil {
			return exceptions.New(
				"FailedToCommitTransaction",
				"RoutineTask",
				"ApplyPreparedRoutineTasks",
				"Failed to commit the idempotent routine task result",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}
		return nil
	}

	if exception := s.applyPreparedRoutineTasks(ctx, tx, request); exception != nil {
		tx.Rollback()
		return exception
	}
	if exception := finalizeCompletedRoutineTasks(tx, request); exception != nil {
		tx.Rollback()
		return exception
	}
	completionEvents := make([]eventcontract.EventEnvelope[coreeventscontract.RoutineTaskCompletedData], len(request.Tasks))
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
		coreeventscontract.CoreLifecycleTopic,
		completionEvents,
	); err != nil {
		tx.Rollback()
		return exceptions.New(
			"FailedToEnqueueCompletionEvent",
			"RoutineTask",
			"ApplyPreparedRoutineTasks",
			"Failed to enqueue routine task completion events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		return exceptions.New(
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
	request *durablejobcontract.MarkCompletedRoutineTasksRequestDto,
) *exceptions.Exception {
	taskIds := make([]uuid.UUID, len(request.Tasks))
	for index, task := range request.Tasks {
		if task.PreparedTask == nil {
			return exceptions.New(
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
		return exceptions.New(
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
	groupedTasks := make(map[coreenums.RoutineTaskPurpose][]schemas.RoutineTask)
	actorsByTaskId := make(map[coreenums.RoutineTaskPurpose]map[uuid.UUID]uuid.UUID)
	for _, completedTask := range request.Tasks {
		preparedTask := completedTask.PreparedTask
		storedTask, exists := storedTaskById[completedTask.RoutineTaskId]
		if !exists || storedTask.ActorUserId != preparedTask.ActorUserId || storedTask.Attempts != preparedTask.Attempt {
			return exceptions.New(
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
		purpose := coreenums.RoutineTaskPurpose(preparedTask.Purpose)
		groupedTasks[purpose] = append(groupedTasks[purpose], storedTask)
		if actorsByTaskId[purpose] == nil {
			actorsByTaskId[purpose] = make(map[uuid.UUID]uuid.UUID)
		}
		actorsByTaskId[purpose][storedTask.Id] = preparedTask.ActorUserId
	}

	for purpose, tasks := range groupedTasks {
		var (
			successes          []bool
			exception          *exceptions.Exception
			allowedPermissions []coreenums.AccessControlPermission
		)
		switch purpose {
		case coreenums.RoutineTaskPurpose_CreateRootShelf:
			allowedPermissions = []coreenums.AccessControlPermission{
				coreenums.AccessControlPermission_Owner,
				coreenums.AccessControlPermission_Admin,
				coreenums.AccessControlPermission_Write,
				coreenums.AccessControlPermission_Read,
			}
			successes, exception = s.rootShelfHandler.HandleCreateRootShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case coreenums.RoutineTaskPurpose_UpdateRootShelf:
			allowedPermissions = []coreenums.AccessControlPermission{
				coreenums.AccessControlPermission_Owner,
				coreenums.AccessControlPermission_Admin,
			}
			successes, exception = s.rootShelfHandler.HandleUpdateRootShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case coreenums.RoutineTaskPurpose_ResetRootShelf:
			allowedPermissions = []coreenums.AccessControlPermission{
				coreenums.AccessControlPermission_Owner,
				coreenums.AccessControlPermission_Admin,
			}
			successes, exception = s.rootShelfHandler.HandleResetRootShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case coreenums.RoutineTaskPurpose_CreateSubShelf:
			allowedPermissions = []coreenums.AccessControlPermission{
				coreenums.AccessControlPermission_Owner,
				coreenums.AccessControlPermission_Admin,
			}
			successes, exception = s.subShelfHandler.HandleCreateSubShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case coreenums.RoutineTaskPurpose_UpdateSubShelf:
			allowedPermissions = []coreenums.AccessControlPermission{
				coreenums.AccessControlPermission_Owner,
				coreenums.AccessControlPermission_Admin,
			}
			successes, exception = s.subShelfHandler.HandleUpdateSubShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case coreenums.RoutineTaskPurpose_ResetSubShelf:
			allowedPermissions = []coreenums.AccessControlPermission{
				coreenums.AccessControlPermission_Owner,
				coreenums.AccessControlPermission_Admin,
			}
			successes, exception = s.subShelfHandler.HandleResetSubShelf(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case coreenums.RoutineTaskPurpose_CreateBlockPack:
			allowedPermissions = []coreenums.AccessControlPermission{
				coreenums.AccessControlPermission_Owner,
				coreenums.AccessControlPermission_Admin,
				coreenums.AccessControlPermission_Write,
			}
			successes, exception = s.blockPackHandler.HandleCreateBlockPack(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case coreenums.RoutineTaskPurpose_UpdateBlockPack:
			allowedPermissions = []coreenums.AccessControlPermission{
				coreenums.AccessControlPermission_Owner,
				coreenums.AccessControlPermission_Admin,
				coreenums.AccessControlPermission_Write,
			}
			successes, exception = s.blockPackHandler.HandleUpdateBlockPack(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case coreenums.RoutineTaskPurpose_ResetBlockPack:
			allowedPermissions = []coreenums.AccessControlPermission{
				coreenums.AccessControlPermission_Owner,
				coreenums.AccessControlPermission_Admin,
				coreenums.AccessControlPermission_Write,
			}
			successes, exception = s.blockPackHandler.HandleResetBlockPack(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case coreenums.RoutineTaskPurpose_CreateRoutine:
			allowedPermissions = []coreenums.AccessControlPermission{
				coreenums.AccessControlPermission_Owner,
				coreenums.AccessControlPermission_Admin,
				coreenums.AccessControlPermission_Write,
			}
			successes, exception = s.routineHandler.HandleCreateRoutine(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		case coreenums.RoutineTaskPurpose_UpdateRoutine:
			allowedPermissions = []coreenums.AccessControlPermission{
				coreenums.AccessControlPermission_Owner,
				coreenums.AccessControlPermission_Admin,
				coreenums.AccessControlPermission_Write,
			}
			successes, exception = s.routineHandler.HandleUpdateRoutine(ctx, db, tasks, actorsByTaskId[purpose], allowedPermissions)
		default:
			return exceptions.New(
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
				return exceptions.New(
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
	request *durablejobcontract.MarkCompletedRoutineTasksRequestDto,
) *exceptions.Exception {
	now := time.Now().UTC()
	taskIds := make([]uuid.UUID, len(request.Tasks))
	recordIds := make([]uuid.UUID, len(request.Tasks))
	for index, task := range request.Tasks {
		taskIds[index] = task.RoutineTaskId
		recordIds[index] = task.RoutineTaskRecordId
	}

	result := tx.Model(&schemas.RoutineTask{}).
		Where("id IN ? AND status = ?", taskIds, coreenums.RoutineTaskStatus_Running).
		Updates(map[string]any{
			"status":          coreenums.RoutineTaskStatus_Idle,
			"attempts":        0,
			"actual_ended_at": now,
			"updated_at":      now,
		})
	if result.Error != nil {
		return exceptions.New(
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
			Where("id IN ? AND status = ?", taskIds, coreenums.RoutineTaskStatus_Idle).
			Count(&finalizedTaskCount)
		if finalizedTaskCount != int64(len(taskIds)) {
			return exceptions.New(
				"ResultStateMismatch",
				"RoutineTask",
				"MarkCompletedRoutineTasks",
				"Routine task completion count does not match the claimed batch",
				http.StatusConflict,
				true,
			)
		}
	}

	result = tx.Model(&schemas.RoutineTaskRecord{}).
		Where("id IN ? AND status = ?", recordIds, coreenums.RoutineTaskRecordStatus_Running).
		Updates(map[string]any{
			"status":          coreenums.RoutineTaskRecordStatus_Success,
			"actual_ended_at": now,
			"error_code":      nil,
			"error_reason":    nil,
			"updated_at":      now,
		})
	if result.Error != nil {
		return exceptions.New(
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
		tx.Model(&schemas.RoutineTaskRecord{}).
			Where("id IN ? AND status = ?", recordIds, coreenums.RoutineTaskRecordStatus_Success).
			Count(&finalizedRecordCount)
		if finalizedRecordCount != int64(len(recordIds)) {
			return exceptions.New(
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
