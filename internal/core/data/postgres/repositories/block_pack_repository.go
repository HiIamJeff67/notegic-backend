package repositories

import (
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	pg "github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	array "github.com/HiIamJeff67/notegic-backend/shared/lib/array"
	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"

	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	scopes "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/scopes"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
)

type BlockPackRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.BlockPackRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) (*schemas.BlockPack, *exceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.BlockPackRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]schemas.BlockPack, *exceptions.Exception)
	CheckPermissionAndGetOneWithOwnerIdById(id uuid.UUID, userId uuid.UUID, preloads []schemas.BlockPackRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) (*uuid.UUID, *schemas.BlockPack, *exceptions.Exception)
	CheckPermissionsAndGetManyWithOwnerIdsByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.BlockPackRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]uuid.UUID, []schemas.BlockPack, *exceptions.Exception)
	GetOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) (*schemas.BlockPack, *exceptions.Exception)
	GetManyByRootShelfIds(rootShelfIds []uuid.UUID, opts ...options.RepositoryOptions) ([]schemas.BlockPack, *exceptions.Exception)
	GetIdsByParentSubShelfIds(parentSubShelfIds []uuid.UUID, opts ...options.RepositoryOptions) ([]uuid.UUID, *exceptions.Exception)
	GetIdsBySubShelfIdsAndDescendants(subShelfIds []uuid.UUID, opts ...options.RepositoryOptions) ([]uuid.UUID, *exceptions.Exception)
	CreateOneBySubShelfId(subShelfId uuid.UUID, userId uuid.UUID, input inputs.CreateBlockPackInput, opts ...options.RepositoryOptions) (*uuid.UUID, *exceptions.Exception)
	CreateManyBySubShelfIds(userId uuid.UUID, input []inputs.CreateBlockPackBySubShelfIdInput, opts ...options.RepositoryOptions) ([]uuid.UUID, *exceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateBlockPackInput, opts ...options.RepositoryOptions) (*schemas.BlockPack, *exceptions.Exception)
	UpdateManyByIds(userId uuid.UUID, input []inputs.UpdateBlockPackByIdInput, opts ...options.RepositoryOptions) *exceptions.Exception
	RestoreSoftDeletedOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) (*schemas.BlockPack, *exceptions.Exception)
	RestoreSoftDeletedManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) ([]schemas.BlockPack, *exceptions.Exception)
	SoftDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *exceptions.Exception
	SoftDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *exceptions.Exception
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *exceptions.Exception
	HardDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *exceptions.Exception

	/* ============================== System Only Method ============================== */

	BulkCheckPermissionsAndGetManyByIds(inputs []inputs.BulkCheckBlockPackPermissionInput, preloads []schemas.BlockPackRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]bool, []schemas.BlockPack, *exceptions.Exception)
	BulkCreateMany(inputs []inputs.BulkCreateBlockPackInput, opts ...options.RepositoryOptions) ([]bool, *exceptions.Exception)
	BulkUpdateMany(inputs []inputs.BulkUpdateBlockPackInput, opts ...options.RepositoryOptions) ([]bool, *exceptions.Exception)
	BulkDeleteMany(inputs []inputs.BulkDeleteBlockPackInput, opts ...options.RepositoryOptions) ([]bool, *exceptions.Exception)
}

type BlockPackRepository struct {
	blockPackScope scopes.BlockPackScopeInterface
}

func NewBlockPackRepository(blockPackScope scopes.BlockPackScopeInterface) BlockPackRepositoryInterface {
	return &BlockPackRepository{
		blockPackScope: blockPackScope,
	}
}

func (r *BlockPackRepository) HasPermission(
	id uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var marker int
	result := parsedOptions.DB.
		Model(&schemas.BlockPack{}).
		Select("1").
		Scopes(r.blockPackScope.PassPermissionCheck(id, userId, allowedPermissions)).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Limit(1).
		Scan(&marker)
	if err := result.Error; err != nil {
		return false
	}

	return marker == 1
}

func (r *BlockPackRepository) HavePermissions(
	ids []uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var permittedIds []uuid.UUID
	result := parsedOptions.DB.
		Model(&schemas.BlockPack{}).
		Select(`DISTINCT "BlockPackTable".id`).
		Scopes(r.blockPackScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&permittedIds)
	if err := result.Error; err != nil {
		return false
	}

	return array.GetDistinctCount(ids) == array.GetDistinctCount(permittedIds)
}

