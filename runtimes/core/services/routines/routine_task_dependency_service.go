package routines

import (
	"context"
	"fmt"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/routine-task-dependencies"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/routine-task-dependencies"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/core/exceptions"
)

type RoutineTaskDependencyServiceInterface interface {
	GetRoutineTaskDependenciesByRoutineId(
		ctx context.Context,
		request *capi.GetRoutineTaskDependenciesByRoutineIdRequestDto,
	) (*capi.GetRoutineTaskDependenciesByRoutineIdResponseDto, *cexceptions.Exception)
	CreateRoutineTaskDependencyByRoutineId(
		ctx context.Context,
		request *capi.CreateRoutineTaskDependencyByRoutineIdRequestDto,
	) (*capi.CreateRoutineTaskDependencyByRoutineIdResponseDto, *cexceptions.Exception)
	CreateRoutineTaskDependenciesByRoutineId(
		ctx context.Context,
		request *capi.CreateRoutineTaskDependenciesByRoutineIdRequestDto,
	) (*capi.CreateRoutineTaskDependenciesByRoutineIdResponseDto, *cexceptions.Exception)
	UpdateRoutineTaskDependencyByRoutineId(
		ctx context.Context,
		request *capi.UpdateRoutineTaskDependencyByRoutineIdRequestDto,
	) (*capi.UpdateRoutineTaskDependencyByRoutineIdResponseDto, *cexceptions.Exception)
	UpdateRoutineTaskDependenciesByRoutineId(
		ctx context.Context,
		request *capi.UpdateRoutineTaskDependenciesByRoutineIdRequestDto,
	) (*capi.UpdateRoutineTaskDependenciesByRoutineIdResponseDto, *cexceptions.Exception)
	DeleteRoutineTaskDependencyByRoutineId(
		ctx context.Context,
		request *capi.DeleteRoutineTaskDependencyByRoutineIdRequestDto,
	) (*capi.DeleteRoutineTaskDependencyByRoutineIdResponseDto, *cexceptions.Exception)
	DeleteRoutineTaskDependenciesByRoutineId(
		ctx context.Context,
		request *capi.DeleteRoutineTaskDependenciesByRoutineIdRequestDto,
	) (*capi.DeleteRoutineTaskDependenciesByRoutineIdResponseDto, *cexceptions.Exception)
}

type RoutineTaskDependencyService struct {
	validator                       *validator.Validate
	db                              *gorm.DB
	routineRepository               srepositories.RoutineRepositoryInterface
	routineTaskRepository           srepositories.RoutineTaskRepositoryInterface
	routineTaskDependencyRepository srepositories.RoutineTaskDependencyRepositoryInterface
}

func NewRoutineTaskDependencyService(
	validator *validator.Validate,
	db *gorm.DB,
	routineRepository srepositories.RoutineRepositoryInterface,
	routineTaskRepository srepositories.RoutineTaskRepositoryInterface,
	routineTaskDependencyRepository srepositories.RoutineTaskDependencyRepositoryInterface,
) RoutineTaskDependencyServiceInterface {
	return &RoutineTaskDependencyService{
		validator:                       validator,
		db:                              db,
		routineRepository:               routineRepository,
		routineTaskRepository:           routineTaskRepository,
		routineTaskDependencyRepository: routineTaskDependencyRepository,
	}
}

/* ============================== Auxiliary Functions ============================== */

