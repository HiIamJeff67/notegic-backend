package handlers

import (
	"context"
	"fmt"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"
	"net/http"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	constants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/block-packs"
	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	matchers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/matchers"
	parsers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/parsers"
	resolvers "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines/resolvers"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type YjsDocumentInitializer interface {
	InitializeDocuments(
		context.Context,
		[]capi.InitializeBlockPackYjsDocumentReqDto,
	) ([]capi.InitializeBlockPackYjsDocumentResDto, error)
}

type BlockPackHandlerInterface interface {
	HandleCreateBlockPack(ctx context.Context, db *gorm.DB, tasks []schemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []enums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleUpdateBlockPack(ctx context.Context, db *gorm.DB, tasks []schemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []enums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleResetBlockPack(ctx context.Context, db *gorm.DB, tasks []schemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []enums.AccessControlPermission) ([]bool, *cexceptions.Exception)
}

type BlockPackHandler struct {
	db                   *gorm.DB
	validator            *validator.Validate
	patternResolver      resolvers.RoutineTaskPatternResolverInterface
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface
	yjsWorkerClient      YjsDocumentInitializer
	blockPackRepository  repositories.BlockPackRepositoryInterface
	blockRepository      repositories.BlockRepositoryInterface
}

func NewBlockPackHandler(
	db *gorm.DB,
	validatorInstance *validator.Validate,
	yjsDocumentInitializer YjsDocumentInitializer,
	patternResolver resolvers.RoutineTaskPatternResolverInterface,
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface,
) BlockPackHandlerInterface {
	if validatorInstance == nil {
		validatorInstance = validator.New()
	}
	if patternResolver == nil {
		patternResolver = resolvers.NewRoutineTaskPatternResolver(db)
	}
	if templateBlockMatcher == nil {
		templateBlockMatcher = matchers.NewRoutineTaskTemplateMatcher()
	}
	return &BlockPackHandler{
		db:                   db,
		validator:            validatorInstance,
		patternResolver:      patternResolver,
		templateBlockMatcher: templateBlockMatcher,
		yjsWorkerClient:      yjsDocumentInitializer,
		blockPackRepository:  repositories.NewBlockPackRepository(scopes.NewBlockPackScope()),
		blockRepository:      repositories.NewBlockRepository(scopes.NewBlockScope()),
	}
}

func (s *BlockPackHandler) HandleCreateBlockPack(
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
	candidatePayloads := make([]croutinetasktypes.CreateBlockPackRoutineTaskPayload, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, exception := parsers.DecodePayload[croutinetasktypes.CreateBlockPackRoutineTaskPayload](s.validator, task)
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

	blockPackInputs := make([]inputs.BulkCreateBlockPackInput, 0, len(candidateTasks))
	blockContentInputs := make([]inputs.BulkCreateBlockPackContentInput, 0, len(candidateTasks))
	initializationReqDtos := make([]capi.InitializeBlockPackYjsDocumentReqDto, 0, len(candidateTasks))
	preparedTaskIndexes := make([]int, 0, len(candidateTasks))

	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] {
			continue
		}
		patternValues := patternValuesByCandidate[candidateIndex]
		blockPackId := uuid.New()
		name := s.templateBlockMatcher.MatchString(payload.Template.Name, patternValues)
		var prevRootId *uuid.UUID
		taskFailed := false
		taskBlocks := make([]inputs.CreateBlockInput, 0)
		matchedRootBlocks := make([]cblocknote.ArborizedEditableBlock, 0, len(payload.Template.Blocks))
		prevRootInputIndex := -1
		for _, block := range payload.Template.Blocks {
			matchedBlock, exception := s.templateBlockMatcher.MatchArborizedEditableBlock(block.ArborizedEditableBlock, patternValues)
			if exception != nil {
				taskFailed = true
				break
			}
			matchedRootBlocks = append(matchedRootBlocks, matchedBlock)
			blocks, _, _, exception := parsers.FlattenArborizedBlock(blockPackId, &matchedBlock)
			if exception != nil || len(blocks) == 0 {
				taskFailed = true
				break
			}
			blocks[0].PrevBlockId = prevRootId
			if prevRootInputIndex >= 0 {
				nextBlockId := blocks[0].Id
				taskBlocks[prevRootInputIndex].NextBlockId = &nextBlockId
			}
			prevRootId = &blocks[0].Id
			prevRootInputIndex = len(taskBlocks)
			for _, block := range blocks {
				taskBlocks = append(taskBlocks, inputs.CreateBlockInput{
					Id:            block.Id,
					BlockPackId:   block.BlockPackId,
					ParentBlockId: block.ParentBlockId,
					PrevBlockId:   block.PrevBlockId,
					NextBlockId:   block.NextBlockId,
					Type:          block.Type,
					Props:         block.Props,
					Content:       block.Content,
				})
			}
		}
		if taskFailed || len(taskBlocks) == 0 {
			continue
		}
		blockPackInputs = append(blockPackInputs, inputs.BulkCreateBlockPackInput{
			UserId:              candidateActorUserIds[candidateIndex],
			Id:                  &blockPackId,
			ParentSubShelfId:    payload.TargetSubShelfId,
			Name:                name,
			Icon:                (*enums.SupportedIcon)(payload.Template.Icon),
			HeaderBackgroundURL: payload.Template.HeaderBackgroundURL,
		})
		blockContentInputs = append(blockContentInputs, inputs.BulkCreateBlockPackContentInput{
			UserId:      candidateActorUserIds[candidateIndex],
			BlockPackId: blockPackId,
			Blocks:      taskBlocks,
		})
		initializationReqDtos = append(initializationReqDtos, capi.InitializeBlockPackYjsDocumentReqDto{
			Blocks: matchedRootBlocks,
		})
		preparedTaskIndexes = append(preparedTaskIndexes, candidateTaskIndexes[candidateIndex])
	}
	if len(blockPackInputs) == 0 {
		return successes, nil
	}
	if s.yjsWorkerClient == nil {
		return successes, cexceptions.New(
			"DependencyUnavailable",
			"BlockPack",
			"Create",
			"The Yjs worker document initializer is not configured",
			http.StatusServiceUnavailable,
			true,
		)
	}
	initializationResDtos, err := s.yjsWorkerClient.InitializeDocuments(ctx, initializationReqDtos)
	if err != nil {
		return successes, cexceptions.New(
			"FailedToCreate",
			"BlockPack",
			"Create",
			"Failed to initialize block pack documents",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	tx := db.WithContext(ctx)

	blockPackSuccesses, exception := s.blockPackRepository.BulkCreateMany(
		blockPackInputs,
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(types.Ternary_Negative),
	)
	if exception != nil {
		return successes, exception
	}

	successfulBlockContentInputs := make([]inputs.BulkCreateBlockPackContentInput, 0, len(blockContentInputs))
	successfulInitializationResDtos := make([]capi.InitializeBlockPackYjsDocumentResDto, 0, len(initializationResDtos))
	successfulTaskIndexes := make([]int, 0, len(preparedTaskIndexes))
	for index, success := range blockPackSuccesses {
		if success {
			successfulBlockContentInputs = append(successfulBlockContentInputs, blockContentInputs[index])
			successfulInitializationResDtos = append(successfulInitializationResDtos, initializationResDtos[index])
			successfulTaskIndexes = append(successfulTaskIndexes, preparedTaskIndexes[index])
		}
	}
	if len(successfulBlockContentInputs) == 0 {
		return successes, nil
	}

	documents := make([]schemas.BlockPackYjsDocument, len(successfulBlockContentInputs))
	for index, successfulBlockContentInput := range successfulBlockContentInputs {
		documents[index] = schemas.BlockPackYjsDocument{
			BlockPackId:            successfulBlockContentInput.BlockPackId,
			Snapshot:               successfulInitializationResDtos[index].Snapshot,
			StateVector:            successfulInitializationResDtos[index].StateVector,
			ProjectedUntilSequence: 0,
		}
	}
	if err := tx.CreateInBatches(&documents, constants.MaxBatchCreateBlockSize).Error; err != nil {
		return successes, cexceptions.New(
			"FailedToCreate",
			"BlockPack",
			"Create",
			"Failed to create block pack documents",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	blockSuccesses, exception := s.blockRepository.BulkCreateMany(
		successfulBlockContentInputs,
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(types.Ternary_Negative),
	)
	if exception != nil {
		return successes, exception
	}
	for _, success := range blockSuccesses {
		if !success {
			return successes, nil
		}
	}

	for _, taskIndex := range successfulTaskIndexes {
		successes[taskIndex] = true
	}

	return successes, nil
}

func (s *BlockPackHandler) HandleUpdateBlockPack(
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
	candidatePayloads := make([]croutinetasktypes.UpdateBlockPackRoutineTaskPayload, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, exception := parsers.DecodePayload[croutinetasktypes.UpdateBlockPackRoutineTaskPayload](s.validator, task)
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

	preparedInputs := make([]inputs.BulkUpdateBlockInput, 0)
	taskIndexes := make([]int, 0)
	pairPlaceholders := make([]string, 0)
	pairArgs := make([]any, 0)

	for candidateIndex, payload := range candidatePayloads {
		if !patternSuccesses[candidateIndex] {
			continue
		}
		actorUserId := candidateActorUserIds[candidateIndex]
		patternValues := patternValuesByCandidate[candidateIndex]
		for _, block := range payload.UpdatedBlocks {
			if block.ArborizedEditableBlock == nil {
				continue
			}
			matchedBlock, exception := s.templateBlockMatcher.MatchArborizedEditableBlock(*block.ArborizedEditableBlock, patternValues)
			if exception != nil {
				continue
			}
			flattenedBlocks, _, _, exception := parsers.FlattenArborizedBlock(payload.BlockPackId, &matchedBlock)
			if exception != nil || len(flattenedBlocks) != 1 {
				continue
			}
			blockType := flattenedBlocks[0].Type
			props := datatypes.JSON(flattenedBlocks[0].Props)
			content := datatypes.JSON(flattenedBlocks[0].Content)
			pairPlaceholders = append(pairPlaceholders, "(?::uuid, ?::uuid)")
			pairArgs = append(pairArgs, block.BlockId, payload.BlockPackId)
			preparedInputs = append(preparedInputs, inputs.BulkUpdateBlockInput{
				UserId: actorUserId,
				Id:     block.BlockId,
				PartialUpdateInput: inputs.PartialUpdateBlockInput{Values: inputs.UpdateBlockInput{
					Type:    &blockType,
					Props:   &props,
					Content: &content,
				}},
			})
			taskIndexes = append(taskIndexes, candidateTaskIndexes[candidateIndex])
		}
	}
	if len(preparedInputs) == 0 {
		return successes, nil
	}

	var validRows []struct {
		BlockId     uuid.UUID `gorm:"column:block_id"`
		BlockPackId uuid.UUID `gorm:"column:block_pack_id"`
	}
	sql := fmt.Sprintf(`
		WITH pairs(block_id, block_pack_id) AS (VALUES %s)
		SELECT p.block_id::uuid, p.block_pack_id::uuid
		FROM pairs p
		INNER JOIN "BlockTable" b ON b.id = p.block_id::uuid AND b.block_pack_id = p.block_pack_id::uuid
	`, strings.Join(pairPlaceholders, ","))
	if err := db.WithContext(ctx).Raw(sql, pairArgs...).Scan(&validRows).Error; err != nil {
		return successes, cexceptions.New(
			"QueryFailed",
			"Block",
			"Update",
			"Failed to validate block pack blocks",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	valid := make(map[[2]uuid.UUID]bool, len(validRows))
	for _, row := range validRows {
		valid[[2]uuid.UUID{row.BlockId, row.BlockPackId}] = true
	}
	filteredInputs := make([]inputs.BulkUpdateBlockInput, 0, len(preparedInputs))
	filteredTaskIndexes := make([]int, 0, len(taskIndexes))
	for index, input := range preparedInputs {
		blockPackId := pairArgs[index*2+1].(uuid.UUID)
		if valid[[2]uuid.UUID{input.Id, blockPackId}] {
			filteredInputs = append(filteredInputs, input)
			filteredTaskIndexes = append(filteredTaskIndexes, taskIndexes[index])
		}
	}
	if len(filteredInputs) == 0 {
		return successes, nil
	}

	bulkSuccesses, exception := s.blockRepository.BulkUpdateMany(
		filteredInputs,
		options.WithTransactionDB(db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(types.Ternary_Negative),
	)
	if exception != nil {
		return successes, exception
	}
	for index, success := range bulkSuccesses {
		successes[filteredTaskIndexes[index]] = success
	}

	return successes, nil
}

func (s *BlockPackHandler) HandleResetBlockPack(
	ctx context.Context,
	db *gorm.DB,
	tasks []schemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	checkInputs := make([]inputs.BulkCheckBlockPackPermissionInput, 0, len(tasks))
	taskIndexes := make([]int, 0, len(tasks))
	blockPackIds := make([]uuid.UUID, 0, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, exception := parsers.DecodePayload[croutinetasktypes.ResetBlockPackRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		checkInputs = append(checkInputs, inputs.BulkCheckBlockPackPermissionInput{
			UserId: actorUserId,
			Id:     payload.BlockPackId,
		})
		taskIndexes = append(taskIndexes, taskIndex)
		blockPackIds = append(blockPackIds, payload.BlockPackId)
	}
	if len(checkInputs) == 0 {
		return successes, nil
	}

	tx := db.WithContext(ctx)

	checkSuccesses, _, exception := s.blockPackRepository.BulkCheckPermissionsAndGetManyByIds(
		checkInputs,
		nil,
		allowedPermissions,
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(types.Ternary_Negative),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		return successes, exception
	}

	validBlockPackIds := make([]uuid.UUID, 0, len(blockPackIds))
	for index, success := range checkSuccesses {
		if success {
			validBlockPackIds = append(validBlockPackIds, blockPackIds[index])
		}
	}
	if len(validBlockPackIds) == 0 {
		return successes, nil
	}

	if err := tx.Model(&schemas.Block{}).
		Where("block_pack_id IN ? AND deleted_at IS NULL", validBlockPackIds).
		Updates(map[string]any{"deleted_at": time.Now(), "prev_block_id": nil, "next_block_id": nil}).Error; err != nil {
		return successes, cexceptions.New(
			"FailedToUpdate",
			"Block",
			"Reset",
			"Failed to reset block pack blocks",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	for index, success := range checkSuccesses {
		successes[taskIndexes[index]] = success
	}

	return successes, nil
}
