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

	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"
	"github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"
)

type RoutineBulkRepositoryInterface interface {
	BulkCreateMany(inputs []inputs.BulkCreateRoutineInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkUpdateMany(inputs []inputs.BulkUpdateRoutineInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkDeleteMany(inputs []inputs.BulkDeleteRoutineInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
}

func (r *RoutineBulkRepository) BulkCheckPermissionsAndGetManyByIds(
	inputs []inputs.BulkCheckRoutinePermissionInput,
	preloads []schemas.RoutineRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]bool, []schemas.Routine, *cexceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, []schemas.Routine{}, nil
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

	var validTargets []struct {
		Id     uuid.UUID `gorm:"column:id"`
		UserId uuid.UUID `gorm:"column:user_id"`
	}
	result := parsedOptions.DB.Model(&schemas.Routine{}).
		Select(`"RoutineTable".id, uts.user_id`).
		Joins(`INNER JOIN "UsersToStationsTable" AS uts ON uts.station_id = "RoutineTable".station_id`).
		Where(`"RoutineTable".id IN ?`, ids).
		Where("uts.user_id IN ? AND uts.permission IN ?", userIds, allowedPermissions).
		Scopes(r.routineScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scan(&validTargets)
	if result.Error != nil {
		return nil, nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	validTargetByUserId := make(map[[2]uuid.UUID]bool, len(validTargets))
	for _, validTarget := range validTargets {
		validTargetByUserId[[2]uuid.UUID{validTarget.Id, validTarget.UserId}] = true
	}

	validIdSet := make(map[uuid.UUID]bool, len(validTargets))
	for _, in := range inputs {
		if validTargetByUserId[[2]uuid.UUID{in.Id, in.UserId}] {
			validIdSet[in.Id] = true
		}
	}

	validIds := make([]uuid.UUID, 0, len(validIdSet))
	for validId := range validIdSet {
		validIds = append(validIds, validId)
	}
	if len(validIds) == 0 {
		return successes, []schemas.Routine{}, nil
	}

	var routines []schemas.Routine
	result = parsedOptions.DB.Model(&schemas.Routine{}).
		Where(`"RoutineTable".id IN ?`, validIds).
		Scopes(r.routineScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.routineScope.IncludePreloads(preloads, nil)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&routines)
	if result.Error != nil {
		return nil, nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	foundIdSet := make(map[uuid.UUID]bool, len(routines))
	for _, routine := range routines {
		foundIdSet[routine.Id] = true
	}
	for index, in := range inputs {
		if validTargetByUserId[[2]uuid.UUID{in.Id, in.UserId}] && foundIdSet[in.Id] {
			successes[index] = true
		}
	}

	return successes, routines, nil
}

type RoutineBulkRepository struct {
	db           *gorm.DB
	routineScope scopes.RoutineScopeInterface
	exceptions   exceptions.RoutineException
}

func NewRoutineBulkRepository(
	routineScope scopes.RoutineScopeInterface,
	repositoryExceptions ...exceptions.RoutineException,
) *RoutineBulkRepository {
	return NewRoutineBulkRepositoryWithDB(nil, routineScope, repositoryExceptions...)
}

func NewRoutineBulkRepositoryWithDB(
	db *gorm.DB,
	routineScope scopes.RoutineScopeInterface,
	repositoryExceptions ...exceptions.RoutineException,
) *RoutineBulkRepository {
	repositoryException := exceptions.NewRoutineException()
	if len(repositoryExceptions) > 0 {
		repositoryException = repositoryExceptions[0]
	}

	return &RoutineBulkRepository{
		db:           db,
		routineScope: routineScope,
		exceptions:   repositoryException,
	}
}

func (r *RoutineBulkRepository) BulkCreateMany(
	inputs []inputs.BulkCreateRoutineInput,
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

	now := time.Now().Truncate(time.Minute)
	successes := make([]bool, len(inputs))
	stationIds := make([]uuid.UUID, 0, len(inputs))
	userIds := make([]uuid.UUID, 0, len(inputs))
	for _, in := range inputs {
		stationIds = append(stationIds, in.StationId)
		userIds = append(userIds, in.UserId)
	}

	var validTargets []struct {
		Id     uuid.UUID `gorm:"column:id"`
		UserId uuid.UUID `gorm:"column:user_id"`
	}
	result := parsedOptions.DB.Model(&schemas.Station{}).
		Select(`"StationTable".id, uts.user_id`).
		Joins(`INNER JOIN "UsersToStationsTable" AS uts ON uts.station_id = "StationTable".id`).
		Where(`"StationTable".id IN ? AND "StationTable".deleted_at IS NULL`, stationIds).
		Where("uts.user_id IN ? AND uts.permission IN ?", userIds, parsedOptions.AllowedPermissions).
		Scan(&validTargets)
	if result.Error != nil {
		parsedOptions.DB.Rollback()
		return nil, r.exceptions.FailedToCreate().WithOrigin(result.Error)
	}

	validTargetByUserId := make(map[[2]uuid.UUID]bool, len(validTargets))
	for _, validTarget := range validTargets {
		validTargetByUserId[[2]uuid.UUID{validTarget.Id, validTarget.UserId}] = true
	}

	newRoutines := make([]schemas.Routine, 0, len(inputs))
	successIndexes := make([]int, 0, len(inputs))
	for index, in := range inputs {
		if !validTargetByUserId[[2]uuid.UUID{in.StationId, in.UserId}] {
			continue
		}

		newRoutineId := uuid.New()
		if in.Id != nil && *in.Id != uuid.Nil {
			newRoutineId = *in.Id
		}

		scheduledStartAt := in.ScheduledStartAt
		if scheduledStartAt == nil {
			scheduledStartAt = &now
		} else {
			truncatedScheduledStartAt := scheduledStartAt.Truncate(time.Minute)
			scheduledStartAt = &truncatedScheduledStartAt
		}

		scheduledEndAt := in.ScheduledEndAt
		if scheduledEndAt == nil {
			defaultScheduledEndAt := scheduledStartAt.Add(time.Hour)
			scheduledEndAt = &defaultScheduledEndAt
		} else {
			truncatedScheduledEndAt := scheduledEndAt.Truncate(time.Minute)
			scheduledEndAt = &truncatedScheduledEndAt
		}

		status := cenums.RoutineStatus_Scheduled
		if in.Status != nil {
			status = *in.Status
		}
		isPinned := false
		if in.IsPinned != nil {
			isPinned = *in.IsPinned
		}
		timezone := "UTC"
		if in.Timezone != nil {
			timezone = *in.Timezone
		}

		newRoutines = append(newRoutines, schemas.Routine{
			Id:               newRoutineId,
			StationId:        in.StationId,
			Title:            in.Title,
			Description:      in.Description,
			Status:           status,
			IsPinned:         isPinned,
			ScheduledStartAt: *scheduledStartAt,
			ScheduledEndAt:   *scheduledEndAt,
			Period:           in.Period,
			Timezone:         timezone,
		})
		successIndexes = append(successIndexes, index)
	}

	if len(newRoutines) == 0 {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return successes, nil
	}

	result = parsedOptions.DB.Model(&schemas.Routine{}).
		CreateInBatches(&newRoutines, parsedOptions.BatchSize)
	if result.Error != nil {
		parsedOptions.DB.Rollback()
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

func (r *RoutineBulkRepository) BulkUpdateMany(
	bulkInputs []inputs.BulkUpdateRoutineInput,
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

	checkInputs := make([]inputs.BulkCheckRoutinePermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckRoutinePermissionInput{
			UserId: in.UserId,
			Id:     in.Id,
		}
	}
	checkOptions := append(opts, WithTransactionDB(parsedOptions.DB))
	checkOptions = append(checkOptions, WithOnlyDeleted(types.Ternary_Negative))
	checkOptions = append(checkOptions, WithLockingStrength(LockingStrengthNoKeyUpdate))
	successes, _, exception := r.BulkCheckPermissionsAndGetManyByIds(checkInputs, nil, parsedOptions.AllowedPermissions, checkOptions...)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	targetStationIds := make([]uuid.UUID, 0, len(bulkInputs))
	targetUserIds := make([]uuid.UUID, 0, len(bulkInputs))
	for index, in := range bulkInputs {
		if !successes[index] ||
			in.PartialUpdateInput.Values.StationId == nil ||
			partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "StationId") {
			continue
		}
		targetStationIds = append(targetStationIds, *in.PartialUpdateInput.Values.StationId)
		targetUserIds = append(targetUserIds, in.UserId)
	}
	if len(targetStationIds) > 0 {
		var validTargets []struct {
			Id     uuid.UUID `gorm:"column:id"`
			UserId uuid.UUID `gorm:"column:user_id"`
		}
		result := parsedOptions.DB.Model(&schemas.Station{}).
			Select(`"StationTable".id, uts.user_id`).
			Joins(`INNER JOIN "UsersToStationsTable" AS uts ON uts.station_id = "StationTable".id`).
			Where(`"StationTable".id IN ? AND "StationTable".deleted_at IS NULL`, targetStationIds).
			Where("uts.user_id IN ? AND uts.permission IN ?", targetUserIds, parsedOptions.AllowedPermissions).
			Scan(&validTargets)
		if result.Error != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToUpdate().WithOrigin(result.Error)
		}

		validTargetByUserId := make(map[[2]uuid.UUID]bool, len(validTargets))
		for _, validTarget := range validTargets {
			validTargetByUserId[[2]uuid.UUID{validTarget.Id, validTarget.UserId}] = true
		}
		for index, in := range bulkInputs {
			if !successes[index] ||
				in.PartialUpdateInput.Values.StationId == nil ||
				partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "StationId") {
				continue
			}
			if !validTargetByUserId[[2]uuid.UUID{*in.PartialUpdateInput.Values.StationId, in.UserId}] {
				successes[index] = false
			}
		}
	}

	valuePlaceholders := make([]string, 0, len(bulkInputs))
	valueArgs := make([]interface{}, 0, len(bulkInputs)*12)
	for index, in := range bulkInputs {
		if !successes[index] {
			continue
		}

		setPeriodNull := partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "Period")

		scheduledStartAt := in.PartialUpdateInput.Values.ScheduledStartAt
		if scheduledStartAt != nil {
			truncatedScheduledStartAt := scheduledStartAt.Truncate(time.Minute)
			scheduledStartAt = &truncatedScheduledStartAt
		}

		scheduledEndAt := in.PartialUpdateInput.Values.ScheduledEndAt
		if scheduledEndAt != nil {
			truncatedScheduledEndAt := scheduledEndAt.Truncate(time.Minute)
			scheduledEndAt = &truncatedScheduledEndAt
		}

		valuePlaceholders = append(valuePlaceholders, `(?::int, ?::uuid, ?::uuid, ?::text, ?::text, ?::"RoutineStatus", ?::boolean, ?::timestamptz, ?::timestamptz, ?::"RoutinePeriod", ?::text, ?::boolean)`)
		valueArgs = append(valueArgs,
			index,
			in.Id,
			in.PartialUpdateInput.Values.StationId,
			in.PartialUpdateInput.Values.Title,
			in.PartialUpdateInput.Values.Description,
			in.PartialUpdateInput.Values.Status,
			in.PartialUpdateInput.Values.IsPinned,
			scheduledStartAt,
			scheduledEndAt,
			in.PartialUpdateInput.Values.Period,
			in.PartialUpdateInput.Values.Timezone,
			setPeriodNull,
		)
	}
	if len(valuePlaceholders) == 0 {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return successes, nil
	}

	sql := fmt.Sprintf(`
		WITH payload(idx, id, station_id, title, description, status, is_pinned, scheduled_start_at, scheduled_end_at, period, timezone, set_period_null) AS (
			VALUES %s
		),
		updated AS (
			UPDATE "RoutineTable" AS r
			SET
				station_id = COALESCE(v.station_id::uuid, r.station_id),
				title = COALESCE(v.title::text, r.title),
				description = COALESCE(v.description::text, r.description),
				status = COALESCE(v.status::"RoutineStatus", r.status),
				is_pinned = COALESCE(v.is_pinned::boolean, r.is_pinned),
				scheduled_start_at = COALESCE(v.scheduled_start_at::timestamptz, r.scheduled_start_at),
				scheduled_end_at = COALESCE(v.scheduled_end_at::timestamptz, r.scheduled_end_at),
				period = CASE
					WHEN v.set_period_null::boolean THEN NULL
					ELSE COALESCE(v.period::"RoutinePeriod", r.period)
				END,
				timezone = COALESCE(v.timezone::text, r.timezone),
				updated_at = NOW()
			FROM payload AS v
			WHERE r.id = v.id::uuid
				AND r.deleted_at IS NULL
			RETURNING r.id
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

func (r *RoutineBulkRepository) BulkDeleteMany(
	bulkInputs []inputs.BulkDeleteRoutineInput,
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

	checkInputs := make([]inputs.BulkCheckRoutinePermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckRoutinePermissionInput{
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

	var deletedRoutines []schemas.Routine
	result := parsedOptions.DB.Model(&deletedRoutines).
		Clauses(clause.Returning{}).
		Where(`id IN ? AND deleted_at IS NULL`, validIds).
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

	deletedIdSet := make(map[uuid.UUID]bool, len(deletedRoutines))
	for _, routine := range deletedRoutines {
		deletedIdSet[routine.Id] = true
	}
	for index, in := range bulkInputs {
		successes[index] = successes[index] && deletedIdSet[in.Id]
	}

	return successes, nil
}