func validateRoutineTaskDependencyBatch(
	routineTasks []sschemas.RoutineTask,
	inputs []coretypes.CreatableRoutineTaskDependency,
	dependencies []sschemas.RoutineTaskDependency,
) *cexceptions.Exception {
	taskIds := make(map[uuid.UUID]struct{}, len(routineTasks))
	graph := make(map[uuid.UUID][]uuid.UUID, len(routineTasks))
	for _, routineTask := range routineTasks {
		taskIds[routineTask.Id] = struct{}{}
		for _, previousTask := range routineTask.PreviousTasks {
			graph[routineTask.Id] = append(graph[routineTask.Id], previousTask.Id)
		}
	}
	for _, dependency := range dependencies {
		graph[dependency.RoutineTaskId] = append(graph[dependency.RoutineTaskId], dependency.PreviousRoutineTaskId)
	}
	knownDependencies := make(map[sinputs.RoutineTaskDependencyKey]struct{}, len(dependencies))
	for _, dependency := range dependencies {
		knownDependencies[sinputs.RoutineTaskDependencyKey{
			RoutineTaskId:         dependency.RoutineTaskId,
			PreviousRoutineTaskId: dependency.PreviousRoutineTaskId,
		}] = struct{}{}
	}
	for _, input := range inputs {
		if _, exists := taskIds[input.RoutineTaskId]; !exists {
			return apiexceptions.NewRoutineTaskDependencyException().
				InvalidInput().
				WithOrigin(fmt.Errorf("routine task does not belong to the routine"))
		}
		if _, exists := taskIds[input.PreviousRoutineTaskId]; !exists {
			return apiexceptions.NewRoutineTaskDependencyException().
				InvalidInput().
				WithOrigin(fmt.Errorf("previous routine task does not belong to the routine"))
		}
		if input.RoutineTaskId == input.PreviousRoutineTaskId {
			return apiexceptions.NewRoutineTaskDependencyException().
				InvalidInput().
				WithOrigin(fmt.Errorf("a routine task cannot depend on itself"))
		}
		key := sinputs.RoutineTaskDependencyKey{
			RoutineTaskId:         input.RoutineTaskId,
			PreviousRoutineTaskId: input.PreviousRoutineTaskId,
		}
		if _, exists := knownDependencies[key]; exists {
			return apiexceptions.NewRoutineTaskDependencyException().DependencyAlreadyExists()
		}
		knownDependencies[key] = struct{}{}
		graph[input.RoutineTaskId] = append(graph[input.RoutineTaskId], input.PreviousRoutineTaskId)
	}
	visitState := make(map[uuid.UUID]uint8, len(taskIds))
	var visit func(uuid.UUID) bool
	visit = func(taskId uuid.UUID) bool {
		switch visitState[taskId] {
		case 1:
			return true
		case 2:
			return false
		}
		visitState[taskId] = 1
		for _, previousTaskId := range graph[taskId] {
			if visit(previousTaskId) {
				return true
			}
		}
		visitState[taskId] = 2
		return false
	}
	for taskId := range taskIds {
		if visit(taskId) {
			return apiexceptions.NewRoutineTaskDependencyException().
				InvalidInput().
				WithOrigin(fmt.Errorf("routine task dependencies cannot contain a cycle"))
		}
	}
	return nil
}

/* ============================== Service Methods for Routine Task Dependency ============================== */

