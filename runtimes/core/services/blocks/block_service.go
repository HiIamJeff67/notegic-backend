package blocks

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/blocks"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	cyjsworker "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	ssearchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"
	smetrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"
	seditableblock "github.com/HiIamJeff67/notegic-backend/shared/util/editableblock"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/core/exceptions"
)

type BlockServiceInterface interface {
	GetMyBlockById(ctx context.Context, requestDto *capi.GetMyBlockByIdRequestDto) (*capi.GetMyBlockByIdResponseDto, *cexceptions.Exception)
	GetMyBlocksByIds(ctx context.Context, requestDto *capi.GetMyBlocksByIdsRequestDto) (*capi.GetMyBlocksByIdsResponseDto, *cexceptions.Exception)
	GetMyBlocksByBlockPackId(ctx context.Context, requestDto *capi.GetMyBlocksByBlockPackIdRequestDto) (*capi.GetMyBlocksByBlockPackIdResponseDto, *cexceptions.Exception)

	Apply(ctx context.Context, blockPackId uuid.UUID, requestDto capi.ApplyBlockProjectionRequestDto) (*capi.ApplyBlockProjectionResponseDto, error)
	ApplyWithTransaction(ctx context.Context, tx *gorm.DB, blockPackId uuid.UUID, requestDto capi.ApplyBlockProjectionRequestDto) (*capi.ApplyBlockProjectionResponseDto, error)
	ApplyMany(ctx context.Context, requestDtos []capi.ApplyBlockProjectionDocumentRequestDto) (capi.ApplyBlockProjectionDocumentResponseDto, error)

	SearchPrivateBlocks(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchBlockInput) (*cgqlmodels.SearchBlockConnection, *cexceptions.Exception)
}

type BlockService struct {
	validator           *validator.Validate
	db                  *gorm.DB
	blockScope          sscopes.BlockScopeInterface
	blockPackScope      sscopes.BlockPackScopeInterface
	subShelfScope       sscopes.SubShelfScopeInterface
	blockPackRepository srepositories.BlockPackRepositoryInterface
	blockRepository     srepositories.BlockRepositoryInterface
}

func NewBlockService(
	validator *validator.Validate,
	db *gorm.DB,
	blockScope sscopes.BlockScopeInterface,
	blockPackScope sscopes.BlockPackScopeInterface,
	subShelfScope sscopes.SubShelfScopeInterface,
	blockPackRepository srepositories.BlockPackRepositoryInterface,
	blockRepository srepositories.BlockRepositoryInterface,
) BlockServiceInterface {
	return &BlockService{
		validator:           validator,
		db:                  db,
		blockScope:          blockScope,
		blockPackScope:      blockPackScope,
		subShelfScope:       subShelfScope,
		blockPackRepository: blockPackRepository,
		blockRepository:     blockRepository,
	}
}

/* ============================== Auxiliary Functions ============================== */

func newBlockResponseDto(block sschemas.Block) capi.BlockResponseDto {
	return capi.BlockResponseDto{
		Id:            block.Id,
		BlockPackId:   block.BlockPackId,
		ParentBlockId: block.ParentBlockId,
		PrevBlockId:   block.PrevBlockId,
		NextBlockId:   block.NextBlockId,
		Type:          block.Type.String(),
		Props:         json.RawMessage(block.Props),
		Content:       json.RawMessage(block.Content),
		UpdatedAt:     block.UpdatedAt,
		CreatedAt:     block.CreatedAt,
	}
}

/* ============================== Service Methods for Block ============================== */

func (s *BlockService) GetMyBlockById(
	ctx context.Context, requestDto *capi.GetMyBlockByIdRequestDto,
) (*capi.GetMyBlockByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewBlockException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	block, exception := s.blockRepository.GetOneById(
		requestDto.Param.BlockId,
		actorUserId,
		nil,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		return nil, exception
	}

	responseDto := newBlockResponseDto(*block)
	return &responseDto, nil
}

func (s *BlockService) GetMyBlocksByIds(
	ctx context.Context, requestDto *capi.GetMyBlocksByIdsRequestDto,
) (*capi.GetMyBlocksByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewBlockException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	blocks, exception := s.blockRepository.CheckPermissionsAndGetManyByIds(
		requestDto.Param.BlockIds,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		return nil, exception
	}

	responseDto := make(capi.GetMyBlocksByIdsResponseDto, len(blocks))
	for index, block := range blocks {
		responseDto[index] = newBlockResponseDto(block)
	}

	return &responseDto, nil
}

