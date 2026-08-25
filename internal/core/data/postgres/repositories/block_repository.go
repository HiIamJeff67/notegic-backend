package repositories

import (
	"fmt"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	array "github.com/HiIamJeff67/notegic-backend/shared/lib/array"
	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type BlockRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.BlockRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) (*schemas.Block, *cexceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.BlockRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]schemas.Block, *cexceptions.Exception)
	GetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.BlockRelation, opts ...options.RepositoryOptions) (*schemas.Block, *cexceptions.Exception)

	/* ============================== System Only Method ============================== */

	BulkCheckPermissionsAndGetManyByIds(inputs []inputs.BulkCheckBlockPermissionInput, preloads []schemas.BlockRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]bool, []schemas.Block, *cexceptions.Exception)
	BulkCreateMany(inputs []inputs.BulkCreateBlockPackContentInput, opts ...options.RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkUpdateMany(inputs []inputs.BulkUpdateBlockInput, opts ...options.RepositoryOptions) ([]bool, *cexceptions.Exception)
}

type BlockRepository struct {
	blockScope scopes.BlockScopeInterface
}

func NewBlockRepository(blockScope scopes.BlockScopeInterface) BlockRepositoryInterface {
	return &BlockRepository{
		blockScope: blockScope,
	}
}

func (r *BlockRepository) HasPermission(
	id uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	if parsedOptions.DB == nil {
		parsedOptions.DB = data.DB
	}

	var marker int
	result := parsedOptions.DB.
		Model(&schemas.Block{}).
		Select("1").
		Scopes(r.blockScope.PassPermissionCheck(id, userId, allowedPermissions)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Limit(1).
		Scan(&marker)
	if result.Error != nil {
		return false
	}

	return marker == 1
}

func (r *BlockRepository) HavePermissions(
	ids []uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	if parsedOptions.DB == nil {
		parsedOptions.DB = data.DB
	}

	var permittedIds []uuid.UUID
	result := parsedOptions.DB.
		Model(&schemas.Block{}).
		Select(`DISTINCT "BlockTable".id`).
		Scopes(r.blockScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&permittedIds)
	if result.Error != nil {
		return false
	}

	return array.GetDistinctCount(ids) == array.GetDistinctCount(permittedIds)
}

func (r *BlockRepository) CheckPermissionAndGetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.BlockRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) (*schemas.Block, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	if parsedOptions.DB == nil {
		parsedOptions.DB = data.DB
	}

	var block schemas.Block
	query := parsedOptions.DB.
		Model(&schemas.Block{}).
		Where(`"BlockTable".id = ?`, id)
	if allowedPermissions != nil && len(allowedPermissions) > 0 {
		query = query.Scopes(
			r.blockScope.PassPermissionCheck(id, userId, allowedPermissions),
		)
	}

	result := query.
		Scopes(r.blockScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&block)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockException().NotFound().WithOrigin(result.Error)},
		{First: block.Id == uuid.Nil, Second: apiexceptions.NewBlockException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &block, nil
}

func (r *BlockRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.BlockRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]schemas.Block, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	if parsedOptions.DB == nil {
		parsedOptions.DB = data.DB
	}

	var blocks []schemas.Block
	result := parsedOptions.DB.
		Model(&schemas.Block{}).
		Scopes(r.blockScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.blockScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&blocks)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockException().NotFound().WithOrigin(result.Error)},
		{First: len(blocks) == 0, Second: apiexceptions.NewBlockException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return blocks, nil
}

func (r *BlockRepository) GetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.BlockRelation,
	opts ...options.RepositoryOptions,
) (*schemas.Block, *cexceptions.Exception) {
	return r.CheckPermissionAndGetOneById(
		id,
		userId,
		preloads,
		options.ParseRepositoryOptions(opts...).AllowedPermissions,
		opts...,
	)
}

/* ============================== System Only Method ============================== */

