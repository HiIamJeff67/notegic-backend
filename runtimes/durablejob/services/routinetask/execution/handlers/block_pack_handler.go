package handlers

import (
	"context"
	"net/http"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/block-packs"
	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"

	matchers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/matchers"
	parsers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/parsers"
	resolvers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/execution/resolvers"
)

type YjsDocumentInitializer interface {
	InitializeDocuments(
		context.Context,
		[]capi.InitializeBlockPackYjsDocumentReqDto,
	) ([]capi.InitializeBlockPackYjsDocumentResDto, error)
}

type YjsBlockPackUpdater interface {
	UpdateBlockPack(
		context.Context,
		capi.UpdateBlockPackYjsDocumentRequestDto,
	) (*capi.UpdateBlockPackYjsDocumentResponseDto, error)
}

type BlockPackHandlerInterface interface {
	HandleGetBlockPack(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleCreateBlockPack(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleUpdateBlockPack(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
	HandleDeleteBlockPack(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, *cexceptions.Exception)
}

type BlockPackDetailedExecutionHandlerInterface interface {
	HandleUpdateBlockPackWithResults(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, map[uuid.UUID]croutinetasktypes.ExecutionResult, *cexceptions.Exception)
}

type BlockPackGetDetailedExecutionHandlerInterface interface {
	HandleGetBlockPackWithResults(ctx context.Context, db *gorm.DB, tasks []sschemas.RoutineTask, taskIdToActorUserId map[uuid.UUID]uuid.UUID, allowedPermissions []cenums.AccessControlPermission) ([]bool, map[uuid.UUID]croutinetasktypes.ExecutionResult, *cexceptions.Exception)
}

type BlockPackHandler struct {
	Handler
	db                   *gorm.DB
	validator            *validator.Validate
	patternResolver      resolvers.RoutineTaskPatternResolverInterface
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface
	yjsWorkerClient      YjsDocumentInitializer
	yjsBlockPackUpdater  YjsBlockPackUpdater
	blockPackRepository  srepositories.BlockPackRepositoryInterface
	blockRepository      srepositories.BlockRepositoryInterface
}

func NewBlockPackHandler(
	db *gorm.DB,
	validatorInstance *validator.Validate,
	patternResolver resolvers.RoutineTaskPatternResolverInterface,
	templateBlockMatcher matchers.RoutineTaskTemplateMatcherInterface,
	yjsDocumentInitializer YjsDocumentInitializer,
	yjsBlockPackUpdater YjsBlockPackUpdater,
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
		yjsBlockPackUpdater:  yjsBlockPackUpdater,
		blockPackRepository:  srepositories.NewBlockPackRepository(db, sscopes.NewBlockPackScope()),
		blockRepository:      srepositories.NewBlockRepository(db, sscopes.NewBlockScope()),
	}
}

func (s *BlockPackHandler) HandleGetBlockPack(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes, _, exception := s.HandleGetBlockPackWithResults(ctx, db, tasks, taskIdToActorUserId, allowedPermissions)
	return successes, exception
}

func (s *BlockPackHandler) HandleCreateBlockPack(
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

	blockPackInputs := make([]sinputs.BulkCreateBlockPackInput, 0, len(candidateTasks))
	blockContentInputs := make([]sinputs.BulkCreateBlockPackContentInput, 0, len(candidateTasks))
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
		taskBlocks := make([]sinputs.CreateBlockInput, 0)
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
				taskBlocks = append(taskBlocks, sinputs.CreateBlockInput{
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
		blockPackInputs = append(blockPackInputs, sinputs.BulkCreateBlockPackInput{
			UserId:              candidateActorUserIds[candidateIndex],
			Id:                  &blockPackId,
			ParentSubShelfId:    payload.TargetSubShelfId,
			Name:                name,
			Icon:                (*cenums.SupportedIcon)(payload.Template.Icon),
			HeaderBackgroundURL: payload.Template.HeaderBackgroundURL,
		})
		blockContentInputs = append(blockContentInputs, sinputs.BulkCreateBlockPackContentInput{
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
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		return successes, exception
	}

	successfulBlockContentInputs := make([]sinputs.BulkCreateBlockPackContentInput, 0, len(blockContentInputs))
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

	documents := make([]sschemas.BlockPackYjsDocument, len(successfulBlockContentInputs))
	for index, successfulBlockContentInput := range successfulBlockContentInputs {
		documents[index] = sschemas.BlockPackYjsDocument{
			BlockPackId:            successfulBlockContentInput.BlockPackId,
			Snapshot:               successfulInitializationResDtos[index].Snapshot,
			StateVector:            successfulInitializationResDtos[index].StateVector,
			ProjectedUntilSequence: 0,
		}
	}
	if err := tx.CreateInBatches(&documents, sconstants.MaxBatchCreateBlockSize).Error; err != nil {
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
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
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
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes, _, exception := s.HandleUpdateBlockPackWithResults(
		ctx,
		db,
		tasks,
		taskIdToActorUserId,
		allowedPermissions,
	)

	return successes, exception
}

func (s *BlockPackHandler) HandleDeleteBlockPack(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	deleteInputs := make([]sinputs.BulkDeleteBlockPackInput, 0, len(tasks))
	taskIndexes := make([]int, 0, len(tasks))
	for index, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, exception := parsers.DecodePayload[croutinetasktypes.DeleteBlockPackRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		deleteInputs = append(deleteInputs, sinputs.BulkDeleteBlockPackInput{Id: payload.BlockPackId, UserId: actorUserId})
		taskIndexes = append(taskIndexes, index)
	}
	if len(deleteInputs) == 0 {
		return successes, nil
	}
	deleteSuccesses, exception := s.blockPackRepository.BulkDeleteMany(
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

func (s *BlockPackHandler) HandleUpdateBlockPackWithResults(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, map[uuid.UUID]croutinetasktypes.ExecutionResult, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	results := make(map[uuid.UUID]croutinetasktypes.ExecutionResult)
	candidateTaskIndexes := make([]int, 0, len(tasks))
	candidateTasks := make([]sschemas.RoutineTask, 0, len(tasks))
	candidateActorUserIds := make([]uuid.UUID, 0, len(tasks))
	candidatePayloads := make([]croutinetasktypes.UpdateBlockPackRoutineTaskPayload, 0, len(tasks))
	candidatePatterns := make([]croutinetasktypes.RoutineTaskPattern, 0, len(tasks))

	for taskIndex, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, decodeException := parsers.DecodePayload[croutinetasktypes.UpdateBlockPackRoutineTaskPayload](s.validator, task)
		if decodeException != nil {
			continue
		}
		candidateTaskIndexes = append(candidateTaskIndexes, taskIndex)
		candidateTasks = append(candidateTasks, task)
		candidateActorUserIds = append(candidateActorUserIds, actorUserId)
		candidatePayloads = append(candidatePayloads, *payload)
		candidatePatterns = append(candidatePatterns, payload.Pattern)
	}
	if len(candidateTasks) == 0 {
		return successes, results, nil
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
		return successes, results, exception
	}
	if s.yjsBlockPackUpdater == nil {
		return successes, results, cexceptions.New(
			"DependencyUnavailable",
			"BlockPack",
			"Update",
			"The Yjs worker block pack updater is not configured",
			http.StatusServiceUnavailable,
			true,
		)
	}

	for candidateIndex, payload := range candidatePayloads {
		taskIndex := candidateTaskIndexes[candidateIndex]
		result := croutinetasktypes.ExecutionResult{At: time.Now().UTC()}
		if !patternSuccesses[candidateIndex] {
			continue
		}

		items := make([]croutinetasktypes.ExecutionItemResult, len(payload.Blocks))
		requestBlocks := make([]capi.UpdateBlockPackYjsDocumentBlockRequestDto, 0, len(payload.Blocks))
		requestBlockIndexes := make([]int, 0, len(payload.Blocks))
		for blockIndex, block := range payload.Blocks {
			item := croutinetasktypes.ExecutionItemResult{ItemId: block.BlockId.String()}
			if block.ArborizedEditableBlock == nil {
				item.Status = croutinetasktypes.ExecutionItemStatus_Failed
				item.Reason = "block_payload_invalid"
				result.Failed++
				items[blockIndex] = item
				continue
			}
			matchedBlock, matchException := s.templateBlockMatcher.MatchArborizedEditableBlock(
				*block.ArborizedEditableBlock,
				patternValuesByCandidate[candidateIndex],
			)
			if matchException != nil {
				item.Status = croutinetasktypes.ExecutionItemStatus_Failed
				item.Reason = "block_payload_invalid"
				result.Failed++
				items[blockIndex] = item
				continue
			}
			flattenedBlocks, _, _, flattenException := parsers.FlattenArborizedBlock(
				payload.BlockPackId,
				&matchedBlock,
			)
			if flattenException != nil || len(flattenedBlocks) != 1 {
				item.Status = croutinetasktypes.ExecutionItemStatus_Failed
				item.Reason = "block_payload_must_be_single_block"
				result.Failed++
				items[blockIndex] = item
				continue
			}
			requestBlocks = append(requestBlocks, capi.UpdateBlockPackYjsDocumentBlockRequestDto{
				BlockId: block.BlockId,
				Block:   matchedBlock,
			})
			requestBlockIndexes = append(requestBlockIndexes, blockIndex)
		}

		if len(requestBlocks) > 0 {
			responseDto, err := s.yjsBlockPackUpdater.UpdateBlockPack(ctx, capi.UpdateBlockPackYjsDocumentRequestDto{
				BlockPackId: payload.BlockPackId,
				Blocks:      requestBlocks,
			})
			if err != nil {
				return successes, results, cexceptions.New(
					"FailedToUpdate",
					"BlockPack",
					"Update",
					"Failed to update block pack through the Yjs worker",
					http.StatusInternalServerError,
					true,
				).WithOrigin(err)
			}
			if responseDto == nil || len(responseDto.Blocks) != len(requestBlocks) {
				return successes, results, cexceptions.New(
					"InvalidResponse",
					"BlockPack",
					"Update",
					"The Yjs worker returned an incomplete block pack update response",
					http.StatusBadGateway,
					true,
				)
			}
			for responseIndex, blockResult := range responseDto.Blocks {
				item := croutinetasktypes.ExecutionItemResult{ItemId: blockResult.BlockId.String()}
				switch blockResult.Status {
				case "updated":
					item.Status = croutinetasktypes.ExecutionItemStatus_Updated
					result.Updated++
				case "skipped":
					item.Status = croutinetasktypes.ExecutionItemStatus_Skipped
					result.Skipped++
				default:
					item.Status = croutinetasktypes.ExecutionItemStatus_Failed
					result.Failed++
				}
				item.Reason = blockResult.Reason
				items[requestBlockIndexes[responseIndex]] = item
			}
		}

		result.Items = items
		successes[taskIndex] = result.Failed == 0
		results[tasks[taskIndex].Id] = result
	}

	return successes, results, nil
}

func (s *BlockPackHandler) HandleGetBlockPackWithResults(
	ctx context.Context,
	db *gorm.DB,
	tasks []sschemas.RoutineTask,
	taskIdToActorUserId map[uuid.UUID]uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
) ([]bool, map[uuid.UUID]croutinetasktypes.ExecutionResult, *cexceptions.Exception) {
	successes := make([]bool, len(tasks))
	results := make(map[uuid.UUID]croutinetasktypes.ExecutionResult)
	checkInputs := make([]sinputs.BulkCheckBlockPackPermissionInput, 0, len(tasks))
	taskIndexes := make([]int, 0, len(tasks))
	taskObjectIds := make([]uuid.UUID, 0, len(tasks))
	for index, task := range tasks {
		actorUserId, exists := taskIdToActorUserId[task.Id]
		if !exists {
			continue
		}
		payload, exception := parsers.DecodePayload[croutinetasktypes.GetBlockPackRoutineTaskPayload](s.validator, task)
		if exception != nil {
			continue
		}
		checkInputs = append(checkInputs, sinputs.BulkCheckBlockPackPermissionInput{Id: payload.BlockPackId, UserId: actorUserId})
		taskIndexes = append(taskIndexes, index)
		taskObjectIds = append(taskObjectIds, payload.BlockPackId)
	}
	if len(checkInputs) == 0 {
		return successes, results, nil
	}
	checkSuccesses, objects, exception := s.blockPackRepository.BulkCheckPermissionsAndGetManyByIds(
		checkInputs, nil, allowedPermissions,
		srepositories.WithDB(db.WithContext(ctx)), srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		return successes, nil, exception
	}
	objectsById := make(map[uuid.UUID]sschemas.BlockPack, len(objects))
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
