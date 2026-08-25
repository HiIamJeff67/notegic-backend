package repositories

import (
	"fmt"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm/clause"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	array "github.com/HiIamJeff67/notegic-backend/shared/lib/array"
	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type StationRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.StationRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) (*schemas.Station, enums.AccessControlPermission, *cexceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.StationRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]schemas.Station, []enums.AccessControlPermission, *cexceptions.Exception)
	GetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.StationRelation, opts ...options.RepositoryOptions) (*schemas.Station, enums.AccessControlPermission, *cexceptions.Exception)
	GetAllByUserId(userId uuid.UUID, preloads []schemas.StationRelation, opts ...options.RepositoryOptions) ([]schemas.Station, []enums.AccessControlPermission, *cexceptions.Exception)
	CreateOne(ownerId uuid.UUID, input inputs.CreateStationInput, opts ...options.RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	CreateMany(ownerId uuid.UUID, input []inputs.CreateStationInput, opts ...options.RepositoryOptions) ([]uuid.UUID, *cexceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateStationInput, opts ...options.RepositoryOptions) (*schemas.Station, *cexceptions.Exception)
	UpdateManyByIds(userId uuid.UUID, input []inputs.UpdateStationByIdInput, opts ...options.RepositoryOptions) *cexceptions.Exception
	RestoreSoftDeletedOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) (*schemas.Station, *cexceptions.Exception)
	RestoreSoftDeletedManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) ([]schemas.Station, *cexceptions.Exception)
	SoftDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	SoftDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	SoftDeleteManyByUserId(userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByUserId(userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception

	/* ============================== System Only Method ============================== */
	BulkCheckPermissionsAndGetManyByIds(inputs []inputs.BulkCheckStationPermissionInput, preloads []schemas.StationRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]bool, []schemas.Station, *cexceptions.Exception)
	BulkCreateMany(inputs []inputs.BulkCreateStationInput, opts ...options.RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkUpdateMany(inputs []inputs.BulkUpdateStationInput, opts ...options.RepositoryOptions) ([]bool, *cexceptions.Exception)
}

type StationRepository struct {
	stationScope scopes.StationScopeInterface
}

func NewStationRepository(stationScope scopes.StationScopeInterface) StationRepositoryInterface {
	return &StationRepository{
		stationScope: stationScope,
	}
}

func (r *StationRepository) HasPermission(
	id uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var marker int
	result := parsedOptions.DB.
		Model(&schemas.Station{}).
		Select("1").
		Scopes(r.stationScope.PassPermissionCheck(id, userId, allowedPermissions)).
		Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Limit(1).
		Scan(&marker)
	if err := result.Error; err != nil {
		return false
	}

	return marker == 1
}

func (r *StationRepository) HavePermissions(
	ids []uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var permittedIds []uuid.UUID
	result := parsedOptions.DB.
		Model(&schemas.Station{}).
		Select(`DISTINCT "StationTable".id`).
		Scopes(r.stationScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&permittedIds)
	if err := result.Error; err != nil {
		return false
	}

	return array.GetDistinctCount(ids) == array.GetDistinctCount(permittedIds)
}

func (r *StationRepository) CheckPermissionAndGetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.StationRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) (*schemas.Station, enums.AccessControlPermission, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	type stationWithPermission struct {
		schemas.Station
		Permission enums.AccessControlPermission `gorm:"column:permission"`
	}

	var station stationWithPermission
	query := parsedOptions.DB.
		Model(&schemas.Station{}).
		Select(`"StationTable".*, users_to_station.permission AS permission`).
		Joins(`
			INNER JOIN "UsersToStationsTable" AS users_to_station
				ON users_to_station.station_id = "StationTable".id
				AND users_to_station.user_id = ?
		`, userId).
		Where(`"StationTable".id = ?`, id)
	if allowedPermissions != nil && len(allowedPermissions) > 0 {
		query = query.Scopes(
			r.stationScope.PassPermissionCheck(id, userId, allowedPermissions),
		)
	}

	result := query.
		Scopes(r.stationScope.IncludePreloads(preloads)).
		Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&station)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().NotFound().WithOrigin(result.Error)},
		{First: station.Id == uuid.Nil, Second: apiexceptions.NewStationException().NotFound()},
	}); exception != nil {
		return nil, "", exception
	}

	return &station.Station, station.Permission, nil
}

