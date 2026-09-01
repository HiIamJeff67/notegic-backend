package repositories

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
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

type RootShelfRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.RootShelfRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) (*schemas.RootShelf, cenums.AccessControlPermission, *cexceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.RootShelfRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]schemas.RootShelf, []cenums.AccessControlPermission, *cexceptions.Exception)
	GetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.RootShelfRelation, opts ...RepositoryOptions) (*schemas.RootShelf, cenums.AccessControlPermission, *cexceptions.Exception)
	CreateOne(ownerId uuid.UUID, input inputs.CreateRootShelfInput, opts ...RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	CreateMany(ownerId uuid.UUID, input []inputs.CreateRootShelfInput, opts ...RepositoryOptions) ([]uuid.UUID, *cexceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateRootShelfInput, opts ...RepositoryOptions) (*schemas.RootShelf, *cexceptions.Exception)
	UpdateManyByIds(userId uuid.UUID, input []inputs.UpdateRootShelfByIdInput, opts ...RepositoryOptions) *cexceptions.Exception
	RestoreSoftDeletedOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) (*schemas.RootShelf, *cexceptions.Exception)
	RestoreSoftDeletedManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) ([]schemas.RootShelf, *cexceptions.Exception)
	SoftDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	SoftDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	SoftDeleteManyByUserId(userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByIds(sids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByUserId(userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception

	/* ============================== System Only Method ============================== */
	BulkCheckPermissionsAndGetManyByIds(inputs []inputs.BulkCheckRootShelfPermissionInput, preloads []schemas.RootShelfRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]bool, []schemas.RootShelf, *cexceptions.Exception)
	BulkCreateMany(inputs []inputs.BulkCreateRootShelfInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkUpdateMany(inputs []inputs.BulkUpdateRootShelfInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
}

type RootShelfRepository struct {
	db *gorm.DB
	BulkRootShelfRepository
	rootShelfScope scopes.RootShelfScopeInterface
	exceptions     exceptions.ShelfException
}

func NewRootShelfRepository(
	db *gorm.DB,
	rootShelfScope scopes.RootShelfScopeInterface,
) RootShelfRepositoryInterface {
	return &RootShelfRepository{
		db:                      db,
		BulkRootShelfRepository: *NewBulkRootShelfRepository(db, rootShelfScope),
		rootShelfScope:          rootShelfScope,
		exceptions:              exceptions.NewShelfException(),
	}
}

func (r *RootShelfRepository) HasPermission(
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
		Model(&schemas.RootShelf{}).
		Select("1").
		Scopes(r.rootShelfScope.PassPermissionCheck(id, userId, allowedPermissions)).
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Limit(1).
		Scan(&marker)
	if err := result.Error; err != nil {
		return false
	}

	return marker == 1
}

func (r *RootShelfRepository) HavePermissions(
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
		Model(&schemas.RootShelf{}).
		Select(`DISTINCT "RootShelfTable".id`).
		Scopes(r.rootShelfScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&permittedIds)
	if err := result.Error; err != nil {
		return false
	}

	return array.GetDistinctCount(ids) == array.GetDistinctCount(permittedIds)
}

func (r *RootShelfRepository) CheckPermissionAndGetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RootShelfRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) (*schemas.RootShelf, cenums.AccessControlPermission, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	type rootShelfWithPermission struct {
		schemas.RootShelf
		Permission cenums.AccessControlPermission `gorm:"column:permission"`
	}

	var rootShelf rootShelfWithPermission
	query := parsedOptions.DB.
		Model(&schemas.RootShelf{}).
		Select(`"RootShelfTable".*, users_to_shelf.permission AS permission`).
		Joins(`
			INNER JOIN "UsersToShelvesTable" AS users_to_shelf
				ON users_to_shelf.root_shelf_id = "RootShelfTable".id
				AND users_to_shelf.user_id = ?
		`, userId).
		Where(`"RootShelfTable".id = ?`, id)
	if allowedPermissions != nil && len(allowedPermissions) > 0 {
		query = query.Scopes(
			r.rootShelfScope.PassPermissionCheck(id, userId, allowedPermissions),
		)
	}

	result := query.
		Scopes(r.rootShelfScope.IncludePreloads(preloads)).
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&rootShelf)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: rootShelf.Id == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, "", exception
	}

	return &rootShelf.RootShelf, rootShelf.Permission, nil
}

