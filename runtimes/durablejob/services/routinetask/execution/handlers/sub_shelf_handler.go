package handlers

import (
	"context"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"

	matchers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/matchers"
	parsers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/parsers"
	resolvers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/resolvers"
)

type SubShelfHandlerInterface interface {
	HandleGetSubShelf(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleCreateSubShelf(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleUpdateSubShelf(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleDeleteSubShelf(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
}

type SubShelfDetailedExecutionHandlerInterface interface {
	HandleGetSubShelfWithResults(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, map[uuid.UUID]croutinetasktypes.ExecutionResult, *cexceptions.Exception)
}

type SubShelfHandler struct {
	Handler
	validator            *validator.Validate
	patternResolver      resolvers.RoutineTaskPatternResolverInterface
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface
	subShelfRepository   srepositories.SubShelfRepositoryInterface
	blockPackRepository  srepositories.BlockPackRepositoryInterface
	materialRepository   srepositories.MaterialRepositoryInterface
}

func NewSubShelfHandler(
	validatorInstance *validator.Validate,
	patternResolver resolvers.RoutineTaskPatternResolverInterface,
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface,
	subShelfRepository srepositories.SubShelfRepositoryInterface,
	blockPackRepository srepositories.BlockPackRepositoryInterface,
	materialRepository srepositories.MaterialRepositoryInterface,
) SubShelfHandlerInterface {
	return &SubShelfHandler{
		validator:            validatorInstance,
		patternResolver:      patternResolver,
		templateBlockMatcher: templateBlockMatcher,
		subShelfRepository:   subShelfRepository,
		blockPackRepository:  blockPackRepository,
		materialRepository:   materialRepository,
	}
}

func (s *SubShelfHandler) HandleGetSubShelf(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes, _, exception := s.HandleGetSubShelfWithResults(ctx, db, tasks, taskIdToActorUserId, allowedPermissions)
	return successes, exception
}

func (s *SubShelfHandler) HandleCreateSubShelf(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	candidateTaskIndexes := make([]int, 0, len(tasks))
	candidateTasks := make([]sschemas.RoutineTask, 0, len(tasks))
	candidateActorUserIds := make([]uuid.UUID, 0, len(tasks))
	candidatePayloads := make([]croutinetasktypes.CreateSubShelfRoutineTaskPayload, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))
	candidatePrevSubShelfIds := make([]*uuid.UUID, 0, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}

		payload, exception := parsers.DecodePayload[croutinetasktypes.CreateSubShelfRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		var prevSubShelfId *uuid.UUID
		if payload.PrevSubShelfId != nil {
			resolvedId, err := payload.PrevSubShelfId.Resolve(nil)
			if err != nil {
				continue
			}
			prevSubShelfId = &resolvedId
		}
		candidateTaskIndexes = append(candidateTaskIndexes, taskIndex)
		candidateTasks = append(candidateTasks, task)
		candidateActorUserIds = append(candidateActorUserIds, actorUserId)
		candidatePayloads = append(candidatePayloads, *payload)
		candidatePatterns = append(candidatePatterns, payload.Pattern)
		candidatePrevSubShelfIds = append(candidatePrevSubShelfIds, prevSubShelfId)
	}
	if len(candidateTasks) == 0 {
		return successes, nil
	}

	patternValuesByCandidate, patternSuccesses, exception := s.patternResolver.ResolveMany(
		ctx,
		db,
		candidateTasks,
		candidateActorUserIds,
		candidatePatterns,
		allowedPermissions,
	)
	if exception != nil {
		return successes, exception
	}

	bulkInputs := make([]sinputs.BulkCreateSubShelfInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		name := s.templateBlockMatcher.MatchString(payload.Name, patternValues)
		var path *stypes.UUIDArray
		if payload.Path != nil {
			pathValue := stypes.UUIDArray(payload.Path)
			path = &pathValue
		}
		bulkInputs = append(bulkInputs, sinputs.BulkCreateSubShelfInput{
			UserId:         candidateActorUserIds[candidateIndex],
			Id:             payload.Id,
			RootShelfId:    payload.RootShelfId,
			PrevSubShelfId: candidatePrevSubShelfIds[candidateIndex],
			Path:           path,
			Name:           name,
		})
		taskIndexes = append(taskIndexes, candidateTaskIndexes[candidateIndex])
	}
	if len(bulkInputs) == 0 {
		return successes, nil
	}

	bulkSuccesses, exception := s.subShelfRepository.BulkCreateMany(
		bulkInputs,
		srepositories.WithTransactionDB(db.WithContext(ctx)),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		return successes, exception
	}

	for index, success := range bulkSuccesses {
		successes[taskIndexes[index]] = success
	}

	return successes, nil
}

func (s *SubShelfHandler) HandleUpdateSubShelf(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	candidateTaskIndexes := make([]int, 0, len(tasks))
	candidateTasks := make([]sschemas.RoutineTask, 0, len(tasks))
	candidateActorUserIds := make([]uuid.UUID, 0, len(tasks))
	candidatePayloads := make([]croutinetasktypes.UpdateSubShelfRoutineTaskPayload, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}

		payload, exception := parsers.DecodePayload[croutinetasktypes.UpdateSubShelfRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		if payload.Name == nil {
			continue
		}
		candidateTaskIndexes = append(candidateTaskIndexes, taskIndex)
		candidateTasks = append(candidateTasks, task)
		candidateActorUserIds = append(candidateActorUserIds, actorUserId)
		candidatePayloads = append(candidatePayloads, *payload)
		candidatePatterns = append(candidatePatterns, payload.Pattern)
	}
	if len(candidateTasks) == 0 {
		return successes, nil
	}

	patternValuesByCandidate, patternSuccesses, exception := s.patternResolver.ResolveMany(
		ctx,
		db,
		candidateTasks,
		candidateActorUserIds,
		candidatePatterns,
		allowedPermissions,
	)
	if exception != nil {
		return successes, exception
	}

	bulkInputs := make([]sinputs.BulkUpdateSubShelfInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] || payload.Name == nil {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		name := s.templateBlockMatcher.MatchString(*payload.Name, patternValues)
		bulkInputs = append(bulkInputs, sinputs.BulkUpdateSubShelfInput{
			UserId: candidateActorUserIds[candidateIndex],
			Id:     payload.SubShelfId,
			PartialUpdateInput: sinputs.PartialUpdateSubShelfInput{
				Values: sinputs.UpdateSubShelfInput{
					Name: &name,
				},
			},
		})
		taskIndexes = append(taskIndexes, candidateTaskIndexes[candidateIndex])
	}

	if len(bulkInputs) == 0 {
		return successes, nil
	}
	bulkSuccesses, exception := s.subShelfRepository.BulkUpdateMany(
		bulkInputs,
		srepositories.WithTransactionDB(db.WithContext(ctx)),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		return successes, exception
	}

	for index, success := range bulkSuccesses {
		successes[taskIndexes[index]] = success
	}

	return successes, nil
}

func (s *SubShelfHandler) HandleDeleteSubShelf(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	deleteInputs := make([]sinputs.BulkDeleteSubShelfInput, 0, len(tasks))
	taskIndexes := make([]int, 0, len(tasks))
	for index, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, exception := parsers.DecodePayload[croutinetasktypes.DeleteSubShelfRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		deleteInputs = append(deleteInputs, sinputs.BulkDeleteSubShelfInput{Id: payload.SubShelfId, UserId: actorUserId})
		taskIndexes = append(taskIndexes, index)
	}
	if len(deleteInputs) == 0 {
		return successes, nil
	}
	deleteSuccesses, exception := s.subShelfRepository.BulkDeleteMany(
		deleteInputs,
		srepositories.WithDB(db.WithContext(ctx)), srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return successes, exception
	}
	for index, success := range deleteSuccesses {
		successes[taskIndexes[index]] = success
	}
	return successes, nil
}

func (s *SubShelfHandler) HandleGetSubShelfWithResults(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, map[uuid.UUID]croutinetasktypes.ExecutionResult, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	results := make(map[uuid.UUID]croutinetasktypes.ExecutionResult)
	checkInputs := make([]sinputs.BulkCheckSubShelfPermissionInput, 0, len(tasks))
	taskIndexes := make([]int, 0, len(tasks))
	taskObjectIds := make([]uuid.UUID, 0, len(tasks))
	for index, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, exception := parsers.DecodePayload[croutinetasktypes.GetSubShelfRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		checkInputs = append(checkInputs, sinputs.BulkCheckSubShelfPermissionInput{Id: payload.SubShelfId, UserId: actorUserId})
		taskIndexes = append(taskIndexes, index)
		taskObjectIds = append(taskObjectIds, payload.SubShelfId)
	}
	if len(checkInputs) == 0 {
		return successes, results, nil
	}
	checkSuccesses, objects, exception := s.subShelfRepository.BulkCheckPermissionsAndGetManyByIds(
		checkInputs, nil, allowedPermissions,
		srepositories.WithDB(db.WithContext(ctx)), srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		return successes, nil, exception
	}
	objectsById := make(map[uuid.UUID]sschemas.SubShelf, len(objects))
	for _, object := range objects {
		objectsById[object.Id] = object
	}
	for index, success := range checkSuccesses {
		taskIndex := taskIndexes[index]
		successes[taskIndex] = success
		result, resultException := s.BuildGetResult(taskObjectIds[index], success, objectsById[taskObjectIds[index]])
		if resultException != nil {
			return successes, nil, resultException
		}
		results[tasks[taskIndex].Id] = result
	}
	return successes, results, nil
}
