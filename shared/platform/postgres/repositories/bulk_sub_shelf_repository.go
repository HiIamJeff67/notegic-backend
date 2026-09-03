package repositories

import (
	"fmt"
	"strings"
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

type BulkSubShelfRepositoryInterface interface {
	BulkCheckPermissionsAndGetManyByIds(
		inputs []inputs.BulkCheckSubShelfPermissionInput,
		preloads []schemas.SubShelfRelation,
		allowedPermissions []cenums.AccessControlPermission,
		opts ...RepositoryOptions,
	) ([]bool, []schemas.SubShelf, *cexceptions.Exception)
	BulkCreateMany(
		inputs []inputs.BulkCreateSubShelfInput,
		opts ...RepositoryOptions,
	) ([]bool, *cexceptions.Exception)
	BulkUpdateMany(
		inputs []inputs.BulkUpdateSubShelfInput,
		opts ...RepositoryOptions,
	) ([]bool, *cexceptions.Exception)
	BulkDeleteMany(
		inputs []inputs.BulkDeleteSubShelfInput,
		opts ...RepositoryOptions,
	) ([]bool, *cexceptions.Exception)
}

type BulkSubShelfRepository struct {
	db            *gorm.DB
	subShelfScope scopes.SubShelfScopeInterface
	exceptions    exceptions.ShelfException
}

func NewBulkSubShelfRepository(
	db *gorm.DB,
	subShelfScope scopes.SubShelfScopeInterface,
) *BulkSubShelfRepository {
	return &BulkSubShelfRepository{
		db:            db,
		subShelfScope: subShelfScope,
		exceptions:    exceptions.NewShelfException(),
	}
}

func (r *BulkSubShelfRepository) BulkCheckPermissionsAndGetManyByIds(
	inputs []inputs.BulkCheckSubShelfPermissionInput,
	preloads []schemas.SubShelfRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]bool, []schemas.SubShelf, *cexceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, []schemas.SubShelf{}, nil
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
		result := parsedOptions.DB.Model(&schemas.SubShelf{}).
			Select(`"SubShelfTable".id, uts.user_id`).
			Joins(`INNER JOIN "UsersToShelvesTable" AS uts ON uts.root_shelf_id = "SubShelfTable".root_shelf_id`).
			Where(`"SubShelfTable".id IN ?`, ids).
			Where("uts.user_id IN ? AND uts.permission IN ?", userIds, allowedPermissions).
			Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
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
		result := parsedOptions.DB.Model(&schemas.SubShelf{}).
			Select(`"SubShelfTable".id`).
			Where(`"SubShelfTable".id IN ?`, ids).
			Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
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
		return successes, []schemas.SubShelf{}, nil
	}

	var subShelves []schemas.SubShelf
	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Where(`"SubShelfTable".id IN ?`, validIds).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.subShelfScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&subShelves)
	if result.Error != nil {
		return nil, nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	foundIdSet := make(map[uuid.UUID]bool, len(subShelves))
	for _, subShelf := range subShelves {
		foundIdSet[subShelf.Id] = true
	}
	for index, in := range inputs {
		if validIdSet[in.Id] &&
			foundIdSet[in.Id] &&
			(allowedPermissions == nil || validTargetByUserId[[2]uuid.UUID{in.Id, in.UserId}]) {
			successes[index] = true
		}
	}

	return successes, subShelves, nil
}