func (r *RootShelfRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RootShelfRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]schemas.RootShelf, []cenums.AccessControlPermission, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var rootShelves []schemas.RootShelf
	result := parsedOptions.DB.
		Model(&schemas.RootShelf{}).
		Scopes(r.rootShelfScope.IncludePreloads(preloads)).
		Scopes(r.rootShelfScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&rootShelves)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: len(rootShelves) == 0, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, nil, exception
	}

	permissions := make([]cenums.AccessControlPermission, len(rootShelves))
	if allowedPermissions != nil {
		var usersToShelves []schemas.UsersToShelves
		result = parsedOptions.DB.
			Model(&schemas.UsersToShelves{}).
			Select("root_shelf_id, permission").
			Where(
				"root_shelf_id IN ? AND user_id = ? AND permission IN ?",
				ids,
				userId,
				allowedPermissions,
			).
			Scopes(scopes.Locking(parsedOptions.LockingStrength)).
			Find(&usersToShelves)
		if exception := cexceptions.Cover(nil, []cexceptions.Pair{
			{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
			{First: len(usersToShelves) == 0, Second: r.exceptions.NotFound()},
		}); exception != nil {
			return nil, nil, exception
		}

		permissionByRootShelfId := make(map[uuid.UUID]cenums.AccessControlPermission, len(usersToShelves))
		for _, usersToShelf := range usersToShelves {
			permissionByRootShelfId[usersToShelf.RootShelfId] = usersToShelf.Permission
		}

		for index, rootShelf := range rootShelves {
			permission, exist := permissionByRootShelfId[rootShelf.Id]
			if !exist {
				return nil, nil, r.exceptions.NotFound()
			}
			permissions[index] = permission
		}
	}

	return rootShelves, permissions, nil
}

func (r *RootShelfRepository) GetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RootShelfRelation,
	opts ...RepositoryOptions,
) (*schemas.RootShelf, cenums.AccessControlPermission, *cexceptions.Exception) {
	return r.CheckPermissionAndGetOneById(
		id,
		userId,
		preloads,
		ParseRepositoryOptions(
			append([]RepositoryOptions{
				WithDB(r.db),
			}, opts...)...,
		).AllowedPermissions,
		opts...,
	)
}

func (r *RootShelfRepository) CreateOne(
	ownerId uuid.UUID,
	input inputs.CreateRootShelfInput,
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

	var newRootShelf schemas.RootShelf
	newRootShelf.OwnerId = ownerId
	if err := copier.Copy(&newRootShelf, &input); err != nil {
		parsedOptions.DB.Rollback()
		return nil, r.exceptions.FailedToCreate().WithOrigin(err)
	}
	if newRootShelf.Id == uuid.Nil {
		newRootShelf.Id = uuid.New()
	}

	result := parsedOptions.DB.Model(&schemas.RootShelf{}).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Create(&newRootShelf)
	if err := result.Error; err != nil {
		parsedOptions.DB.Rollback()
		switch err.Error() {
		case "ERROR: duplicate key value violates unique constraint \"shelf_idx_owner_id_name\" (SQLSTATE 23505)":
			return nil, r.exceptions.DuplicateName(input.Name)
		default:
			return nil, r.exceptions.FailedToCreate().WithOrigin(err)
		}
	}

	// create the users to shelves relation with the permission of admin
	newUsersToShelves := schemas.UsersToShelves{
		UserId:      ownerId,
		RootShelfId: newRootShelf.Id,
		Permission:  cenums.AccessControlPermission_Owner,
	}
	result = parsedOptions.DB.Model(&schemas.UsersToShelves{}).
		Create(&newUsersToShelves)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToCreate().WithOrigin(result.Error)},
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

	return &newRootShelf.Id, nil
}

func (r *RootShelfRepository) CreateMany(
	ownerId uuid.UUID,
	input []inputs.CreateRootShelfInput,
	opts ...RepositoryOptions,
) ([]uuid.UUID, *cexceptions.Exception) {
	if len(input) == 0 {
		return nil, r.exceptions.NoChanges()
	}

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

	var newRootShelves []schemas.RootShelf
	for _, in := range input {
		var newRootShelf schemas.RootShelf
		newRootShelf.OwnerId = ownerId
		if err := copier.Copy(&newRootShelf, &in); err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.InvalidDto().WithOrigin(err)
		}
		newRootShelves = append(newRootShelves, newRootShelf)
	}

	result := parsedOptions.DB.Model(&schemas.RootShelf{}).
		CreateInBatches(&newRootShelves, parsedOptions.BatchSize)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	newRootShelfIds := make([]uuid.UUID, len(newRootShelves))
	newUsersToShelves := make([]schemas.UsersToShelves, len(newRootShelves))
	for index, newRootShelf := range newRootShelves {
		newRootShelfIds[index] = newRootShelf.Id
		newUsersToShelves[index] = schemas.UsersToShelves{
			UserId:      ownerId,
			RootShelfId: newRootShelf.Id,
			Permission:  cenums.AccessControlPermission_Owner,
		}
	}
	result = parsedOptions.DB.Model(&schemas.UsersToShelves{}).
		CreateInBatches(&newUsersToShelves, parsedOptions.BatchSize)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToCreate().WithOrigin(result.Error)},
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

	return newRootShelfIds, nil
}

