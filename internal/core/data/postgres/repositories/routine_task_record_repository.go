package repositories

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	array "github.com/HiIamJeff67/notegic-backend/shared/lib/array"

	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	scopes "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/scopes"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
)

type RoutineTaskRecordRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTaskRecordRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) (*schemas.RoutineTaskRecord, *exceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTaskRecordRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]schemas.RoutineTaskRecord, *exceptions.Exception)
	GetAllByRoutineTaskId(routineTaskId uuid.UUID, userId uuid.UUID, limit int, preloads []schemas.RoutineTaskRecordRelation, opts ...options.RepositoryOptions) ([]schemas.RoutineTaskRecord, *exceptions.Exception)
	UpdateManyAsFailed(failureInputs []inputs.UpdateRoutineTaskRecordFailureInput, failedAt time.Time, opts ...options.RepositoryOptions) (int64, *exceptions.Exception)
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *exceptions.Exception
	HardDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *exceptions.Exception
}

type RoutineTaskRecordRepository struct {
	routineTaskRecordScope scopes.RoutineTaskRecordScopeInterface
}

func NewRoutineTaskRecordRepository(
	routineTaskRecordScope scopes.RoutineTaskRecordScopeInterface,
) RoutineTaskRecordRepositoryInterface {
	return &RoutineTaskRecordRepository{
		routineTaskRecordScope: routineTaskRecordScope,
	}
}

func (r *RoutineTaskRecordRepository) HasPermission(
	id uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var marker int
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskRecord{}).
		Select("1").
		Scopes(r.routineTaskRecordScope.PassPermissionCheck(id, userId, allowedPermissions)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Limit(1).
		Scan(&marker)
	if result.Error != nil {
		return false
	}

	return marker == 1
}

func (r *RoutineTaskRecordRepository) HavePermissions(
	ids []uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var permittedIds []uuid.UUID
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskRecord{}).
		Select(`DISTINCT "RoutineTaskRecordTable".id`).
		Scopes(r.routineTaskRecordScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&permittedIds)
	if result.Error != nil {
		return false
	}

	return array.GetDistinctCount(ids) == array.GetDistinctCount(permittedIds)
}

func (r *RoutineTaskRecordRepository) CheckPermissionAndGetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineTaskRecordRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) (*schemas.RoutineTaskRecord, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var routineTaskRecord schemas.RoutineTaskRecord
	query := parsedOptions.DB.
		Model(&schemas.RoutineTaskRecord{}).
		Where(`"RoutineTaskRecordTable".id = ?`, id)
	if allowedPermissions != nil && len(allowedPermissions) > 0 {
		query = query.Scopes(
			r.routineTaskRecordScope.PassPermissionCheck(id, userId, allowedPermissions),
		)
	}

	result := query.
		Scopes(r.routineTaskRecordScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&routineTaskRecord)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTaskException().NotFound().WithOrigin(result.Error)},
		{First: routineTaskRecord.Id == uuid.Nil, Second: apiexceptions.NewRoutineTaskException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &routineTaskRecord, nil
}

func (r *RoutineTaskRecordRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineTaskRecordRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]schemas.RoutineTaskRecord, *exceptions.Exception) {
	if len(ids) == 0 {
		return []schemas.RoutineTaskRecord{}, nil
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	var routineTaskRecords []schemas.RoutineTaskRecord
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskRecord{}).
		Scopes(r.routineTaskRecordScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.routineTaskRecordScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&routineTaskRecords)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTaskException().NotFound().WithOrigin(result.Error)},
		{First: len(routineTaskRecords) == 0, Second: apiexceptions.NewRoutineTaskException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return routineTaskRecords, nil
}

func (r *RoutineTaskRecordRepository) GetAllByRoutineTaskId(
	routineTaskId uuid.UUID,
	userId uuid.UUID,
	limit int,
	preloads []schemas.RoutineTaskRecordRelation,
	opts ...options.RepositoryOptions,
) ([]schemas.RoutineTaskRecord, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	if limit <= 0 {
		limit = 100
	}

	var routineTaskRecords []schemas.RoutineTaskRecord
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskRecord{}).
		Select(`"RoutineTaskRecordTable".*`).
		Joins(`INNER JOIN "RoutineTaskTable" routine_task ON routine_task.id = "RoutineTaskRecordTable".routine_task_id`).
		Joins(`INNER JOIN "RoutineTable" routine ON routine.id = routine_task.routine_id AND routine.deleted_at IS NULL`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = routine.station_id`).
		Where(`"RoutineTaskRecordTable".routine_task_id = ?`, routineTaskId).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, parsedOptions.AllowedPermissions).
		Scopes(r.routineTaskRecordScope.IncludePreloads(preloads)).
		Order(`"RoutineTaskRecordTable".created_at DESC`).
		Limit(limit).
		Find(&routineTaskRecords)
	if result.Error != nil {
		return nil, apiexceptions.NewRoutineTaskException().NotFound().WithOrigin(result.Error)
	}

	return routineTaskRecords, nil
}

func (r *RoutineTaskRecordRepository) UpdateManyAsFailed(
	failureInputs []inputs.UpdateRoutineTaskRecordFailureInput,
	failedAt time.Time,
	opts ...options.RepositoryOptions,
) (int64, *exceptions.Exception) {
	if len(failureInputs) == 0 {
		return 0, nil
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)
	valuePlaceholders := make([]string, 0, len(failureInputs))
	valueArgs := make([]any, 0, len(failureInputs)*3+4)
	for _, failureInput := range failureInputs {
		valuePlaceholders = append(valuePlaceholders, "(?::uuid, ?::\"RoutineTaskRecordErrorCode\", ?::varchar)")
		valueArgs = append(
			valueArgs,
			failureInput.Id,
			failureInput.ErrorCode.String(),
			failureInput.ErrorReason,
		)
	}

	query := fmt.Sprintf(`
		UPDATE "RoutineTaskRecordTable" AS routine_task_record
		SET
			status = ?::"RoutineTaskRecordStatus",
			actual_ended_at = ?::timestamptz,
			error_code = value.error_code,
			error_reason = value.error_reason,
			updated_at = ?::timestamptz
		FROM (VALUES %s) AS value(id, error_code, error_reason)
		WHERE routine_task_record.id = value.id
			AND routine_task_record.status = ?::"RoutineTaskRecordStatus"
	`, strings.Join(valuePlaceholders, ","))
	valueArgs = append(
		[]any{
			enums.RoutineTaskRecordStatus_Failed.String(),
			failedAt,
			failedAt,
			enums.RoutineTaskRecordStatus_Running.String(),
		},
		valueArgs...,
	)

	result := parsedOptions.DB.Exec(query, valueArgs...)
	if result.Error != nil {
		return 0, apiexceptions.NewRoutineTaskException().FailedToUpdate().WithOrigin(result.Error)
	}

	return result.RowsAffected, nil
}

func (r *RoutineTaskRecordRepository) HardDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskRecord{}).
		Scopes(r.routineTaskRecordScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Where(`"RoutineTaskRecordTable".id = ?`, id).
		Delete(&schemas.RoutineTaskRecord{})
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTaskException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTaskException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RoutineTaskRecordRepository) HardDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	if len(ids) == 0 {
		return apiexceptions.NewRoutineTaskException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskRecord{}).
		Scopes(r.routineTaskRecordScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Where(`"RoutineTaskRecordTable".id IN ?`, ids).
		Delete(&schemas.RoutineTaskRecord{})
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTaskException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTaskException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}