func (r *StationRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.StationRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]schemas.Station, []enums.AccessControlPermission, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var stations []schemas.Station
	result := parsedOptions.DB.
		Model(&schemas.Station{}).
		Scopes(r.stationScope.IncludePreloads(preloads)).
		Scopes(r.stationScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&stations)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().NotFound().WithOrigin(result.Error)},
		{First: len(stations) == 0, Second: apiexceptions.NewStationException().NotFound()},
	}); exception != nil {
		return nil, nil, exception
	}

	permissions := make([]enums.AccessControlPermission, len(stations))
	if allowedPermissions != nil {
		var usersToStations []schemas.UsersToStations
		result = parsedOptions.DB.
			Model(&schemas.UsersToStations{}).
			Select("station_id, permission").
			Where(
				"station_id IN ? AND user_id = ? AND permission IN ?",
				ids,
				userId,
				allowedPermissions,
			).
			Scopes(scopes.Locking(parsedOptions.LockingStrength)).
			Find(&usersToStations)
		if exception := cexceptions.Cover(nil, []cexceptions.Pair{
			{First: result.Error != nil, Second: apiexceptions.NewStationException().NotFound().WithOrigin(result.Error)},
			{First: len(usersToStations) == 0, Second: apiexceptions.NewStationException().NotFound()},
		}); exception != nil {
			return nil, nil, exception
		}

		permissionByStationId := make(map[uuid.UUID]enums.AccessControlPermission, len(usersToStations))
		for _, usersToStation := range usersToStations {
			permissionByStationId[usersToStation.StationId] = usersToStation.Permission
		}

		for index, station := range stations {
			permission, exist := permissionByStationId[station.Id]
			if !exist {
				return nil, nil, apiexceptions.NewStationException().NotFound()
			}
			permissions[index] = permission
		}
	}

	return stations, permissions, nil
}

func (r *StationRepository) GetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.StationRelation,
	opts ...options.RepositoryOptions,
) (*schemas.Station, enums.AccessControlPermission, *cexceptions.Exception) {
	return r.CheckPermissionAndGetOneById(
		id,
		userId,
		preloads,
		options.ParseRepositoryOptions(opts...).AllowedPermissions,
		opts...,
	)
}