func (r *RootShelfRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateRootShelfInput,
	opts ...RepositoryOptions,
) (*schemas.RootShelf, *cexceptions.Exception) {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if !parsedOptions.IsTransactionStarted {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, WithTransactionDB(parsedOptions.DB))
		opts = append(opts, WithLockingStrength(LockingStrengthNoKeyUpdate))
	}

	existingRootShelf, _, exception := r.CheckPermissionAndGetOneById(
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

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingRootShelf)
	if err != nil {
		parsedOptions.DB.Rollback()
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true).WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.RootShelf{}).
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

func (r *RootShelfRepository) UpdateManyByIds(
	userId uuid.UUID,
	input []inputs.UpdateRootShelfByIdInput,
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

	isRootShelfValid := make(map[uuid.UUID]bool)
	if parsedOptions.HasAllowedPermissions() {
		ids := make([]uuid.UUID, len(input))
		for index, in := range input {
			ids[index] = in.Id
		}

		validRootShelves, _, exception := r.CheckPermissionsAndGetManyByIds(
			ids,
			userId,
			nil,
			parsedOptions.AllowedPermissions,
			opts...,
		)
		if exception != nil {
			parsedOptions.DB.Rollback()
			return r.exceptions.NoPermission("update these root shelves")
		}

		for _, validRootShelf := range validRootShelves {
			isRootShelfValid[validRootShelf.Id] = true
		}
	}

	var valuePlaceholders []string
	var valueArgs []interface{}
	for _, in := range input {
		if parsedOptions.HasAllowedPermissions() && !isRootShelfValid[in.Id] {
			continue
		}

		valuePlaceholders = append(valuePlaceholders, "(?::uuid, ?::text, ?::bigint, ?::bigint, ?::timestamptz)")
		valueArgs = append(valueArgs,
			in.Id,
			in.PartialUpdateInput.Values.Name,
			in.PartialUpdateInput.Values.SubShelfCount,
			in.PartialUpdateInput.Values.ItemCount,
			in.PartialUpdateInput.Values.LastAnalyzedAt,
		)
	}

	sql := fmt.Sprintf(`
		UPDATE "RootShelfTable" AS r
		SET
			name = COALESCE(v.name::text, r.name),
			sub_shelf_count = COALESCE(v.sub_shelf_count::bigint, r.sub_shelf_count),
			item_count = COALESCE(v.item_count::bigint, r.item_count),
			last_analyzed_at = COALESCE(v.last_analyzed_at, r.last_analyzed_at),
			updated_at = NOW()
		FROM (VALUES %s) AS v(id, name, sub_shelf_count, item_count, last_analyzed_at)
		WHERE r.id = v.id::uuid AND r.deleted_at IS NULL
	`, strings.Join(valuePlaceholders, ","))
	result := parsedOptions.DB.Exec(sql, valueArgs...)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
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

func (r *RootShelfRepository) RestoreSoftDeletedOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) (*schemas.RootShelf, *cexceptions.Exception) {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var restoredRootShelf schemas.RootShelf
	query := parsedOptions.DB.Model(&restoredRootShelf)
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.rootShelfScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions))
	} else {
		query = query.Where(`"RootShelfTable".id = ?`, id)
	}
	result := query.
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Clauses(clause.Returning{}).
		Updates(map[string]interface{}{"deleted_at": nil}) // force to assign null value
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: restoredRootShelf.Id == uuid.Nil, Second: r.exceptions.FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &restoredRootShelf, nil
}

func (r *RootShelfRepository) RestoreSoftDeletedManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) ([]schemas.RootShelf, *cexceptions.Exception) {
	if len(ids) == 0 {
		return nil, r.exceptions.NoChanges()
	}

	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var restoredRootShelves []schemas.RootShelf
	query := parsedOptions.DB.Model(&restoredRootShelves)
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.rootShelfScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions))
	} else {
		query = query.Where(`"RootShelfTable".id IN ?`, ids)
	}
	result := query.
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Clauses(clause.Returning{}).
		Updates(map[string]interface{}{"deleted_at": nil}) // force to assign null value
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: len(restoredRootShelves) != len(ids), Second: r.exceptions.FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return restoredRootShelves, nil
}

