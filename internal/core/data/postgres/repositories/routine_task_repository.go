package repositories

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	array "github.com/HiIamJeff67/notegic-backend/shared/lib/array"
	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"

	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	scopes "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/scopes"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
)

type RoutineTaskRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTaskRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) (*schemas.RoutineTask, *exceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTaskRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]schemas.RoutineTask, *exceptions.Exception)
	GetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTaskRelation, opts ...options.RepositoryOptions) (*schemas.RoutineTask, *exceptions.Exception)
	GetAllByUserId(userId uuid.UUID, preloads []schemas.RoutineTaskRelation, opts ...options.RepositoryOptions) ([]schemas.RoutineTask, *exceptions.Exception)
	GetAllByRoutineIds(routineIds []uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTaskRelation, opts ...options.RepositoryOptions) ([]schemas.RoutineTask, *exceptions.Exception)
	CreateOneByRoutineId(routineId uuid.UUID, userId uuid.UUID, input inputs.CreateRoutineTaskInput, opts ...options.RepositoryOptions) (*uuid.UUID, *exceptions.Exception)
	CreateManyByRoutineIds(userId uuid.UUID, input []inputs.CreateRoutineTaskByRoutineIdInput, opts ...options.RepositoryOptions) ([]uuid.UUID, *exceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateRoutineTaskInput, opts ...options.RepositoryOptions) (*schemas.RoutineTask, *exceptions.Exception)
	UpdateManyByIds(userId uuid.UUID, input []inputs.UpdateRoutineTaskByIdInput, opts ...options.RepositoryOptions) *exceptions.Exception
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *exceptions.Exception
	HardDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *exceptions.Exception
}

type RoutineTaskRepository struct {
	routineTaskScope scopes.RoutineTaskScopeInterface
}

func NewRoutineTaskRepository(routineTaskScope scopes.RoutineTaskScopeInterface) RoutineTaskRepositoryInterface {
	return &RoutineTaskRepository{
		routineTaskScope: routineTaskScope,
	}
}

func (r *RoutineTaskRepository) HasPermission(
	id uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var marker int
	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		Select("1").
		Scopes(r.routineTaskScope.PassPermissionCheck(id, userId, allowedPermissions)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Limit(1).
		Scan(&marker)
	if err := result.Error; err != nil {
		return false
	}

	return marker == 1
}

func (r *RoutineTaskRepository) HavePermissions(
	ids []uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var permittedIds []uuid.UUID
	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		Select(`DISTINCT "RoutineTaskTable".id`).
		Scopes(r.routineTaskScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&permittedIds)
	if err := result.Error; err != nil {
		return false
	}

	return array.GetDistinctCount(ids) == array.GetDistinctCount(permittedIds)
}

func (r *RoutineTaskRepository) CheckPermissionAndGetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineTaskRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) (*schemas.RoutineTask, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var routineTask schemas.RoutineTask
	query := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		Where(`"RoutineTaskTable".id = ?`, id)
	if allowedPermissions != nil && len(allowedPermissions) > 0 {
		query = query.Scopes(
			r.routineTaskScope.PassPermissionCheck(id, userId, allowedPermissions),
		)
	}

	result := query.
		Scopes(r.routineTaskScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&routineTask)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTaskException().NotFound().WithOrigin(result.Error)},
		{First: routineTask.Id == uuid.Nil, Second: apiexceptions.NewRoutineTaskException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &routineTask, nil
}

func (r *RoutineTaskRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineTaskRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]schemas.RoutineTask, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var routineTasks []schemas.RoutineTask
	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		Scopes(r.routineTaskScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.routineTaskScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&routineTasks)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTaskException().NotFound().WithOrigin(result.Error)},
		{First: len(routineTasks) == 0, Second: apiexceptions.NewRoutineTaskException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return routineTasks, nil
}

func (r *RoutineTaskRepository) GetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineTaskRelation,
	opts ...options.RepositoryOptions,
) (*schemas.RoutineTask, *exceptions.Exception) {
	return r.CheckPermissionAndGetOneById(
		id,
		userId,
		preloads,
		options.ParseRepositoryOptions(opts...).AllowedPermissions,
		opts...,
	)
}

