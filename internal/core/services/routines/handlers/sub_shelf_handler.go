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

type SubShelfHandlerInterface interface {
	HandleCreateSubShelf(ctx context.Context, db *gorm.DB, tasks []schemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []enums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleUpdateSubShelf(ctx context.Context, db *gorm.DB, tasks []schemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []enums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleResetSubShelf(ctx context.Context, db *gorm.DB, tasks []schemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []enums.AccessControlPermission) ([]bool, *cexceptions.Exception)
}

type SubShelfHandler struct {
	db                   *gorm.DB
	validator            *validator.Validate
	patternResolver      resolvers.RoutineTaskPatternResolverInterface
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface
	subShelfRepository   repositories.SubShelfRepositoryInterface
	blockPackRepository  repositories.BlockPackRepositoryInterface
	materialRepository   repositories.MaterialRepositoryInterface
}

func NewSubShelfHandler(
	db *gorm.DB,
	validatorInstance *validator.Validate,
	patternResolver resolvers.RoutineTaskPatternResolverInterface,
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface,
) SubShelfHandlerInterface {
	if validatorInstance == nil {
		validatorInstance = validator.New()
	}
	if patternResolver == nil {
		patternResolver = resolvers.NewRoutineTaskPatternResolver(db)
	}
	if templateBlockMatcher == nil {
		templateBlockMatcher = matchers.NewRoutineTaskTemplateMatcher()
	}
	return &SubShelfHandler{
		db:                   db,
		validator:            validatorInstance,
		patternResolver:      patternResolver,
		templateBlockMatcher: templateBlockMatcher,
		subShelfRepository:   repositories.NewSubShelfRepository(scopes.NewSubShelfScope()),
		blockPackRepository:  repositories.NewBlockPackRepository(scopes.NewBlockPackScope()),
		materialRepository:   repositories.NewMaterialRepository(scopes.NewMaterialScope()),
	}
}

func (s *SubShelfHandler) HandleCreateSubShelf(
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
	candidatePayloads := make([]croutinetasktypes.CreateSubShelfRoutineTaskPayload, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}

		payload, exception := parsers.DecodePayload[croutinetasktypes.CreateSubShelfRoutineTaskPayload](s.validator, task)
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

	bulkInputs := make([]inputs.BulkCreateSubShelfInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		name := s.templateBlockMatcher.MatchString(payload.Name, patternValues)
		bulkInputs = append(bulkInputs, inputs.BulkCreateSubShelfInput{
			UserId:         candidateActorUserIds[candidateIndex],
			Id:             payload.Id,
			RootShelfId:    payload.RootShelfId,
			PrevSubShelfId: payload.PrevSubShelfId,
			Name:           name,
		})
		taskIndexes = append(taskIndexes, candidateTaskIndexes[candidateIndex])
	}
	if len(bulkInputs) == 0 {
		return successes, nil
	}

	bulkSuccesses, exception := s.subShelfRepository.BulkCreateMany(
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

func (s *SubShelfHandler) HandleUpdateSubShelf(
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

	bulkInputs := make([]inputs.BulkUpdateSubShelfInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] || payload.Name == nil {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		name := s.templateBlockMatcher.MatchString(*payload.Name, patternValues)
		bulkInputs = append(bulkInputs, inputs.BulkUpdateSubShelfInput{
			UserId: candidateActorUserIds[candidateIndex],
			Id:     payload.SubShelfId,
			PartialUpdateInput: inputs.PartialUpdateSubShelfInput{
				Values: inputs.UpdateSubShelfInput{
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

func (s *SubShelfHandler) HandleResetSubShelf(
	ctx context.Context,
	db *gorm.DB,
	tasks []schemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	subShelfIds := make([]uuid.UUID, 0, len(tasks))
	actorUserIdBySubShelfId := make(map[uuid.UUID]uuid.UUID, len(tasks))
	taskIndexesBySubShelfId := make(map[uuid.UUID][]int, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}

		payload, exception := parsers.DecodePayload[croutinetasktypes.ResetSubShelfRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		subShelfIds = append(subShelfIds, payload.SubShelfId)
		actorUserIdBySubShelfId[payload.SubShelfId] = actorUserId
		taskIndexesBySubShelfId[payload.SubShelfId] = append(taskIndexesBySubShelfId[payload.SubShelfId], taskIndex)
	}

	if len(subShelfIds) == 0 {
		return successes, nil
	}

	tx := db.WithContext(ctx)

	var childSubShelves []struct {
		Id             uuid.UUID `gorm:"column:id"`
		PrevSubShelfId uuid.UUID `gorm:"column:prev_sub_shelf_id"`
	}
	if err := tx.Model(&schemas.SubShelf{}).
		Select("id, prev_sub_shelf_id").
		Where("prev_sub_shelf_id IN ? AND deleted_at IS NULL", subShelfIds).
		Find(&childSubShelves).Error; err != nil {
		return successes, cexceptions.New(
			"QueryFailed",
			"SubShelf",
			"Reset",
			"Failed to retrieve child sub shelves",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	var blockPacks []struct {
		Id               uuid.UUID `gorm:"column:id"`
		ParentSubShelfId uuid.UUID `gorm:"column:parent_sub_shelf_id"`
	}
	if err := tx.Model(&schemas.BlockPack{}).
		Select("id, parent_sub_shelf_id").
		Where("parent_sub_shelf_id IN ? AND deleted_at IS NULL", subShelfIds).
		Find(&blockPacks).Error; err != nil {
		return successes, cexceptions.New(
			"QueryFailed",
			"BlockPack",
			"Reset",
			"Failed to retrieve child block packs",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	var materials []struct {
		Id               uuid.UUID `gorm:"column:id"`
		ParentSubShelfId uuid.UUID `gorm:"column:parent_sub_shelf_id"`
	}
	if err := tx.Model(&schemas.Material{}).
		Select("id, parent_sub_shelf_id").
		Where("parent_sub_shelf_id IN ? AND deleted_at IS NULL", subShelfIds).
		Find(&materials).Error; err != nil {
		return successes, cexceptions.New(
			"QueryFailed",
			"Material",
			"Reset",
			"Failed to retrieve child materials",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	for _, taskIndexes := range taskIndexesBySubShelfId {
		for _, taskIndex := range taskIndexes {
			successes[taskIndex] = true
		}
	}

	if len(childSubShelves) > 0 {
		bulkInputs := make([]inputs.BulkDeleteSubShelfInput, 0, len(childSubShelves))
		taskIndexes := make([][]int, 0, len(childSubShelves))
		for _, childSubShelf := range childSubShelves {
			bulkInputs = append(bulkInputs, inputs.BulkDeleteSubShelfInput{
				UserId: actorUserIdBySubShelfId[childSubShelf.PrevSubShelfId],
				Id:     childSubShelf.Id,
			})
			taskIndexes = append(taskIndexes, taskIndexesBySubShelfId[childSubShelf.PrevSubShelfId])
		}
		bulkSuccesses, exception := s.subShelfRepository.BulkDeleteMany(
			bulkInputs,
			options.WithTransactionDB(tx),
			options.WithAllowedPermissions(allowedPermissions),
			options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
			options.WithOnlyDeleted(types.Ternary_Negative),
		)
		if exception != nil {
			return successes, exception
		}
		for index, success := range bulkSuccesses {
			if !success {
				for _, taskIndex := range taskIndexes[index] {
					successes[taskIndex] = false
				}
			}
		}
	}

	if len(blockPacks) > 0 {
		bulkInputs := make([]inputs.BulkDeleteBlockPackInput, 0, len(blockPacks))
		taskIndexes := make([][]int, 0, len(blockPacks))
		for _, blockPack := range blockPacks {
			bulkInputs = append(bulkInputs, inputs.BulkDeleteBlockPackInput{
				UserId: actorUserIdBySubShelfId[blockPack.ParentSubShelfId],
				Id:     blockPack.Id,
			})
			taskIndexes = append(taskIndexes, taskIndexesBySubShelfId[blockPack.ParentSubShelfId])
		}
		bulkSuccesses, exception := s.blockPackRepository.BulkDeleteMany(
			bulkInputs,
			options.WithTransactionDB(tx),
			options.WithAllowedPermissions(allowedPermissions),
			options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
			options.WithOnlyDeleted(types.Ternary_Negative),
		)
		if exception != nil {
			return successes, exception
		}
		for index, success := range bulkSuccesses {
			if !success {
				for _, taskIndex := range taskIndexes[index] {
					successes[taskIndex] = false
				}
			}
		}
	}

	if len(materials) > 0 {
		bulkInputs := make([]inputs.BulkDeleteMaterialInput, 0, len(materials))
		taskIndexes := make([][]int, 0, len(materials))
		for _, material := range materials {
			bulkInputs = append(bulkInputs, inputs.BulkDeleteMaterialInput{
				UserId: actorUserIdBySubShelfId[material.ParentSubShelfId],
				Id:     material.Id,
			})
			taskIndexes = append(taskIndexes, taskIndexesBySubShelfId[material.ParentSubShelfId])
		}
		bulkSuccesses, exception := s.materialRepository.BulkDeleteMany(
			bulkInputs,
			options.WithTransactionDB(tx),
			options.WithAllowedPermissions(allowedPermissions),
			options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
			options.WithOnlyDeleted(types.Ternary_Negative),
		)
		if exception != nil {
			return successes, exception
		}
		for index, success := range bulkSuccesses {
			if !success {
				for _, taskIndex := range taskIndexes[index] {
					successes[taskIndex] = false
				}
			}
		}
	}

	return successes, nil
}