func (r *BlockPackRepository) CheckPermissionAndGetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.BlockPackRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) (*schemas.BlockPack, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var blockPack schemas.BlockPack
	query := parsedOptions.DB.
		Model(&schemas.BlockPack{}).
		Where(`"BlockPackTable".id = ?`, id)
	if allowedPermissions != nil && len(allowedPermissions) > 0 {
		query = query.Scopes(
			r.blockPackScope.PassPermissionCheck(id, userId, allowedPermissions),
		)
	}

	result := query.
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.blockPackScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&blockPack)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockPackException().NotFound().WithOrigin(result.Error)},
		{First: blockPack.Id == uuid.Nil, Second: apiexceptions.NewBlockPackException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &blockPack, nil
}

func (r *BlockPackRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.BlockPackRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]schemas.BlockPack, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var blockPacks []schemas.BlockPack
	result := parsedOptions.DB.
		Model(&schemas.BlockPack{}).
		Scopes(r.blockPackScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.blockPackScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&blockPacks)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockPackException().NotFound().WithOrigin(result.Error)},
		{First: len(blockPacks) == 0, Second: apiexceptions.NewBlockPackException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return blockPacks, nil
}

func (r *BlockPackRepository) CheckPermissionAndGetOneWithOwnerIdById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.BlockPackRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) (*uuid.UUID, *schemas.BlockPack, *exceptions.Exception) { // we should also return the owner id for the block groups and blocks
	parsedOptions := options.ParseRepositoryOptions(opts...)
	query := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Select(`"BlockPackTable".*, owner_uts.user_id AS owner_id`).
		Joins(`INNER JOIN "SubShelfTable" ss ON parent_sub_shelf_id = ss.id`).
		// inner join the owner's user to shelves table to extract owner's id
		// note that this should be attach AFTER we have join the SubShelfTable of ss
		// so we can't use PassPermissionCheck scope
		Joins(`INNER JOIN "UsersToShelvesTable" owner_uts ON ss.root_shelf_id = owner_uts.root_shelf_id AND owner_uts.permission = 'Owner'`).
		Where(`"BlockPackTable".id = ?`, id).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.blockPackScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength))
	if allowedPermissions != nil && len(allowedPermissions) > 0 {
		subQuery := parsedOptions.DB.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToShelves{}).
			Select("1").
			Where("root_shelf_id = ss.root_shelf_id").
			Where("user_id = ? AND permission IN ?", userId, allowedPermissions)
		query = query.Where("EXISTS (?)", subQuery)
	}

	var blockPackWithOwnerId struct {
		schemas.BlockPack
		OwnerId uuid.UUID `gorm:"column:owner_id;"`
	}
	result := query.First(&blockPackWithOwnerId)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockPackException().NotFound().WithOrigin(result.Error)},
		{First: blockPackWithOwnerId.OwnerId == uuid.Nil, Second: apiexceptions.NewBlockPackException().NotFound()},
	}); exception != nil {
		return nil, nil, exception
	}

	return &blockPackWithOwnerId.OwnerId, &blockPackWithOwnerId.BlockPack, nil
}

func (r *BlockPackRepository) CheckPermissionsAndGetManyWithOwnerIdsByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.BlockPackRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]uuid.UUID, []schemas.BlockPack, *exceptions.Exception) { // we should also return the owner id for the block groups and blocks
	parsedOptions := options.ParseRepositoryOptions(opts...)
	query := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Select(`"BlockPackTable".*, owner_uts.user_id AS owner_id`).
		Joins(`INNER JOIN "SubShelfTable" ss ON parent_sub_shelf_id = ss.id`).
		// inner join the owner's user to shelves table to extract owner's id
		// note that this should be attach AFTER we have join the SubShelfTable of ss
		// so we can't use PassPermissionChecks scope
		Joins(`INNER JOIN "UsersToShelvesTable" owner_uts ON ss.root_shelf_id = owner_uts.root_shelf_id AND owner_uts.permission = 'Owner'`).
		Where(`"BlockPackTable".id IN ?`, ids).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.blockPackScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength))
	if allowedPermissions != nil && len(allowedPermissions) > 0 {
		subQuery := parsedOptions.DB.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToShelves{}).
			Select("1").
			Where("root_shelf_id = ss.root_shelf_id").
			Where("user_id = ? AND permission IN ?", userId, allowedPermissions)
		query = query.Where("EXISTS (?)", subQuery)
	}

	var blockPacksWithOwnerIds []struct {
		schemas.BlockPack
		ownerId uuid.UUID `gorm:"column:owner_id;"`
	}
	result := query.Find(&blockPacksWithOwnerIds)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockPackException().NotFound().WithOrigin(result.Error)},
		{First: len(blockPacksWithOwnerIds) == 0, Second: apiexceptions.NewBlockPackException().NotFound()},
	}); exception != nil {
		return nil, nil, exception
	}

	ownerIds := make([]uuid.UUID, len(blockPacksWithOwnerIds))
	blockPacks := make([]schemas.BlockPack, len(blockPacksWithOwnerIds))
	for index, element := range blockPacksWithOwnerIds {
		ownerIds[index] = element.ownerId
		blockPacks[index] = element.BlockPack
	}

	return ownerIds, blockPacks, nil
}

