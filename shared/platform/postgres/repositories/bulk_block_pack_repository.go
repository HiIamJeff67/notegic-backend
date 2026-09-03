package repositories

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"
	exceptions "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"
)

type BulkBlockPackRepositoryInterface interface {
	BulkCheckPermissionsAndGetManyByIds(
		inputs []inputs.BulkCheckBlockPackPermissionInput,
		preloads []schemas.BlockPackRelation,
		allowedPermissions []cenums.AccessControlPermission,
		opts ...RepositoryOptions,
	) ([]bool, []schemas.BlockPack, *cexceptions.Exception)
	BulkCreateMany(
		inputs []inputs.BulkCreateBlockPackInput,
		opts ...RepositoryOptions,
	) ([]bool, *cexceptions.Exception)
	BulkUpdateMany(
		inputs []inputs.BulkUpdateBlockPackInput,
		opts ...RepositoryOptions,
	) ([]bool, *cexceptions.Exception)
	BulkDeleteMany(
		inputs []inputs.BulkDeleteBlockPackInput,
		opts ...RepositoryOptions,
	) ([]bool, *cexceptions.Exception)
}

type BulkBlockPackRepository struct {
	db             *gorm.DB
	blockPackScope scopes.BlockPackScopeInterface
	exceptions     exceptions.BlockPackException
}

func NewBulkBlockPackRepository(
	db *gorm.DB,
	blockPackScope scopes.BlockPackScopeInterface,
) *BulkBlockPackRepository {
	return &BulkBlockPackRepository{
		db:             db,
		blockPackScope: blockPackScope,
		exceptions:     exceptions.NewBlockPackException(),
	}
}

func (r *BulkBlockPackRepository) BulkCheckPermissionsAndGetManyByIds(
	inputs []inputs.BulkCheckBlockPackPermissionInput,
	preloads []schemas.BlockPackRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]bool, []schemas.BlockPack, *cexceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, []schemas.BlockPack{}, nil
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
		result := parsedOptions.DB.Model(&schemas.BlockPack{}).
			Select(`"BlockPackTable".id, uts.user_id`).
			Joins(`INNER JOIN "SubShelfTable" AS ss ON ss.id = "BlockPackTable".parent_sub_shelf_id`).
			Joins(`INNER JOIN "UsersToShelvesTable" AS uts ON uts.root_shelf_id = ss.root_shelf_id`).
			Where(`"BlockPackTable".id IN ?`, ids).
			Where("uts.user_id IN ? AND uts.permission IN ?", userIds, allowedPermissions).
			Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
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
		result := parsedOptions.DB.Model(&schemas.BlockPack{}).
			Select(`"BlockPackTable".id`).
			Where(`"BlockPackTable".id IN ?`, ids).
			Scopes(r.blockPackScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
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
		return nil, nil, r.exceptions.NotFound().WithOrigin(result.Error)
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

func (r *BulkBlockPackRepository) BulkCreateMany(
	inputs []inputs.BulkCreateBlockPackInput,
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
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return nil, r.exceptions.FailedToCreate().WithOrigin(result.Error)
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

func (r *BulkBlockPackRepository) BulkUpdateMany(
	bulkInputs []inputs.BulkUpdateBlockPackInput,
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

	checkInputs := make([]inputs.BulkCheckBlockPackPermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckBlockPackPermissionInput{
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
			if shouldStartTransaction {
				parsedOptions.DB.Rollback()
			}
			return nil, r.exceptions.FailedToUpdate().WithOrigin(result.Error)
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

func (r *BulkBlockPackRepository) BulkDeleteMany(
	bulkInputs []inputs.BulkDeleteBlockPackInput,
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

	checkInputs := make([]inputs.BulkCheckBlockPackPermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckBlockPackPermissionInput{
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

	var deletedBlockPacks []schemas.BlockPack
	result := parsedOptions.DB.Model(&deletedBlockPacks).
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