func (s *RoutineTaskDependencyService) GetRoutineTaskDependenciesByRoutineId(
	ctx context.Context,
	request *capi.GetRoutineTaskDependenciesByRoutineIdRequestDto,
) (*capi.GetRoutineTaskDependenciesByRoutineIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(request); err != nil {
		return nil, apiexceptions.NewRoutineTaskDependencyException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	db := s.db.WithContext(ctx)
	if _, exception := s.routineRepository.CheckPermissionAndGetOneById(
		request.Param.RoutineId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	); exception != nil {
		return nil, exception
	}
	dependencies, exception := s.routineTaskDependencyRepository.GetAllByRoutineId(
		request.Param.RoutineId,
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}
	result := make(capi.GetRoutineTaskDependenciesByRoutineIdResponseDto, len(dependencies))
	for index, dependency := range dependencies {
		result[index] = coretypes.RoutineTaskDependency{
			RoutineTaskId:         dependency.RoutineTaskId,
			PreviousRoutineTaskId: dependency.PreviousRoutineTaskId,
			Description:           dependency.Description,
			Progress:              dependency.Progress,
			UpdatedAt:             dependency.UpdatedAt,
			CreatedAt:             dependency.CreatedAt,
		}
	}
	return &result, nil
}

func (s *RoutineTaskDependencyService) CreateRoutineTaskDependencyByRoutineId(
	ctx context.Context,
	request *capi.CreateRoutineTaskDependencyByRoutineIdRequestDto,
) (*capi.CreateRoutineTaskDependencyByRoutineIdResponseDto, *cexceptions.Exception) {
	result, exception := s.createRoutineTaskDependencies(
		ctx,
		request.Param.RoutineId,
		[]coretypes.CreatableRoutineTaskDependency{request.Body},
		request,
	)
	if exception != nil {
		return nil, exception
	}
	response := capi.CreateRoutineTaskDependencyByRoutineIdResponseDto(result[0])
	return &response, nil
}

func (s *RoutineTaskDependencyService) CreateRoutineTaskDependenciesByRoutineId(
	ctx context.Context,
	request *capi.CreateRoutineTaskDependenciesByRoutineIdRequestDto,
) (*capi.CreateRoutineTaskDependenciesByRoutineIdResponseDto, *cexceptions.Exception) {
	result, exception := s.createRoutineTaskDependencies(
		ctx,
		request.Param.RoutineId,
		request.Body.Dependencies,
		request,
	)
	if exception != nil {
		return nil, exception
	}
	response := capi.CreateRoutineTaskDependenciesByRoutineIdResponseDto(result)
	return &response, nil
}

func (s *RoutineTaskDependencyService) createRoutineTaskDependencies(
	ctx context.Context,
	routineId uuid.UUID,
	inputs []coretypes.CreatableRoutineTaskDependency,
	request any,
) ([]coretypes.RoutineTaskDependency, *cexceptions.Exception) {
	if err := s.validator.Struct(request); err != nil {
		return nil, apiexceptions.NewRoutineTaskDependencyException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	tx := s.db.WithContext(ctx).Begin()
	if _, exception := s.routineRepository.CheckPermissionAndGetOneById(
		routineId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	routineTasks, exception := s.routineTaskRepository.GetAllByRoutineIds(
		[]uuid.UUID{routineId},
		actorUserId,
		[]sschemas.RoutineTaskRelation{sschemas.RoutineTaskRelation_PreviousTasks},
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	dependencies, exception := s.routineTaskDependencyRepository.GetAllByRoutineId(
		routineId,
		srepositories.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if exception := validateRoutineTaskDependencyBatch(routineTasks, inputs, dependencies); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	createInputs := make([]sinputs.CreateRoutineTaskDependencyInput, len(inputs))
	for index, input := range inputs {
		createInputs[index] = sinputs.CreateRoutineTaskDependencyInput{
			RoutineTaskId:         input.RoutineTaskId,
			PreviousRoutineTaskId: input.PreviousRoutineTaskId,
			Description:           input.Description,
			Progress:              input.Progress,
		}
	}
	if exception := s.routineTaskDependencyRepository.CreateManyByRoutineId(
		routineId,
		createInputs,
		srepositories.WithTransactionDB(tx),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	dependencies, exception = s.routineTaskDependencyRepository.GetAllByRoutineId(
		routineId,
		srepositories.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	keys := make([]sinputs.RoutineTaskDependencyKey, len(inputs))
	for index, input := range inputs {
		keys[index] = sinputs.RoutineTaskDependencyKey{
			RoutineTaskId:         input.RoutineTaskId,
			PreviousRoutineTaskId: input.PreviousRoutineTaskId,
		}
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineTaskDependencyException().FailedToCreate().WithOrigin(err)
	}
	byKey := make(map[sinputs.RoutineTaskDependencyKey]sschemas.RoutineTaskDependency, len(dependencies))
	for _, dependency := range dependencies {
		byKey[sinputs.RoutineTaskDependencyKey{
			RoutineTaskId:         dependency.RoutineTaskId,
			PreviousRoutineTaskId: dependency.PreviousRoutineTaskId,
		}] = dependency
	}
	result := make([]coretypes.RoutineTaskDependency, 0, len(keys))
	for _, key := range keys {
		if dependency, exists := byKey[key]; exists {
			result = append(result, coretypes.RoutineTaskDependency{
				RoutineTaskId:         dependency.RoutineTaskId,
				PreviousRoutineTaskId: dependency.PreviousRoutineTaskId,
				Description:           dependency.Description,
				Progress:              dependency.Progress,
				UpdatedAt:             dependency.UpdatedAt,
				CreatedAt:             dependency.CreatedAt,
			})
		}
	}
	return result, nil
}

func (s *RoutineTaskDependencyService) UpdateRoutineTaskDependencyByRoutineId(
	ctx context.Context,
	request *capi.UpdateRoutineTaskDependencyByRoutineIdRequestDto,
) (*capi.UpdateRoutineTaskDependencyByRoutineIdResponseDto, *cexceptions.Exception) {
	result, exception := s.updateRoutineTaskDependencies(
		ctx,
		request.Param.RoutineId,
		[]coretypes.UpdatableRoutineTaskDependency{request.Body},
		request,
	)
	if exception != nil {
		return nil, exception
	}
	response := capi.UpdateRoutineTaskDependencyByRoutineIdResponseDto(result[0])
	return &response, nil
}

func (s *RoutineTaskDependencyService) UpdateRoutineTaskDependenciesByRoutineId(
	ctx context.Context,
	request *capi.UpdateRoutineTaskDependenciesByRoutineIdRequestDto,
) (*capi.UpdateRoutineTaskDependenciesByRoutineIdResponseDto, *cexceptions.Exception) {
	result, exception := s.updateRoutineTaskDependencies(
		ctx,
		request.Param.RoutineId,
		request.Body.Dependencies,
		request,
	)
	if exception != nil {
		return nil, exception
	}
	response := capi.UpdateRoutineTaskDependenciesByRoutineIdResponseDto(result)
	return &response, nil
}

func (s *RoutineTaskDependencyService) updateRoutineTaskDependencies(
	ctx context.Context,
	routineId uuid.UUID,
	inputs []coretypes.UpdatableRoutineTaskDependency,
	request any,
) ([]coretypes.RoutineTaskDependency, *cexceptions.Exception) {
	if err := s.validator.Struct(request); err != nil {
		return nil, apiexceptions.NewRoutineTaskDependencyException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	tx := s.db.WithContext(ctx).Begin()
	if _, exception := s.routineRepository.CheckPermissionAndGetOneById(
		routineId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	_, exception = s.routineTaskRepository.GetAllByRoutineIds(
		[]uuid.UUID{routineId},
		actorUserId,
		[]sschemas.RoutineTaskRelation{sschemas.RoutineTaskRelation_PreviousTasks},
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	dependencies, exception := s.routineTaskDependencyRepository.GetAllByRoutineId(
		routineId,
		srepositories.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	keys := make([]sinputs.RoutineTaskDependencyKey, len(inputs))
	updates := make([]sinputs.UpdateRoutineTaskDependencyInput, len(inputs))
	knownDependencies := make(map[sinputs.RoutineTaskDependencyKey]sschemas.RoutineTaskDependency, len(dependencies))
	for _, dependency := range dependencies {
		knownDependencies[sinputs.RoutineTaskDependencyKey{
			RoutineTaskId:         dependency.RoutineTaskId,
			PreviousRoutineTaskId: dependency.PreviousRoutineTaskId,
		}] = dependency
	}
	seenKeys := make(map[sinputs.RoutineTaskDependencyKey]struct{}, len(inputs))
	for index, input := range inputs {
		key := sinputs.RoutineTaskDependencyKey{
			RoutineTaskId:         input.RoutineTaskId,
			PreviousRoutineTaskId: input.PreviousRoutineTaskId,
		}
		if _, exists := seenKeys[key]; exists {
			tx.Rollback()
			return nil, apiexceptions.NewRoutineTaskDependencyException().DependencyAlreadyExists()
		}
		seenKeys[key] = struct{}{}
		dependency, exists := knownDependencies[key]
		if !exists {
			tx.Rollback()
			return nil, apiexceptions.NewRoutineTaskDependencyException().NotFound()
		}
		if input.Description == nil && input.Progress == nil {
			tx.Rollback()
			return nil, apiexceptions.NewRoutineTaskDependencyException().
				InvalidInput().
				WithOrigin(fmt.Errorf("at least one dependency field must be updated"))
		}
		if input.Description != nil {
			dependency.Description = *input.Description
		}
		if input.Progress != nil {
			dependency.Progress = *input.Progress
		}
		keys[index] = key
		updates[index] = sinputs.UpdateRoutineTaskDependencyInput{
			RoutineTaskId:         dependency.RoutineTaskId,
			PreviousRoutineTaskId: dependency.PreviousRoutineTaskId,
			Description:           dependency.Description,
			Progress:              dependency.Progress,
		}
	}
	if exception := s.routineTaskDependencyRepository.UpdateManyByRoutineId(
		routineId,
		updates,
		srepositories.WithTransactionDB(tx),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	dependencies, exception = s.routineTaskDependencyRepository.GetAllByRoutineId(
		routineId,
		srepositories.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineTaskDependencyException().FailedToUpdate().WithOrigin(err)
	}
	byKey := make(map[sinputs.RoutineTaskDependencyKey]sschemas.RoutineTaskDependency, len(dependencies))
	for _, dependency := range dependencies {
		byKey[sinputs.RoutineTaskDependencyKey{
			RoutineTaskId:         dependency.RoutineTaskId,
			PreviousRoutineTaskId: dependency.PreviousRoutineTaskId,
		}] = dependency
	}
	result := make([]coretypes.RoutineTaskDependency, 0, len(keys))
	for _, key := range keys {
		if dependency, exists := byKey[key]; exists {
			result = append(result, coretypes.RoutineTaskDependency{
				RoutineTaskId:         dependency.RoutineTaskId,
				PreviousRoutineTaskId: dependency.PreviousRoutineTaskId,
				Description:           dependency.Description,
				Progress:              dependency.Progress,
				UpdatedAt:             dependency.UpdatedAt,
				CreatedAt:             dependency.CreatedAt,
			})
		}
	}
	return result, nil
}

func (s *RoutineTaskDependencyService) DeleteRoutineTaskDependencyByRoutineId(
	ctx context.Context,
	request *capi.DeleteRoutineTaskDependencyByRoutineIdRequestDto,
) (*capi.DeleteRoutineTaskDependencyByRoutineIdResponseDto, *cexceptions.Exception) {
	return s.deleteRoutineTaskDependencies(
		ctx,
		request.Param.RoutineId,
		[]coretypes.DeletableRoutineTaskDependency{request.Body},
		request,
	)
}

func (s *RoutineTaskDependencyService) DeleteRoutineTaskDependenciesByRoutineId(
	ctx context.Context,
	request *capi.DeleteRoutineTaskDependenciesByRoutineIdRequestDto,
) (*capi.DeleteRoutineTaskDependenciesByRoutineIdResponseDto, *cexceptions.Exception) {
	return s.deleteRoutineTaskDependencies(
		ctx,
		request.Param.RoutineId,
		request.Body.Dependencies,
		request,
	)
}

func (s *RoutineTaskDependencyService) deleteRoutineTaskDependencies(
	ctx context.Context,
	routineId uuid.UUID,
	inputs []coretypes.DeletableRoutineTaskDependency,
	request any,
) (*capi.DeleteRoutineTaskDependenciesByRoutineIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(request); err != nil {
		return nil, apiexceptions.NewRoutineTaskDependencyException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	tx := s.db.WithContext(ctx).Begin()
	if _, exception := s.routineRepository.CheckPermissionAndGetOneById(
		routineId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	dependencies, exception := s.routineTaskDependencyRepository.GetAllByRoutineId(
		routineId,
		srepositories.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	keys := make([]sinputs.RoutineTaskDependencyKey, len(inputs))
	for index, input := range inputs {
		keys[index] = sinputs.RoutineTaskDependencyKey{
			RoutineTaskId:         input.RoutineTaskId,
			PreviousRoutineTaskId: input.PreviousRoutineTaskId,
		}
	}
	knownKeys := make(map[sinputs.RoutineTaskDependencyKey]struct{}, len(dependencies))
	for _, dependency := range dependencies {
		knownKeys[sinputs.RoutineTaskDependencyKey{
			RoutineTaskId:         dependency.RoutineTaskId,
			PreviousRoutineTaskId: dependency.PreviousRoutineTaskId,
		}] = struct{}{}
	}
	seenKeys := make(map[sinputs.RoutineTaskDependencyKey]struct{}, len(keys))
	for _, key := range keys {
		if key.RoutineTaskId == uuid.Nil || key.PreviousRoutineTaskId == uuid.Nil {
			tx.Rollback()
			return nil, apiexceptions.NewRoutineTaskDependencyException().
				InvalidInput().
				WithOrigin(fmt.Errorf("routine task dependency ids are required"))
		}
		if _, exists := knownKeys[key]; !exists {
			tx.Rollback()
			return nil, apiexceptions.NewRoutineTaskDependencyException().NotFound()
		}
		if _, exists := seenKeys[key]; exists {
			tx.Rollback()
			return nil, apiexceptions.NewRoutineTaskDependencyException().
				InvalidInput().
				WithOrigin(fmt.Errorf("duplicate routine task dependency"))
		}
		seenKeys[key] = struct{}{}
	}
	deletedCount, exception := s.routineTaskDependencyRepository.DeleteManyByRoutineId(
		routineId,
		keys,
		srepositories.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewRoutineTaskDependencyException().FailedToDelete().WithOrigin(err)
	}
	return &capi.DeleteRoutineTaskDependenciesByRoutineIdResponseDto{
		DeletedCount: deletedCount,
	}, nil
}
