package routinetask

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

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"

	routinetasksql "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/sqls/routinetask"
	handlers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/handlers"
	matchers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/matchers"
	resolvers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/resolvers"
)

type RoutineTaskExecutionServiceInterface interface {
	ApplyResult(
		ctx context.Context,
		eventId uuid.UUID,
		result croutinetasktypes.Result,
	) *cexceptions.Exception
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

func (s *RoutineTaskExecutionService) ApplyResult(
	ctx context.Context,
	eventId uuid.UUID,
	result croutinetasktypes.Result,
) *cexceptions.Exception {
	switch result.Kind {
	case croutinetasktypes.ResultKind_Completed:
		request, ok := result.Data.(cdurablejob.MarkCompletedRoutineTasksRequestDto)
		if !ok {
			return cexceptions.New(
				"InvalidDto",
				"RoutineTask",
				"ApplyResult",
				fmt.Sprintf("The completed routine task result payload is invalid: %T", result.Data),
				http.StatusBadRequest,
			)
		}
		return s.ApplyPreparedRoutineTasks(ctx, eventId, &request)
	case croutinetasktypes.ResultKind_Failed:
		request, ok := result.Data.(cdurablejob.MarkFailedRoutineTasksRequestDto)
		if !ok {
			return cexceptions.New(
				"InvalidDto",
				"RoutineTask",
				"ApplyResult",
				fmt.Sprintf("The failed routine task result payload is invalid: %T", result.Data),
				http.StatusBadRequest,
			)
		}
		return s.ApplyFailedRoutineTasks(ctx, eventId, &request)
	default:
		return cexceptions.New(
			"InvalidDto",
			"RoutineTask",
			"ApplyResult",
			fmt.Sprintf("The routine task result kind %q is unsupported", result.Kind),
			http.StatusBadRequest,
		)
	}
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
	lockingStrength := srepositories.LockingStrengthUpdate
	var storedRecords []sschemas.RoutineTaskRecord
	if err := tx.WithContext(ctx).
		Scopes(sscopes.Locking(&lockingStrength)).
		Where("id IN ?", recordIds).
		Find(&storedRecords).Error; err != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToRead",
			"RoutineTaskRecord",
			"ApplyPreparedRoutineTasks",
			"Failed to read routine task records for execution",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	recordById := make(map[uuid.UUID]sschemas.RoutineTaskRecord, len(storedRecords))
	for _, record := range storedRecords {
		recordById[record.Id] = record
	}
	var routineRecords []sschemas.RoutineRecord
	if err := tx.WithContext(ctx).Where("id IN ?", routineRecordIds).Find(&routineRecords).Error; err != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToRead",
			"RoutineRecord",
			"ApplyPreparedRoutineTasks",
			"Failed to read routine records for execution",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	routineRecordById := make(map[uuid.UUID]sschemas.RoutineRecord, len(routineRecords))
	routineTaskPlanByRecordId := make(map[uuid.UUID]*croutinetasktypes.RoutineTaskPlan, len(routineRecords))
	for _, record := range routineRecords {
		routineRecordById[record.Id] = record
		if len(record.Snapshot) == 0 || string(record.Snapshot) == "{}" {
			continue
		}
		var snapshot struct {
			RoutineTaskPlan *croutinetasktypes.RoutineTaskPlan `json:"routineTaskPlan"`
		}
		if err := json.Unmarshal(record.Snapshot, &snapshot); err != nil {
			tx.Rollback()
			return cexceptions.New(
				"InvalidRoutinePlan",
				"RoutineRecord",
				"ApplyPreparedRoutineTasks",
				"The routine record snapshot contains an invalid routine task plan",
				http.StatusConflict,
			).WithOrigin(err)
		}
		routineTaskPlanByRecordId[record.Id] = snapshot.RoutineTaskPlan
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

		payload := preparedTask.Payload
		plan := routineTaskPlanByRecordId[completedTask.RoutineRecordId]
		facts := map[string]uuid.UUID(nil)
		if plan != nil {
			facts = plan.Facts
		}
		switch preparedTask.Purpose {
		case cenums.RoutineTaskPurpose_CreateSubShelf:
			if plan == nil {
				tx.Rollback()
				return cexceptions.New(
					"InvalidRoutinePlan",
					"Routine",
					"ApplyPreparedRoutineTasks",
					"The routine task does not have a persisted deterministic plan",
					http.StatusConflict,
				)
			}
			var createPayload croutinetasktypes.CreateSubShelfRoutineTaskPayload
			if err := json.Unmarshal(payload, &createPayload); err != nil {
				tx.Rollback()
				return cexceptions.New(
					"InvalidRoutinePlan",
					"Routine",
					"ApplyPreparedRoutineTasks",
					"The create sub shelf payload is invalid",
					http.StatusConflict,
				).WithOrigin(err)
			}
			precreatedSubShelf, exists := plan.PrecreatedSubShelves[string(createPayload.FakeId)]
			if !exists {
				tx.Rollback()
				return cexceptions.New(
					"InvalidRoutinePlan",
					"Routine",
					"ApplyPreparedRoutineTasks",
					"The create sub shelf fake id is not in the persisted plan",
					http.StatusConflict,
				)
			}
			createPayload.Id = &precreatedSubShelf.RealId
			if createPayload.PrevSubShelfId != nil {
				resolvedId, err := createPayload.PrevSubShelfId.Resolve(facts)
				if err != nil {
					tx.Rollback()
					return cexceptions.New(
						"InvalidRoutinePlan",
						"Routine",
						"ApplyPreparedRoutineTasks",
						"The create sub shelf parent cannot be resolved",
						http.StatusConflict,
					).WithOrigin(err)
				}
				resolvedReference := croutinetasktypes.RoutineTaskObjectReference(resolvedId.String())
				createPayload.PrevSubShelfId = &resolvedReference
			}
			if precreatedSubShelf.Path != nil {
				createPayload.Path = append([]uuid.UUID{}, precreatedSubShelf.Path...)
			}
			var err error
			payload, err = json.Marshal(createPayload)
			if err != nil {
				tx.Rollback()
				return cexceptions.New(
					"InvalidRoutinePlan",
					"Routine",
					"ApplyPreparedRoutineTasks",
					"The create sub shelf payload cannot be normalized",
					http.StatusConflict,
				).WithOrigin(err)
			}
		case cenums.RoutineTaskPurpose_CreateBlockPack:
			var createPayload croutinetasktypes.CreateBlockPackRoutineTaskPayload
			if err := json.Unmarshal(payload, &createPayload); err != nil {
				tx.Rollback()
				return cexceptions.New(
					"InvalidRoutinePlan",
					"Routine",
					"ApplyPreparedRoutineTasks",
					"The create block pack payload is invalid",
					http.StatusConflict,
				).WithOrigin(err)
			}
			resolvedId, err := createPayload.TargetSubShelfId.Resolve(facts)
			if err != nil {
				tx.Rollback()
				return cexceptions.New(
					"InvalidRoutinePlan",
					"Routine",
					"ApplyPreparedRoutineTasks",
					"The create block pack parent cannot be resolved",
					http.StatusConflict,
				).WithOrigin(err)
			}
			createPayload.TargetSubShelfId = croutinetasktypes.RoutineTaskObjectReference(resolvedId.String())
			if plan != nil {
				if plannedId, exists := plan.PlannedObjectIds[completedTask.RoutineTaskId.String()]; exists {
					createPayload.Id = &plannedId
				}
			}
			payload, err = json.Marshal(createPayload)
			if err != nil {
				tx.Rollback()
				return cexceptions.New(
					"InvalidRoutinePlan",
					"Routine",
					"ApplyPreparedRoutineTasks",
					"The create block pack payload cannot be normalized",
					http.StatusConflict,
				).WithOrigin(err)
			}
		case cenums.RoutineTaskPurpose_CreateMaterial:
			var createPayload croutinetasktypes.CreateMaterialRoutineTaskPayload
			if err := json.Unmarshal(payload, &createPayload); err != nil {
				tx.Rollback()
				return cexceptions.New(
					"InvalidRoutinePlan",
					"Routine",
					"ApplyPreparedRoutineTasks",
					"The create material payload is invalid",
					http.StatusConflict,
				).WithOrigin(err)
			}
			resolvedId, err := createPayload.ParentSubShelfId.Resolve(facts)
			if err != nil {
				tx.Rollback()
				return cexceptions.New(
					"InvalidRoutinePlan",
					"Routine",
					"ApplyPreparedRoutineTasks",
					"The create material parent cannot be resolved",
					http.StatusConflict,
				).WithOrigin(err)
			}
			createPayload.ParentSubShelfId = croutinetasktypes.RoutineTaskObjectReference(resolvedId.String())
			if plan != nil {
				if plannedId, exists := plan.PlannedObjectIds[completedTask.RoutineTaskId.String()]; exists {
					createPayload.Id = &plannedId
				}
			}
			payload, err = json.Marshal(createPayload)
			if err != nil {
				tx.Rollback()
				return cexceptions.New(
					"InvalidRoutinePlan",
					"Routine",
					"ApplyPreparedRoutineTasks",
					"The create material payload cannot be normalized",
					http.StatusConflict,
				).WithOrigin(err)
			}
		}
		storedTask.Payload = datatypes.JSON(payload)
		storedTask.RecordId = completedTask.RoutineTaskRecordId
		storedTask.RecordScheduledAt = routineRecord.ScheduledAt
		purpose := cenums.RoutineTaskPurpose(preparedTask.Purpose)
		groupedTasks[purpose] = append(groupedTasks[purpose], storedTask)
		if actorsByTaskId[purpose] == nil {
			actorsByTaskId[purpose] = make(map[uuid.UUID]uuid.UUID)
		}
		actorsByTaskId[purpose][storedTask.Id] = preparedTask.ActorUserId
	}

	purposeOrder := []cenums.RoutineTaskPurpose{
		cenums.RoutineTaskPurpose_GetSubShelf,
		cenums.RoutineTaskPurpose_CreateSubShelf,
		cenums.RoutineTaskPurpose_UpdateSubShelf,
		cenums.RoutineTaskPurpose_DeleteSubShelf,
		cenums.RoutineTaskPurpose_GetBlockPack,
		cenums.RoutineTaskPurpose_CreateBlockPack,
		cenums.RoutineTaskPurpose_UpdateBlockPack,
		cenums.RoutineTaskPurpose_DeleteBlockPack,
		cenums.RoutineTaskPurpose_GetRoutine,
		cenums.RoutineTaskPurpose_CreateRoutine,
		cenums.RoutineTaskPurpose_UpdateRoutine,
		cenums.RoutineTaskPurpose_DeleteRoutine,
		cenums.RoutineTaskPurpose_GetMaterial,
		cenums.RoutineTaskPurpose_CreateMaterial,
		cenums.RoutineTaskPurpose_UpdateMaterial,
		cenums.RoutineTaskPurpose_DeleteMaterial,
	}
	for _, purpose := range purposeOrder {
		tasks, exists := groupedTasks[purpose]
		if !exists {
			continue
		}
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
				successes, executionResults, exception = detailedHandler.HandleGetSubShelfWithResults(
					ctx,
					tx,
					tasks,
					actorsByTaskId[purpose],
					allowedPermissions,
				)
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
				successes, executionResults, exception = detailedHandler.HandleGetBlockPackWithResults(
					ctx,
					tx,
					tasks,
					actorsByTaskId[purpose],
					allowedPermissions,
				)
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
				successes, executionResults, exception = detailedHandler.HandleUpdateBlockPackWithResults(
					ctx,
					tx,
					tasks,
					actorsByTaskId[purpose],
					allowedPermissions,
				)
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
				successes, executionResults, exception = detailedHandler.HandleGetRoutineWithResults(
					ctx,
					tx,
					tasks,
					actorsByTaskId[purpose],
					allowedPermissions,
				)
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
				successes, executionResults, exception = detailedHandler.HandleGetMaterialWithResults(
					ctx,
					tx,
					tasks,
					actorsByTaskId[purpose],
					allowedPermissions,
				)
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
				return cexceptions.New(
					"FailedToEncode",
					"RoutineTaskRecord",
					"MarkCompletedRoutineTasks",
					"Failed to encode the routine task execution result",
					http.StatusInternalServerError,
					true,
				).WithOrigin(err)
			}
			snapshot = encodedSnapshot
		}
		resultSnapshotPlaceholders = append(resultSnapshotPlaceholders, "(?::uuid, ?::jsonb)")
		resultSnapshotArgs = append(resultSnapshotArgs, task.RoutineTaskRecordId, snapshot)
	}
	if len(resultSnapshotPlaceholders) > 0 {
		query := fmt.Sprintf(
			routinetasksql.UpdateRoutineTaskRecordResultSnapshotSQL,
			strings.Join(resultSnapshotPlaceholders, ","),
		)
		args := append([]any{now}, resultSnapshotArgs...)
		result = tx.Exec(query, args...)
		if result.Error != nil {
			tx.Rollback()
			return cexceptions.New(
				"FailedToUpdate",
				"RoutineTaskRecord",
				"MarkCompletedRoutineTasks",
				"Failed to store routine task execution results",
				http.StatusInternalServerError,
				true,
			).WithOrigin(result.Error)
		}
	}
	result = tx.Model(&sschemas.RoutineTaskRecord{}).
		Where("status = ? AND routine_record_id IN ?", cenums.RoutineTaskRecordStatus_Waiting, routineRecordIds).
		Where(`NOT EXISTS (
			SELECT 1
			FROM "RoutineDependencyTable" dependency
			INNER JOIN "RoutineTaskRecordTable" previous
				ON previous.routine_task_id = dependency.previous_routine_task_id
				AND previous.routine_record_id = "RoutineTaskRecordTable".routine_record_id
			WHERE dependency.routine_task_id = "RoutineTaskRecordTable".routine_task_id
				AND previous.status <> ?
		)`, cenums.RoutineTaskRecordStatus_Success).
		Updates(map[string]any{
			"status":     cenums.RoutineTaskRecordStatus_Ready,
			"updated_at": now,
		})
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToUpdate",
			"RoutineTaskRecord",
			"MarkCompletedRoutineTasks",
			"Failed to release dependent routine tasks",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	result = tx.Exec(
		routinetasksql.UpdateRoutineRecordAggregateSQL,
		cenums.RoutineRecordStatus_Running,
		cenums.RoutineRecordStatus_Blocked,
		cenums.RoutineRecordStatus_Failed,
		cenums.RoutineRecordStatus_Success,
		now,
		now,
		routineRecordIds,
	)
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToUpdate",
			"RoutineRecord",
			"MarkCompletedRoutineTasks",
			"Failed to update routine record aggregates",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	routineIdsToFinalize := tx.Model(&sschemas.RoutineRecord{}).
		Select("routine_id").
		Where("id IN ? AND status IN ?", routineRecordIds, []cenums.RoutineRecordStatus{
			cenums.RoutineRecordStatus_Success,
			cenums.RoutineRecordStatus_Failed,
			cenums.RoutineRecordStatus_Blocked,
		})
	result = tx.Model(&sschemas.Routine{}).
		Where("id IN (?)", routineIdsToFinalize).
		Updates(map[string]any{
			"status":     gorm.Expr("CASE WHEN period IS NULL THEN ? ELSE ? END", cenums.RoutineStatus_Completed, cenums.RoutineStatus_Scheduled),
			"updated_at": now,
		})
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToUpdate",
			"Routine",
			"MarkCompletedRoutineTasks",
			"Failed to finalize routine schedule",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
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
			return cexceptions.New(
				"InvalidDto",
				"RoutineTask",
				"ApplyFailedRoutineTasks",
				"The failed routine task result contains duplicate routine task ids",
				http.StatusBadRequest,
			)
		}
		if _, exists := seenRecordIds[task.RoutineTaskRecordId]; exists {
			return cexceptions.New(
				"InvalidDto",
				"RoutineTaskRecord",
				"ApplyFailedRoutineTasks",
				"The failed routine task result contains duplicate routine task record ids",
				http.StatusBadRequest,
			)
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
	lockingStrength := srepositories.LockingStrengthUpdate
	var runningRecords []sschemas.RoutineTaskRecord
	if result := tx.Model(&sschemas.RoutineTaskRecord{}).
		Scopes(sscopes.Locking(&lockingStrength)).
		Where("id IN ?", recordIds).
		Find(&runningRecords); result.Error != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToRead",
			"RoutineTaskRecord",
			"ApplyFailedRoutineTasks",
			"Failed to read routine task records for failure finalization",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
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
			return cexceptions.New(
				"ResultStateMismatch",
				"RoutineTaskRecord",
				"ApplyFailedRoutineTasks",
				"The failed routine task does not match the running task record",
				http.StatusConflict,
				true,
			)
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
	result := tx.Exec(
		routinetasksql.BlockRoutineTaskRecordDependenciesSQL,
		recordIds,
		cenums.RoutineTaskRecordStatus_Blocked,
		now,
		cenums.RoutineTaskRecordStatus_Waiting,
		cenums.RoutineTaskRecordStatus_Ready,
	)
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToUpdate",
			"RoutineTaskRecord",
			"ApplyFailedRoutineTasks",
			"Failed to block dependent routine tasks",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	routineRecordIdsQuery := tx.Model(&sschemas.RoutineTaskRecord{}).
		Select("routine_record_id").
		Where("id IN ?", recordIds)
	result = tx.Model(&sschemas.RoutineTaskRecord{}).
		Where("routine_record_id IN (?)", routineRecordIdsQuery).
		Where("status IN ?", []cenums.RoutineTaskRecordStatus{
			cenums.RoutineTaskRecordStatus_Waiting,
			cenums.RoutineTaskRecordStatus_Ready,
		}).
		Where(`EXISTS (
			SELECT 1
			FROM "RoutineTaskRecordTable" barrier_record
			INNER JOIN "RoutineTaskTable" barrier_task
				ON barrier_task.id = barrier_record.routine_task_id
				WHERE barrier_record.routine_record_id = "RoutineTaskRecordTable".routine_record_id
					AND barrier_task.purpose IN (?, ?, ?)
					AND barrier_record.status IN (?, ?)
		)`,
			cenums.RoutineTaskPurpose_CreateSubShelf,
			cenums.RoutineTaskPurpose_CreateBlockPack,
			cenums.RoutineTaskPurpose_CreateMaterial,
			cenums.RoutineTaskRecordStatus_Failed,
			cenums.RoutineTaskRecordStatus_Blocked,
		).
		Updates(map[string]any{
			"status":     cenums.RoutineTaskRecordStatus_Blocked,
			"updated_at": now,
		})
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToUpdate",
			"RoutineTaskRecord",
			"ApplyFailedRoutineTasks",
			"Failed to block tasks behind deterministic creation failure",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	result = tx.Exec(
		routinetasksql.UpdateRoutineRecordAggregateSQL,
		cenums.RoutineRecordStatus_Running,
		cenums.RoutineRecordStatus_Blocked,
		cenums.RoutineRecordStatus_Failed,
		cenums.RoutineRecordStatus_Success,
		now,
		now,
		routineRecordIds,
	)
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToUpdate",
			"RoutineRecord",
			"ApplyFailedRoutineTasks",
			"Failed to update routine record aggregates",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	routineIdsToFinalize := tx.Model(&sschemas.RoutineRecord{}).
		Select("routine_id").
		Where("id IN ? AND status IN ?", routineRecordIds, []cenums.RoutineRecordStatus{
			cenums.RoutineRecordStatus_Success,
			cenums.RoutineRecordStatus_Failed,
			cenums.RoutineRecordStatus_Blocked,
		})
	result = tx.Model(&sschemas.Routine{}).
		Where("id IN (?)", routineIdsToFinalize).
		Updates(map[string]any{
			"status":     gorm.Expr("CASE WHEN period IS NULL THEN ? ELSE ? END", cenums.RoutineStatus_Completed, cenums.RoutineStatus_Scheduled),
			"updated_at": now,
		})
	if result.Error != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToUpdate",
			"Routine",
			"ApplyFailedRoutineTasks",
			"Failed to finalize routine schedule",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
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