func (s *BlockService) GetMyBlocksByBlockPackId(
	ctx context.Context, requestDto *capi.GetMyBlocksByBlockPackIdRequestDto,
) (*capi.GetMyBlocksByBlockPackIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewBlockException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	if !s.blockPackRepository.HasPermission(
		requestDto.Param.BlockPackId,
		actorUserId,
		allowedPermissions,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	) {
		return nil, apiexceptions.NewBlockException().NoPermission("get the block pack of blocks")
	}

	var blocks []sschemas.Block
	if err := db.Model(&sschemas.Block{}).
		Where("block_pack_id = ?", requestDto.Param.BlockPackId).
		Order("created_at ASC").
		Order("id ASC").
		Find(&blocks).Error; err != nil {
		return nil, apiexceptions.NewBlockException().NotFound().WithOrigin(err)
	}

	responseDto := make(capi.GetMyBlocksByBlockPackIdResponseDto, len(blocks))
	for index, block := range blocks {
		responseDto[index] = newBlockResponseDto(block)
	}

	return &responseDto, nil
}

func (s *BlockService) Apply(
	ctx context.Context,
	blockPackId uuid.UUID,
	requestDto capi.ApplyBlockProjectionRequestDto,
) (*capi.ApplyBlockProjectionResponseDto, error) {
	tx := s.db.WithContext(ctx).Begin()

	responseDto, err := s.ApplyWithTransaction(ctx, tx, blockPackId, requestDto)
	if err != nil {
		tx.Rollback()

		return nil, err
	}
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("commit block projection: %w", err)
	}

	return responseDto, nil
}

