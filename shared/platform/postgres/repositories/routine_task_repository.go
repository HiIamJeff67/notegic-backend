package repositories

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"github.com/lib/pq"
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
)

type RoutineTaskRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTaskRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) (*schemas.RoutineTask, *cexceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTaskRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]schemas.RoutineTask, *cexceptions.Exception)
	GetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTaskRelation, opts ...RepositoryOptions) (*schemas.RoutineTask, *cexceptions.Exception)
	GetAllByUserId(userId uuid.UUID, preloads []schemas.RoutineTaskRelation, opts ...RepositoryOptions) ([]schemas.RoutineTask, *cexceptions.Exception)
	GetAllByRoutineIds(routineIds []uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTaskRelation, opts ...RepositoryOptions) ([]schemas.RoutineTask, *cexceptions.Exception)
	CreateOneByRoutineId(routineId uuid.UUID, userId uuid.UUID, input inputs.CreateRoutineTaskInput, opts ...RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	CreateManyByRoutineIds(userId uuid.UUID, input []inputs.CreateRoutineTaskByRoutineIdInput, opts ...RepositoryOptions) ([]uuid.UUID, *cexceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateRoutineTaskInput, opts ...RepositoryOptions) (*schemas.RoutineTask, *cexceptions.Exception)
	UpdateManyByIds(userId uuid.UUID, input []inputs.UpdateRoutineTaskByIdInput, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
}

type RoutineTaskRepository struct {
	db               *gorm.DB
	routineTaskScope scopes.RoutineTaskScopeInterface
	exceptions       exceptions.RoutineTaskException
}

func NewRoutineTaskRepository(db *gorm.DB,
	routineTaskScope scopes.RoutineTaskScopeInterface) RoutineTaskRepositoryInterface {
	return &RoutineTaskRepository{
		db:               db,
		routineTaskScope: routineTaskScope,
		exceptions:       exceptions.NewRoutineTaskException(),
	}
}

func (r *RoutineTaskRepository) replaceDependencies(
	db *gorm.DB,
	routineTaskId uuid.UUID,
	routineId uuid.UUID,
	previousRoutineTaskIds []uuid.UUID,
	batchSize int,
) *cexceptions.Exception {
	if routineTaskId == uuid.Nil || routineId == uuid.Nil {
		return r.exceptions.InvalidInput().WithOrigin(fmt.Errorf("routine task and routine ids are required"))
	}

	var routineTaskIds []uuid.UUID
	result := db.Model(&schemas.RoutineTask{}).
		Where("routine_id = ?", routineId).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Pluck("id", &routineTaskIds)
	if result.Error != nil {
		return r.exceptions.FailedToUpdate().WithOrigin(result.Error)
	}

	seenPreviousRoutineTaskIds := make(map[uuid.UUID]struct{}, len(previousRoutineTaskIds))
	for _, previousRoutineTaskId := range previousRoutineTaskIds {
		if previousRoutineTaskId == uuid.Nil {
			return r.exceptions.InvalidInput().WithOrigin(fmt.Errorf("previous routine task id is required"))
		}
		if previousRoutineTaskId == routineTaskId {
			return r.exceptions.InvalidInput().WithOrigin(fmt.Errorf("a routine task cannot depend on itself"))
		}
		if _, exists := seenPreviousRoutineTaskIds[previousRoutineTaskId]; exists {
			return r.exceptions.InvalidInput().WithOrigin(fmt.Errorf("duplicate previous routine task id"))
		}
		seenPreviousRoutineTaskIds[previousRoutineTaskId] = struct{}{}
	}

	if len(previousRoutineTaskIds) > 0 {
		var validPreviousRoutineTaskCount int64
		result = db.Model(&schemas.RoutineTask{}).
			Where("routine_id = ? AND id IN ?", routineId, previousRoutineTaskIds).
			Count(&validPreviousRoutineTaskCount)
		if result.Error != nil {
			return r.exceptions.FailedToUpdate().WithOrigin(result.Error)
		}
		if validPreviousRoutineTaskCount != int64(len(previousRoutineTaskIds)) {
			return r.exceptions.InvalidInput().WithOrigin(fmt.Errorf("previous routine tasks must belong to the same routine"))
		}

		var createsCycle bool
		result = db.Raw(`
			WITH RECURSIVE previous_tasks(id) AS (
				SELECT unnest(?::uuid[])
				UNION
				SELECT dependency.previous_routine_task_id
				FROM "RoutineDependencyTable" dependency
				INNER JOIN previous_tasks
					ON dependency.routine_task_id = previous_tasks.id
			)
			SELECT EXISTS (
				SELECT 1 FROM previous_tasks WHERE id = ?
			)
		`, pq.Array(previousRoutineTaskIds), routineTaskId).Scan(&createsCycle)
		if result.Error != nil {
			return r.exceptions.FailedToUpdate().WithOrigin(result.Error)
		}
		if createsCycle {
			return r.exceptions.InvalidInput().WithOrigin(fmt.Errorf("routine task dependencies cannot contain a cycle"))
		}
	}

	result = db.Where("routine_task_id = ?", routineTaskId).Delete(&schemas.RoutineTaskDependency{})
	if result.Error != nil {
		return r.exceptions.FailedToUpdate().WithOrigin(result.Error)
	}
	if len(previousRoutineTaskIds) == 0 {
		return nil
	}

	dependencies := make([]schemas.RoutineTaskDependency, len(previousRoutineTaskIds))
	for index, previousRoutineTaskId := range previousRoutineTaskIds {
		dependencies[index] = schemas.RoutineTaskDependency{
			RoutineTaskId:         routineTaskId,
			PreviousRoutineTaskId: previousRoutineTaskId,
		}
	}
	if result = db.CreateInBatches(&dependencies, batchSize); result.Error != nil {
		return r.exceptions.FailedToUpdate().WithOrigin(result.Error)
	}

	return nil
}

func (r *RoutineTaskRepository) HasPermission(
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
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) (*schemas.RoutineTask, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

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
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: routineTask.Id == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &routineTask, nil
}

func (r *RoutineTaskRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineTaskRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]schemas.RoutineTask, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var routineTasks []schemas.RoutineTask
	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		Scopes(r.routineTaskScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.routineTaskScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&routineTasks)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: len(routineTasks) == 0, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return routineTasks, nil
}

