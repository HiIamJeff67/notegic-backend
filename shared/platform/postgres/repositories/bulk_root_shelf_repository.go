package repositories

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	"github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"
)

type BulkRootShelfRepositoryInterface interface {
	BulkCheckPermissionsAndGetManyByIds(
		inputs []inputs.BulkCheckRootShelfPermissionInput,
		preloads []schemas.RootShelfRelation,
		allowedPermissions []cenums.AccessControlPermission,
		opts ...RepositoryOptions,
	) ([]bool, []schemas.RootShelf, *cexceptions.Exception)
	BulkCreateMany(
		inputs []inputs.BulkCreateRootShelfInput,
		opts ...RepositoryOptions,
	) ([]bool, *cexceptions.Exception)
	BulkUpdateMany(
		inputs []inputs.BulkUpdateRootShelfInput,
		opts ...RepositoryOptions,
	) ([]bool, *cexceptions.Exception)
}

type BulkRootShelfRepository struct {
	db             *gorm.DB
	rootShelfScope scopes.RootShelfScopeInterface
	exceptions     exceptions.ShelfException
}

func NewBulkRootShelfRepository(
	db *gorm.DB,
	rootShelfScope scopes.RootShelfScopeInterface,
) *BulkRootShelfRepository {
	return &BulkRootShelfRepository{
		db:             db,
		rootShelfScope: rootShelfScope,
		exceptions:     exceptions.NewShelfException(),
	}
}

func (r *BulkRootShelfRepository) BulkCheckPermissionsAndGetManyByIds(
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

func (r *BulkRootShelfRepository) BulkCreateMany(
	inputs []inputs.BulkCreateRootShelfInput,
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
			Permission:  cenums.AccessControlPermission_Owner,
		}
	}

	result := parsedOptions.DB.Model(&schemas.RootShelf{}).
		CreateInBatches(&newRootShelves, parsedOptions.BatchSize)
	if result.Error != nil {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return nil, r.exceptions.FailedToCreate().WithOrigin(result.Error)
	}

	result = parsedOptions.DB.Model(&schemas.UsersToShelves{}).
		CreateInBatches(&newUsersToShelves, parsedOptions.BatchSize)
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

	successes := make([]bool, len(inputs))
	for index := range successes {
		successes[index] = true
	}

	return successes, nil
}

func (r *BulkRootShelfRepository) BulkUpdateMany(
	bulkInputs []inputs.BulkUpdateRootShelfInput,
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

	checkInputs := make([]inputs.BulkCheckRootShelfPermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckRootShelfPermissionInput{
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