func (s *BlockService) ApplyWithTransaction(
	ctx context.Context,
	tx *gorm.DB,
	blockPackId uuid.UUID,
	requestDto capi.ApplyBlockProjectionRequestDto,
) (*capi.ApplyBlockProjectionResponseDto, error) {
	if blockPackId == uuid.Nil {
		return nil, fmt.Errorf("block projection requires a block pack id")
	}
	if requestDto.SchemaId != cyjsworker.YjsBlockPackSchemaId ||
		requestDto.SchemaVersion != cyjsworker.YjsBlockPackSchemaVersion {
		return nil, fmt.Errorf("block projection source schema is not supported")
	}
	if requestDto.ProjectedSequence < 0 {
		return nil, fmt.Errorf("block projection target update sequence must not be negative")
	}

	flattenedBlocks, _, err := seditableblock.FlattenEditableBlocks(requestDto.Blocks)
	if err != nil {
		return nil, fmt.Errorf("failed to flatten block projection: %w", err)
	}

	blockIds := make([]uuid.UUID, len(flattenedBlocks))
	projectedBlocks := make([]sschemas.Block, len(flattenedBlocks))
	for index, flattenedBlock := range flattenedBlocks {
		blockIds[index] = flattenedBlock.Id
		projectedBlocks[index] = sschemas.Block{
			Id:            flattenedBlock.Id,
			BlockPackId:   blockPackId,
			ParentBlockId: flattenedBlock.ParentBlockId,
			PrevBlockId:   flattenedBlock.PrevBlockId,
			NextBlockId:   flattenedBlock.NextBlockId,
			Type:          cenums.BlockType(flattenedBlock.Type),
			Props:         datatypes.JSON(flattenedBlock.Props),
			Content:       datatypes.JSON(flattenedBlock.Content),
		}
	}

	lockingStrength := "UPDATE"
	var blockPack sschemas.BlockPack
	if err := tx.Model(&sschemas.BlockPack{}).
		Select("id").
		Scopes(sscopes.Locking(&lockingStrength)).
		Where("id = ? AND deleted_at IS NULL", blockPackId).
		First(&blockPack).Error; err != nil {
		return nil, fmt.Errorf("failed to lock block pack for projection: %w", err)
	}

	var document sschemas.BlockPackYjsDocument
	if err := tx.Model(&sschemas.BlockPackYjsDocument{}).
		Scopes(sscopes.Locking(&lockingStrength)).
		Where("block_pack_id = ? AND deleted_at IS NULL", blockPackId).
		First(&document).Error; err != nil {
		return nil, fmt.Errorf("failed to lock yjs document for projection: %w", err)
	}
	smetrics.NotegicMeter.Value(ctx, "yjs.projection.lag", document.LastUpdateSequence-document.ProjectedUntilSequence)

	if requestDto.ProjectedSequence <= document.ProjectedUntilSequence {
		return &capi.ApplyBlockProjectionResponseDto{
			Applied:                false,
			ProjectedUntilSequence: document.ProjectedUntilSequence,
		}, nil
	}

	if requestDto.ProjectedSequence > document.LastUpdateSequence {
		return nil, fmt.Errorf("block projection target update sequence exceeds durable yjs state")
	}

	type existingBlock struct {
		Id          uuid.UUID `gorm:"column:id"`
		BlockPackId uuid.UUID `gorm:"column:block_pack_id"`
	}

	existingBlocks := []existingBlock{}
	if len(blockIds) > 0 {
		if err := tx.Model(&sschemas.Block{}).
			Select("id, block_pack_id").
			Scopes(sscopes.Locking(&lockingStrength)).
			Where("id IN ?", blockIds).
			Find(&existingBlocks).Error; err != nil {
			return nil, fmt.Errorf("failed to lock projected blocks: %w", err)
		}
	}

	for _, existingBlock := range existingBlocks {
		if existingBlock.BlockPackId != blockPackId {
			return nil, fmt.Errorf("block projection contains an id owned by another block pack")
		}
	}

	now := time.Now()

	if len(projectedBlocks) > 0 {
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"block_pack_id":   blockPackId,
				"parent_block_id": gorm.Expr("EXCLUDED.parent_block_id"),
				"prev_block_id":   gorm.Expr("EXCLUDED.prev_block_id"),
				"next_block_id":   gorm.Expr("EXCLUDED.next_block_id"),
				"type":            gorm.Expr("EXCLUDED.type"),
				"props":           gorm.Expr("EXCLUDED.props"),
				"content":         gorm.Expr("EXCLUDED.content"),
				"updated_at":      now,
			}),
		}).CreateInBatches(&projectedBlocks, sconstants.MaxBatchCreateBlockSize).Error; err != nil {
			return nil, fmt.Errorf("failed to bulk upsert block projection: %w", err)
		}
	}

	deleteQuery := tx.Where("block_pack_id = ?", blockPackId)
	if len(blockIds) > 0 {
		deleteQuery = deleteQuery.Where("id NOT IN ?", blockIds)
	}
	if err := deleteQuery.Delete(&sschemas.Block{}).Error; err != nil {
		return nil, fmt.Errorf("failed to delete removed projected blocks: %w", err)
	}

	if err := tx.Model(&sschemas.BlockPackYjsDocument{}).
		Where("id = ?", document.Id).
		Updates(map[string]any{
			"projected_until_sequence": requestDto.ProjectedSequence,
			"updated_at":               now,
		}).Error; err != nil {
		return nil, fmt.Errorf("failed to update block projection checkpoint: %w", err)
	}

	return &capi.ApplyBlockProjectionResponseDto{
		Applied:                true,
		ProjectedUntilSequence: requestDto.ProjectedSequence,
	}, nil
}