func (r *RoutineTaskRepository) GetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineTaskRelation,
	opts ...RepositoryOptions,
) (*schemas.RoutineTask, *cexceptions.Exception) {
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

func (r *RoutineTaskRepository) GetAllByUserId(
	userId uuid.UUID,
	preloads []schemas.RoutineTaskRelation,
	opts ...RepositoryOptions,
) ([]schemas.RoutineTask, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
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
		return nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	return routineTasks, nil
}

func (r *RoutineTaskRepository) GetAllByRoutineIds(
	routineIds []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineTaskRelation,
	opts ...RepositoryOptions,
) ([]schemas.RoutineTask, *cexceptions.Exception) {
	if len(routineIds) == 0 {
		return []schemas.RoutineTask{}, nil
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
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
		return nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	return routineTasks, nil
}

func (r *RoutineTaskRepository) CreateOneByRoutineId(
	routineId uuid.UUID,
	userId uuid.UUID,
	input inputs.CreateRoutineTaskInput,
	opts ...RepositoryOptions,
) (*uuid.UUID, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	if input.ActorUserId == uuid.Nil || input.ActorUserId != userId {
		return nil, r.exceptions.InvalidInput().WithOrigin(fmt.Errorf("actorUserId must match userId"))
	}

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, WithTransactionDB(parsedOptions.DB))
		opts = append(opts, WithLockingStrength(LockingStrengthNoKeyUpdate))
	}

	routineRepository := NewRoutineRepository(r.db, scopes.NewRoutineScope())
	if !routineRepository.HasPermission(
		routineId,
		userId,
		parsedOptions.AllowedPermissions,
		opts...,
	) {
		parsedOptions.DB.Rollback()
		return nil, r.exceptions.NoPermission("create a routine task under this routine")
	}

	newRoutineTask := schemas.RoutineTask{}
	if err := copier.Copy(&newRoutineTask, &input); err != nil {
		parsedOptions.DB.Rollback()
		return nil, r.exceptions.InvalidInput().WithOrigin(err)
	}
	newRoutineTask.RoutineId = routineId
	newRoutineTask.ActorUserId = input.ActorUserId

	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		Create(&newRoutineTask)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}
	if exception := r.replaceDependencies(
		parsedOptions.DB,
		newRoutineTask.Id,
		routineId,
		input.PreviousRoutineTaskIds,
		parsedOptions.BatchSize,
	); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &newRoutineTask.Id, nil
}