func (r *BlockPackRepository) GetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) (*schemas.BlockPack, *exceptions.Exception) {
	return r.CheckPermissionAndGetOneById(
		id,
		userId,
		nil,
		options.ParseRepositoryOptions(opts...).AllowedPermissions,
		opts...,
	)
}

func (r *BlockPackRepository) GetManyByRootShelfIds(
	rootShelfIds []uuid.UUID,
	opts ...options.RepositoryOptions,
) ([]schemas.BlockPack, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var blockPacks []schemas.BlockPack
	result := parsedOptions.DB.
		Model(&schemas.BlockPack{}).
		Joins(`INNER JOIN "SubShelfTable" ON "SubShelfTable".id = "BlockPackTable".parent_sub_shelf_id`).
		Where(`"SubShelfTable".root_shelf_id IN ?`, rootShelfIds).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&blockPacks)
	if result.Error != nil {
		return nil, apiexceptions.NewBlockPackException().NotFound().WithOrigin(result.Error)
	}

	return blockPacks, nil
}

func (r *BlockPackRepository) GetIdsByParentSubShelfIds(
	parentSubShelfIds []uuid.UUID,
	opts ...options.RepositoryOptions,
) ([]uuid.UUID, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var blockPackIds []uuid.UUID
	result := parsedOptions.DB.
		Model(&schemas.BlockPack{}).
		Select("id").
		Where("parent_sub_shelf_id IN ?", parentSubShelfIds).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Find(&blockPackIds)
	if result.Error != nil {
		return nil, apiexceptions.NewBlockPackException().NotFound().WithOrigin(result.Error)
	}

	return blockPackIds, nil
}

func (r *BlockPackRepository) GetIdsBySubShelfIdsAndDescendants(
	subShelfIds []uuid.UUID,
	opts ...options.RepositoryOptions,
) ([]uuid.UUID, *exceptions.Exception) {
	if len(subShelfIds) == 0 {
		return []uuid.UUID{}, nil
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)
	var blockPackIds []uuid.UUID
	result := parsedOptions.DB.
		Model(&schemas.BlockPack{}).
		Select(`"BlockPackTable".id`).
		Joins(`INNER JOIN "SubShelfTable" ON "SubShelfTable".id = "BlockPackTable".parent_sub_shelf_id`).
		Where(`"SubShelfTable".id IN ? OR "SubShelfTable".path && ?`, subShelfIds, pg.Array(subShelfIds)).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&blockPackIds)
	if result.Error != nil {
		return nil, apiexceptions.NewBlockPackException().NotFound().WithOrigin(result.Error)
	}

	return blockPackIds, nil
}

func (r *BlockPackRepository) CreateOneBySubShelfId(
	subShelfId uuid.UUID,
	userId uuid.UUID,
	input inputs.CreateBlockPackInput,
	opts ...options.RepositoryOptions,
) (*uuid.UUID, *exceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	if parsedOptions.HasAllowedPermissions() {
		subShelfRepository := NewSubShelfRepository(scopes.NewSubShelfScope())

		if !subShelfRepository.HasPermission(
			subShelfId,
			userId,
			parsedOptions.AllowedPermissions,
			opts...,
		) {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().NoPermission("create a block pack under this shelf")
		}
	}

	var newBlockPack schemas.BlockPack
	if err := copier.Copy(&newBlockPack, &input); err != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewBlockPackException().FailedToCreate().WithOrigin(err)
	}
	if newBlockPack.Id == uuid.Nil {
		newBlockPack.Id = uuid.New()
	}
	newBlockPack.ParentSubShelfId = subShelfId

	result := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Create(&newBlockPack)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockPackException().FailedToCreate().WithOrigin(result.Error)},
		{First: newBlockPack.Id == uuid.Nil, Second: apiexceptions.NewBlockPackException().FailedToCreate()},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewBlockPackException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewBlockPackException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &newBlockPack.Id, nil
}