func (s *BlockService) ApplyMany(
	ctx context.Context,
	requestDtos []capi.ApplyBlockProjectionDocumentRequestDto,
) (capi.ApplyBlockProjectionDocumentResponseDto, error) {
	if len(requestDtos) == 0 {
		return capi.ApplyBlockProjectionDocumentResponseDto{}, nil
	}

	type preparedProjection struct {
		blockPackId  uuid.UUID
		projectedSeq int64
		blocks       []sschemas.Block
		blockIds     []uuid.UUID
	}

	preparedProjections := make([]preparedProjection, 0, len(requestDtos))
	blockPackIdSet := make(map[uuid.UUID]bool, len(requestDtos))
	blockIdSet := make(map[uuid.UUID]bool)
	for _, requestDto := range requestDtos {
		if requestDto.BlockPackId == uuid.Nil {
			return nil, fmt.Errorf("block projection requires a block pack id")
		}
		if requestDto.Projection.SchemaId != cyjsworker.YjsBlockPackSchemaId ||
			requestDto.Projection.SchemaVersion != cyjsworker.YjsBlockPackSchemaVersion {
			return nil, fmt.Errorf("block projection source schema is not supported")
		}
		if requestDto.Projection.ProjectedSequence < 0 {
			return nil, fmt.Errorf("block projection target update sequence must not be negative")
		}
		if blockPackIdSet[requestDto.BlockPackId] {
			return nil, fmt.Errorf("duplicate block projection block pack id")
		}
		blockPackIdSet[requestDto.BlockPackId] = true

		flattenedBlocks, _, err := seditableblock.FlattenEditableBlocks(requestDto.Projection.Blocks)
		if err != nil {
			return nil, fmt.Errorf("failed to flatten block projection: %w", err)
		}

		blocks := make([]sschemas.Block, len(flattenedBlocks))
		blockIds := make([]uuid.UUID, len(flattenedBlocks))
		for index, flattenedBlock := range flattenedBlocks {
			if blockIdSet[flattenedBlock.Id] {
				return nil, fmt.Errorf("duplicate block id in projection batch")
			}
			blockIdSet[flattenedBlock.Id] = true
			blockIds[index] = flattenedBlock.Id
			blocks[index] = sschemas.Block{
				Id:            flattenedBlock.Id,
				BlockPackId:   requestDto.BlockPackId,
				ParentBlockId: flattenedBlock.ParentBlockId,
				PrevBlockId:   flattenedBlock.PrevBlockId,
				NextBlockId:   flattenedBlock.NextBlockId,
				Type:          cenums.BlockType(flattenedBlock.Type),
				Props:         datatypes.JSON(flattenedBlock.Props),
				Content:       datatypes.JSON(flattenedBlock.Content),
			}
		}

		preparedProjections = append(preparedProjections, preparedProjection{
			blockPackId:  requestDto.BlockPackId,
			projectedSeq: requestDto.Projection.ProjectedSequence,
			blocks:       blocks,
			blockIds:     blockIds,
		})
	}

	blockPackIds := make([]uuid.UUID, 0, len(blockPackIdSet))
	for blockPackId := range blockPackIdSet {
		blockPackIds = append(blockPackIds, blockPackId)
	}
	sort.Slice(blockPackIds, func(left, right int) bool {
		return string(blockPackIds[left][:]) < string(blockPackIds[right][:])
	})

	tx := s.db.WithContext(ctx).Begin()

	lockingStrength := srepositories.LockingStrengthUpdate
	var blockPacks []sschemas.BlockPack
	if err := tx.Model(&sschemas.BlockPack{}).
		Select("id").
		Where("id IN ? AND deleted_at IS NULL", blockPackIds).
		Scopes(sscopes.Locking(&lockingStrength)).
		Order("id ASC").
		Find(&blockPacks).Error; err != nil {
		tx.Rollback()

		return nil, fmt.Errorf("failed to lock block packs for projection: %w", err)
	}

	activeBlockPackIdSet := make(map[uuid.UUID]bool, len(blockPacks))
	for _, blockPack := range blockPacks {
		activeBlockPackIdSet[blockPack.Id] = true
	}

	var documents []sschemas.BlockPackYjsDocument
	if err := tx.Model(&sschemas.BlockPackYjsDocument{}).
		Where("block_pack_id IN ? AND deleted_at IS NULL", blockPackIds).
		Scopes(sscopes.Locking(&lockingStrength)).
		Order("block_pack_id ASC").
		Find(&documents).Error; err != nil {
		tx.Rollback()

		return nil, fmt.Errorf("failed to lock yjs documents for projection: %w", err)
	}

	documentByBlockPackId := make(map[uuid.UUID]sschemas.BlockPackYjsDocument, len(documents))
	for _, document := range documents {
		documentByBlockPackId[document.BlockPackId] = document
		smetrics.NotegicMeter.Value(ctx, "yjs.projection.lag", document.LastUpdateSequence-document.ProjectedUntilSequence)
	}

	applicableProjections := make([]preparedProjection, 0, len(preparedProjections))
	for _, projection := range preparedProjections {
		if !activeBlockPackIdSet[projection.blockPackId] {
			continue
		}
		document, exists := documentByBlockPackId[projection.blockPackId]
		if !exists {
			continue
		}
		if projection.projectedSeq > document.LastUpdateSequence {
			tx.Rollback()

			return nil, fmt.Errorf("block projection target update sequence exceeds durable yjs state")
		}
		if projection.projectedSeq <= document.ProjectedUntilSequence {
			continue
		}
		applicableProjections = append(applicableProjections, projection)
	}
	if len(applicableProjections) == 0 {
		if err := tx.Commit().Error; err != nil {
			return nil, fmt.Errorf("failed to commit stale block projections: %w", err)
		}

		return capi.ApplyBlockProjectionDocumentResponseDto{}, nil
	}

	projectedBlockIds := make([]uuid.UUID, 0)
	allProjectedBlocks := make([]sschemas.Block, 0)
	applicableBlockPackIds := make([]uuid.UUID, 0, len(applicableProjections))
	for _, projection := range applicableProjections {
		applicableBlockPackIds = append(applicableBlockPackIds, projection.blockPackId)
		projectedBlockIds = append(projectedBlockIds, projection.blockIds...)
		allProjectedBlocks = append(allProjectedBlocks, projection.blocks...)
	}

	type existingBlock struct {
		Id          uuid.UUID `gorm:"column:id"`
		BlockPackId uuid.UUID `gorm:"column:block_pack_id"`
	}
	var existingBlocks []existingBlock
	if len(projectedBlockIds) > 0 {
		if err := tx.Model(&sschemas.Block{}).
			Select("id, block_pack_id").
			Where("id IN ?", projectedBlockIds).
			Scopes(sscopes.Locking(&lockingStrength)).
			Find(&existingBlocks).Error; err != nil {
			tx.Rollback()

			return nil, fmt.Errorf("failed to lock projected blocks: %w", err)
		}
	}

	projectedBlockPackIdById := make(map[uuid.UUID]uuid.UUID, len(allProjectedBlocks))
	for _, block := range allProjectedBlocks {
		projectedBlockPackIdById[block.Id] = block.BlockPackId
	}
	for _, existingBlock := range existingBlocks {
		if projectedBlockPackIdById[existingBlock.Id] != existingBlock.BlockPackId {
			tx.Rollback()

			return nil, fmt.Errorf("block projection contains an id owned by another block pack")
		}
	}

	deleteQuery := tx.Where("block_pack_id IN ?", applicableBlockPackIds)
	if len(projectedBlockIds) > 0 {
		deleteQuery = deleteQuery.Where("id NOT IN ?", projectedBlockIds)
	}
	result := deleteQuery.Delete(&sschemas.Block{})
	if err := result.Error; err != nil {
		tx.Rollback()

		return nil, fmt.Errorf("failed to delete removed projected blocks: %w", err)
	}

	now := time.Now()
	if len(allProjectedBlocks) > 0 {
		if err := tx.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"block_pack_id":   gorm.Expr("EXCLUDED.block_pack_id"),
				"parent_block_id": gorm.Expr("EXCLUDED.parent_block_id"),
				"prev_block_id":   gorm.Expr("EXCLUDED.prev_block_id"),
				"next_block_id":   gorm.Expr("EXCLUDED.next_block_id"),
				"type":            gorm.Expr("EXCLUDED.type"),
				"props":           gorm.Expr("EXCLUDED.props"),
				"content":         gorm.Expr("EXCLUDED.content"),
				"updated_at":      now,
			}),
		}).CreateInBatches(&allProjectedBlocks, sconstants.MaxBatchCreateBlockSize).Error; err != nil {
			tx.Rollback()

			return nil, fmt.Errorf("failed to bulk upsert block projections: %w", err)
		}
	}

	valueRows := make([]string, 0, len(applicableProjections))
	args := make([]any, 0, len(applicableProjections)*2)
	for _, projection := range applicableProjections {
		valueRows = append(valueRows, "(?::uuid, ?::bigint)")
		args = append(args, projection.blockPackId, projection.projectedSeq)
	}
	args = append(args, now)
	updateQuery := `
		WITH target(block_pack_id, projected_until_sequence) AS (
			VALUES ` + strings.Join(valueRows, ",") + `
		)
		UPDATE "BlockPackYjsDocumentTable" AS document
		SET projected_until_sequence = target.projected_until_sequence, updated_at = ?
		FROM target
		WHERE document.block_pack_id = target.block_pack_id
			AND document.deleted_at IS NULL
			AND document.projected_until_sequence < target.projected_until_sequence
			AND document.last_update_sequence >= target.projected_until_sequence
		RETURNING document.block_pack_id`

	var appliedDocuments []struct {
		BlockPackId uuid.UUID `gorm:"column:block_pack_id"`
	}
	result = tx.Raw(updateQuery, args...).Scan(&appliedDocuments)
	if err := result.Error; err != nil {
		tx.Rollback()

		return nil, fmt.Errorf("failed to update block projection checkpoints: %w", err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit block projections: %w", err)
	}

	appliedBlockPackIds := make([]uuid.UUID, len(appliedDocuments))
	for index, document := range appliedDocuments {
		appliedBlockPackIds[index] = document.BlockPackId
	}

	return capi.ApplyBlockProjectionDocumentResponseDto(appliedBlockPackIds), nil
}