func (r *RoutineTaskRepository) GetAllByUserId(
	userId uuid.UUID,
	preloads []schemas.RoutineTaskRelation,
	opts ...options.RepositoryOptions,
) ([]schemas.RoutineTask, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	var routineTasks []schemas.RoutineTask
	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		Select(`"RoutineTaskTable".*`).
		Joins(`INNER JOIN "RoutineTable" routine ON routine.id = "RoutineTaskTable".routine_id AND routine.deleted_at IS NULL`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = routine.station_id`).
		Joins(`INNER JOIN "StationTable" station ON station.id = routine.station_id AND station.deleted_at IS NULL`).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, parsedOptions.AllowedPermissions).
		Scopes(r.routineTaskScope.IncludePreloads(preloads)).
		Find(&routineTasks)
	if result.Error != nil {
		return nil, apiexceptions.NewRoutineTaskException().NotFound().WithOrigin(result.Error)
	}

	return routineTasks, nil
}

func (r *RoutineTaskRepository) GetAllByRoutineIds(
	routineIds []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineTaskRelation,
	opts ...options.RepositoryOptions,
) ([]schemas.RoutineTask, *exceptions.Exception) {
	if len(routineIds) == 0 {
		return []schemas.RoutineTask{}, nil
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)
	var routineTasks []schemas.RoutineTask
	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		Select(`"RoutineTaskTable".*`).
		Joins(`INNER JOIN "RoutineTable" routine ON routine.id = "RoutineTaskTable".routine_id AND routine.deleted_at IS NULL`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = routine.station_id`).
		Joins(`INNER JOIN "StationTable" station ON station.id = routine.station_id AND station.deleted_at IS NULL`).
		Where(`"RoutineTaskTable".routine_id IN ?`, routineIds).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, parsedOptions.AllowedPermissions).
		Scopes(r.routineTaskScope.IncludePreloads(preloads)).
		Find(&routineTasks)
	if result.Error != nil {
		return nil, apiexceptions.NewRoutineTaskException().NotFound().WithOrigin(result.Error)
	}

	return routineTasks, nil
}

func (r *RoutineTaskRepository) CreateOneByRoutineId(
	routineId uuid.UUID,
	userId uuid.UUID,
	input inputs.CreateRoutineTaskInput,
	opts ...options.RepositoryOptions,
) (*uuid.UUID, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	if input.ActorUserId == uuid.Nil || input.ActorUserId != userId {
		return nil, apiexceptions.NewRoutineTaskException().InvalidInput().WithOrigin(fmt.Errorf("actorUserId must match userId"))
	}

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	routineRepository := NewRoutineRepository(scopes.NewRoutineScope())
	if !routineRepository.HasPermission(
		routineId,
		userId,
		parsedOptions.AllowedPermissions,
		opts...,
	) {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewRoutineTaskException().NoPermission("create a routine task under this routine")
	}

	if input.NextScheduledAt.IsZero() {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewRoutineTaskException().InvalidInput().WithOrigin(fmt.Errorf("nextScheduledAt is required"))
	}

	newRoutineTask := schemas.RoutineTask{}
	if err := copier.Copy(&newRoutineTask, &input); err != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewRoutineTaskException().InvalidInput().WithOrigin(err)
	}
	newRoutineTask.RoutineId = routineId
	newRoutineTask.ActorUserId = input.ActorUserId
	newRoutineTask.NextScheduledAt = input.NextScheduledAt.Truncate(time.Minute)
	newRoutineTask.ScheduledAt = newRoutineTask.NextScheduledAt

	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		Create(&newRoutineTask)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTaskException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTaskException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &newRoutineTask.Id, nil
}