func (r *BlockPackRepository) CreateManyBySubShelfIds(
	userId uuid.UUID,
	input []inputs.CreateBlockPackBySubShelfIdInput,
	opts ...options.RepositoryOptions,
) ([]uuid.UUID, *exceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	isParentSubShelfIdValid := make(map[uuid.UUID]bool)
	if parsedOptions.HasAllowedPermissions() {
		isParentSubShelfExist := make(map[uuid.UUID]bool)
		var parentSubShelfIds []uuid.UUID
		for _, in := range input {
			if isParentSubShelfExist[in.ParentSubShelfId] {
				continue
			}
			isParentSubShelfExist[in.ParentSubShelfId] = true
			parentSubShelfIds = append(parentSubShelfIds, in.ParentSubShelfId)
		}

		subShelfRepository := NewSubShelfRepository(scopes.NewSubShelfScope())
		validParentSubShelves, exception := subShelfRepository.CheckPermissionsAndGetManyByIds(
			parentSubShelfIds,
			userId,
			nil,
			parsedOptions.AllowedPermissions,
			opts...,
		)
		if exception != nil {
			parsedOptions.DB.Rollback()
			return nil, exception
		}

		for _, validParentSubShelf := range validParentSubShelves {
			isParentSubShelfIdValid[validParentSubShelf.Id] = true
		}
	}

	var newBlockPacks []schemas.BlockPack
	for _, in := range input {
		if parsedOptions.HasAllowedPermissions() && !isParentSubShelfIdValid[in.ParentSubShelfId] {
			continue
		}
		var newBlockPack schemas.BlockPack
		if err := copier.Copy(&newBlockPack, &in); err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewBlockPackException().InvalidInput().WithOrigin(err)
		}
		if newBlockPack.Id == uuid.Nil {
			newBlockPack.Id = uuid.New()
		}
		newBlockPacks = append(newBlockPacks, newBlockPack)
	}

	result := parsedOptions.DB.Model(&schemas.BlockPack{}).
		CreateInBatches(&newBlockPacks, parsedOptions.BatchSize)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewBlockPackException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	newBlockPackIds := make([]uuid.UUID, len(newBlockPacks))
	for index, newBlockPack := range newBlockPacks {
		newBlockPackIds[index] = newBlockPack.Id
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewBlockPackException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return newBlockPackIds, nil
}

func (r *BlockPackRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateBlockPackInput,
	opts ...options.RepositoryOptions,
) (*schemas.BlockPack, *exceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	existingBlockPack, exception := r.CheckPermissionAndGetOneById(
		id,
		userId,
		nil,
		parsedOptions.AllowedPermissions,
		opts...,
	)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if input.Values.ParentSubShelfId != nil && !partialupdate.CheckSetNull(input.SetNull, "ParentSubShelfId") {
		subShelfRepository := NewSubShelfRepository(scopes.NewSubShelfScope())

		if !subShelfRepository.HasPermission(
			*input.Values.ParentSubShelfId,
			userId,
			parsedOptions.AllowedPermissions,
			opts...,
		) {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().NoPermission("move a block pack to this shelf")
		}
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingBlockPack)
	if err != nil {
		parsedOptions.DB.Rollback()
		return nil, exceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true).WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Select("*").
		Updates(&updates)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockPackException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewBlockPackException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewBlockPackException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &updates, nil
}