func (r *RoutineTaskRepository) CreateManyByRoutineIds(
	userId uuid.UUID,
	input []inputs.CreateRoutineTaskByRoutineIdInput,
	opts ...RepositoryOptions,
) ([]uuid.UUID, *cexceptions.Exception) {
	if len(input) == 0 {
		return nil, r.exceptions.NoChanges()
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	for _, in := range input {
		if in.ActorUserId == uuid.Nil || in.ActorUserId != userId {
			return nil, r.exceptions.InvalidInput().WithOrigin(fmt.Errorf("actorUserId must match userId"))
		}
	}

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, WithTransactionDB(parsedOptions.DB))
		opts = append(opts, WithLockingStrength(LockingStrengthNoKeyUpdate))
	}

	routineIds := make([]uuid.UUID, len(input))
	for index, in := range input {
		routineIds[index] = in.RoutineId
	}

	routineRepository := NewRoutineRepository(r.db, scopes.NewRoutineScope())
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
		newRoutineTask := schemas.RoutineTask{
			Id:          uuid.New(),
			RoutineId:   in.RoutineId,
			ActorUserId: in.ActorUserId,
		}
		if err := copier.Copy(&newRoutineTask, &in); err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.InvalidInput().WithOrigin(err)
		}
		newRoutineTask.RoutineId = in.RoutineId
		newRoutineTasks = append(newRoutineTasks, newRoutineTask)
	}

	if len(newRoutineTasks) == 0 {
		parsedOptions.DB.Rollback()
		return nil, r.exceptions.NoChanges()
	}

	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		CreateInBatches(&newRoutineTasks, parsedOptions.BatchSize)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
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
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return newRoutineTaskIds, nil
}

func (r *RoutineTaskRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateRoutineTaskInput,
	opts ...RepositoryOptions,
) (*schemas.RoutineTask, *cexceptions.Exception) {
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

	existingRoutineTask, exception := r.CheckPermissionAndGetOneById(id, userId, nil, parsedOptions.AllowedPermissions, opts...)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}
	if input.Values.RoutineId != nil && !partialupdate.CheckSetNull(input.SetNull, "RoutineId") {
		routineRepository := NewRoutineRepository(r.db, scopes.NewRoutineScope())
		if !routineRepository.HasPermission(*input.Values.RoutineId, userId, parsedOptions.AllowedPermissions, opts...) {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.NoPermission("move a routine task to this routine")
		}
	}
	previousRoutineTaskIds := input.Values.PreviousRoutineTaskIds
	input.Values.PreviousRoutineTaskIds = nil
	if input.SetNull != nil {
		for fieldName := range *input.SetNull {
			normalizedFieldName := strings.ToLower(strings.ReplaceAll(fieldName, "_", ""))
			if normalizedFieldName == "previousroutinetaskids" {
				delete(*input.SetNull, fieldName)
			}
		}
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingRoutineTask)
	if err != nil {
		parsedOptions.DB.Rollback()
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true).WithOrigin(err)
	}

	result := parsedOptions.DB.
		Model(&schemas.RoutineTask{}).
		Where(`"RoutineTaskTable".id = ?`, id).
		Select("*").
		Updates(&updates)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0 && previousRoutineTaskIds == nil && input.Values.RoutineId == nil, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}
	if previousRoutineTaskIds != nil || input.Values.RoutineId != nil {
		routineId := existingRoutineTask.RoutineId
		if input.Values.RoutineId != nil {
			routineId = *input.Values.RoutineId
		}
		dependencyIds := []uuid.UUID{}
		if previousRoutineTaskIds != nil {
			dependencyIds = *previousRoutineTaskIds
		}
		if exception := r.replaceDependencies(
			parsedOptions.DB,
			id,
			routineId,
			dependencyIds,
			parsedOptions.BatchSize,
		); exception != nil {
			parsedOptions.DB.Rollback()
			return nil, exception
		}
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &updates, nil
}