func (r *RoutineTaskRepository) CreateManyByRoutineIds(
	userId uuid.UUID,
	input []inputs.CreateRoutineTaskByRoutineIdInput,
	opts ...options.RepositoryOptions,
) ([]uuid.UUID, *exceptions.Exception) {
	if len(input) == 0 {
		return nil, apiexceptions.NewRoutineTaskException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)
	for _, in := range input {
		if in.ActorUserId == uuid.Nil || in.ActorUserId != userId {
			return nil, apiexceptions.NewRoutineTaskException().InvalidInput().WithOrigin(fmt.Errorf("actorUserId must match userId"))
		}
	}

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	routineIds := make([]uuid.UUID, len(input))
	for index, in := range input {
		routineIds[index] = in.RoutineId
	}

	routineRepository := NewRoutineRepository(scopes.NewRoutineScope())
	validRoutines, exception := routineRepository.CheckPermissionsAndGetManyByIds(routineIds, userId, nil, parsedOptions.AllowedPermissions, opts...)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	isRoutineValid := make(map[uuid.UUID]bool, len(validRoutines))
	for _, validRoutine := range validRoutines {
		isRoutineValid[validRoutine.Id] = true
	}

	newRoutineTasks := make([]schemas.RoutineTask, 0, len(input))
	for _, in := range input {
		if !isRoutineValid[in.RoutineId] {
			continue
		}
		if in.NextScheduledAt.IsZero() {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewRoutineTaskException().InvalidInput().WithOrigin(fmt.Errorf("nextScheduledAt is required"))
		}
		newRoutineTask := schemas.RoutineTask{
			Id:          uuid.New(),
			RoutineId:   in.RoutineId,
			ActorUserId: in.ActorUserId,
		}
		if err := copier.Copy(&newRoutineTask, &in); err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewRoutineTaskException().InvalidInput().WithOrigin(err)
		}
		newRoutineTask.RoutineId = in.RoutineId
		newRoutineTask.NextScheduledAt = in.NextScheduledAt.Truncate(time.Minute)
		newRoutineTask.ScheduledAt = newRoutineTask.NextScheduledAt
		newRoutineTasks = append(newRoutineTasks, newRoutineTask)
	}

	if len(newRoutineTasks) == 0 {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewRoutineTaskException().NoChanges()
	}

	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		CreateInBatches(&newRoutineTasks, parsedOptions.BatchSize)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTaskException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTaskException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}
	newRoutineTaskIds := make([]uuid.UUID, len(newRoutineTasks))
	for index, newRoutineTask := range newRoutineTasks {
		newRoutineTaskIds[index] = newRoutineTask.Id
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return newRoutineTaskIds, nil
}

func (r *RoutineTaskRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateRoutineTaskInput,
	opts ...options.RepositoryOptions,
) (*schemas.RoutineTask, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	existingRoutineTask, exception := r.CheckPermissionAndGetOneById(id, userId, nil, parsedOptions.AllowedPermissions, opts...)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}
	if input.Values.RoutineId != nil && !partialupdate.CheckSetNull(input.SetNull, "RoutineId") {
		routineRepository := NewRoutineRepository(scopes.NewRoutineScope())
		if !routineRepository.HasPermission(*input.Values.RoutineId, userId, parsedOptions.AllowedPermissions, opts...) {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewRoutineTaskException().NoPermission("move a routine task to this routine")
		}
	}
	if input.Values.NextScheduledAt != nil {
		truncatedNextScheduledAt := input.Values.NextScheduledAt.Truncate(time.Minute)
		input.Values.NextScheduledAt = &truncatedNextScheduledAt
		scheduledAt := existingRoutineTask.ScheduledAt
		if truncatedNextScheduledAt.After(scheduledAt) {
			scheduledAt = truncatedNextScheduledAt
		}
		input.Values.ScheduledAt = &scheduledAt
	} else if input.Values.ScheduledAt != nil {
		truncatedScheduledAt := input.Values.ScheduledAt.Truncate(time.Minute)
		input.Values.ScheduledAt = &truncatedScheduledAt
	}
	if input.SetNull != nil {
		for fieldName := range *input.SetNull {
			normalizedFieldName := strings.ToLower(strings.ReplaceAll(fieldName, "_", ""))
			if normalizedFieldName == "scheduledat" || normalizedFieldName == "nextscheduledat" {
				delete(*input.SetNull, fieldName)
			}
		}
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingRoutineTask)
	if err != nil {
		parsedOptions.DB.Rollback()
		return nil, exceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true).WithOrigin(err)
	}

	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		Where(`"RoutineTaskTable".id = ?`, id).
		Select("*").
		Updates(&updates)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTaskException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTaskException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &updates, nil
}