func (r *BlockPackRepository) UpdateManyByIds(
	userId uuid.UUID,
	input []inputs.UpdateBlockPackByIdInput,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	isSubShelfValid := make(map[uuid.UUID]bool)
	isBlockPackValid := make(map[uuid.UUID]bool)
	if parsedOptions.HasAllowedPermissions() {
		blockPackIds := make([]uuid.UUID, len(input))
		isParentSubShelfExist := make(map[uuid.UUID]bool)
		var parentSubShelfIds []uuid.UUID
		for index, in := range input {
			blockPackIds[index] = in.Id
			if in.PartialUpdateInput.Values.ParentSubShelfId == nil ||
				partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "ParentSubShelfId") {
				continue
			}
			parentSubShelfId := *in.PartialUpdateInput.Values.ParentSubShelfId

			if isParentSubShelfExist[parentSubShelfId] {
				continue
			}

			parentSubShelfIds = append(parentSubShelfIds, parentSubShelfId)
			isParentSubShelfExist[parentSubShelfId] = true
		}

		subShelfRepository := NewSubShelfRepository(scopes.NewSubShelfScope())
		validSubShelves, exception := subShelfRepository.CheckPermissionsAndGetManyByIds(
			parentSubShelfIds,
			userId,
			nil,
			parsedOptions.AllowedPermissions,
			opts...,
		)
		if exception != nil {
			parsedOptions.DB.Rollback()
			return exception
		}

		for _, validSubShelf := range validSubShelves {
			isSubShelfValid[validSubShelf.Id] = true
		}

		validBlockPacks, exception := r.CheckPermissionsAndGetManyByIds(
			blockPackIds,
			userId,
			nil,
			parsedOptions.AllowedPermissions,
			opts...,
		)
		if exception != nil {
			parsedOptions.DB.Rollback()
			return exception
		}

		for _, validBlockPack := range validBlockPacks {
			isBlockPackValid[validBlockPack.Id] = true
		}
	}

	var valuePlaceholders []string
	var valueArgs []interface{}
	for _, in := range input {
		if parsedOptions.HasAllowedPermissions() &&
			((in.PartialUpdateInput.Values.ParentSubShelfId != nil &&
				!partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "ParentSubShelfId") &&
				!isSubShelfValid[*in.PartialUpdateInput.Values.ParentSubShelfId]) || // check if the updated sub shelf is valid when it is given
				(!isBlockPackValid[in.Id])) { // check the block pack is valid
			continue
		}

		setIconNull := partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "Icon")
		setHeaderBackgroundNull := partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "HeaderBackgroundURL")

		valuePlaceholders = append(valuePlaceholders, `(?::uuid, ?::uuid, ?::text, ?::"SupportedIcon", ?::text, ?::boolean, ?::boolean)`)
		valueArgs = append(valueArgs,
			in.Id,
			in.PartialUpdateInput.Values.ParentSubShelfId,
			in.PartialUpdateInput.Values.Name,
			in.PartialUpdateInput.Values.Icon,
			in.PartialUpdateInput.Values.HeaderBackgroundURL,
			setIconNull,
			setHeaderBackgroundNull,
		)
	}

	sql := fmt.Sprintf(`
		UPDATE "BlockPackTable" bp
		SET
			parent_sub_shelf_id = COALESCE(v.parent_sub_shelf_id::uuid, bp.parent_sub_shelf_id),
			name = COALESCE(v.name::text, bp.name),
			icon = CASE
				WHEN v.set_icon_null::boolean THEN NULL
				ELSE COALESCE(v.icon::"SupportedIcon", bp.icon)
			END,
			header_background_url = CASE
				WHEN v.set_header_background_url_null::boolean THEN NULL
				ELSE COALESCE(v.header_background_url::text, bp.header_background_url)
			END,
			updated_at = NOW()
		FROM (VALUES %s) AS v(id, parent_sub_shelf_id, name, icon, header_background_url, set_icon_null, set_header_background_url_null)
		WHERE bp.id = v.id::uuid AND bp.deleted_at IS NULL
	`, strings.Join(valuePlaceholders, ","))
	result := parsedOptions.DB.Exec(sql, valueArgs...)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockPackException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewBlockPackException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return apiexceptions.NewBlockPackException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return nil
}

func (r *BlockPackRepository) RestoreSoftDeletedOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) (*schemas.BlockPack, *exceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var restoredBlockPack schemas.BlockPack
	query := parsedOptions.DB.Model(&restoredBlockPack).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted))
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.blockPackScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions))
	}

	result := query.
		Clauses(clause.Returning{}).
		Where(`"BlockPackTable".id = ?`, id).
		Updates(map[string]interface{}{"deleted_at": nil})
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockPackException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewBlockPackException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &restoredBlockPack, nil
}

func (r *BlockPackRepository) RestoreSoftDeletedManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) ([]schemas.BlockPack, *exceptions.Exception) {
	if len(ids) == 0 {
		return nil, apiexceptions.NewBlockPackException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var restoredBlockPacks []schemas.BlockPack
	query := parsedOptions.DB.Model(&restoredBlockPacks).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted))
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.blockPackScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions))
	}

	result := query.
		Clauses(&clause.Returning{}).
		Where(`"BlockPackTable".id IN ?`, ids).
		Updates(map[string]interface{}{"deleted_at": nil})
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockPackException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewBlockPackException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return restoredBlockPacks, nil
}

