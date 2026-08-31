package repositories

import (
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	array "github.com/HiIamJeff67/notegic-backend/shared/lib/array"
	exceptions "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type RoutineTaskRecordRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTaskRecordRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) (*schemas.RoutineTaskRecord, *cexceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTaskRecordRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]schemas.RoutineTaskRecord, *cexceptions.Exception)
	GetAllByRoutineTaskId(routineTaskId uuid.UUID, userId uuid.UUID, limit int, preloads []schemas.RoutineTaskRecordRelation, opts ...RepositoryOptions) ([]schemas.RoutineTaskRecord, *cexceptions.Exception)
	UpdateManyAsFailed(failureInputs []inputs.UpdateRoutineTaskRecordFailureInput, opts ...RepositoryOptions) (int64, *cexceptions.Exception)
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
}

type RoutineTaskRecordRepository struct {
	db                     *gorm.DB
	routineTaskRecordScope scopes.RoutineTaskRecordScopeInterface
	exceptions             exceptions.RoutineTaskRecordException
}

func NewRoutineTaskRecordRepository(
	db *gorm.DB,
	routineTaskRecordScope scopes.RoutineTaskRecordScopeInterface,
) RoutineTaskRecordRepositoryInterface {
	return &RoutineTaskRecordRepository{
		db:                     db,
		routineTaskRecordScope: routineTaskRecordScope,
		exceptions:             exceptions.NewRoutineTaskRecordException(),
	}
}

func (r *RoutineTaskRecordRepository) HasPermission(
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
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) (*schemas.RoutineTaskRecord, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

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
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: routineTaskRecord.Id == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &routineTaskRecord, nil
}

func (r *RoutineTaskRecordRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineTaskRecordRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]schemas.RoutineTaskRecord, *cexceptions.Exception) {
	if len(ids) == 0 {
		return []schemas.RoutineTaskRecord{}, nil
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var routineTaskRecords []schemas.RoutineTaskRecord
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskRecord{}).
		Scopes(r.routineTaskRecordScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.routineTaskRecordScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&routineTaskRecords)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: len(routineTaskRecords) == 0, Second: r.exceptions.NotFound()},
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
	opts ...RepositoryOptions,
) ([]schemas.RoutineTaskRecord, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	if limit <= 0 {
		limit = 100
	}

	var routineTaskRecords []schemas.RoutineTaskRecord
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskRecord{}).
		Select(`
			"RoutineTaskRecordTable".id,
			"RoutineTaskRecordTable".routine_record_id,
			"RoutineTaskRecordTable".routine_task_id,
			"RoutineTaskRecordTable".purpose,
			"RoutineTaskRecordTable".status,
			"RoutineTaskRecordTable".error_code,
			"RoutineTaskRecordTable".error_reason,
			"RoutineTaskRecordTable".cost_unit,
			"RoutineTaskRecordTable".attempts,
			"RoutineTaskRecordTable".payload_snapshot,
			"RoutineTaskRecordTable".result_snapshot,
			"RoutineTaskRecordTable".actual_started_at,
			"RoutineTaskRecordTable".actual_ended_at,
			"RoutineTaskRecordTable".updated_at,
			"RoutineTaskRecordTable".created_at`).
		Joins(`INNER JOIN "RoutineRecordTable" routine_record ON routine_record.id = "RoutineTaskRecordTable".routine_record_id`).
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
		return nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	return routineTaskRecords, nil
}

func (r *RoutineTaskRecordRepository) UpdateManyAsFailed(
	failureInputs []inputs.UpdateRoutineTaskRecordFailureInput,
	opts ...RepositoryOptions,
) (int64, *cexceptions.Exception) {
	if len(failureInputs) == 0 {
		return 0, nil
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	valuePlaceholders := make([]string, 0, len(failureInputs))
	valueArgs := make([]any, 0, len(failureInputs)*4+2)
	for _, failureInput := range failureInputs {
		valuePlaceholders = append(valuePlaceholders, "(?::uuid, ?::\"RoutineTaskRecordErrorCode\", ?::varchar, ?::timestamptz)")
		valueArgs = append(
			valueArgs,
			failureInput.Id,
			failureInput.ErrorCode.String(),
			failureInput.ErrorReason,
			failureInput.FailedAt,
		)
	}

	query := fmt.Sprintf(`
		UPDATE "RoutineTaskRecordTable" AS routine_task_record
		SET
			status = ?::"RoutineTaskRecordStatus",
			actual_ended_at = value.failed_at,
			error_code = value.error_code,
			error_reason = value.error_reason,
			updated_at = ?::timestamptz
		FROM (VALUES %s) AS value(id, error_code, error_reason, failed_at)
		WHERE routine_task_record.id = value.id
			AND routine_task_record.status = ?::"RoutineTaskRecordStatus"
	`, strings.Join(valuePlaceholders, ","))
	valueArgs = append(
		[]any{
			cenums.RoutineTaskRecordStatus_Failed.String(),
			time.Now().UTC(),
			cenums.RoutineTaskRecordStatus_Running.String(),
		},
		valueArgs...,
	)

	result := parsedOptions.DB.Exec(query, valueArgs...)
	if result.Error != nil {
		return 0, r.exceptions.FailedToUpdate().WithOrigin(result.Error)
	}

	return result.RowsAffected, nil
}

func (r *RoutineTaskRecordRepository) HardDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskRecord{}).
		Scopes(r.routineTaskRecordScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Where(`"RoutineTaskRecordTable".id = ?`, id).
		Delete(&schemas.RoutineTaskRecord{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RoutineTaskRecordRepository) HardDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	if len(ids) == 0 {
		return r.exceptions.NoChanges()
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	result := parsedOptions.DB.
		Model(&schemas.RoutineTaskRecord{}).
		Scopes(r.routineTaskRecordScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Where(`"RoutineTaskRecordTable".id IN ?`, ids).
		Delete(&schemas.RoutineTaskRecord{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}
