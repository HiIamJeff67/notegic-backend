package handlers

import (
	"context"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"
	"net/http"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	matchers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/matchers"
	parsers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/parsers"
	resolvers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/resolvers"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type RootShelfHandlerInterface interface {
	HandleCreateRootShelf(ctx context.Context, db *gorm.DB, tasks []schemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []enums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleUpdateRootShelf(ctx context.Context, db *gorm.DB, tasks []schemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []enums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleResetRootShelf(ctx context.Context, db *gorm.DB, tasks []schemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []enums.AccessControlPermission) ([]bool, *cexceptions.Exception)
}

type RootShelfHandler struct {
	db                   *gorm.DB
	validator            *validator.Validate
	patternResolver      resolvers.RoutineTaskPatternResolverInterface
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface
	rootShelfRepository  repositories.RootShelfRepositoryInterface
	subShelfRepository   repositories.SubShelfRepositoryInterface
}

func NewRootShelfHandler(
	db *gorm.DB,
	validatorInstance *validator.Validate,
	patternResolver resolvers.RoutineTaskPatternResolverInterface,
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface,
) RootShelfHandlerInterface {
	if validatorInstance == nil {
		validatorInstance = validator.New()
	}
	if patternResolver == nil {
		patternResolver = resolvers.NewRoutineTaskPatternResolver(db)
	}
	if templateBlockMatcher == nil {
		templateBlockMatcher = matchers.NewRoutineTaskTemplateMatcher()
	}
	return &RootShelfHandler{
		db:                   db,
		validator:            validatorInstance,
		patternResolver:      patternResolver,
		templateBlockMatcher: templateBlockMatcher,
		rootShelfRepository:  repositories.NewRootShelfRepository(scopes.NewRootShelfScope()),
		subShelfRepository:   repositories.NewSubShelfRepository(scopes.NewSubShelfScope()),
	}
}

func (s *RootShelfHandler) HandleCreateRootShelf(
	ctx context.Context,
	db *gorm.DB,
	tasks []schemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	candidateTaskIndexes := make([]int, 0, len(tasks))
	candidateTasks := make([]schemas.RoutineTask, 0, len(tasks))
	candidateActorUserIds := make([]uuid.UUID, 0, len(tasks))
	candidatePayloads := make([]croutinetasktypes.CreateRootShelfRoutineTaskPayload, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}

		payload, exception := parsers.DecodePayload[croutinetasktypes.CreateRootShelfRoutineTaskPayload](s.validator, task)
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

	bulkInputs := make([]inputs.BulkCreateRootShelfInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		name := s.templateBlockMatcher.MatchString(payload.Name, patternValues)
		bulkInputs = append(bulkInputs, inputs.BulkCreateRootShelfInput{
			UserId: candidateActorUserIds[candidateIndex],
			Id:     payload.Id,
			Name:   name,
		})
		taskIndexes = append(taskIndexes, candidateTaskIndexes[candidateIndex])
	}

	if len(bulkInputs) == 0 {
		return successes, nil
	}
	bulkSuccesses, exception := s.rootShelfRepository.BulkCreateMany(
		bulkInputs,
		options.WithTransactionDB(db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
		options.WithOnlyDeleted(types.Ternary_Negative),
	)
	if exception != nil {
		return successes, exception
	}

	for index, success := range bulkSuccesses {
		successes[taskIndexes[index]] = success
	}

	return successes, nil
}

func (s *RootShelfHandler) HandleUpdateRootShelf(
	ctx context.Context,
	db *gorm.DB,
	tasks []schemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	candidateTaskIndexes := make([]int, 0, len(tasks))
	candidateTasks := make([]schemas.RoutineTask, 0, len(tasks))
	candidateActorUserIds := make([]uuid.UUID, 0, len(tasks))
	candidatePayloads := make([]croutinetasktypes.UpdateRootShelfRoutineTaskPayload, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}

		payload, exception := parsers.DecodePayload[croutinetasktypes.UpdateRootShelfRoutineTaskPayload](s.validator, task)
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

	bulkInputs := make([]inputs.BulkUpdateRootShelfInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] || payload.Name == nil {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		name := s.templateBlockMatcher.MatchString(*payload.Name, patternValues)
		bulkInputs = append(bulkInputs, inputs.BulkUpdateRootShelfInput{
			UserId: candidateActorUserIds[candidateIndex],
			Id:     payload.RootShelfId,
			PartialUpdateInput: inputs.PartialUpdateRootShelfInput{
				Values: inputs.UpdateRootShelfInput{
					Name: &name,
				},
			},
		})
		taskIndexes = append(taskIndexes, candidateTaskIndexes[candidateIndex])
	}

	if len(bulkInputs) == 0 {
		return successes, nil
	}

	bulkSuccesses, exception := s.rootShelfRepository.BulkUpdateMany(
		bulkInputs,
		options.WithTransactionDB(db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
		options.WithOnlyDeleted(types.Ternary_Negative),
	)
	if exception != nil {
		return successes, exception
	}

	for index, success := range bulkSuccesses {
		successes[taskIndexes[index]] = success
	}

	return successes, nil
}

func (s *RootShelfHandler) HandleResetRootShelf(
	ctx context.Context,
	db *gorm.DB,
	tasks []schemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	rootShelfIds := make([]uuid.UUID, 0, len(tasks))
	taskIndexesByRootShelfId := make(map[uuid.UUID][]int, len(tasks))
	actorUserIdByRootShelfId := make(map[uuid.UUID]uuid.UUID, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}

		payload, exception := parsers.DecodePayload[croutinetasktypes.ResetRootShelfRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		rootShelfIds = append(rootShelfIds, payload.RootShelfId)
		taskIndexesByRootShelfId[payload.RootShelfId] = append(taskIndexesByRootShelfId[payload.RootShelfId], taskIndex)
		actorUserIdByRootShelfId[payload.RootShelfId] = actorUserId
	}

	if len(rootShelfIds) == 0 {
		return successes, nil
	}

	var rows []struct {
		Id          uuid.UUID `gorm:"column:id"`
		RootShelfId uuid.UUID `gorm:"column:root_shelf_id"`
	}
	if err := db.WithContext(ctx).
		Model(&schemas.SubShelf{}).
		Select("id, root_shelf_id").
		Where("root_shelf_id IN ? AND deleted_at IS NULL", rootShelfIds).
		Find(&rows).Error; err != nil {
		return successes, cexceptions.New(
			"QueryFailed",
			"RootShelf",
			"Reset",
			"Failed to retrieve root shelf descendants",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	bulkInputs := make([]inputs.BulkDeleteSubShelfInput, 0, len(rows))
	taskIndexes := make([][]int, 0, len(rows))
	for _, row := range rows {
		bulkInputs = append(bulkInputs, inputs.BulkDeleteSubShelfInput{
			UserId: actorUserIdByRootShelfId[row.RootShelfId],
			Id:     row.Id,
		})
		taskIndexes = append(taskIndexes, taskIndexesByRootShelfId[row.RootShelfId])
	}
	if len(bulkInputs) == 0 {
		for _, indexes := range taskIndexesByRootShelfId {
			for _, taskIndex := range indexes {
				successes[taskIndex] = true
			}
		}
		return successes, nil
	}

	bulkSuccesses, exception := s.subShelfRepository.BulkDeleteMany(
		bulkInputs,
		options.WithTransactionDB(db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
		options.WithOnlyDeleted(types.Ternary_Negative),
	)
	if exception != nil {
		return successes, exception
	}

	for _, indexes := range taskIndexesByRootShelfId {
		for _, taskIndex := range indexes {
			successes[taskIndex] = true
		}
	}
	for index, success := range bulkSuccesses {
		if !success {
			for _, taskIndex := range taskIndexes[index] {
				successes[taskIndex] = false
			}
		}
	}

	return successes, nil
}