func (r *RoutineTaskRepository) UpdateManyByIds(
	userId uuid.UUID,
	input []inputs.UpdateRoutineTaskByIdInput,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	if len(input) == 0 {
		return apiexceptions.NewRoutineTaskException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	ids := make([]uuid.UUID, len(input))
	for index, in := range input {
		ids[index] = in.Id
	}
	validRoutineTasks, exception := r.CheckPermissionsAndGetManyByIds(ids, userId, nil, parsedOptions.AllowedPermissions, opts...)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return apiexceptions.NewRoutineTaskException().NoPermission("update these routine tasks")
	}

	isRoutineTaskValid := make(map[uuid.UUID]bool, len(validRoutineTasks))
	for _, validRoutineTask := range validRoutineTasks {
		isRoutineTaskValid[validRoutineTask.Id] = true
	}

	targetRoutineIdSet := make(map[uuid.UUID]bool)
	for _, in := range input {
		if in.PartialUpdateInput.Values.RoutineId == nil ||
			partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "RoutineId") {
			continue
		}
		targetRoutineIdSet[*in.PartialUpdateInput.Values.RoutineId] = true
	}
	if len(targetRoutineIdSet) > 0 {
		targetRoutineIds := make([]uuid.UUID, 0, len(targetRoutineIdSet))
		for targetRoutineId := range targetRoutineIdSet {
			targetRoutineIds = append(targetRoutineIds, targetRoutineId)
		}
		routineRepository := NewRoutineRepository(scopes.NewRoutineScope())
		if !routineRepository.HavePermissions(targetRoutineIds, userId, parsedOptions.AllowedPermissions, opts...) {
			parsedOptions.DB.Rollback()
			return apiexceptions.NewRoutineTaskException().NoPermission("move these routine tasks to the given routines")
		}
	}

	var valuePlaceholders []string
	var valueArgs []interface{}
	for _, in := range input {
		if !isRoutineTaskValid[in.Id] {
			continue
		}

		setPeriodNull := partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "Period")

		nextScheduledAt := in.PartialUpdateInput.Values.NextScheduledAt
		if nextScheduledAt != nil {
			truncatedNextScheduledAt := nextScheduledAt.Truncate(time.Minute)
			nextScheduledAt = &truncatedNextScheduledAt
		}

		scheduledAt := in.PartialUpdateInput.Values.ScheduledAt
		if scheduledAt != nil {
			truncatedScheduledAt := scheduledAt.Truncate(time.Minute)
			scheduledAt = &truncatedScheduledAt
		}

		valuePlaceholders = append(valuePlaceholders, `(?::uuid, ?::uuid, ?::text, ?::"RoutineTaskPurpose", ?::jsonb, ?::integer, ?::integer, ?::"RoutinePeriod", ?::timestamptz, ?::timestamptz, ?::boolean)`)
		valueArgs = append(valueArgs,
			in.Id,
			in.PartialUpdateInput.Values.RoutineId,
			in.PartialUpdateInput.Values.Title,
			in.PartialUpdateInput.Values.Purpose,
			in.PartialUpdateInput.Values.Payload,
			in.PartialUpdateInput.Values.Priority,
			in.PartialUpdateInput.Values.MaxAttempts,
			in.PartialUpdateInput.Values.Period,
			nextScheduledAt,
			scheduledAt,
			setPeriodNull,
		)
	}

	if len(valuePlaceholders) == 0 {
		parsedOptions.DB.Rollback()
		return apiexceptions.NewRoutineTaskException().NoChanges()
	}

	sql := fmt.Sprintf(`
		UPDATE "RoutineTaskTable" AS rt
		SET
			routine_id = COALESCE(v.routine_id::uuid, rt.routine_id),
			title = COALESCE(v.title::text, rt.title),
			purpose = COALESCE(v.purpose::"RoutineTaskPurpose", rt.purpose),
			payload = COALESCE(v.payload::jsonb, rt.payload),
			priority = COALESCE(v.priority::integer, rt.priority),
			max_attempts = COALESCE(v.max_attempts::integer, rt.max_attempts),
			period = CASE
				WHEN v.set_period_null::boolean THEN NULL
				ELSE COALESCE(v.period::"RoutinePeriod", rt.period)
			END,
			next_scheduled_at = COALESCE(v.next_scheduled_at::timestamptz, rt.next_scheduled_at),
			scheduled_at = CASE
				WHEN v.scheduled_at IS NOT NULL THEN v.scheduled_at::timestamptz
				WHEN v.next_scheduled_at IS NOT NULL THEN GREATEST(rt.scheduled_at, v.next_scheduled_at::timestamptz)
				ELSE rt.scheduled_at
			END,
			updated_at = NOW()
		FROM (VALUES %s) AS v(id, routine_id, title, purpose, payload, priority, max_attempts, period, next_scheduled_at, scheduled_at, set_period_null)
		WHERE rt.id = v.id::uuid
	`, strings.Join(valuePlaceholders, ","))
	result := parsedOptions.DB.Exec(sql, valueArgs...)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTaskException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTaskException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return apiexceptions.NewRoutineTaskException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return nil
}

func (r *RoutineTaskRepository) HardDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		Scopes(r.routineTaskScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Where(`"RoutineTaskTable".id = ?`, id).
		Delete(&schemas.RoutineTask{})
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTaskException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTaskException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RoutineTaskRepository) HardDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	if len(ids) == 0 {
		return apiexceptions.NewRoutineTaskException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		Scopes(r.routineTaskScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Where(`"RoutineTaskTable".id IN ?`, ids).
		Delete(&schemas.RoutineTask{})
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTaskException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTaskException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}
