package repositories

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	pg "github.com/lib/pq"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	array "github.com/HiIamJeff67/notegic-backend/shared/lib/array"
	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"
	exceptions "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"
)

type BlockPackRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.BlockPackRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) (*schemas.BlockPack, *cexceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.BlockPackRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]schemas.BlockPack, *cexceptions.Exception)
	CheckPermissionAndGetOneWithOwnerIdById(id uuid.UUID, userId uuid.UUID, preloads []schemas.BlockPackRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) (*uuid.UUID, *schemas.BlockPack, *cexceptions.Exception)
	CheckPermissionsAndGetManyWithOwnerIdsByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.BlockPackRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]uuid.UUID, []schemas.BlockPack, *cexceptions.Exception)
	GetOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) (*schemas.BlockPack, *cexceptions.Exception)
	GetManyByRootShelfIds(rootShelfIds []uuid.UUID, opts ...RepositoryOptions) ([]schemas.BlockPack, *cexceptions.Exception)
	GetIdsByParentSubShelfIds(parentSubShelfIds []uuid.UUID, opts ...RepositoryOptions) ([]uuid.UUID, *cexceptions.Exception)
	GetIdsBySubShelfIdsAndDescendants(subShelfIds []uuid.UUID, opts ...RepositoryOptions) ([]uuid.UUID, *cexceptions.Exception)
	CreateOneBySubShelfId(subShelfId uuid.UUID, userId uuid.UUID, input inputs.CreateBlockPackInput, opts ...RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	CreateManyBySubShelfIds(userId uuid.UUID, input []inputs.CreateBlockPackBySubShelfIdInput, opts ...RepositoryOptions) ([]uuid.UUID, *cexceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateBlockPackInput, opts ...RepositoryOptions) (*schemas.BlockPack, *cexceptions.Exception)
	UpdateManyByIds(userId uuid.UUID, input []inputs.UpdateBlockPackByIdInput, opts ...RepositoryOptions) *cexceptions.Exception
	RestoreSoftDeletedOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) (*schemas.BlockPack, *cexceptions.Exception)
	RestoreSoftDeletedManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) ([]schemas.BlockPack, *cexceptions.Exception)
	SoftDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	SoftDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception

	/* ============================== System Only Method ============================== */

	BulkCheckPermissionsAndGetManyByIds(inputs []inputs.BulkCheckBlockPackPermissionInput, preloads []schemas.BlockPackRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]bool, []schemas.BlockPack, *cexceptions.Exception)
	BulkCreateMany(inputs []inputs.BulkCreateBlockPackInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkUpdateMany(inputs []inputs.BulkUpdateBlockPackInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkDeleteMany(inputs []inputs.BulkDeleteBlockPackInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
}

type BlockPackRepository struct {
	db *gorm.DB
	BulkBlockPackRepository
	blockPackScope scopes.BlockPackScopeInterface
	exceptions     exceptions.BlockPackException
}

func NewBlockPackRepository(
	db *gorm.DB,
	blockPackScope scopes.BlockPackScopeInterface,
) BlockPackRepositoryInterface {
	return &BlockPackRepository{
		db:                      db,
		BulkBlockPackRepository: *NewBulkBlockPackRepository(db, blockPackScope),
		blockPackScope:          blockPackScope,
		exceptions:              exceptions.NewBlockPackException(),
	}
}

func (r *BlockPackRepository) HasPermission(
	id uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) bool {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

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
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) bool {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

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
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) (*schemas.BlockPack, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

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
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: blockPack.Id == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &blockPack, nil
}

func (r *BlockPackRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.BlockPackRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]schemas.BlockPack, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var blockPacks []schemas.BlockPack
	result := parsedOptions.DB.
		Model(&schemas.BlockPack{}).
		Scopes(r.blockPackScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.blockPackScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&blockPacks)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: len(blockPacks) == 0, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return blockPacks, nil
}

func (r *BlockPackRepository) CheckPermissionAndGetOneWithOwnerIdById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.BlockPackRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) (*uuid.UUID, *schemas.BlockPack, *cexceptions.Exception) { // we should also return the owner id for the block groups and blocks
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
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
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: blockPackWithOwnerId.OwnerId == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, nil, exception
	}

	return &blockPackWithOwnerId.OwnerId, &blockPackWithOwnerId.BlockPack, nil
}

