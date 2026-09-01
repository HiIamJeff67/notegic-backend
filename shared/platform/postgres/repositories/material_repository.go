package repositories

import (
	"net/http"
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

type MaterialRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.MaterialRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) (*schemas.Material, *cexceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.MaterialRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]schemas.Material, *cexceptions.Exception)
	GetOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) (*schemas.Material, *cexceptions.Exception)
	CreateOneBySubShelfId(subShelfId uuid.UUID, userId uuid.UUID, input inputs.CreateMaterialInput, opts ...RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateMaterialInput, opts ...RepositoryOptions) (*schemas.Material, *cexceptions.Exception)
	RestoreSoftDeletedOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) (*schemas.Material, *cexceptions.Exception)
	RestoreSoftDeletedManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) ([]schemas.Material, *cexceptions.Exception)
	SoftDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	SoftDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception

	/* ============================== System Only Method ============================== */

	BulkCheckPermissionsAndGetManyByIds(inputs []inputs.BulkCheckMaterialPermissionInput, preloads []schemas.MaterialRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]bool, []schemas.Material, *cexceptions.Exception)
	BulkCreateMany(inputs []inputs.BulkCreateMaterialInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkUpdateMany(inputs []inputs.BulkUpdateMaterialInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkDeleteMany(inputs []inputs.BulkDeleteMaterialInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
}

type MaterialRepository struct {
	db *gorm.DB
	BulkMaterialRepository
	materialScope scopes.MaterialScopeInterface
	exceptions    exceptions.MaterialException
}

func NewMaterialRepository(
	db *gorm.DB,
	materialScope scopes.MaterialScopeInterface,
) MaterialRepositoryInterface {
	return &MaterialRepository{
		db:                     db,
		BulkMaterialRepository: *NewBulkMaterialRepository(db, materialScope),
		materialScope:          materialScope,
		exceptions:             exceptions.NewMaterialException(),
	}
}

func (r *MaterialRepository) HasPermission(
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
		Model(&schemas.Material{}).
		Select("1").
		Scopes(r.materialScope.PassPermissionCheck(id, userId, allowedPermissions)).
		Scopes(r.materialScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Limit(1).
		Scan(&marker)
	if err := result.Error; err != nil {
		return false
	}

	return marker == 1
}

func (r *MaterialRepository) HavePermissions(
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
		Model(&schemas.Material{}).
		Select(`DISTINCT "MaterialTable".id`).
		Scopes(r.materialScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.materialScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&permittedIds)
	if err := result.Error; err != nil {
		return false
	}

	return array.GetDistinctCount(ids) == array.GetDistinctCount(permittedIds)
}

func (r *MaterialRepository) CheckPermissionAndGetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.MaterialRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) (*schemas.Material, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var material schemas.Material
	query := parsedOptions.DB.
		Model(&schemas.Material{}).
		Where(`"MaterialTable".id = ?`, id)
	if allowedPermissions != nil && len(allowedPermissions) > 0 {
		query = query.Scopes(
			r.materialScope.PassPermissionCheck(id, userId, allowedPermissions),
		)
	}

	result := query.
		Scopes(r.materialScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.materialScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&material)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: material.Id == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &material, nil
}

func (r *MaterialRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.MaterialRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]schemas.Material, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var materials []schemas.Material
	result := parsedOptions.DB.
		Model(&schemas.Material{}).
		Scopes(r.materialScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.materialScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.materialScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&materials)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: len(materials) == 0, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return materials, nil
}

func (r *MaterialRepository) GetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) (*schemas.Material, *cexceptions.Exception) {
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

func (r *MaterialRepository) CreateOneBySubShelfId(
	subShelfId uuid.UUID,
	userId uuid.UUID,
	input inputs.CreateMaterialInput,
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
			return nil, r.exceptions.NoPermission("create a material under this shelf")
		}
	}

	var newMaterial schemas.Material
	if err := copier.Copy(&newMaterial, &input); err != nil {
		parsedOptions.DB.Rollback()
		return nil, r.exceptions.FailedToCreate().WithOrigin(err)
	}
	newMaterial.ParentSubShelfId = subShelfId

	result := parsedOptions.DB.Model(&schemas.Material{}).
		Create(&newMaterial)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToCreate().WithOrigin(result.Error)},
		{First: newMaterial.Id == uuid.Nil, Second: r.exceptions.FailedToCreate()},
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

	return &newMaterial.Id, nil
}

