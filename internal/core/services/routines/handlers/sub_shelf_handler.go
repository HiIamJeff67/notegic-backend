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

	matchers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/matchers"
	parsers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/parsers"
	resolvers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/resolvers"
)

type SubShelfHandlerInterface interface {
	HandleCreateSubShelf(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleUpdateSubShelf(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleResetSubShelf(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
}

type SubShelfHandler struct {
	db                   *gorm.DB
	validator            *validator.Validate
	patternResolver      resolvers.RoutineTaskPatternResolverInterface
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface
	subShelfRepository   srepositories.SubShelfRepositoryInterface
	blockPackRepository  srepositories.BlockPackRepositoryInterface
	materialRepository   srepositories.MaterialRepositoryInterface
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
		subShelfRepository:   srepositories.NewSubShelfRepository(db, sscopes.NewSubShelfScope()),
		blockPackRepository:  srepositories.NewBlockPackRepository(db, sscopes.NewBlockPackScope()),
		materialRepository:   srepositories.NewMaterialRepository(db, sscopes.NewMaterialScope()),
	}
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

	bulkInputs := make([]sinputs.BulkCreateSubShelfInput, 0, len(candidateTasks))
	taskIndexes := make([]int, 0, len(candidateTasks))
	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		name := s.templateBlockMatcher.MatchString(payload.Name, patternValues)
		bulkInputs = append(bulkInputs, sinputs.BulkCreateSubShelfInput{
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

func (s *SubShelfHandler) HandleResetSubShelf(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
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
	if err := tx.Model(&sschemas.SubShelf{}).
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
	if err := tx.Model(&sschemas.BlockPack{}).
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
	if err := tx.Model(&sschemas.Material{}).
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
		bulkInputs := make([]sinputs.BulkDeleteSubShelfInput, 0, len(childSubShelves))
		taskIndexes := make([][]int, 0, len(childSubShelves))
		for _, childSubShelf := range childSubShelves {
			bulkInputs = append(bulkInputs, sinputs.BulkDeleteSubShelfInput{
				UserId: actorUserIdBySubShelfId[childSubShelf.PrevSubShelfId],
				Id:     childSubShelf.Id,
			})
			taskIndexes = append(taskIndexes, taskIndexesBySubShelfId[childSubShelf.PrevSubShelfId])
		}
		bulkSuccesses, exception := s.subShelfRepository.BulkDeleteMany(
			bulkInputs,
			srepositories.WithTransactionDB(tx),
			srepositories.WithAllowedPermissions(allowedPermissions),
			srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
			srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
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
		bulkInputs := make([]sinputs.BulkDeleteBlockPackInput, 0, len(blockPacks))
		taskIndexes := make([][]int, 0, len(blockPacks))
		for _, blockPack := range blockPacks {
			bulkInputs = append(bulkInputs, sinputs.BulkDeleteBlockPackInput{
				UserId: actorUserIdBySubShelfId[blockPack.ParentSubShelfId],
				Id:     blockPack.Id,
			})
			taskIndexes = append(taskIndexes, taskIndexesBySubShelfId[blockPack.ParentSubShelfId])
		}
		bulkSuccesses, exception := s.blockPackRepository.BulkDeleteMany(
			bulkInputs,
			srepositories.WithTransactionDB(tx),
			srepositories.WithAllowedPermissions(allowedPermissions),
			srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
			srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
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
		bulkInputs := make([]sinputs.BulkDeleteMaterialInput, 0, len(materials))
		taskIndexes := make([][]int, 0, len(materials))
		for _, material := range materials {
			bulkInputs = append(bulkInputs, sinputs.BulkDeleteMaterialInput{
				UserId: actorUserIdBySubShelfId[material.ParentSubShelfId],
				Id:     material.Id,
			})
			taskIndexes = append(taskIndexes, taskIndexesBySubShelfId[material.ParentSubShelfId])
		}
		bulkSuccesses, exception := s.materialRepository.BulkDeleteMany(
			bulkInputs,
			srepositories.WithTransactionDB(tx),
			srepositories.WithAllowedPermissions(allowedPermissions),
			srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
			srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
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