func (r *BlockPackRepository) CheckPermissionsAndGetManyWithOwnerIdsByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.BlockPackRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]uuid.UUID, []schemas.BlockPack, *cexceptions.Exception) { // we should also return the owner id for the block groups and blocks
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
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
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: len(blockPacksWithOwnerIds) == 0, Second: r.exceptions.NotFound()},
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
	opts ...RepositoryOptions,
) (*schemas.BlockPack, *cexceptions.Exception) {
	return r.CheckPermissionAndGetOneById(
		id,
		userId,
		nil,
		ParseRepositoryOptions(
			append([]RepositoryOptions{
				WithDB(r.db),
			}, opts...)...,
		).AllowedPermissions,
		opts...,
	)
}

func (r *BlockPackRepository) GetManyByRootShelfIds(
	rootShelfIds []uuid.UUID,
	opts ...RepositoryOptions,
) ([]schemas.BlockPack, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var blockPacks []schemas.BlockPack
	result := parsedOptions.DB.
		Model(&schemas.BlockPack{}).
		Joins(`INNER JOIN "SubShelfTable" ON "SubShelfTable".id = "BlockPackTable".parent_sub_shelf_id`).
		Where(`"SubShelfTable".root_shelf_id IN ?`, rootShelfIds).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&blockPacks)
	if result.Error != nil {
		return nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	return blockPacks, nil
}

func (r *BlockPackRepository) GetIdsByParentSubShelfIds(
	parentSubShelfIds []uuid.UUID,
	opts ...RepositoryOptions,
) ([]uuid.UUID, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var blockPackIds []uuid.UUID
	result := parsedOptions.DB.
		Model(&schemas.BlockPack{}).
		Select("id").
		Where("parent_sub_shelf_id IN ?", parentSubShelfIds).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Find(&blockPackIds)
	if result.Error != nil {
		return nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	return blockPackIds, nil
}

func (r *BlockPackRepository) GetIdsBySubShelfIdsAndDescendants(
	subShelfIds []uuid.UUID,
	opts ...RepositoryOptions,
) ([]uuid.UUID, *cexceptions.Exception) {
	if len(subShelfIds) == 0 {
		return []uuid.UUID{}, nil
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
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
		return nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	return blockPackIds, nil
}

func (r *BlockPackRepository) CreateOneBySubShelfId(
	subShelfId uuid.UUID,
	userId uuid.UUID,
	input inputs.CreateBlockPackInput,
	opts ...RepositoryOptions,
) (*uuid.UUID, *cexceptions.Exception) {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, WithTransactionDB(parsedOptions.DB))
		opts = append(opts, WithLockingStrength(LockingStrengthNoKeyUpdate))
	}

	if parsedOptions.HasAllowedPermissions() {
		subShelfRepository := NewSubShelfRepository(r.db, scopes.NewSubShelfScope())

		if !subShelfRepository.HasPermission(
			subShelfId,
			userId,
			parsedOptions.AllowedPermissions,
			opts...,
		) {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.NoPermission("create a block pack under this shelf")
		}
	}

	var newBlockPack schemas.BlockPack
	if err := copier.Copy(&newBlockPack, &input); err != nil {
		parsedOptions.DB.Rollback()
		return nil, r.exceptions.FailedToCreate().WithOrigin(err)
	}
	if newBlockPack.Id == uuid.Nil {
		newBlockPack.Id = uuid.New()
	}
	newBlockPack.ParentSubShelfId = subShelfId

	result := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Create(&newBlockPack)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToCreate().WithOrigin(result.Error)},
		{First: newBlockPack.Id == uuid.Nil, Second: r.exceptions.FailedToCreate()},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &newBlockPack.Id, nil
}

