package handlers

import (
	"context"
	"net/http"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"

	matchers "github.com/HiIamJeff67/notegic-backend/internal/durablejob/routinetask/execution/matchers"
	parsers "github.com/HiIamJeff67/notegic-backend/internal/durablejob/routinetask/execution/parsers"
	resolvers "github.com/HiIamJeff67/notegic-backend/internal/durablejob/routinetask/execution/resolvers"
)

type RootShelfHandlerInterface interface {
	HandleCreateRootShelf(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleUpdateRootShelf(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleResetRootShelf(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
}

type RootShelfHandler struct {
	db                   *gorm.DB
	validator            *validator.Validate
	patternResolver      resolvers.RoutineTaskPatternResolverInterface
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface
	rootShelfRepository  srepositories.RootShelfRepositoryInterface
	subShelfRepository   srepositories.SubShelfRepositoryInterface
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
		rootShelfRepository:  srepositories.NewRootShelfRepository(db, sscopes.NewRootShelfScope()),
		subShelfRepository:   srepositories.NewSubShelfRepository(db, sscopes.NewSubShelfScope()),
	}
}

func (s *RootShelfHandler) HandleCreateRootShelf(
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

	bulkInputs := make([]sinputs.BulkCreateRootShelfInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		name := s.templateBlockMatcher.MatchString(payload.Name, patternValues)
		bulkInputs = append(bulkInputs, sinputs.BulkCreateRootShelfInput{
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

func (s *RootShelfHandler) HandleUpdateRootShelf(
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

	bulkInputs := make([]sinputs.BulkUpdateRootShelfInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] || payload.Name == nil {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		name := s.templateBlockMatcher.MatchString(*payload.Name, patternValues)
		bulkInputs = append(bulkInputs, sinputs.BulkUpdateRootShelfInput{
			UserId: candidateActorUserIds[candidateIndex],
			Id:     payload.RootShelfId,
			PartialUpdateInput: sinputs.PartialUpdateRootShelfInput{
				Values: sinputs.UpdateRootShelfInput{
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

func (s *RootShelfHandler) HandleResetRootShelf(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
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
		Model(&sschemas.SubShelf{}).
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

	bulkInputs := make([]sinputs.BulkDeleteSubShelfInput, 0, len(rows))
	taskIndexes := make([][]int, 0, len(rows))
	for _, row := range rows {
		bulkInputs = append(bulkInputs, sinputs.BulkDeleteSubShelfInput{
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
		srepositories.WithTransactionDB(db.WithContext(ctx)),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
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
