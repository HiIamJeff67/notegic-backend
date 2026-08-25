package repositories

import (
	"database/sql"
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

type RoutineRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) (*schemas.Routine, *cexceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]schemas.Routine, *cexceptions.Exception)
	GetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineRelation, opts ...RepositoryOptions) (*schemas.Routine, *cexceptions.Exception)
	GetAllByTimeRange(from time.Time, to time.Time, stationIds []uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineRelation, opts ...RepositoryOptions) ([]schemas.Routine, *cexceptions.Exception)
	CreateOneByStationId(stationId uuid.UUID, userId uuid.UUID, input inputs.CreateRoutineInput, opts ...RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	CreateManyByStationIds(userId uuid.UUID, input []inputs.CreateRoutineByStationIdInput, opts ...RepositoryOptions) ([]uuid.UUID, *cexceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateRoutineInput, opts ...RepositoryOptions) (*schemas.Routine, *cexceptions.Exception)
	UpdateManyByIds(userId uuid.UUID, input []inputs.UpdateRoutineByIdInput, opts ...RepositoryOptions) *cexceptions.Exception
	RestoreSoftDeletedOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) (*schemas.Routine, *cexceptions.Exception)
	RestoreSoftDeletedManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) ([]schemas.Routine, *cexceptions.Exception)
	SoftDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	SoftDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception

	/* ============================== System Only Method ============================== */

	BulkCheckPermissionsAndGetManyByIds(inputs []inputs.BulkCheckRoutinePermissionInput, preloads []schemas.RoutineRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]bool, []schemas.Routine, *cexceptions.Exception)
	BulkCreateMany(inputs []inputs.BulkCreateRoutineInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkUpdateMany(inputs []inputs.BulkUpdateRoutineInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
}

type RoutineRepository struct {
	db *gorm.DB
	RoutineBulkRepository
	routineScope scopes.RoutineScopeInterface
	exceptions   exceptions.RoutineException
}

func NewRoutineRepository(
	db *gorm.DB,
	routineScope scopes.RoutineScopeInterface,
) RoutineRepositoryInterface {
	return &RoutineRepository{
		db:                    db,
		RoutineBulkRepository: *NewRoutineBulkRepositoryWithDB(db, routineScope),
		routineScope:          routineScope,
		exceptions:            exceptions.NewRoutineException(),
	}
}

func (r *RoutineRepository) HasPermission(
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
		Model(&schemas.Routine{}).
		Select("1").
		Scopes(r.routineScope.PassPermissionCheck(id, userId, allowedPermissions)).
		Scopes(r.routineScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Limit(1).
		Scan(&marker)
	if err := result.Error; err != nil {
		return false
	}

	return marker == 1
}

func (r *RoutineRepository) HavePermissions(
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
		Model(&schemas.Routine{}).
		Select(`DISTINCT "RoutineTable".id`).
		Scopes(r.routineScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.routineScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&permittedIds)
	if err := result.Error; err != nil {
		return false
	}

	return array.GetDistinctCount(ids) == array.GetDistinctCount(permittedIds)
}

func (r *RoutineRepository) CheckPermissionAndGetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) (*schemas.Routine, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var routine schemas.Routine
	query := parsedOptions.DB.
		Model(&schemas.Routine{}).
		Where(`"RoutineTable".id = ?`, id)
	if allowedPermissions != nil && len(allowedPermissions) > 0 {
		query = query.Scopes(
			r.routineScope.PassPermissionCheck(id, userId, allowedPermissions),
		)
	}

	result := query.
		Scopes(r.routineScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.routineScope.IncludePreloads(preloads, &userId)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&routine)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: routine.Id == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &routine, nil
}

func (r *RoutineRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]schemas.Routine, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var routines []schemas.Routine
	result := parsedOptions.DB.
		Model(&schemas.Routine{}).
		Scopes(r.routineScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.routineScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.routineScope.IncludePreloads(preloads, &userId)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&routines)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: len(routines) == 0, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return routines, nil
}