func (r *MaterialRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateMaterialInput,
	opts ...RepositoryOptions,
) (*schemas.Material, *cexceptions.Exception) {
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

	// get and check the permission of the current user to the source shelf
	existingMaterial, exception := r.CheckPermissionAndGetOneById(
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
	if existingMaterial == nil {
		parsedOptions.DB.Rollback()
		return nil, r.exceptions.NotFound()
	}

	// if the root shelf id is required to be updated in the database
	if input.Values.ParentSubShelfId != nil && !partialupdate.CheckSetNull(input.SetNull, "ParentSubShelfId") {
		subShelfRepository := NewSubShelfRepository(r.db, scopes.NewSubShelfScope())
		// check if the user has the enough permission to the destination shelf
		if !subShelfRepository.HasPermission(
			*input.Values.ParentSubShelfId,
			userId,
			parsedOptions.AllowedPermissions,
			opts...,
		) {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.NoPermission("move a material to this shelf")
		}
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingMaterial)
	if err != nil {
		parsedOptions.DB.Rollback()
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true).WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.Material{}).
		Where("id = ? AND deleted_at IS NULL", id). // no need to check the permission here, since we have done that part on the above
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

func (r *MaterialRepository) RestoreSoftDeletedOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) (*schemas.Material, *cexceptions.Exception) {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var restoredMaterial schemas.Material
	query := parsedOptions.DB.Model(&restoredMaterial).
		Scopes(r.materialScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted))
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.materialScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions))
	}

	result := query.
		Clauses(clause.Returning{}).
		Where(`"MaterialTable".id = ?`, id).
		Updates(map[string]interface{}{"deleted_at": nil}) // force to assign null value
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: restoredMaterial.Id == uuid.Nil, Second: r.exceptions.FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &restoredMaterial, nil
}

func (r *MaterialRepository) RestoreSoftDeletedManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) ([]schemas.Material, *cexceptions.Exception) {
	if len(ids) == 0 {
		return nil, r.exceptions.NoChanges()
	}

	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	var restoredMaterials []schemas.Material
	query := parsedOptions.DB.Model(&restoredMaterials).
		Scopes(r.materialScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted))
	if parsedOptions.HasAllowedPermissions() {
		query = query.Scopes(r.materialScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions))
	}

	result := query.
		Clauses(clause.Returning{}).
		Where(`"MaterialTable".id IN ?`, ids).
		Updates(map[string]interface{}{"deleted_at": nil}) // force to assign null value
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: len(restoredMaterials) != len(ids), Second: r.exceptions.FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return restoredMaterials, nil
}

func (r *MaterialRepository) SoftDeleteOneById(
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

	result := parsedOptions.DB.Model(&schemas.Material{}).
		Scopes(r.materialScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.materialScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"MaterialTable".id = ?`, id).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *MaterialRepository) SoftDeleteManyByIds(
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

	result := parsedOptions.DB.Model(&schemas.Material{}).
		Scopes(r.materialScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.materialScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"MaterialTable".id IN ?`, ids).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *MaterialRepository) HardDeleteOneById(
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

	result := parsedOptions.DB.Model(&schemas.Material{}).
		Scopes(r.materialScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.materialScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"MaterialTable".id = ?`, id).
		Delete(&schemas.Material{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *MaterialRepository) HardDeleteManyByIds(
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

	result := parsedOptions.DB.Model(&schemas.Material{}).
		Scopes(r.materialScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.materialScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"MaterialTable".id IN ?`, ids).
		Delete(&schemas.Material{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

/* ============================== System Only Method ============================== */