func (r *BlockPackRepository) SoftDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)
	query := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted))
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.blockPackScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions))
	}

	result := query.
		Where(`"BlockPackTable".id = ?`, id).
		Update("deleted_at", time.Now())
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockPackException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewBlockPackException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *BlockPackRepository) SoftDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	if len(ids) == 0 {
		return apiexceptions.NewBlockPackException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	query := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted))
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.blockPackScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions))
	}

	result := query.
		Where(`"BlockPackTable".id IN ?`, ids).
		Update("deleted_at", time.Now())
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockPackException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewBlockPackException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *BlockPackRepository) HardDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)
	query := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted))
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.blockPackScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions))
	}

	result := query.
		Where(`"BlockPackTable".id = ?`, id).
		Delete(&schemas.BlockPack{})
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockPackException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewBlockPackException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *BlockPackRepository) HardDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	if len(ids) == 0 {
		return apiexceptions.NewBlockPackException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	query := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted))
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.blockPackScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions))
	}

	result := query.
		Where(`"BlockPackTable".id IN ?`, ids).
		Delete(&schemas.BlockPack{})
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBlockPackException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewBlockPackException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

/* ============================== System Only Method ============================== */

func (r *BlockPackRepository) BulkCheckPermissionsAndGetManyByIds(
	inputs []inputs.BulkCheckBlockPackPermissionInput,
	preloads []schemas.BlockPackRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]bool, []schemas.BlockPack, *exceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, []schemas.BlockPack{}, nil
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	successes := make([]bool, len(inputs))
	ids := make([]uuid.UUID, 0, len(inputs))
	userIds := make([]uuid.UUID, 0, len(inputs))
	for _, in := range inputs {
		ids = append(ids, in.Id)
		userIds = append(userIds, in.UserId)
	}

	validIdSet := make(map[uuid.UUID]bool, len(ids))
	validTargetByUserId := make(map[[2]uuid.UUID]bool)
	if allowedPermissions != nil {
		var validTargets []struct {
			Id     uuid.UUID `gorm:"column:id"`
			UserId uuid.UUID `gorm:"column:user_id"`
		}
		result := parsedOptions.DB.Model(&schemas.BlockPack{}).
			Select(`"BlockPackTable".id, uts.user_id`).
			Joins(`INNER JOIN "SubShelfTable" AS ss ON ss.id = "BlockPackTable".parent_sub_shelf_id`).
			Joins(`INNER JOIN "UsersToShelvesTable" AS uts ON uts.root_shelf_id = ss.root_shelf_id`).
			Where(`"BlockPackTable".id IN ?`, ids).
			Where("uts.user_id IN ? AND uts.permission IN ?", userIds, allowedPermissions).
			Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
			Scan(&validTargets)
		if result.Error != nil {
			return nil, nil, apiexceptions.NewBlockPackException().NotFound().WithOrigin(result.Error)
		}

		for _, validTarget := range validTargets {
			validTargetByUserId[[2]uuid.UUID{validTarget.Id, validTarget.UserId}] = true
			validIdSet[validTarget.Id] = true
		}
	} else {
		var validIds []uuid.UUID
		result := parsedOptions.DB.Model(&schemas.BlockPack{}).
			Select(`"BlockPackTable".id`).
			Where(`"BlockPackTable".id IN ?`, ids).
			Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
			Scan(&validIds)
		if result.Error != nil {
			return nil, nil, apiexceptions.NewBlockPackException().NotFound().WithOrigin(result.Error)
		}

		for _, validId := range validIds {
			validIdSet[validId] = true
		}
	}

	validIds := make([]uuid.UUID, 0, len(validIdSet))
	for validId := range validIdSet {
		validIds = append(validIds, validId)
	}
	sort.Slice(validIds, func(left int, right int) bool {
		return validIds[left].String() < validIds[right].String()
	})
	if len(validIds) == 0 {
		return successes, []schemas.BlockPack{}, nil
	}

	var blockPacks []schemas.BlockPack
	result := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Where(`"BlockPackTable".id IN ?`, validIds).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.blockPackScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Order(`"BlockPackTable".id ASC`).
		Find(&blockPacks)
	if result.Error != nil {
		return nil, nil, apiexceptions.NewBlockPackException().NotFound().WithOrigin(result.Error)
	}

	foundIdSet := make(map[uuid.UUID]bool, len(blockPacks))
	for _, blockPack := range blockPacks {
		foundIdSet[blockPack.Id] = true
	}
	for index, in := range inputs {
		if validIdSet[in.Id] &&
			foundIdSet[in.Id] &&
			(allowedPermissions == nil || validTargetByUserId[[2]uuid.UUID{in.Id, in.UserId}]) {
			successes[index] = true
		}
	}

	return successes, blockPacks, nil
}