func (r *RoutineRepository) GetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineRelation,
	opts ...RepositoryOptions,
) (*schemas.Routine, *cexceptions.Exception) {
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

func (r *RoutineRepository) GetAllByTimeRange(
	from time.Time,
	to time.Time,
	stationIds []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineRelation,
	opts ...RepositoryOptions,
) ([]schemas.Routine, *cexceptions.Exception) {
	if len(stationIds) == 0 {
		return []schemas.Routine{}, nil
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	var routines []schemas.Routine
	timeRangeCondition := `
		(
			(
				"RoutineTable".period IS NULL
				AND "RoutineTable".scheduled_start_at < @query_to
				AND "RoutineTable".scheduled_end_at > @query_from
			)
			OR (
				"RoutineTable".period = 'Daily'::"RoutinePeriod"
				AND EXISTS (
					SELECT 1
					FROM generate_series(
						date_trunc('day', CAST(@query_from AS timestamptz) AT TIME ZONE "RoutineTable".timezone) - interval '1 day',
						date_trunc('day', CAST(@query_to AS timestamptz) AT TIME ZONE "RoutineTable".timezone),
						interval '1 day'
					) AS occurrence(bucket_start)
					CROSS JOIN LATERAL (
						SELECT occurrence.bucket_start + (
							("RoutineTable".scheduled_start_at AT TIME ZONE "RoutineTable".timezone)
							- date_trunc('day', "RoutineTable".scheduled_start_at AT TIME ZONE "RoutineTable".timezone)
						) AS occurrence_start_at
					) daily_occurrence
					WHERE (daily_occurrence.occurrence_start_at AT TIME ZONE "RoutineTable".timezone) >= "RoutineTable".scheduled_start_at
						AND (daily_occurrence.occurrence_start_at AT TIME ZONE "RoutineTable".timezone) < @query_to
						AND ((daily_occurrence.occurrence_start_at AT TIME ZONE "RoutineTable".timezone) + ("RoutineTable".scheduled_end_at - "RoutineTable".scheduled_start_at)) > @query_from
				)
			)
			OR (
				"RoutineTable".period = 'Weekly'::"RoutinePeriod"
				AND EXISTS (
					SELECT 1
					FROM generate_series(
						date_trunc('week', CAST(@query_from AS timestamptz) AT TIME ZONE "RoutineTable".timezone) - interval '1 week',
						date_trunc('week', CAST(@query_to AS timestamptz) AT TIME ZONE "RoutineTable".timezone),
						interval '1 week'
					) AS occurrence(bucket_start)
					CROSS JOIN LATERAL (
						SELECT occurrence.bucket_start + (
							("RoutineTable".scheduled_start_at AT TIME ZONE "RoutineTable".timezone)
							- date_trunc('week', "RoutineTable".scheduled_start_at AT TIME ZONE "RoutineTable".timezone)
						) AS occurrence_start_at
					) weekly_occurrence
					WHERE (weekly_occurrence.occurrence_start_at AT TIME ZONE "RoutineTable".timezone) >= "RoutineTable".scheduled_start_at
						AND (weekly_occurrence.occurrence_start_at AT TIME ZONE "RoutineTable".timezone) < @query_to
						AND ((weekly_occurrence.occurrence_start_at AT TIME ZONE "RoutineTable".timezone) + ("RoutineTable".scheduled_end_at - "RoutineTable".scheduled_start_at)) > @query_from
				)
			)
			OR (
				"RoutineTable".period = 'Monthly'::"RoutinePeriod"
				AND EXISTS (
					SELECT 1
					FROM generate_series(
						date_trunc('month', CAST(@query_from AS timestamptz) AT TIME ZONE "RoutineTable".timezone) - interval '1 month',
						date_trunc('month', CAST(@query_to AS timestamptz) AT TIME ZONE "RoutineTable".timezone),
						interval '1 month'
					) AS occurrence(bucket_start)
					CROSS JOIN LATERAL (
						SELECT "RoutineTable".scheduled_start_at AT TIME ZONE "RoutineTable".timezone AS routine_start_at
					) routine_local
					CROSS JOIN LATERAL (
						SELECT make_timestamp(
							EXTRACT(YEAR FROM occurrence.bucket_start)::integer,
							EXTRACT(MONTH FROM occurrence.bucket_start)::integer,
							LEAST(
								EXTRACT(DAY FROM routine_local.routine_start_at)::integer,
								EXTRACT(DAY FROM (date_trunc('month', occurrence.bucket_start) + interval '1 month' - interval '1 day'))::integer
							),
							EXTRACT(HOUR FROM routine_local.routine_start_at)::integer,
							EXTRACT(MINUTE FROM routine_local.routine_start_at)::integer,
							EXTRACT(SECOND FROM routine_local.routine_start_at)::double precision
						) AS occurrence_start_at
					) monthly_occurrence
					WHERE (monthly_occurrence.occurrence_start_at AT TIME ZONE "RoutineTable".timezone) >= "RoutineTable".scheduled_start_at
						AND (monthly_occurrence.occurrence_start_at AT TIME ZONE "RoutineTable".timezone) < @query_to
						AND ((monthly_occurrence.occurrence_start_at AT TIME ZONE "RoutineTable".timezone) + ("RoutineTable".scheduled_end_at - "RoutineTable".scheduled_start_at)) > @query_from
				)
			)
		)
	`
	result := parsedOptions.DB.
		Model(&schemas.Routine{}).
		Select(`"RoutineTable".*`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = "RoutineTable".station_id`).
		Joins(`INNER JOIN "StationTable" station ON station.id = "RoutineTable".station_id AND station.deleted_at IS NULL`).
		Where(`"RoutineTable".station_id IN ?`, stationIds).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, parsedOptions.AllowedPermissions).
		Where(timeRangeCondition, sql.Named("query_from", from), sql.Named("query_to", to)).
		Scopes(r.routineScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.routineScope.IncludePreloads(preloads, &userId)).
		Order(`"RoutineTable".scheduled_start_at ASC`).
		Order(`"RoutineTable".scheduled_end_at ASC`).
		Order(`"RoutineTable".id ASC`).
		Find(&routines)
	if result.Error != nil {
		return nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	return routines, nil
}

func (r *RoutineRepository) CreateOneByStationId(
	stationId uuid.UUID,
	userId uuid.UUID,
	input inputs.CreateRoutineInput,
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

	stationRepository := NewStationRepository(r.db, scopes.NewStationScope())
	if !stationRepository.HasPermission(
		stationId,
		userId,
		parsedOptions.AllowedPermissions,
		append(opts, WithAllowedPermissions(parsedOptions.AllowedPermissions))...,
	) {
		parsedOptions.DB.Rollback()
		return nil, r.exceptions.NoPermission("create a routine under this station")
	}

	startAt := time.Now().Truncate(time.Minute)
	newRoutine := schemas.Routine{
		Id:               uuid.New(),
		StationId:        stationId,
		Status:           cenums.RoutineStatus_Scheduled,
		ScheduledStartAt: startAt,
		ScheduledEndAt:   startAt.Add(time.Hour),
		Timezone:         "UTC",
	}
	if err := copier.Copy(&newRoutine, &input); err != nil {
		parsedOptions.DB.Rollback()
		return nil, r.exceptions.InvalidInput().WithOrigin(err)
	}
	newRoutine.StationId = stationId

	result := parsedOptions.DB.Model(&schemas.Routine{}).
		Create(&newRoutine)
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

	return &newRoutine.Id, nil
}

func (r *RoutineRepository) CreateManyByStationIds(
	userId uuid.UUID,
	input []inputs.CreateRoutineByStationIdInput,
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

	stationIds := make([]uuid.UUID, len(input))
	for index, in := range input {
		stationIds[index] = in.StationId
	}
	stationRepository := NewStationRepository(r.db, scopes.NewStationScope())
	validStations, _, exception := stationRepository.CheckPermissionsAndGetManyByIds(
		stationIds,
		userId,
		nil,
		parsedOptions.AllowedPermissions,
		append(opts, WithAllowedPermissions(parsedOptions.AllowedPermissions))...,
	)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}
	isStationValid := make(map[uuid.UUID]bool, len(validStations))
	for _, validStation := range validStations {
		isStationValid[validStation.Id] = true
	}

	newRoutines := make([]schemas.Routine, 0, len(input))
	for _, in := range input {
		if !isStationValid[in.StationId] {
			continue
		}
		startAt := time.Now().Truncate(time.Minute)
		newRoutine := schemas.Routine{
			Id:               uuid.New(),
			StationId:        in.StationId,
			Status:           cenums.RoutineStatus_Scheduled,
			ScheduledStartAt: startAt,
			ScheduledEndAt:   startAt.Add(time.Hour),
			Timezone:         "UTC",
		}
		if err := copier.Copy(&newRoutine, &in); err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.InvalidInput().WithOrigin(err)
		}
		newRoutine.StationId = in.StationId
		newRoutines = append(newRoutines, newRoutine)
	}
	if len(newRoutines) == 0 {
		parsedOptions.DB.Rollback()
		return nil, r.exceptions.NoChanges()
	}

	result := parsedOptions.DB.Model(&schemas.Routine{}).
		CreateInBatches(&newRoutines, parsedOptions.BatchSize)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	newRoutineIds := make([]uuid.UUID, len(newRoutines))
	for index, newRoutine := range newRoutines {
		newRoutineIds[index] = newRoutine.Id
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return newRoutineIds, nil
}

func (r *RoutineRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateRoutineInput,
	opts ...RepositoryOptions,
) (*schemas.Routine, *cexceptions.Exception) {
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

	existingRoutine, exception := r.CheckPermissionAndGetOneById(
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
	if input.Values.StationId != nil && !partialupdate.CheckSetNull(input.SetNull, "StationId") {
		stationRepository := NewStationRepository(r.db, scopes.NewStationScope())
		if !stationRepository.HasPermission(
			*input.Values.StationId,
			userId,
			parsedOptions.AllowedPermissions,
			append(opts, WithAllowedPermissions(parsedOptions.AllowedPermissions))...,
		) {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.NoPermission("move a routine to this station")
		}
	}
	if input.Values.ScheduledStartAt != nil {
		truncatedScheduledStartAt := input.Values.ScheduledStartAt.Truncate(time.Minute)
		input.Values.ScheduledStartAt = &truncatedScheduledStartAt
	}
	if input.Values.ScheduledEndAt != nil {
		truncatedScheduledEndAt := input.Values.ScheduledEndAt.Truncate(time.Minute)
		input.Values.ScheduledEndAt = &truncatedScheduledEndAt
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingRoutine)
	if err != nil {
		parsedOptions.DB.Rollback()
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true).WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.Routine{}).
		Where(`"RoutineTable".id = ? AND "RoutineTable".deleted_at IS NULL`, id).
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

func (r *RoutineRepository) UpdateManyByIds(
	userId uuid.UUID,
	input []inputs.UpdateRoutineByIdInput,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	if len(input) == 0 {
		return r.exceptions.NoChanges()
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

	ids := make([]uuid.UUID, len(input))
	for index, in := range input {
		ids[index] = in.Id
	}
	validRoutines, exception := r.CheckPermissionsAndGetManyByIds(
		ids,
		userId,
		nil,
		parsedOptions.AllowedPermissions,
		opts...,
	)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return r.exceptions.NoPermission("update these routines")
	}
	isRoutineValid := make(map[uuid.UUID]bool, len(validRoutines))
	for _, validRoutine := range validRoutines {
		isRoutineValid[validRoutine.Id] = true
	}

	targetStationIdSet := make(map[uuid.UUID]bool)
	for _, in := range input {
		if in.PartialUpdateInput.Values.StationId == nil ||
			partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "StationId") {
			continue
		}
		targetStationIdSet[*in.PartialUpdateInput.Values.StationId] = true
	}
	if len(targetStationIdSet) > 0 {
		targetStationIds := make([]uuid.UUID, 0, len(targetStationIdSet))
		for targetStationId := range targetStationIdSet {
			targetStationIds = append(targetStationIds, targetStationId)
		}
		stationRepository := NewStationRepository(r.db, scopes.NewStationScope())
		if !stationRepository.HavePermissions(
			targetStationIds,
			userId,
			parsedOptions.AllowedPermissions,
			append(opts, WithAllowedPermissions(parsedOptions.AllowedPermissions))...,
		) {
			parsedOptions.DB.Rollback()
			return r.exceptions.NoPermission("move these routines to the given stations")
		}
	}

	var valuePlaceholders []string
	var valueArgs []interface{}
	for _, in := range input {
		if !isRoutineValid[in.Id] {
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

		valuePlaceholders = append(valuePlaceholders, `(?::uuid, ?::uuid, ?::text, ?::text, ?::"RoutineStatus", ?::boolean, ?::timestamptz, ?::timestamptz, ?::"RoutinePeriod", ?::text, ?::boolean)`)
		valueArgs = append(valueArgs,
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
		parsedOptions.DB.Rollback()
		return r.exceptions.NoChanges()
	}

	sql := fmt.Sprintf(`
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
		FROM (VALUES %s) AS v(id, station_id, title, description, status, is_pinned, scheduled_start_at, scheduled_end_at, period, timezone, set_period_null)
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

func (r *RoutineRepository) RestoreSoftDeletedOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) (*schemas.Routine, *cexceptions.Exception) {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var restoredRoutine schemas.Routine
	result := parsedOptions.DB.
		Model(&restoredRoutine).
		Scopes(r.routineScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.routineScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Clauses(clause.Returning{}).
		Where(`"RoutineTable".id = ?`, id).
		Updates(map[string]interface{}{"deleted_at": nil})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: restoredRoutine.Id == uuid.Nil, Second: r.exceptions.FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &restoredRoutine, nil
}

func (r *RoutineRepository) RestoreSoftDeletedManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) ([]schemas.Routine, *cexceptions.Exception) {
	if len(ids) == 0 {
		return nil, r.exceptions.NoChanges()
	}

	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var restoredRoutines []schemas.Routine
	result := parsedOptions.DB.
		Model(&restoredRoutines).
		Scopes(r.routineScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.routineScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Clauses(clause.Returning{}).
		Where(`"RoutineTable".id IN ?`, ids).
		Updates(map[string]interface{}{"deleted_at": nil})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: len(restoredRoutines) == 0, Second: r.exceptions.FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return restoredRoutines, nil
}

func (r *RoutineRepository) SoftDeleteOneById(
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

	result := parsedOptions.DB.
		Model(&schemas.Routine{}).
		Scopes(r.routineScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.routineScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"RoutineTable".id = ?`, id).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RoutineRepository) SoftDeleteManyByIds(
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

	result := parsedOptions.DB.
		Model(&schemas.Routine{}).
		Scopes(r.routineScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.routineScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"RoutineTable".id IN ?`, ids).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RoutineRepository) HardDeleteOneById(
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

	result := parsedOptions.DB.
		Model(&schemas.Routine{}).
		Scopes(r.routineScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.routineScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"RoutineTable".id = ?`, id).
		Delete(&schemas.Routine{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RoutineRepository) HardDeleteManyByIds(
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

	result := parsedOptions.DB.
		Model(&schemas.Routine{}).
		Scopes(r.routineScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.routineScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"RoutineTable".id IN ?`, ids).
		Delete(&schemas.Routine{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

/* ============================== System Only Method ============================== */

func (r *RoutineRepository) BulkCheckPermissionsAndGetManyByIds(
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
