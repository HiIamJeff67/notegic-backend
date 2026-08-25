package repositories

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	"github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"
)

type MaterialBulkRepositoryInterface interface {
	BulkCheckPermissionsAndGetManyByIds(inputs []inputs.BulkCheckMaterialPermissionInput, preloads []schemas.MaterialRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]bool, []schemas.Material, *cexceptions.Exception)
	BulkDeleteMany(inputs []inputs.BulkDeleteMaterialInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
}

type MaterialBulkRepository struct {
	db            *gorm.DB
	materialScope scopes.MaterialScopeInterface
	exceptions    exceptions.MaterialException
}

func NewMaterialBulkRepository(
	materialScope scopes.MaterialScopeInterface,
	repositoryExceptions ...exceptions.MaterialException,
) *MaterialBulkRepository {
	return NewMaterialBulkRepositoryWithDB(nil, materialScope, repositoryExceptions...)
}

func NewMaterialBulkRepositoryWithDB(
	db *gorm.DB,
	materialScope scopes.MaterialScopeInterface,
	repositoryExceptions ...exceptions.MaterialException,
) *MaterialBulkRepository {
	repositoryException := exceptions.NewMaterialException()
	if len(repositoryExceptions) > 0 {
		repositoryException = repositoryExceptions[0]
	}

	return &MaterialBulkRepository{
		db:            db,
		materialScope: materialScope,
		exceptions:    repositoryException,
	}
}

func (r *MaterialBulkRepository) BulkCheckPermissionsAndGetManyByIds(
	inputs []inputs.BulkCheckMaterialPermissionInput,
	preloads []schemas.MaterialRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]bool, []schemas.Material, *cexceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, []schemas.Material{}, nil
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
		result := parsedOptions.DB.Model(&schemas.Material{}).
			Select(`"MaterialTable".id, uts.user_id`).
			Joins(`INNER JOIN "SubShelfTable" AS ss ON ss.id = "MaterialTable".parent_sub_shelf_id`).
			Joins(`INNER JOIN "UsersToShelvesTable" AS uts ON uts.root_shelf_id = ss.root_shelf_id`).
			Where(`"MaterialTable".id IN ? AND uts.user_id IN ? AND uts.permission IN ?`, ids, userIds, allowedPermissions).
			Scopes(r.materialScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
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
		result := parsedOptions.DB.Model(&schemas.Material{}).
			Select(`"MaterialTable".id`).
			Where(`"MaterialTable".id IN ?`, ids).
			Scopes(r.materialScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
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
		return successes, []schemas.Material{}, nil
	}

	var materials []schemas.Material
	result := parsedOptions.DB.Model(&schemas.Material{}).
		Where(`"MaterialTable".id IN ?`, validIds).
		Scopes(r.materialScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.materialScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&materials)
	if result.Error != nil {
		return nil, nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	foundIdSet := make(map[uuid.UUID]bool, len(materials))
	for _, material := range materials {
		foundIdSet[material.Id] = true
	}
	for index, in := range inputs {
		if validIdSet[in.Id] &&
			foundIdSet[in.Id] &&
			(allowedPermissions == nil || validTargetByUserId[[2]uuid.UUID{in.Id, in.UserId}]) {
			successes[index] = true
		}
	}

	return successes, materials, nil
}

func (r *MaterialBulkRepository) BulkDeleteMany(
	bulkInputs []inputs.BulkDeleteMaterialInput,
	opts ...RepositoryOptions,
) ([]bool, *cexceptions.Exception) {
	if len(bulkInputs) == 0 {
		return []bool{}, r.exceptions.NoChanges()
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	checkInputs := make([]inputs.BulkCheckMaterialPermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckMaterialPermissionInput{
			UserId: in.UserId,
			Id:     in.Id,
		}
	}
	checkOptions := append(opts, WithTransactionDB(parsedOptions.DB))
	checkOptions = append(checkOptions, WithOnlyDeleted(types.Ternary_Negative))
	checkOptions = append(checkOptions, WithLockingStrength(LockingStrengthNoKeyUpdate))
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

	var deletedMaterials []schemas.Material
	result := parsedOptions.DB.Model(&deletedMaterials).
		Clauses(clause.Returning{}).
		Where("id IN ? AND deleted_at IS NULL", validIds).
		Updates(map[string]interface{}{"deleted_at": time.Now(), "updated_at": time.Now()})
	if result.Error != nil {
		parsedOptions.DB.Rollback()
		return nil, r.exceptions.FailedToDelete().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	deletedIdSet := make(map[uuid.UUID]bool, len(deletedMaterials))
	for _, deletedMaterial := range deletedMaterials {
		deletedIdSet[deletedMaterial.Id] = true
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