func (r *StationRepository) GetAllByUserId(
	userId uuid.UUID,
	preloads []schemas.StationRelation,
	opts ...options.RepositoryOptions,
) ([]schemas.Station, []enums.AccessControlPermission, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	type stationWithPermission struct {
		schemas.Station
		Permission enums.AccessControlPermission `gorm:"column:permission"`
	}

	var stationsWithPermissions []stationWithPermission
	result := parsedOptions.DB.
		Model(&schemas.Station{}).
		Select(`"StationTable".*, uts.permission AS permission`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = "StationTable".id`).
		Where("uts.user_id = ?", userId)
	if parsedOptions.HasAllowedPermissions() {
		result = result.Where("uts.permission IN ?", parsedOptions.AllowedPermissions)
	}
	result = result.Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.stationScope.IncludePreloads(preloads)).
		Order(`"StationTable".created_at ASC`).
		Order(`"StationTable".id ASC`).
		Find(&stationsWithPermissions)
	if result.Error != nil {
		return nil, nil, apiexceptions.NewStationException().NotFound().WithOrigin(result.Error)
	}

	stations := make([]schemas.Station, len(stationsWithPermissions))
	permissions := make([]enums.AccessControlPermission, len(stationsWithPermissions))
	for index, stationWithPermission := range stationsWithPermissions {
		stations[index] = stationWithPermission.Station
		permissions[index] = stationWithPermission.Permission
	}

	return stations, permissions, nil
}

func (r *StationRepository) CreateOne(
	ownerId uuid.UUID,
	input inputs.CreateStationInput,
	opts ...options.RepositoryOptions,
) (*uuid.UUID, *cexceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	var newStation schemas.Station
	newStation.OwnerId = ownerId
	if err := copier.Copy(&newStation, &input); err != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewStationException().InvalidInput().WithOrigin(err)
	}
	if newStation.Id == uuid.Nil {
		newStation.Id = uuid.New()
	}

	result := parsedOptions.DB.Model(&schemas.Station{}).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Create(&newStation)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToCreate().WithOrigin(result.Error)},
		{First: newStation.Id == uuid.Nil, Second: apiexceptions.NewStationException().FailedToCreate()},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	newUsersToStations := schemas.UsersToStations{
		UserId:     ownerId,
		StationId:  newStation.Id,
		Permission: enums.AccessControlPermission_Owner,
	}
	result = parsedOptions.DB.Model(&schemas.UsersToStations{}).
		Create(&newUsersToStations)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewStationException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &newStation.Id, nil
}

func (r *StationRepository) CreateMany(
	ownerId uuid.UUID,
	input []inputs.CreateStationInput,
	opts ...options.RepositoryOptions,
) ([]uuid.UUID, *cexceptions.Exception) {
	if len(input) == 0 {
		return nil, apiexceptions.NewStationException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	newStations := make([]schemas.Station, 0, len(input))
	for _, in := range input {
		var newStation schemas.Station
		newStation.OwnerId = ownerId
		if err := copier.Copy(&newStation, &in); err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewStationException().InvalidInput().WithOrigin(err)
		}
		if newStation.Id == uuid.Nil {
			newStation.Id = uuid.New()
		}
		newStations = append(newStations, newStation)
	}

	result := parsedOptions.DB.Model(&schemas.Station{}).
		CreateInBatches(&newStations, parsedOptions.BatchSize)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	newStationIds := make([]uuid.UUID, len(newStations))
	newUsersToStations := make([]schemas.UsersToStations, len(newStations))
	for index, newStation := range newStations {
		newStationIds[index] = newStation.Id
		newUsersToStations[index] = schemas.UsersToStations{
			UserId:     ownerId,
			StationId:  newStation.Id,
			Permission: enums.AccessControlPermission_Owner,
		}
	}
	result = parsedOptions.DB.Model(&schemas.UsersToStations{}).
		CreateInBatches(&newUsersToStations, parsedOptions.BatchSize)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewStationException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return newStationIds, nil
}

func (r *StationRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateStationInput,
	opts ...options.RepositoryOptions,
) (*schemas.Station, *cexceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	existingStation, _, exception := r.CheckPermissionAndGetOneById(
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

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingStation)
	if err != nil {
		parsedOptions.DB.Rollback()
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true).WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.Station{}).
		Where(`"StationTable".id = ? AND "StationTable".deleted_at IS NULL`, id).
		Select("*").
		Updates(&updates)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewStationException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &updates, nil
}

func (r *StationRepository) UpdateManyByIds(
	userId uuid.UUID,
	input []inputs.UpdateStationByIdInput,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	if len(input) == 0 {
		return apiexceptions.NewStationException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
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

	isStationValid := make(map[uuid.UUID]bool, len(input))
	if parsedOptions.HasAllowedPermissions() {
		validStations, _, exception := r.CheckPermissionsAndGetManyByIds(
			ids,
			userId,
			nil,
			parsedOptions.AllowedPermissions,
			opts...,
		)
		if exception != nil {
			parsedOptions.DB.Rollback()
			return apiexceptions.NewStationException().NoPermission("update these stations")
		}

		for _, validStation := range validStations {
			isStationValid[validStation.Id] = true
		}
	}

	var valuePlaceholders []string
	var valueArgs []interface{}
	for _, in := range input {
		if parsedOptions.HasAllowedPermissions() && !isStationValid[in.Id] {
			continue
		}

		setIconNull := partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "Icon")
		setHeaderBackgroundURLNull := partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "HeaderBackgroundURL")

		valuePlaceholders = append(valuePlaceholders, `(?::uuid, ?::text, ?::text, ?::"SupportedIcon", ?::text, ?::boolean, ?::boolean)`)
		valueArgs = append(valueArgs,
			in.Id,
			in.PartialUpdateInput.Values.Name,
			in.PartialUpdateInput.Values.Description,
			in.PartialUpdateInput.Values.Icon,
			in.PartialUpdateInput.Values.HeaderBackgroundURL,
			setIconNull,
			setHeaderBackgroundURLNull,
		)
	}

	if len(valuePlaceholders) == 0 {
		parsedOptions.DB.Rollback()
		return apiexceptions.NewStationException().NoChanges()
	}

	sql := fmt.Sprintf(`
		UPDATE "StationTable" AS s
		SET
			name = COALESCE(v.name::text, s.name),
			description = COALESCE(v.description::text, s.description),
			icon = CASE
				WHEN v.set_icon_null::boolean THEN NULL
				ELSE COALESCE(v.icon::"SupportedIcon", s.icon)
			END,
			header_background_url = CASE
				WHEN v.set_header_background_url_null::boolean THEN NULL
				ELSE COALESCE(v.header_background_url::text, s.header_background_url)
			END,
			updated_at = NOW()
		FROM (VALUES %s) AS v(id, name, description, icon, header_background_url, set_icon_null, set_header_background_url_null)
		WHERE s.id = v.id::uuid AND s.deleted_at IS NULL
	`, strings.Join(valuePlaceholders, ","))
	result := parsedOptions.DB.Exec(sql, valueArgs...)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return apiexceptions.NewStationException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return nil
}

func (r *StationRepository) RestoreSoftDeletedOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) (*schemas.Station, *cexceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var restoredStation schemas.Station
	result := parsedOptions.DB.Model(&restoredStation).
		Scopes(r.stationScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Clauses(clause.Returning{}).
		Where(`"StationTable".id = ?`, id).
		Updates(map[string]interface{}{"deleted_at": nil})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToUpdate().WithOrigin(result.Error)},
		{First: restoredStation.Id == uuid.Nil, Second: apiexceptions.NewStationException().FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &restoredStation, nil
}

func (r *StationRepository) RestoreSoftDeletedManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) ([]schemas.Station, *cexceptions.Exception) {
	if len(ids) == 0 {
		return nil, apiexceptions.NewStationException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var restoredStations []schemas.Station
	result := parsedOptions.DB.Model(&restoredStations).
		Scopes(r.stationScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Clauses(clause.Returning{}).
		Where(`"StationTable".id IN ?`, ids).
		Updates(map[string]interface{}{"deleted_at": nil})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToUpdate().WithOrigin(result.Error)},
		{First: len(restoredStations) == 0, Second: apiexceptions.NewStationException().FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return restoredStations, nil
}

func (r *StationRepository) SoftDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.Station{}).
		Scopes(r.stationScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"StationTable".id = ?`, id).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *StationRepository) SoftDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	if len(ids) == 0 {
		return apiexceptions.NewStationException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.Station{}).
		Scopes(r.stationScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"StationTable".id IN ?`, ids).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *StationRepository) SoftDeleteManyByUserId(
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.Station{}).
		Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where("owner_id = ?", userId).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *StationRepository) HardDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.Station{}).
		Scopes(r.stationScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"StationTable".id = ?`, id).
		Delete(&schemas.Station{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *StationRepository) HardDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	if len(ids) == 0 {
		return apiexceptions.NewStationException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.Station{}).
		Scopes(r.stationScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"StationTable".id IN ?`, ids).
		Delete(&schemas.Station{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *StationRepository) HardDeleteManyByUserId(
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.Station{}).
		Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where("owner_id = ?", userId).
		Delete(&schemas.Station{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewStationException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewStationException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

/* ============================== System Only Method ============================== */

func (r *StationRepository) BulkCheckPermissionsAndGetManyByIds(
	inputs []inputs.BulkCheckStationPermissionInput,
	preloads []schemas.StationRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]bool, []schemas.Station, *cexceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, []schemas.Station{}, nil
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

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
		result := parsedOptions.DB.Model(&schemas.Station{}).
			Select(`"StationTable".id, uts.user_id`).
			Joins(`INNER JOIN "UsersToStationsTable" AS uts ON uts.station_id = "StationTable".id`).
			Where(`"StationTable".id IN ? AND uts.user_id IN ? AND uts.permission IN ?`, ids, userIds, allowedPermissions).
			Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
			Scan(&validTargets)
		if result.Error != nil {
			return nil, nil, apiexceptions.NewStationException().NotFound().WithOrigin(result.Error)
		}

		for _, validTarget := range validTargets {
			validTargetByUserId[[2]uuid.UUID{validTarget.Id, validTarget.UserId}] = true
			validIdSet[validTarget.Id] = true
		}
	} else {
		var validIds []uuid.UUID
		result := parsedOptions.DB.Model(&schemas.Station{}).
			Select(`"StationTable".id`).
			Where(`"StationTable".id IN ?`, ids).
			Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
			Scan(&validIds)
		if result.Error != nil {
			return nil, nil, apiexceptions.NewStationException().NotFound().WithOrigin(result.Error)
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
		return successes, []schemas.Station{}, nil
	}

	var stations []schemas.Station
	result := parsedOptions.DB.Model(&schemas.Station{}).
		Where(`"StationTable".id IN ?`, validIds).
		Scopes(r.stationScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.stationScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&stations)
	if result.Error != nil {
		return nil, nil, apiexceptions.NewStationException().NotFound().WithOrigin(result.Error)
	}

	foundIdSet := make(map[uuid.UUID]bool, len(stations))
	for _, station := range stations {
		foundIdSet[station.Id] = true
	}
	for index, in := range inputs {
		if validIdSet[in.Id] &&
			foundIdSet[in.Id] &&
			(allowedPermissions == nil || validTargetByUserId[[2]uuid.UUID{in.Id, in.UserId}]) {
			successes[index] = true
		}
	}

	return successes, stations, nil
}

func (r *StationRepository) BulkCreateMany(
	inputs []inputs.BulkCreateStationInput,
	opts ...options.RepositoryOptions,
) ([]bool, *cexceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, apiexceptions.NewStationException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	newStations := make([]schemas.Station, len(inputs))
	newUsersToStations := make([]schemas.UsersToStations, len(inputs))
	for index, in := range inputs {
		newStationId := uuid.New()
		if in.Id != nil && *in.Id != uuid.Nil {
			newStationId = *in.Id
		}

		newStations[index] = schemas.Station{
			Id:                  newStationId,
			OwnerId:             in.UserId,
			Name:                in.Name,
			Description:         in.Description,
			Icon:                in.Icon,
			HeaderBackgroundURL: in.HeaderBackgroundURL,
		}
		newUsersToStations[index] = schemas.UsersToStations{
			UserId:     in.UserId,
			StationId:  newStationId,
			Permission: enums.AccessControlPermission_Owner,
		}
	}

	result := parsedOptions.DB.Model(&schemas.Station{}).
		CreateInBatches(&newStations, parsedOptions.BatchSize)
	if result.Error != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewStationException().FailedToCreate().WithOrigin(result.Error)
	}

	result = parsedOptions.DB.Model(&schemas.UsersToStations{}).
		CreateInBatches(&newUsersToStations, parsedOptions.BatchSize)
	if result.Error != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewStationException().FailedToCreate().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewStationException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return make([]bool, len(inputs)), nil
}

func (r *StationRepository) BulkUpdateMany(
	bulkInputs []inputs.BulkUpdateStationInput,
	opts ...options.RepositoryOptions,
) ([]bool, *cexceptions.Exception) {
	if len(bulkInputs) == 0 {
		return []bool{}, apiexceptions.NewStationException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	checkInputs := make([]inputs.BulkCheckStationPermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckStationPermissionInput{
			UserId: in.UserId,
			Id:     in.Id,
		}
	}
	checkOptions := append(opts, options.WithTransactionDB(parsedOptions.DB))
	checkOptions = append(checkOptions, options.WithOnlyDeleted(types.Ternary_Negative))
	checkOptions = append(checkOptions, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
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

	valuePlaceholders := make([]string, 0, len(bulkInputs))
	valueArgs := make([]interface{}, 0, len(bulkInputs)*8)
	for index, in := range bulkInputs {
		if !successes[index] {
			continue
		}

		setIconNull := partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "Icon")
		setHeaderBackgroundURLNull := partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "HeaderBackgroundURL")

		valuePlaceholders = append(valuePlaceholders, `(?::int, ?::uuid, ?::text, ?::text, ?::"SupportedIcon", ?::text, ?::boolean, ?::boolean)`)
		valueArgs = append(valueArgs,
			index,
			in.Id,
			in.PartialUpdateInput.Values.Name,
			in.PartialUpdateInput.Values.Description,
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
		WITH payload(idx, id, name, description, icon, header_background_url, set_icon_null, set_header_background_url_null) AS (
			VALUES %s
		),
		updated AS (
			UPDATE "StationTable" AS s
			SET
				name = COALESCE(v.name::text, s.name),
				description = COALESCE(v.description::text, s.description),
				icon = CASE
					WHEN v.set_icon_null::boolean THEN NULL
					ELSE COALESCE(v.icon::"SupportedIcon", s.icon)
				END,
				header_background_url = CASE
					WHEN v.set_header_background_url_null::boolean THEN NULL
					ELSE COALESCE(v.header_background_url::text, s.header_background_url)
				END,
				updated_at = NOW()
			FROM payload AS v
			WHERE s.id = v.id::uuid
				AND s.deleted_at IS NULL
			RETURNING s.id
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
		return nil, apiexceptions.NewStationException().FailedToUpdate().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewStationException().FailedToCommitTransaction().WithOrigin(err)
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