func (r *BulkSubShelfRepository) BulkCreateMany(
	inputs []inputs.BulkCreateSubShelfInput,
	opts ...RepositoryOptions,
) ([]bool, *cexceptions.Exception) {
	if len(inputs) == 0 {
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

	successes := make([]bool, len(inputs))
	rootShelfIds := make([]uuid.UUID, 0, len(inputs))
	userIds := make([]uuid.UUID, 0, len(inputs))
	prevSubShelfIds := make([]uuid.UUID, 0, len(inputs))
	for _, in := range inputs {
		rootShelfIds = append(rootShelfIds, in.RootShelfId)
		userIds = append(userIds, in.UserId)
		if in.PrevSubShelfId != nil && *in.PrevSubShelfId != uuid.Nil {
			prevSubShelfIds = append(prevSubShelfIds, *in.PrevSubShelfId)
		}
	}

	var usersToShelves []schemas.UsersToShelves
	result := parsedOptions.DB.Model(&schemas.UsersToShelves{}).
		Where("root_shelf_id IN ? AND user_id IN ? AND permission IN ?", rootShelfIds, userIds, parsedOptions.AllowedPermissions).
		Find(&usersToShelves)
	if result.Error != nil {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return nil, r.exceptions.FailedToCreate().WithOrigin(result.Error)
	}

	validRootShelfByUserId := make(map[[2]uuid.UUID]bool, len(usersToShelves))
	for _, usersToShelf := range usersToShelves {
		validRootShelfByUserId[[2]uuid.UUID{usersToShelf.RootShelfId, usersToShelf.UserId}] = true
	}

	prevSubShelfById := make(map[uuid.UUID]schemas.SubShelf)
	if len(prevSubShelfIds) > 0 {
		var prevSubShelves []schemas.SubShelf
		result = parsedOptions.DB.Model(&schemas.SubShelf{}).
			Where("id IN ? AND deleted_at IS NULL", prevSubShelfIds).
			Find(&prevSubShelves)
		if result.Error != nil {
			if shouldStartTransaction {
				parsedOptions.DB.Rollback()
			}
			return nil, r.exceptions.FailedToCreate().WithOrigin(result.Error)
		}
		for _, prevSubShelf := range prevSubShelves {
			prevSubShelfById[prevSubShelf.Id] = prevSubShelf
		}
	}

	newSubShelves := make([]schemas.SubShelf, 0, len(inputs))
	successIndexes := make([]int, 0, len(inputs))
	for index, in := range inputs {
		if !validRootShelfByUserId[[2]uuid.UUID{in.RootShelfId, in.UserId}] {
			continue
		}

		newSubShelfId := uuid.New()
		if in.Id != nil && *in.Id != uuid.Nil {
			newSubShelfId = *in.Id
		}

		newSubShelf := schemas.SubShelf{
			Id:             newSubShelfId,
			RootShelfId:    in.RootShelfId,
			PrevSubShelfId: in.PrevSubShelfId,
			Name:           in.Name,
			Path:           types.UUIDArray{},
		}

		if in.Path != nil {
			newSubShelf.Path = *in.Path
		} else if in.PrevSubShelfId != nil && *in.PrevSubShelfId != uuid.Nil {
			prevSubShelf, exist := prevSubShelfById[*in.PrevSubShelfId]
			if !exist || prevSubShelf.RootShelfId != in.RootShelfId {
				continue
			}
			newSubShelf.Path = append(prevSubShelf.Path, prevSubShelf.Id)
		}

		newSubShelves = append(newSubShelves, newSubShelf)
		successIndexes = append(successIndexes, index)
	}

	if len(newSubShelves) == 0 {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return successes, nil
	}

	result = parsedOptions.DB.Model(&schemas.SubShelf{}).
		CreateInBatches(&newSubShelves, parsedOptions.BatchSize)
	if result.Error != nil {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return nil, r.exceptions.FailedToCreate().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	for _, successIndex := range successIndexes {
		successes[successIndex] = true
	}

	return successes, nil
}

func (r *BulkSubShelfRepository) BulkUpdateMany(
	bulkInputs []inputs.BulkUpdateSubShelfInput,
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

	checkInputs := make([]inputs.BulkCheckSubShelfPermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckSubShelfPermissionInput{
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
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return nil, exception
	}

	valuePlaceholders := make([]string, 0, len(bulkInputs))
	valueArgs := make([]interface{}, 0, len(bulkInputs)*3)
	for index, in := range bulkInputs {
		if !successes[index] {
			continue
		}

		valuePlaceholders = append(valuePlaceholders, "(?::int, ?::uuid, ?::text)")
		valueArgs = append(valueArgs,
			index,
			in.Id,
			in.PartialUpdateInput.Values.Name,
		)
	}
	if len(valuePlaceholders) == 0 {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return successes, nil
	}

	sql := fmt.Sprintf(`
		WITH payload(idx, id, name) AS (
			VALUES %s
		),
		updated AS (
			UPDATE "SubShelfTable" AS ss
			SET
				name = COALESCE(v.name::text, ss.name),
				updated_at = NOW()
			FROM payload AS v
			WHERE ss.id = v.id::uuid
				AND ss.deleted_at IS NULL
			RETURNING ss.id
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
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return nil, r.exceptions.FailedToUpdate().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
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

func (r *BulkSubShelfRepository) BulkDeleteMany(
	bulkInputs []inputs.BulkDeleteSubShelfInput,
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

	checkInputs := make([]inputs.BulkCheckSubShelfPermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckSubShelfPermissionInput{
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
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
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

	var deletedSubShelves []schemas.SubShelf
	result := parsedOptions.DB.Model(&deletedSubShelves).
		Clauses(clause.Returning{}).
		Where("id IN ? AND deleted_at IS NULL", validIds).
		Updates(map[string]interface{}{"deleted_at": time.Now(), "updated_at": time.Now()})
	if result.Error != nil {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return nil, r.exceptions.FailedToDelete().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	deletedIdSet := make(map[uuid.UUID]bool, len(deletedSubShelves))
	for _, deletedSubShelf := range deletedSubShelves {
		deletedIdSet[deletedSubShelf.Id] = true
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