func (r *BlockRepository) BulkCheckPermissionsAndGetManyByIds(
	bulkInputs []inputs.BulkCheckBlockPermissionInput,
	preloads []schemas.BlockRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]bool, []schemas.Block, *cexceptions.Exception) {
	if len(bulkInputs) == 0 {
		return []bool{}, []schemas.Block{}, nil
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)
	if parsedOptions.DB == nil {
		parsedOptions.DB = data.DB
	}

	successes := make([]bool, len(bulkInputs))
	ids := make([]uuid.UUID, 0, len(bulkInputs))
	userIds := make([]uuid.UUID, 0, len(bulkInputs))
	for _, bulkInput := range bulkInputs {
		ids = append(ids, bulkInput.Id)
		userIds = append(userIds, bulkInput.UserId)
	}

	var validTargets []struct {
		Id     uuid.UUID `gorm:"column:id"`
		UserId uuid.UUID `gorm:"column:user_id"`
	}
	result := parsedOptions.DB.Model(&schemas.Block{}).
		Select(`"BlockTable".id, uts.user_id`).
		Joins(`INNER JOIN "BlockPackTable" AS bp ON bp.id = "BlockTable".block_pack_id`).
		Joins(`INNER JOIN "SubShelfTable" AS ss ON ss.id = bp.parent_sub_shelf_id`).
		Joins(`INNER JOIN "UsersToShelvesTable" AS uts ON uts.root_shelf_id = ss.root_shelf_id`).
		Where(`"BlockTable".id IN ?`, ids).
		Where("bp.deleted_at IS NULL").
		Where("uts.user_id IN ? AND uts.permission IN ?", userIds, allowedPermissions).
		Scan(&validTargets)
	if result.Error != nil {
		return nil, nil, apiexceptions.NewBlockException().NotFound().WithOrigin(result.Error)
	}

	validTargetByUserId := make(map[[2]uuid.UUID]bool, len(validTargets))
	for _, validTarget := range validTargets {
		validTargetByUserId[[2]uuid.UUID{validTarget.Id, validTarget.UserId}] = true
	}

	validIdSet := make(map[uuid.UUID]bool, len(validTargets))
	for _, bulkInput := range bulkInputs {
		if validTargetByUserId[[2]uuid.UUID{bulkInput.Id, bulkInput.UserId}] {
			validIdSet[bulkInput.Id] = true
		}
	}

	validIds := make([]uuid.UUID, 0, len(validIdSet))
	for validId := range validIdSet {
		validIds = append(validIds, validId)
	}
	if len(validIds) == 0 {
		return successes, []schemas.Block{}, nil
	}

	var blocks []schemas.Block
	result = parsedOptions.DB.Model(&schemas.Block{}).
		Where(`"BlockTable".id IN ?`, validIds).
		Scopes(r.blockScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&blocks)
	if result.Error != nil {
		return nil, nil, apiexceptions.NewBlockException().NotFound().WithOrigin(result.Error)
	}

	foundIdSet := make(map[uuid.UUID]bool, len(blocks))
	for _, block := range blocks {
		foundIdSet[block.Id] = true
	}
	for index, bulkInput := range bulkInputs {
		if validTargetByUserId[[2]uuid.UUID{bulkInput.Id, bulkInput.UserId}] && foundIdSet[bulkInput.Id] {
			successes[index] = true
		}
	}

	return successes, blocks, nil
}