func (r *BlockPackRepository) BulkCreateMany(
	inputs []inputs.BulkCreateBlockPackInput,
	opts ...options.RepositoryOptions,
) ([]bool, *exceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, apiexceptions.NewBlockPackException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	successes := make([]bool, len(inputs))
	parentSubShelfIds := make([]uuid.UUID, 0, len(inputs))
	userIds := make([]uuid.UUID, 0, len(inputs))
	for _, in := range inputs {
		parentSubShelfIds = append(parentSubShelfIds, in.ParentSubShelfId)
		userIds = append(userIds, in.UserId)
	}

	var validTargets []struct {
		Id     uuid.UUID `gorm:"column:id"`
		UserId uuid.UUID `gorm:"column:user_id"`
	}
	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Select(`"SubShelfTable".id, uts.user_id`).
		Joins(`INNER JOIN "UsersToShelvesTable" AS uts ON uts.root_shelf_id = "SubShelfTable".root_shelf_id`).
		Where(`"SubShelfTable".id IN ? AND "SubShelfTable".deleted_at IS NULL`, parentSubShelfIds).
		Where("uts.user_id IN ? AND uts.permission IN ?", userIds, parsedOptions.AllowedPermissions).
		Scan(&validTargets)
	if result.Error != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewBlockPackException().FailedToCreate().WithOrigin(result.Error)
	}

	validTargetByUserId := make(map[[2]uuid.UUID]bool, len(validTargets))
	for _, validTarget := range validTargets {
		validTargetByUserId[[2]uuid.UUID{validTarget.Id, validTarget.UserId}] = true
	}

	newBlockPacks := make([]schemas.BlockPack, 0, len(inputs))
	successIndexes := make([]int, 0, len(inputs))
	for index, in := range inputs {
		if !validTargetByUserId[[2]uuid.UUID{in.ParentSubShelfId, in.UserId}] {
			continue
		}

		newBlockPackId := uuid.New()
		if in.Id != nil && *in.Id != uuid.Nil {
			newBlockPackId = *in.Id
		}

		newBlockPacks = append(newBlockPacks, schemas.BlockPack{
			Id:                  newBlockPackId,
			ParentSubShelfId:    in.ParentSubShelfId,
			Name:                in.Name,
			Icon:                in.Icon,
			HeaderBackgroundURL: in.HeaderBackgroundURL,
		})
		successIndexes = append(successIndexes, index)
	}

	if len(newBlockPacks) == 0 {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return successes, nil
	}

	result = parsedOptions.DB.Model(&schemas.BlockPack{}).
		CreateInBatches(&newBlockPacks, parsedOptions.BatchSize)
	if result.Error != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewBlockPackException().FailedToCreate().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewBlockPackException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	for _, successIndex := range successIndexes {
		successes[successIndex] = true
	}

	return successes, nil
}

