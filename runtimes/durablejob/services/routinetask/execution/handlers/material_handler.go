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

type MaterialHandlerInterface interface {
	HandleGetMaterial(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleCreateMaterial(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleUpdateMaterial(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleDeleteMaterial(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
}

type MaterialDetailedExecutionHandlerInterface interface {
	HandleGetMaterialWithResults(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, map[uuid.UUID]croutinetasktypes.ExecutionResult, *cexceptions.Exception)
}

type MaterialHandler struct {
	Handler
	validator          *validator.Validate
	patternResolver    resolvers.RoutineTaskPatternResolverInterface
	templateMatcher    matchers.RoutineTaskTemplateMatcherInterface
	materialRepository srepositories.MaterialRepositoryInterface
}

func NewMaterialHandler(
	validatorInstance *validator.Validate,
	patternResolver resolvers.RoutineTaskPatternResolverInterface,
	templateMatcher matchers.RoutineTaskTemplateMatcherInterface,
	materialRepository srepositories.MaterialRepositoryInterface,
) MaterialHandlerInterface {
	return &MaterialHandler{
		validator:          validatorInstance,
		patternResolver:    patternResolver,
		templateMatcher:    templateMatcher,
		materialRepository: materialRepository,
	}
}

func (s *MaterialHandler) HandleGetMaterial(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes, _, exception := s.HandleGetMaterialWithResults(ctx, db, tasks, taskIdToActorUserId, allowedPermissions)
	return successes, exception
}

func (s *MaterialHandler) HandleCreateMaterial(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	candidateTasks := make([]sschemas.RoutineTask, 0, len(tasks))
	candidateActors := make([]uuid.UUID, 0, len(tasks))
	candidatePayloads := make([]croutinetasktypes.CreateMaterialRoutineTaskPayload, 0, len(tasks))
	candidateIndexes := make([]int, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))

	for index, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, exception := parsers.DecodePayload[croutinetasktypes.CreateMaterialRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		if _, err := payload.ParentSubShelfId.Resolve(nil); err != nil {
			continue
		}
		candidateIndexes = append(candidateIndexes, index)
		candidateTasks = append(candidateTasks, task)
		candidateActors = append(candidateActors, actorUserId)
		candidatePayloads = append(candidatePayloads, *payload)
		candidatePatterns = append(candidatePatterns, payload.Pattern)
	}
	if len(candidateTasks) == 0 {
		return successes, nil
	}

	patternValues, patternSuccesses, exception := s.patternResolver.ResolveMany(
		ctx, db, candidateTasks, candidateActors, candidatePatterns, allowedPermissions,
	)
	if exception != nil {
		return successes, exception
	}

	bulkInputs := make([]sinputs.BulkCreateMaterialInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for index, payload := range candidatePayloads {
		if !patternSuccesses[index] {
			continue
		}
		parentSubShelfId, err := payload.ParentSubShelfId.Resolve(nil)
		if err != nil {
			continue
		}
		id := uuid.New()
		if payload.Id != nil {
			id = *payload.Id
		}
		contentType := cenums.MaterialContentType_None
		if payload.ContentType != nil {
			contentType = *payload.ContentType
		}
		values := sinputs.CreateMaterialInput{
			Id:             id,
			Name:           s.templateMatcher.MatchString(payload.Name, patternValues[index]),
			ContentKey:     s.templateMatcher.MatchString(payload.ContentKey, patternValues[index]),
			ContentType:    contentType,
			ParseMediaType: s.templateMatcher.MatchString(payload.ParseMediaType, patternValues[index]),
		}
		bulkInputs = append(bulkInputs, sinputs.BulkCreateMaterialInput{
			UserId:           candidateActors[index],
			Id:               &values.Id,
			ParentSubShelfId: parentSubShelfId,
			Name:             values.Name,
			Size:             values.Size,
			ContentKey:       values.ContentKey,
			ContentType:      values.ContentType,
			ParseMediaType:   values.ParseMediaType,
		})
		taskIndexes = append(taskIndexes, candidateIndexes[index])
	}
	if len(bulkInputs) == 0 {
		return successes, nil
	}

	bulkSuccesses, exception := s.materialRepository.BulkCreateMany(
		bulkInputs,
		srepositories.WithDB(db.WithContext(ctx)),
		srepositories.WithTransactionDB(db.WithContext(ctx)),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return successes, exception
	}
	for index, success := range bulkSuccesses {
		successes[taskIndexes[index]] = success
	}
	return successes, nil
}

func (s *MaterialHandler) HandleUpdateMaterial(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	candidateTasks := make([]sschemas.RoutineTask, 0, len(tasks))
	candidateActors := make([]uuid.UUID, 0, len(tasks))
	candidatePayloads := make([]croutinetasktypes.UpdateMaterialRoutineTaskPayload, 0, len(tasks))
	candidateIndexes := make([]int, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))

	for index, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, exception := parsers.DecodePayload[croutinetasktypes.UpdateMaterialRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		candidateIndexes = append(candidateIndexes, index)
		candidateTasks = append(candidateTasks, task)
		candidateActors = append(candidateActors, actorUserId)
		candidatePayloads = append(candidatePayloads, *payload)
		candidatePatterns = append(candidatePatterns, payload.Pattern)
	}
	if len(candidateTasks) == 0 {
		return successes, nil
	}

	patternValues, patternSuccesses, exception := s.patternResolver.ResolveMany(
		ctx, db, candidateTasks, candidateActors, candidatePatterns, allowedPermissions,
	)
	if exception != nil {
		return successes, exception
	}

	bulkInputs := make([]sinputs.BulkUpdateMaterialInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for index, payload := range candidatePayloads {
		if !patternSuccesses[index] {
			continue
		}
		values := sinputs.UpdateMaterialInput{
			Name:        payload.Name,
			Size:        payload.Size,
			ContentKey:  payload.ContentKey,
			ContentType: payload.ContentType,
		}
		if payload.Name != nil {
			matched := s.templateMatcher.MatchString(*payload.Name, patternValues[index])
			values.Name = &matched
		}
		if payload.ContentKey != nil {
			matched := s.templateMatcher.MatchString(*payload.ContentKey, patternValues[index])
			values.ContentKey = &matched
		}
		if payload.ParseMediaType != nil {
			matched := s.templateMatcher.MatchString(*payload.ParseMediaType, patternValues[index])
			values.ParseMediaType = &matched
		}
		bulkInputs = append(bulkInputs, sinputs.BulkUpdateMaterialInput{
			UserId: candidateActors[index],
			Id:     payload.MaterialId,
			PartialUpdateInput: sinputs.PartialUpdateMaterialInput{
				Values: values,
			},
		})
		taskIndexes = append(taskIndexes, candidateIndexes[index])
	}
	if len(bulkInputs) == 0 {
		return successes, nil
	}

	bulkSuccesses, exception := s.materialRepository.BulkUpdateMany(
		bulkInputs,
		srepositories.WithDB(db.WithContext(ctx)),
		srepositories.WithTransactionDB(db.WithContext(ctx)),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return successes, exception
	}
	for index, success := range bulkSuccesses {
		successes[taskIndexes[index]] = success
	}
	return successes, nil
}

func (s *MaterialHandler) HandleDeleteMaterial(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	deleteInputs := make([]sinputs.BulkDeleteMaterialInput, 0, len(tasks))
	taskIndexes := make([]int, 0, len(tasks))
	for index, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, exception := parsers.DecodePayload[croutinetasktypes.DeleteMaterialRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		deleteInputs = append(deleteInputs, sinputs.BulkDeleteMaterialInput{Id: payload.MaterialId, UserId: actorUserId})
		taskIndexes = append(taskIndexes, index)
	}
	if len(deleteInputs) == 0 {
		return successes, nil
	}
	deleteSuccesses, exception := s.materialRepository.BulkDeleteMany(
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

func (s *MaterialHandler) HandleGetMaterialWithResults(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, map[uuid.UUID]croutinetasktypes.ExecutionResult, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	results := make(map[uuid.UUID]croutinetasktypes.ExecutionResult)
	checkInputs := make([]sinputs.BulkCheckMaterialPermissionInput, 0, len(tasks))
	taskIndexes := make([]int, 0, len(tasks))
	taskObjectIds := make([]uuid.UUID, 0, len(tasks))
	for index, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, exception := parsers.DecodePayload[croutinetasktypes.GetMaterialRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		checkInputs = append(checkInputs, sinputs.BulkCheckMaterialPermissionInput{Id: payload.MaterialId, UserId: actorUserId})
		taskIndexes = append(taskIndexes, index)
		taskObjectIds = append(taskObjectIds, payload.MaterialId)
	}
	if len(checkInputs) == 0 {
		return successes, results, nil
	}
	checkSuccesses, objects, exception := s.materialRepository.BulkCheckPermissionsAndGetManyByIds(
		checkInputs, nil, allowedPermissions,
		srepositories.WithDB(db.WithContext(ctx)), srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		return successes, nil, exception
	}
	objectsById := make(map[uuid.UUID]sschemas.Material, len(objects))
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