func (s *BlockService) SearchPrivateBlocks(
	ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchBlockInput,
) (*cgqlmodels.SearchBlockConnection, *cexceptions.Exception) {
	startTime := time.Now()

	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	query := db.Model(&sschemas.Block{}).
		Select(`"BlockTable".*`).
		Joins(`INNER JOIN "BlockPackTable" ON "BlockPackTable".id = "BlockTable".block_pack_id`).
		Joins(`INNER JOIN "SubShelfTable" ON "SubShelfTable".id = "BlockPackTable".parent_sub_shelf_id`).
		Joins(`INNER JOIN "UsersToShelvesTable" uts ON uts.root_shelf_id = "SubShelfTable".root_shelf_id`).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, allowedPermissions).
		Scopes(s.blockPackScope.FilterOnlyDeleted(stypes.Ternary_Negative)).
		Scopes(s.subShelfScope.FilterOnlyDeleted(stypes.Ternary_Negative)).
		Scopes(s.blockScope.IncludePreloads([]sschemas.BlockRelation{sschemas.BlockRelation_Children}))

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		pattern := "%" + gqlInput.Query + "%"
		query = query.Where(`"BlockTable".content::text ILIKE ? OR "BlockTable".props::text ILIKE ? OR "BlockTable".type::text ILIKE ?`, pattern, pattern, pattern)
	}

	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := ssearchcursor.Decode[cgqlmodels.SearchBlockCursorFields](*gqlInput.After)
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToDecode().WithOrigin(err)
		}

		query = query.Where(`"BlockTable".id > ?`, searchCursor.Fields.ID)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		cending := cgqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == cgqlmodels.SearchSortOrderDesc {
			cending = cgqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case cgqlmodels.SearchBlockSortByType:
			query = query.Order(`"BlockTable".type ` + cending).Order(`"BlockTable".updated_at ` + cending).Order(`"BlockTable".created_at ` + cending)
		case cgqlmodels.SearchBlockSortByLastUpdate:
			query = query.Order(`"BlockTable".updated_at ` + cending).Order(`"BlockTable".type ` + cending).Order(`"BlockTable".created_at ` + cending)
		case cgqlmodels.SearchBlockSortByCreatedAt:
			query = query.Order(`"BlockTable".created_at ` + cending).Order(`"BlockTable".type ` + cending).Order(`"BlockTable".updated_at ` + cending)
		default:
			query = query.Order(`"BlockTable".type ` + cending).Order(`"BlockTable".updated_at ` + cending).Order(`"BlockTable".created_at ` + cending)
		}
	}

	limit := sconstants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, sconstants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var blocks []sschemas.Block
	if err := query.Find(&blocks).Error; err != nil {
		return nil, apiexceptions.NewBlockException().NotFound().WithOrigin(err)
	}

	hasNextPage := len(blocks) > limit
	searchEdges := make([]*cgqlmodels.SearchBlockEdge, len(blocks))
	for index, block := range blocks {
		searchCursor := ssearchcursor.SearchCursor[cgqlmodels.SearchBlockCursorFields]{Fields: cgqlmodels.SearchBlockCursorFields{ID: block.Id}}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToEncode().WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, apiexceptions.NewSearchException().FailedToUnmarshalSearchCursor()
		}

		searchEdges[index] = &cgqlmodels.SearchBlockEdge{EncodedSearchCursor: *encodedSearchCursor, Node: block.ToPrivateBlock()}
	}

	searchPageInfo := &cgqlmodels.SearchPageInfo{
		HasNextPage:     hasNextPage,
		HasPreviousPage: gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0,
	}

	if hasNextPage {
		searchEdges = searchEdges[:limit]
	}

	if len(searchEdges) > 0 {
		searchPageInfo.StartEncodedSearchCursor = &searchEdges[0].EncodedSearchCursor
		searchPageInfo.EndEncodedSearchCursor = &searchEdges[len(searchEdges)-1].EncodedSearchCursor
	}

	searchTime := float64(time.Since(startTime).Nanoseconds()) / 1e6

	return &cgqlmodels.SearchBlockConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