func (r *BlockPackRepository) BulkUpdateMany(
	bulkInputs []inputs.BulkUpdateBlockPackInput,
	opts ...options.RepositoryOptions,
) ([]bool, *exceptions.Exception) {
	if len(bulkInputs) == 0 {
		return []bool{}, apiexceptions.NewBlockPackException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	checkInputs := make([]inputs.BulkCheckBlockPackPermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckBlockPackPermissionInput{
			UserId: in.UserId,
			Id:     in.Id,
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

	targetSubShelfIds := make([]uuid.UUID, 0, len(bulkInputs))
	targetUserIds := make([]uuid.UUID, 0, len(bulkInputs))
	for index, in := range bulkInputs {
		if !successes[index] ||
			in.PartialUpdateInput.Values.ParentSubShelfId == nil ||
			partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "ParentSubShelfId") {
			continue
		}
		targetSubShelfIds = append(targetSubShelfIds, *in.PartialUpdateInput.Values.ParentSubShelfId)
		targetUserIds = append(targetUserIds, in.UserId)
	}
	if len(targetSubShelfIds) > 0 {
		var validTargets []struct {
			Id     uuid.UUID `gorm:"column:id"`
			UserId uuid.UUID `gorm:"column:user_id"`
		}
		result := parsedOptions.DB.Model(&schemas.SubShelf{}).
			Select(`"SubShelfTable".id, uts.user_id`).
			Joins(`INNER JOIN "UsersToShelvesTable" AS uts ON uts.root_shelf_id = "SubShelfTable".root_shelf_id`).
			Where(`"SubShelfTable".id IN ? AND "SubShelfTable".deleted_at IS NULL`, targetSubShelfIds).
			Where("uts.user_id IN ? AND uts.permission IN ?", targetUserIds, parsedOptions.AllowedPermissions).
			Scan(&validTargets)
		if result.Error != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewBlockPackException().FailedToUpdate().WithOrigin(result.Error)
		}

		validTargetByUserId := make(map[[2]uuid.UUID]bool, len(validTargets))
		for _, validTarget := range validTargets {
			validTargetByUserId[[2]uuid.UUID{validTarget.Id, validTarget.UserId}] = true
		}
		for index, in := range bulkInputs {
			if !successes[index] ||
				in.PartialUpdateInput.Values.ParentSubShelfId == nil ||
				partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "ParentSubShelfId") {
				continue
			}
			if !validTargetByUserId[[2]uuid.UUID{*in.PartialUpdateInput.Values.ParentSubShelfId, in.UserId}] {
				successes[index] = false
			}
		}
	}

	valuePlaceholders := make([]string, 0, len(bulkInputs))
	valueArgs := make([]interface{}, 0, len(bulkInputs)*8)
	for index, in := range bulkInputs {
		if !successes[index] {
			continue
		}

		setIconNull := partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "Icon")
		setHeaderBackgroundURLNull := partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "HeaderBackgroundURL")

		valuePlaceholders = append(valuePlaceholders, `(?::int, ?::uuid, ?::uuid, ?::text, ?::"SupportedIcon", ?::text, ?::boolean, ?::boolean)`)
		valueArgs = append(valueArgs,
			index,
			in.Id,
			in.PartialUpdateInput.Values.ParentSubShelfId,
			in.PartialUpdateInput.Values.Name,
			in.PartialUpdateInput.Values.Icon,
			in.PartialUpdateInput.Values.HeaderBackgroundURL,
			setIconNull,
			setHeaderBackgroundURLNull,
		)
	}
	if len(valuePlaceholders) == 0 {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return successes, nil
	}

	sql := fmt.Sprintf(`
		WITH payload(idx, id, parent_sub_shelf_id, name, icon, header_background_url, set_icon_null, set_header_background_url_null) AS (
			VALUES %s
		),
		updated AS (
			UPDATE "BlockPackTable" AS bp
			SET
				parent_sub_shelf_id = COALESCE(v.parent_sub_shelf_id::uuid, bp.parent_sub_shelf_id),
				name = COALESCE(v.name::text, bp.name),
				icon = CASE
					WHEN v.set_icon_null::boolean THEN NULL
					ELSE COALESCE(v.icon::"SupportedIcon", bp.icon)
				END,
				header_background_url = CASE
					WHEN v.set_header_background_url_null::boolean THEN NULL
					ELSE COALESCE(v.header_background_url::text, bp.header_background_url)
				END,
				updated_at = NOW()
			FROM payload AS v
			WHERE bp.id = v.id::uuid
				AND bp.deleted_at IS NULL
			RETURNING bp.id
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
		return nil, apiexceptions.NewBlockPackException().FailedToUpdate().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewBlockPackException().FailedToCommitTransaction().WithOrigin(err)
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

func (r *BlockPackRepository) BulkDeleteMany(
	bulkInputs []inputs.BulkDeleteBlockPackInput,
	opts ...options.RepositoryOptions,
) ([]bool, *exceptions.Exception) {
	if len(bulkInputs) == 0 {
		return []bool{}, apiexceptions.NewBlockPackException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	checkInputs := make([]inputs.BulkCheckBlockPackPermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckBlockPackPermissionInput{
			UserId: in.UserId,
			Id:     in.Id,
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

	validIds := make([]uuid.UUID, 0, len(bulkInputs))
	for index, in := range bulkInputs {
		if successes[index] {
			validIds = append(validIds, in.Id)
		}
	}
	if len(validIds) == 0 {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return successes, nil
	}

	var deletedBlockPacks []schemas.BlockPack
	result := parsedOptions.DB.Model(&deletedBlockPacks).
		Clauses(clause.Returning{}).
		Where("id IN ? AND deleted_at IS NULL", validIds).
		Updates(map[string]interface{}{"deleted_at": time.Now(), "updated_at": time.Now()})
	if result.Error != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewBlockPackException().FailedToDelete().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewBlockPackException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	deletedIdSet := make(map[uuid.UUID]bool, len(deletedBlockPacks))
	for _, deletedBlockPack := range deletedBlockPacks {
		deletedIdSet[deletedBlockPack.Id] = true
	}
	for index, in := range bulkInputs {
		if successes[index] && deletedIdSet[in.Id] {
			successes[index] = true
		} else {
			successes[index] = false
		}
	}

	return successes, nil
}
