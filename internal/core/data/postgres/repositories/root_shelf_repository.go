package repositories

import (
	"fmt"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm/clause"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	array "github.com/HiIamJeff67/notegic-backend/shared/lib/array"
	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type RootShelfRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.RootShelfRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) (*schemas.RootShelf, enums.AccessControlPermission, *cexceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.RootShelfRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]schemas.RootShelf, []enums.AccessControlPermission, *cexceptions.Exception)
	GetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.RootShelfRelation, opts ...options.RepositoryOptions) (*schemas.RootShelf, enums.AccessControlPermission, *cexceptions.Exception)
	CreateOne(ownerId uuid.UUID, input inputs.CreateRootShelfInput, opts ...options.RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	CreateMany(ownerId uuid.UUID, input []inputs.CreateRootShelfInput, opts ...options.RepositoryOptions) ([]uuid.UUID, *cexceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateRootShelfInput, opts ...options.RepositoryOptions) (*schemas.RootShelf, *cexceptions.Exception)
	UpdateManyByIds(userId uuid.UUID, input []inputs.UpdateRootShelfByIdInput, opts ...options.RepositoryOptions) *cexceptions.Exception
	RestoreSoftDeletedOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) (*schemas.RootShelf, *cexceptions.Exception)
	RestoreSoftDeletedManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) ([]schemas.RootShelf, *cexceptions.Exception)
	SoftDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	SoftDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	SoftDeleteManyByUserId(userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByIds(sids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByUserId(userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception

	/* ============================== System Only Method ============================== */
	BulkCheckPermissionsAndGetManyByIds(inputs []inputs.BulkCheckRootShelfPermissionInput, preloads []schemas.RootShelfRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]bool, []schemas.RootShelf, *cexceptions.Exception)
	BulkCreateMany(inputs []inputs.BulkCreateRootShelfInput, opts ...options.RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkUpdateMany(inputs []inputs.BulkUpdateRootShelfInput, opts ...options.RepositoryOptions) ([]bool, *cexceptions.Exception)
}

type RootShelfRepository struct {
	rootShelfScope scopes.RootShelfScopeInterface
}

func NewRootShelfRepository(rootShelfScope scopes.RootShelfScopeInterface) RootShelfRepositoryInterface {
	return &RootShelfRepository{
		rootShelfScope: rootShelfScope,
	}
}

func (r *RootShelfRepository) HasPermission(
	id uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

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
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

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
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) (*schemas.RootShelf, enums.AccessControlPermission, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	type rootShelfWithPermission struct {
		schemas.RootShelf
		Permission enums.AccessControlPermission `gorm:"column:permission"`
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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().NotFound().WithOrigin(result.Error)},
		{First: rootShelf.Id == uuid.Nil, Second: apiexceptions.NewShelfException().NotFound()},
	}); exception != nil {
		return nil, "", exception
	}

	return &rootShelf.RootShelf, rootShelf.Permission, nil
}

func (r *RootShelfRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RootShelfRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]schemas.RootShelf, []enums.AccessControlPermission, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var rootShelves []schemas.RootShelf
	result := parsedOptions.DB.
		Model(&schemas.RootShelf{}).
		Scopes(r.rootShelfScope.IncludePreloads(preloads)).
		Scopes(r.rootShelfScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&rootShelves)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().NotFound().WithOrigin(result.Error)},
		{First: len(rootShelves) == 0, Second: apiexceptions.NewShelfException().NotFound()},
	}); exception != nil {
		return nil, nil, exception
	}

	permissions := make([]enums.AccessControlPermission, len(rootShelves))
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
			{First: result.Error != nil, Second: apiexceptions.NewShelfException().NotFound().WithOrigin(result.Error)},
			{First: len(usersToShelves) == 0, Second: apiexceptions.NewShelfException().NotFound()},
		}); exception != nil {
			return nil, nil, exception
		}

		permissionByRootShelfId := make(map[uuid.UUID]enums.AccessControlPermission, len(usersToShelves))
		for _, usersToShelf := range usersToShelves {
			permissionByRootShelfId[usersToShelf.RootShelfId] = usersToShelf.Permission
		}

		for index, rootShelf := range rootShelves {
			permission, exist := permissionByRootShelfId[rootShelf.Id]
			if !exist {
				return nil, nil, apiexceptions.NewShelfException().NotFound()
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
	opts ...options.RepositoryOptions,
) (*schemas.RootShelf, enums.AccessControlPermission, *cexceptions.Exception) {
	return r.CheckPermissionAndGetOneById(
		id,
		userId,
		preloads,
		options.ParseRepositoryOptions(opts...).AllowedPermissions,
		opts...,
	)
}

func (r *RootShelfRepository) CreateOne(
	ownerId uuid.UUID,
	input inputs.CreateRootShelfInput,
	opts ...options.RepositoryOptions,
) (*uuid.UUID, *cexceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	var newRootShelf schemas.RootShelf
	newRootShelf.OwnerId = ownerId
	if err := copier.Copy(&newRootShelf, &input); err != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewShelfException().FailedToCreate().WithOrigin(err)
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
			return nil, apiexceptions.NewShelfException().DuplicateName(input.Name)
		default:
			return nil, apiexceptions.NewShelfException().FailedToCreate().WithOrigin(err)
		}
	}

	// create the users to shelves relation with the permission of admin
	newUsersToShelves := schemas.UsersToShelves{
		UserId:      ownerId,
		RootShelfId: newRootShelf.Id,
		Permission:  enums.AccessControlPermission_Owner,
	}
	result = parsedOptions.DB.Model(&schemas.UsersToShelves{}).
		Create(&newUsersToShelves)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &newRootShelf.Id, nil
}

