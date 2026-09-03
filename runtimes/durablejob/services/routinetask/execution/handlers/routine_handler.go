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

type RoutineHandlerInterface interface {
	HandleGetRoutine(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleCreateRoutine(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleUpdateRoutine(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleDeleteRoutine(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
}

type RoutineDetailedExecutionHandlerInterface interface {
	HandleGetRoutineWithResults(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, map[uuid.UUID]croutinetasktypes.ExecutionResult, *cexceptions.Exception)
}

type RoutineHandler struct {
	Handler
	db                   *gorm.DB
	validator            *validator.Validate
	patternResolver      resolvers.RoutineTaskPatternResolverInterface
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface
	routineRepository    srepositories.RoutineRepositoryInterface
}

func NewRoutineHandler(
	validatorInstance *validator.Validate,
	patternResolver resolvers.RoutineTaskPatternResolverInterface,
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface,
	routineRepository srepositories.RoutineRepositoryInterface,
) RoutineHandlerInterface {
	return &RoutineHandler{
		validator:            validatorInstance,
		patternResolver:      patternResolver,
		templateBlockMatcher: templateBlockMatcher,
		routineRepository:    routineRepository,
	}
}

func (s *RoutineHandler) HandleGetRoutine(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes, _, exception := s.HandleGetRoutineWithResults(ctx, db, tasks, taskIdToActorUserId, allowedPermissions)
	return successes, exception
}

func (s *RoutineHandler) HandleCreateRoutine(
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
	candidatePayloads := make([]croutinetasktypes.CreateRoutineRoutineTaskPayload, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}

		payload, exception := parsers.DecodePayload[croutinetasktypes.CreateRoutineRoutineTaskPayload](s.validator, task)
		if exception != nil {
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

	bulkInputs := make([]sinputs.BulkCreateRoutineInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		title := s.templateBlockMatcher.MatchString(payload.Title, patternValues)
		description := s.templateBlockMatcher.MatchString(payload.Description, patternValues)
		bulkInputs = append(bulkInputs, sinputs.BulkCreateRoutineInput{
			UserId:           candidateActorUserIds[candidateIndex],
			Id:               payload.Id,
			StationId:        payload.StationId,
			Title:            title,
			Description:      description,
			IsPinned:         payload.IsPinned,
			ScheduledStartAt: payload.ScheduledStartAt,
			ScheduledEndAt:   payload.ScheduledEndAt,
			Period:           (*cenums.RoutinePeriod)(payload.Period),
			Timezone:         payload.Timezone,
		})
		taskIndexes = append(taskIndexes, candidateTaskIndexes[candidateIndex])
	}

	if len(bulkInputs) == 0 {
		return successes, nil
	}
	bulkSuccesses, exception := s.routineRepository.BulkCreateMany(
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

func (s *RoutineHandler) HandleUpdateRoutine(
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
	candidatePayloads := make([]croutinetasktypes.UpdateRoutineRoutineTaskPayload, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}

		payload, exception := parsers.DecodePayload[croutinetasktypes.UpdateRoutineRoutineTaskPayload](s.validator, task)
		if exception != nil {
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

	bulkInputs := make([]sinputs.BulkUpdateRoutineInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		title := payload.Title
		if title != nil {
			matchedTitle := s.templateBlockMatcher.MatchString(*title, patternValues)
			title = &matchedTitle
		}
		description := payload.Description
		if description != nil {
			matchedDescription := s.templateBlockMatcher.MatchString(*description, patternValues)
			description = &matchedDescription
		}
		bulkInputs = append(bulkInputs, sinputs.BulkUpdateRoutineInput{
			UserId: candidateActorUserIds[candidateIndex],
			Id:     payload.RoutineId,
			PartialUpdateInput: sinputs.PartialUpdateRoutineInput{
				Values: sinputs.UpdateRoutineInput{
					Title:            title,
					Description:      description,
					IsPinned:         payload.IsPinned,
					ScheduledStartAt: payload.ScheduledStartAt,
					ScheduledEndAt:   payload.ScheduledEndAt,
					Period:           (*cenums.RoutinePeriod)(payload.Period),
					Timezone:         payload.Timezone,
				},
			},
		})
		taskIndexes = append(taskIndexes, candidateTaskIndexes[candidateIndex])
	}

	if len(bulkInputs) == 0 {
		return successes, nil
	}
	bulkSuccesses, exception := s.routineRepository.BulkUpdateMany(
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

func (s *RoutineHandler) HandleDeleteRoutine(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	deleteInputs := make([]sinputs.BulkDeleteRoutineInput, 0, len(tasks))
	taskIndexes := make([]int, 0, len(tasks))
	for index, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, exception := parsers.DecodePayload[croutinetasktypes.DeleteRoutineRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		deleteInputs = append(deleteInputs, sinputs.BulkDeleteRoutineInput{
			Id:     payload.RoutineId,
			UserId: actorUserId,
		})
		taskIndexes = append(taskIndexes, index)
	}
	if len(deleteInputs) == 0 {
		return successes, nil
	}
	deleteSuccesses, exception := s.routineRepository.BulkDeleteMany(
		deleteInputs,
		srepositories.WithDB(db.WithContext(ctx)),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return successes, exception
	}
	for index, success := range deleteSuccesses {
		successes[taskIndexes[index]] = success
	}
	return successes, nil
}

func (s *RoutineHandler) HandleGetRoutineWithResults(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, map[uuid.UUID]croutinetasktypes.ExecutionResult, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	results := make(map[uuid.UUID]croutinetasktypes.ExecutionResult)
	checkInputs := make([]sinputs.BulkCheckRoutinePermissionInput, 0, len(tasks))
	taskIndexes := make([]int, 0, len(tasks))
	taskObjectIds := make([]uuid.UUID, 0, len(tasks))
	for index, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, exception := parsers.DecodePayload[croutinetasktypes.GetRoutineRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		checkInputs = append(checkInputs, sinputs.BulkCheckRoutinePermissionInput{
			Id:     payload.RoutineId,
			UserId: actorUserId,
		})
		taskIndexes = append(taskIndexes, index)
		taskObjectIds = append(taskObjectIds, payload.RoutineId)
	}
	if len(checkInputs) == 0 {
		return successes, results, nil
	}
	checkSuccesses, objects, exception := s.routineRepository.BulkCheckPermissionsAndGetManyByIds(
		checkInputs, nil, allowedPermissions,
		srepositories.WithDB(db.WithContext(ctx)), srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		return successes, nil, exception
	}
	objectsById := make(map[uuid.UUID]sschemas.Routine, len(objects))
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