func (r *RoutineTaskRepository) UpdateManyByIds(
	userId uuid.UUID,
	input []inputs.UpdateRoutineTaskByIdInput,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	if len(input) == 0 {
		return r.exceptions.NoChanges()
	}

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
	validRoutineTasks, exception := r.CheckPermissionsAndGetManyByIds(ids, userId, nil, parsedOptions.AllowedPermissions, opts...)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return r.exceptions.NoPermission("update these routine tasks")
	}

	isRoutineTaskValid := make(map[uuid.UUID]bool, len(validRoutineTasks))
	for _, validRoutineTask := range validRoutineTasks {
		isRoutineTaskValid[validRoutineTask.Id] = true
	}

	targetRoutineIdSet := make(map[uuid.UUID]bool)
	for _, in := range input {
		if in.PartialUpdateInput.Values.PreviousRoutineTaskIds != nil ||
			partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "PreviousRoutineTaskIds") {
			parsedOptions.DB.Rollback()
			return r.exceptions.InvalidInput().WithOrigin(fmt.Errorf("bulk routine task updates do not support dependency changes"))
		}
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
		routineRepository := NewRoutineRepository(r.db, scopes.NewRoutineScope())
		if !routineRepository.HavePermissions(targetRoutineIds, userId, parsedOptions.AllowedPermissions, opts...) {
			parsedOptions.DB.Rollback()
			return r.exceptions.NoPermission("move these routine tasks to the given routines")
		}
	}

	var valuePlaceholders []string
	var valueArgs []interface{}
	for _, in := range input {
		if !isRoutineTaskValid[in.Id] {
			continue
		}

		valuePlaceholders = append(valuePlaceholders, `(?::uuid, ?::uuid, ?::text, ?::"RoutineTaskPurpose", ?::jsonb, ?::integer, ?::integer)`)
		valueArgs = append(valueArgs,
			in.Id,
			in.PartialUpdateInput.Values.RoutineId,
			in.PartialUpdateInput.Values.Title,
			in.PartialUpdateInput.Values.Purpose,
			in.PartialUpdateInput.Values.Payload,
			in.PartialUpdateInput.Values.Priority,
			in.PartialUpdateInput.Values.MaxAttempts,
		)
	}

	if len(valuePlaceholders) == 0 {
		parsedOptions.DB.Rollback()
		return r.exceptions.NoChanges()
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
			updated_at = NOW()
		FROM (VALUES %s) AS v(id, routine_id, title, purpose, payload, priority, max_attempts)
		WHERE rt.id = v.id::uuid
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

func (r *RoutineTaskRepository) HardDeleteOneById(
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
		Model(&schemas.RoutineTask{}).
		Scopes(r.routineTaskScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Where(`"RoutineTaskTable".id = ?`, id).
		Delete(&schemas.RoutineTask{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RoutineTaskRepository) HardDeleteManyByIds(
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
		Model(&schemas.RoutineTask{}).
		Scopes(r.routineTaskScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Where(`"RoutineTaskTable".id IN ?`, ids).
		Delete(&schemas.RoutineTask{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}