func (r *RootShelfRepository) CreateMany(
	ownerId uuid.UUID,
	input []inputs.CreateRootShelfInput,
	opts ...options.RepositoryOptions,
) ([]uuid.UUID, *cexceptions.Exception) {
	if len(input) == 0 {
		return nil, apiexceptions.NewShelfException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	var newRootShelves []schemas.RootShelf
	for _, in := range input {
		var newRootShelf schemas.RootShelf
		newRootShelf.OwnerId = ownerId
		if err := copier.Copy(&newRootShelf, &in); err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
		}
		newRootShelves = append(newRootShelves, newRootShelf)
	}

	result := parsedOptions.DB.Model(&schemas.RootShelf{}).
		CreateInBatches(&newRootShelves, parsedOptions.BatchSize)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
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
			Permission:  enums.AccessControlPermission_Owner,
		}
	}
	result = parsedOptions.DB.Model(&schemas.UsersToShelves{}).
		CreateInBatches(&newUsersToShelves, parsedOptions.BatchSize)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return newRootShelfIds, nil
}

func (r *RootShelfRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateRootShelfInput,
	opts ...options.RepositoryOptions,
) (*schemas.RootShelf, *cexceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if !parsedOptions.IsTransactionStarted {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &updates, nil
}

func (r *RootShelfRepository) UpdateManyByIds(
	userId uuid.UUID,
	input []inputs.UpdateRootShelfByIdInput,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
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
			return apiexceptions.NewShelfException().NoPermission("update these root shelves")
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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return nil
}

func (r *RootShelfRepository) RestoreSoftDeletedOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) (*schemas.RootShelf, *cexceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)},
		{First: restoredRootShelf.Id == uuid.Nil, Second: apiexceptions.NewShelfException().FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &restoredRootShelf, nil
}

func (r *RootShelfRepository) RestoreSoftDeletedManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) ([]schemas.RootShelf, *cexceptions.Exception) {
	if len(ids) == 0 {
		return nil, apiexceptions.NewShelfException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)},
		{First: len(restoredRootShelves) != len(ids), Second: apiexceptions.NewShelfException().FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return restoredRootShelves, nil
}

func (r *RootShelfRepository) SoftDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RootShelfRepository) SoftDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	if len(ids) == 0 {
		return apiexceptions.NewShelfException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RootShelfRepository) SoftDeleteManyByUserId(
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.RootShelf{}).
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where("owner_id = ?", userId).
		Delete(&schemas.RootShelf{})
	if err := result.Error; err != nil {
		return apiexceptions.NewShelfException().FailedToDelete().WithOrigin(err)
	}
	if result.RowsAffected == 0 {
		return apiexceptions.NewShelfException().NotFound()
	}

	return nil
}