func (r *BlockRepository) BulkCreateMany(
	bulkInputs []inputs.BulkCreateBlockPackContentInput,
	opts ...options.RepositoryOptions,
) ([]bool, *cexceptions.Exception) {
	if len(bulkInputs) == 0 {
		return []bool{}, apiexceptions.NewBlockException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)
	if parsedOptions.DB == nil {
		parsedOptions.DB = data.DB
	}

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	checkInputs := make([]inputs.BulkCheckBlockPackPermissionInput, len(bulkInputs))
	for index, bulkInput := range bulkInputs {
		if bulkInput.BlockPackId == uuid.Nil {
			continue
		}
		checkInputs[index] = inputs.BulkCheckBlockPackPermissionInput{
			UserId: bulkInput.UserId,
			Id:     bulkInput.BlockPackId,
		}
	}

	blockPackRepository := NewBlockPackRepository(scopes.NewBlockPackScope())
	checkOptions := append(opts, options.WithTransactionDB(parsedOptions.DB))
	checkOptions = append(checkOptions, options.WithOnlyDeleted(types.Ternary_Negative))
	checkOptions = append(checkOptions, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	checkOptions = append(checkOptions, options.WithAllowedPermissions(parsedOptions.AllowedPermissions))
	successes, _, exception := blockPackRepository.BulkCheckPermissionsAndGetManyByIds(
		checkInputs,
		nil,
		parsedOptions.AllowedPermissions,
		checkOptions...,
	)
	if exception != nil {
		parsedOptions.DB.Rollback()

		return nil, exception
	}

	newBlocks := make([]schemas.Block, 0)
	for index, bulkInput := range bulkInputs {
		if bulkInput.BlockPackId == uuid.Nil || !successes[index] || len(bulkInput.Blocks) == 0 {
			successes[index] = false
			continue
		}

		for _, inputBlock := range bulkInput.Blocks {
			newBlocks = append(newBlocks, schemas.Block{
				Id:            inputBlock.Id,
				BlockPackId:   bulkInput.BlockPackId,
				ParentBlockId: inputBlock.ParentBlockId,
				PrevBlockId:   inputBlock.PrevBlockId,
				NextBlockId:   inputBlock.NextBlockId,
				Type:          inputBlock.Type,
				Props:         inputBlock.Props,
				Content:       inputBlock.Content,
			})
		}
	}
	if len(newBlocks) == 0 {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}

		return successes, nil
	}

	result := parsedOptions.DB.Model(&schemas.Block{}).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		CreateInBatches(&newBlocks, parsedOptions.BatchSize)
	if result.Error != nil {
		parsedOptions.DB.Rollback()

		return nil, apiexceptions.NewBlockException().FailedToCreate().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()

			return nil, apiexceptions.NewBlockException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return successes, nil
}

func (r *BlockRepository) BulkUpdateMany(
	bulkInputs []inputs.BulkUpdateBlockInput,
	opts ...options.RepositoryOptions,
) ([]bool, *cexceptions.Exception) {
	if len(bulkInputs) == 0 {
		return []bool{}, apiexceptions.NewBlockException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)
	if parsedOptions.DB == nil {
		parsedOptions.DB = data.DB
	}

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	checkInputs := make([]inputs.BulkCheckBlockPermissionInput, len(bulkInputs))
	for index, bulkInput := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckBlockPermissionInput{
			UserId: bulkInput.UserId,
			Id:     bulkInput.Id,
		}
	}
	checkOptions := append(opts, options.WithTransactionDB(parsedOptions.DB))
	checkOptions = append(checkOptions, options.WithOnlyDeleted(types.Ternary_Negative))
	checkOptions = append(checkOptions, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	successes, _, exception := r.BulkCheckPermissionsAndGetManyByIds(
		checkInputs,
		nil,
		parsedOptions.AllowedPermissions,
		checkOptions...,
	)
	if exception != nil {
		parsedOptions.DB.Rollback()

		return nil, exception
	}

	valuePlaceholders := make([]string, 0, len(bulkInputs))
	valueArgs := make([]any, 0, len(bulkInputs)*12)
	for index, bulkInput := range bulkInputs {
		if !successes[index] {
			continue
		}
		if bulkInput.PartialUpdateInput.Values.BlockPackId != nil && *bulkInput.PartialUpdateInput.Values.BlockPackId == uuid.Nil {
			successes[index] = false
			continue
		}

		setParentBlockIdNull := partialupdate.CheckSetNull(bulkInput.PartialUpdateInput.SetNull, "ParentBlockId")
		setPrevBlockIdNull := partialupdate.CheckSetNull(bulkInput.PartialUpdateInput.SetNull, "PrevBlockId")
		setNextBlockIdNull := partialupdate.CheckSetNull(bulkInput.PartialUpdateInput.SetNull, "NextBlockId")
		valuePlaceholders = append(valuePlaceholders, `(?::int, ?::uuid, ?::"BlockType", ?::jsonb, ?::jsonb, ?::uuid, ?::uuid, ?::uuid, ?::uuid, ?::boolean, ?::boolean, ?::boolean)`)
		valueArgs = append(valueArgs,
			index,
			bulkInput.Id,
			bulkInput.PartialUpdateInput.Values.Type,
			bulkInput.PartialUpdateInput.Values.Props,
			bulkInput.PartialUpdateInput.Values.Content,
			bulkInput.PartialUpdateInput.Values.BlockPackId,
			bulkInput.PartialUpdateInput.Values.ParentBlockId,
			bulkInput.PartialUpdateInput.Values.PrevBlockId,
			bulkInput.PartialUpdateInput.Values.NextBlockId,
			setParentBlockIdNull,
			setPrevBlockIdNull,
			setNextBlockIdNull,
		)
	}
	if len(valuePlaceholders) == 0 {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}

		return successes, nil
	}

	sql := fmt.Sprintf(`
		WITH payload(idx, id, type, props, content, block_pack_id, parent_block_id, prev_block_id, next_block_id, set_parent_block_id_null, set_prev_block_id_null, set_next_block_id_null) AS (
			VALUES %s
		),
		updated AS (
			UPDATE "BlockTable" AS b
			SET
				type = COALESCE(v.type::"BlockType", b.type),
				props = COALESCE(v.props::jsonb, b.props),
				content = COALESCE(v.content::jsonb, b.content),
				block_pack_id = COALESCE(v.block_pack_id::uuid, b.block_pack_id),
				parent_block_id = CASE
					WHEN v.set_parent_block_id_null::boolean THEN NULL
					ELSE COALESCE(v.parent_block_id::uuid, b.parent_block_id)
				END,
				prev_block_id = CASE
					WHEN v.set_prev_block_id_null::boolean THEN NULL
					ELSE COALESCE(v.prev_block_id::uuid, b.prev_block_id)
				END,
				next_block_id = CASE
					WHEN v.set_next_block_id_null::boolean THEN NULL
					ELSE COALESCE(v.next_block_id::uuid, b.next_block_id)
				END,
				updated_at = NOW()
			FROM payload AS v
			WHERE b.id = v.id::uuid
			RETURNING b.id
		)
		SELECT v.idx
		FROM payload AS v
		INNER JOIN updated AS u ON u.id = v.id::uuid
	`, strings.Join(valuePlaceholders, ","))

	var updatedIndexes []struct {
		Index int `gorm:"column:idx"`
	}
	result := parsedOptions.DB.Raw(sql, valueArgs...).Scan(&updatedIndexes)
	if result.Error != nil {
		parsedOptions.DB.Rollback()

		return nil, apiexceptions.NewBlockException().FailedToUpdate().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()

			return nil, apiexceptions.NewBlockException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	successes = make([]bool, len(bulkInputs))
	for _, updatedIndex := range updatedIndexes {
		if updatedIndex.Index >= 0 && updatedIndex.Index < len(successes) {
			successes[updatedIndex.Index] = true
		}
	}

	return successes, nil
}