func (r *BlockPackRepository) CreateManyBySubShelfIds(
	userId uuid.UUID,
	input []inputs.CreateBlockPackBySubShelfIdInput,
	opts ...RepositoryOptions,
) ([]uuid.UUID, *cexceptions.Exception) {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, WithTransactionDB(parsedOptions.DB))
		opts = append(opts, WithLockingStrength(LockingStrengthNoKeyUpdate))
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

		subShelfRepository := NewSubShelfRepository(r.db, scopes.NewSubShelfScope())
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
			return nil, r.exceptions.InvalidInput().WithOrigin(err)
		}
		if newBlockPack.Id == uuid.Nil {
			newBlockPack.Id = uuid.New()
		}
		newBlockPacks = append(newBlockPacks, newBlockPack)
	}

	result := parsedOptions.DB.Model(&schemas.BlockPack{}).
		CreateInBatches(&newBlockPacks, parsedOptions.BatchSize)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
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
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return newBlockPackIds, nil
}

func (r *BlockPackRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateBlockPackInput,
	opts ...RepositoryOptions,
) (*schemas.BlockPack, *cexceptions.Exception) {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, WithTransactionDB(parsedOptions.DB))
		opts = append(opts, WithLockingStrength(LockingStrengthNoKeyUpdate))
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
		subShelfRepository := NewSubShelfRepository(r.db, scopes.NewSubShelfScope())

		if !subShelfRepository.HasPermission(
			*input.Values.ParentSubShelfId,
			userId,
			parsedOptions.AllowedPermissions,
			opts...,
		) {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.NoPermission("move a block pack to this shelf")
		}
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingBlockPack)
	if err != nil {
		parsedOptions.DB.Rollback()
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true).WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Select("*").
		Updates(&updates)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &updates, nil
}

func (r *BlockPackRepository) UpdateManyByIds(
	userId uuid.UUID,
	input []inputs.UpdateBlockPackByIdInput,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, WithTransactionDB(parsedOptions.DB))
		opts = append(opts, WithLockingStrength(LockingStrengthNoKeyUpdate))
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

		subShelfRepository := NewSubShelfRepository(r.db, scopes.NewSubShelfScope())
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
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return nil
}

func (r *BlockPackRepository) RestoreSoftDeletedOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) (*schemas.BlockPack, *cexceptions.Exception) {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

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
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &restoredBlockPack, nil
}

func (r *BlockPackRepository) RestoreSoftDeletedManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) ([]schemas.BlockPack, *cexceptions.Exception) {
	if len(ids) == 0 {
		return nil, r.exceptions.NoChanges()
	}

	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

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
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return restoredBlockPacks, nil
}

func (r *BlockPackRepository) SoftDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	query := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted))
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.blockPackScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions))
	}

	result := query.
		Where(`"BlockPackTable".id = ?`, id).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *BlockPackRepository) SoftDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	if len(ids) == 0 {
		return r.exceptions.NoChanges()
	}

	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	query := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted))
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.blockPackScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions))
	}

	result := query.
		Where(`"BlockPackTable".id IN ?`, ids).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *BlockPackRepository) HardDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	query := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted))
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.blockPackScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions))
	}

	result := query.
		Where(`"BlockPackTable".id = ?`, id).
		Delete(&schemas.BlockPack{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *BlockPackRepository) HardDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	if len(ids) == 0 {
		return r.exceptions.NoChanges()
	}

	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	query := parsedOptions.DB.Model(&schemas.BlockPack{}).
		Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted))
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.blockPackScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions))
	}

	result := query.
		Where(`"BlockPackTable".id IN ?`, ids).
		Delete(&schemas.BlockPack{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}