func (r *RootShelfRepository) HardDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RootShelfRepository) HardDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	if len(ids) == 0 {
		return apiexceptions.NewShelfException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RootShelfRepository) HardDeleteManyByUserId(
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.RootShelf{}).
		Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where("owner_id = ?", userId).
		Delete(&schemas.RootShelf{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

/* ============================== System Only Method ============================== */

func (r *RootShelfRepository) BulkCheckPermissionsAndGetManyByIds(
	inputs []inputs.BulkCheckRootShelfPermissionInput,
	preloads []schemas.RootShelfRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]bool, []schemas.RootShelf, *cexceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, []schemas.RootShelf{}, nil
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
		result := parsedOptions.DB.Model(&schemas.RootShelf{}).
			Select(`"RootShelfTable".id, uts.user_id`).
			Joins(`INNER JOIN "UsersToShelvesTable" AS uts ON uts.root_shelf_id = "RootShelfTable".id`).
			Where(`"RootShelfTable".id IN ? AND uts.user_id IN ? AND uts.permission IN ?`, ids, userIds, allowedPermissions).
			Scopes(r.rootShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
			Scan(&validTargets)
		if result.Error != nil {
			return nil, nil, apiexceptions.NewShelfException().NotFound().WithOrigin(result.Error)
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
			return nil, nil, apiexceptions.NewShelfException().NotFound().WithOrigin(result.Error)
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
		return nil, nil, apiexceptions.NewShelfException().NotFound().WithOrigin(result.Error)
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

func (r *RootShelfRepository) BulkCreateMany(
	inputs []inputs.BulkCreateRootShelfInput,
	opts ...options.RepositoryOptions,
) ([]bool, *cexceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, apiexceptions.NewShelfException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	newRootShelves := make([]schemas.RootShelf, len(inputs))
	newUsersToShelves := make([]schemas.UsersToShelves, len(inputs))
	for index, in := range inputs {
		newRootShelfId := uuid.New()
		if in.Id != nil && *in.Id != uuid.Nil {
			newRootShelfId = *in.Id
		}

		newRootShelves[index] = schemas.RootShelf{
			Id:             newRootShelfId,
			OwnerId:        in.UserId,
			Name:           in.Name,
			LastAnalyzedAt: time.Now(),
		}
		if in.LastAnalyzedAt != nil {
			newRootShelves[index].LastAnalyzedAt = *in.LastAnalyzedAt
		}

		newUsersToShelves[index] = schemas.UsersToShelves{
			UserId:      in.UserId,
			RootShelfId: newRootShelfId,
			Permission:  enums.AccessControlPermission_Owner,
		}
	}

	result := parsedOptions.DB.Model(&schemas.RootShelf{}).
		CreateInBatches(&newRootShelves, parsedOptions.BatchSize)
	if result.Error != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewShelfException().FailedToCreate().WithOrigin(result.Error)
	}

	result = parsedOptions.DB.Model(&schemas.UsersToShelves{}).
		CreateInBatches(&newUsersToShelves, parsedOptions.BatchSize)
	if result.Error != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewShelfException().FailedToCreate().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	successes := make([]bool, len(inputs))
	for index := range successes {
		successes[index] = true
	}

	return successes, nil
}

func (r *RootShelfRepository) BulkUpdateMany(
	bulkInputs []inputs.BulkUpdateRootShelfInput,
	opts ...options.RepositoryOptions,
) ([]bool, *cexceptions.Exception) {
	if len(bulkInputs) == 0 {
		return []bool{}, apiexceptions.NewShelfException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	checkInputs := make([]inputs.BulkCheckRootShelfPermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckRootShelfPermissionInput{
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

	valuePlaceholders := make([]string, 0, len(bulkInputs))
	valueArgs := make([]interface{}, 0, len(bulkInputs)*6)
	for index, in := range bulkInputs {
		if !successes[index] {
			continue
		}

		valuePlaceholders = append(valuePlaceholders, "(?::int, ?::uuid, ?::text, ?::bigint, ?::bigint, ?::timestamptz)")
		valueArgs = append(valueArgs,
			index,
			in.Id,
			in.PartialUpdateInput.Values.Name,
			in.PartialUpdateInput.Values.SubShelfCount,
			in.PartialUpdateInput.Values.ItemCount,
			in.PartialUpdateInput.Values.LastAnalyzedAt,
		)
	}
	if len(valuePlaceholders) == 0 {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return successes, nil
	}

	sql := fmt.Sprintf(`
		WITH payload(idx, id, name, sub_shelf_count, item_count, last_analyzed_at) AS (
			VALUES %s
		),
		updated AS (
			UPDATE "RootShelfTable" AS rs
			SET
				name = COALESCE(v.name::text, rs.name),
				sub_shelf_count = COALESCE(v.sub_shelf_count::bigint, rs.sub_shelf_count),
				item_count = COALESCE(v.item_count::bigint, rs.item_count),
				last_analyzed_at = COALESCE(v.last_analyzed_at::timestamptz, rs.last_analyzed_at),
				updated_at = NOW()
			FROM payload AS v
			WHERE rs.id = v.id::uuid
				AND rs.deleted_at IS NULL
			RETURNING rs.id
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
		return nil, apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
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