func (r *RootShelfRepository) SoftDeleteOneById(
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

	query := parsedOptions.DB.Model(&schemas.RootShelf{})
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.rootShelfScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions))
	} else {
		query = query.Where(`"RootShelfTable".id = ?`, id)
	}
	result := query.
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RootShelfRepository) SoftDeleteManyByIds(
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

	query := parsedOptions.DB.Model(&schemas.RootShelf{})
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.rootShelfScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions))
	} else {
		query = query.Where(`"RootShelfTable".id IN ?`, ids)
	}
	result := query.
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RootShelfRepository) SoftDeleteManyByUserId(
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.Model(&schemas.RootShelf{}).
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where("owner_id = ?", userId).
		Delete(&schemas.RootShelf{})
	if err := result.Error; err != nil {
		return r.exceptions.FailedToDelete().WithOrigin(err)
	}
	if result.RowsAffected == 0 {
		return r.exceptions.NotFound()
	}

	return nil
}

func (r *RootShelfRepository) HardDeleteOneById(
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

	query := parsedOptions.DB.Model(&schemas.RootShelf{})
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.rootShelfScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions))
	} else {
		query = query.Where(`"RootShelfTable".id = ?`, id)
	}
	result := query.
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Delete(&schemas.RootShelf{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RootShelfRepository) HardDeleteManyByIds(
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

	query := parsedOptions.DB.Model(&schemas.RootShelf{})
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.rootShelfScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions))
	} else {
		query = query.Where(`"RootShelfTable".id IN ?`, ids)
	}
	result := query.
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Delete(&schemas.RootShelf{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RootShelfRepository) HardDeleteManyByUserId(
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.Model(&schemas.RootShelf{}).
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where("owner_id = ?", userId).
		Delete(&schemas.RootShelf{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

/* ============================== System Only Method ============================== */

func (r *RootShelfRepository) BulkCheckPermissionsAndGetManyByIds(
	inputs []inputs.BulkCheckRootShelfPermissionInput,
	preloads []schemas.RootShelfRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]bool, []schemas.RootShelf, *cexceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, []schemas.RootShelf{}, nil
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

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
		result := parsedOptions.DB.Model(&schemas.RootShelf{}).
			Select(`"RootShelfTable".id, uts.user_id`).
			Joins(`INNER JOIN "UsersToShelvesTable" AS uts ON uts.root_shelf_id = "RootShelfTable".id`).
			Where(`"RootShelfTable".id IN ? AND uts.user_id IN ? AND uts.permission IN ?`, ids, userIds, allowedPermissions).
			Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
			Scan(&validTargets)
		if result.Error != nil {
			return nil, nil, r.exceptions.NotFound().WithOrigin(result.Error)
		}

		for _, validTarget := range validTargets {
			validTargetByUserId[[2]uuid.UUID{validTarget.Id, validTarget.UserId}] = true
			validIdSet[validTarget.Id] = true
		}
	} else {
		var validIds []uuid.UUID
		result := parsedOptions.DB.Model(&schemas.RootShelf{}).
			Select(`"RootShelfTable".id`).
			Where(`"RootShelfTable".id IN ?`, ids).
			Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
			Scan(&validIds)
		if result.Error != nil {
			return nil, nil, r.exceptions.NotFound().WithOrigin(result.Error)
		}

		for _, validId := range validIds {
			validIdSet[validId] = true
		}
	}

	validIds := make([]uuid.UUID, 0, len(validIdSet))
	for validId := range validIdSet {
		validIds = append(validIds, validId)
	}
	if len(validIds) == 0 {
		return successes, []schemas.RootShelf{}, nil
	}

	var rootShelves []schemas.RootShelf
	result := parsedOptions.DB.Model(&schemas.RootShelf{}).
		Where(`"RootShelfTable".id IN ?`, validIds).
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.rootShelfScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&rootShelves)
	if result.Error != nil {
		return nil, nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	foundIdSet := make(map[uuid.UUID]bool, len(rootShelves))
	for _, rootShelf := range rootShelves {
		foundIdSet[rootShelf.Id] = true
	}
	for index, in := range inputs {
		if validIdSet[in.Id] &&
			foundIdSet[in.Id] &&
			(allowedPermissions == nil || validTargetByUserId[[2]uuid.UUID{in.Id, in.UserId}]) {
			successes[index] = true
		}
	}

	return successes, rootShelves, nil
}
