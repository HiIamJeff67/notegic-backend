package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm/clause"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	scopes "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/scopes"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
)

type UsersToStationsRepositoryInterface interface {
	GetOne(stationId uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) (*schemas.UsersToStations, *exceptions.Exception)
	GetMany(stationId uuid.UUID, userIds []uuid.UUID, opts ...options.RepositoryOptions) ([]schemas.UsersToStations, *exceptions.Exception)
	GetManyByStationIdsAndUserId(stationIds []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) ([]schemas.UsersToStations, *exceptions.Exception)
	CreateOne(stationId uuid.UUID, userId uuid.UUID, permission enums.AccessControlPermission, opts ...options.RepositoryOptions) (*schemas.UsersToStations, *exceptions.Exception)
	UpsertMany(stationId uuid.UUID, userIds []uuid.UUID, permissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]schemas.UsersToStations, *exceptions.Exception)
	UpdateOne(stationId uuid.UUID, userId uuid.UUID, permission enums.AccessControlPermission, opts ...options.RepositoryOptions) (*schemas.UsersToStations, *exceptions.Exception)
	DeleteOne(stationId uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *exceptions.Exception
	DeleteMany(stationId uuid.UUID, userIds []uuid.UUID, opts ...options.RepositoryOptions) *exceptions.Exception
	DeleteManyByStationIdsAndUserId(stationIds []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *exceptions.Exception
}

type UsersToStationsRepository struct{}

func NewUsersToStationsRepository() UsersToStationsRepositoryInterface {
	return &UsersToStationsRepository{}
}

func (r *UsersToStationsRepository) GetOne(
	stationId uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) (*schemas.UsersToStations, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var relation schemas.UsersToStations
	result := parsedOptions.DB.
		Model(&schemas.UsersToStations{}).
		Preload(string(schemas.UsersToStationsRelation_User)).
		Where("station_id = ? AND user_id = ?", stationId, userId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&relation)
	if result.Error != nil {
		return nil, apiexceptions.NewStationException().NotFound().WithOrigin(result.Error)
	}

	return &relation, nil
}

func (r *UsersToStationsRepository) GetMany(
	stationId uuid.UUID,
	userIds []uuid.UUID,
	opts ...options.RepositoryOptions,
) ([]schemas.UsersToStations, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var relations []schemas.UsersToStations
	result := parsedOptions.DB.
		Model(&schemas.UsersToStations{}).
		Where("station_id = ? AND user_id IN ?", stationId, userIds).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&relations)
	if result.Error != nil {
		return nil, apiexceptions.NewStationException().NotFound().WithOrigin(result.Error)
	}

	return relations, nil
}

func (r *UsersToStationsRepository) GetManyByStationIdsAndUserId(
	stationIds []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) ([]schemas.UsersToStations, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var relations []schemas.UsersToStations
	result := parsedOptions.DB.
		Model(&schemas.UsersToStations{}).
		Where("station_id IN ? AND user_id = ?", stationIds, userId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&relations)
	if result.Error != nil {
		return nil, apiexceptions.NewStationException().NotFound().WithOrigin(result.Error)
	}

	return relations, nil
}

func (r *UsersToStationsRepository) CreateOne(
	stationId uuid.UUID,
	userId uuid.UUID,
	permission enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) (*schemas.UsersToStations, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	relation := schemas.UsersToStations{
		StationId:  stationId,
		UserId:     userId,
		Permission: permission,
	}
	result := parsedOptions.DB.Create(&relation)
	if result.Error != nil {
		return nil, apiexceptions.NewStationException().FailedToCreate().WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, apiexceptions.NewStationException().NoChanges()
	}

	return r.GetOne(
		stationId,
		userId,
		options.WithDB(parsedOptions.DB),
	)
}

func (r *UsersToStationsRepository) UpsertMany(
	stationId uuid.UUID,
	userIds []uuid.UUID,
	permissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]schemas.UsersToStations, *exceptions.Exception) {
	if len(userIds) != len(permissions) {
		return nil, apiexceptions.NewStationException().InvalidInput("userIds and permissions must have equal lengths")
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	relations := make([]schemas.UsersToStations, len(userIds))
	for index, userId := range userIds {
		relations[index] = schemas.UsersToStations{
			StationId:  stationId,
			UserId:     userId,
			Permission: permissions[index],
		}
	}

	result := parsedOptions.DB.
		Clauses(
			clause.OnConflict{
				Columns: []clause.Column{
					{
						Name: "user_id",
					},
					{
						Name: "station_id",
					},
				},
				DoUpdates: clause.AssignmentColumns([]string{"permission", "updated_at"}),
			},
			clause.Returning{},
		).
		CreateInBatches(&relations, parsedOptions.BatchSize)
	if result.Error != nil {
		return nil, apiexceptions.NewStationException().FailedToUpdate().WithOrigin(result.Error)
	}

	return relations, nil
}

func (r *UsersToStationsRepository) UpdateOne(
	stationId uuid.UUID,
	userId uuid.UUID,
	permission enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) (*schemas.UsersToStations, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	relation := schemas.UsersToStations{
		StationId:  stationId,
		UserId:     userId,
		Permission: permission,
	}
	result := parsedOptions.DB.
		Model(&relation).
		Select("permission").
		Updates(&relation)
	if result.Error != nil {
		return nil, apiexceptions.NewStationException().FailedToUpdate().WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, apiexceptions.NewStationException().NotFound()
	}

	return r.GetOne(
		stationId,
		userId,
		options.WithDB(parsedOptions.DB),
	)
}

func (r *UsersToStationsRepository) DeleteOne(
	stationId uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	parsedOptions := options.ParseRepositoryOptions(opts...)
	result := parsedOptions.DB.
		Where("station_id = ? AND user_id = ?", stationId, userId).
		Delete(&schemas.UsersToStations{})
	if result.Error != nil {
		return apiexceptions.NewStationException().FailedToDelete().WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 {
		return apiexceptions.NewStationException().NotFound()
	}

	return nil
}

func (r *UsersToStationsRepository) DeleteMany(
	stationId uuid.UUID,
	userIds []uuid.UUID,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.
		Where("station_id = ? AND user_id IN ?", stationId, userIds).
		Delete(&schemas.UsersToStations{})
	if result.Error != nil {
		return apiexceptions.NewStationException().FailedToDelete().WithOrigin(result.Error)
	}
	if result.RowsAffected != int64(len(userIds)) {
		return apiexceptions.NewStationException().NotFound()
	}

	return nil
}

func (r *UsersToStationsRepository) DeleteManyByStationIdsAndUserId(
	stationIds []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.
		Where("station_id IN ? AND user_id = ?", stationIds, userId).
		Delete(&schemas.UsersToStations{})
	if result.Error != nil {
		return apiexceptions.NewStationException().FailedToDelete().WithOrigin(result.Error)
	}
	if result.RowsAffected != int64(len(stationIds)) {
		return apiexceptions.NewStationException().NotFound()
	}

	return nil
}
