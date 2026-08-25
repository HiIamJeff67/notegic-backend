package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptions "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type UsersToStationsRepositoryInterface interface {
	GetOne(stationId uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) (*schemas.UsersToStations, *cexceptions.Exception)
	GetMany(stationId uuid.UUID, userIds []uuid.UUID, opts ...RepositoryOptions) ([]schemas.UsersToStations, *cexceptions.Exception)
	GetManyByStationIdsAndUserId(stationIds []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) ([]schemas.UsersToStations, *cexceptions.Exception)
	CreateOne(stationId uuid.UUID, userId uuid.UUID, permission cenums.AccessControlPermission, opts ...RepositoryOptions) (*schemas.UsersToStations, *cexceptions.Exception)
	UpsertMany(stationId uuid.UUID, userIds []uuid.UUID, permissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]schemas.UsersToStations, *cexceptions.Exception)
	UpdateOne(stationId uuid.UUID, userId uuid.UUID, permission cenums.AccessControlPermission, opts ...RepositoryOptions) (*schemas.UsersToStations, *cexceptions.Exception)
	DeleteOne(stationId uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	DeleteMany(stationId uuid.UUID, userIds []uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	DeleteManyByStationIdsAndUserId(stationIds []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
}

type UsersToStationsRepository struct {
	db         *gorm.DB
	exceptions exceptions.StationException
}

func NewUsersToStationsRepository(db *gorm.DB) UsersToStationsRepositoryInterface {
	return &UsersToStationsRepository{
		db: db, exceptions: exceptions.NewStationException()}
}

func (r *UsersToStationsRepository) GetOne(
	stationId uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) (*schemas.UsersToStations, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var relation schemas.UsersToStations
	result := parsedOptions.DB.
		Model(&schemas.UsersToStations{}).
		Preload(string(schemas.UsersToStationsRelation_User)).
		Where("station_id = ? AND user_id = ?", stationId, userId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&relation)
	if result.Error != nil {
		return nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	return &relation, nil
}

func (r *UsersToStationsRepository) GetMany(
	stationId uuid.UUID,
	userIds []uuid.UUID,
	opts ...RepositoryOptions,
) ([]schemas.UsersToStations, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var relations []schemas.UsersToStations
	result := parsedOptions.DB.
		Model(&schemas.UsersToStations{}).
		Where("station_id = ? AND user_id IN ?", stationId, userIds).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&relations)
	if result.Error != nil {
		return nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	return relations, nil
}

func (r *UsersToStationsRepository) GetManyByStationIdsAndUserId(
	stationIds []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) ([]schemas.UsersToStations, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var relations []schemas.UsersToStations
	result := parsedOptions.DB.
		Model(&schemas.UsersToStations{}).
		Where("station_id IN ? AND user_id = ?", stationIds, userId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&relations)
	if result.Error != nil {
		return nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	return relations, nil
}

func (r *UsersToStationsRepository) CreateOne(
	stationId uuid.UUID,
	userId uuid.UUID,
	permission cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) (*schemas.UsersToStations, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	relation := schemas.UsersToStations{
		StationId:  stationId,
		UserId:     userId,
		Permission: permission,
	}
	result := parsedOptions.DB.Create(&relation)
	if result.Error != nil {
		return nil, r.exceptions.FailedToCreate().WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, r.exceptions.NoChanges()
	}

	return r.GetOne(
		stationId,
		userId,
		WithDB(parsedOptions.DB),
	)
}

func (r *UsersToStationsRepository) UpsertMany(
	stationId uuid.UUID,
	userIds []uuid.UUID,
	permissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]schemas.UsersToStations, *cexceptions.Exception) {
	if len(userIds) != len(permissions) {
		return nil, r.exceptions.InvalidInput("userIds and permissions must have equal lengths")
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

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
		return nil, r.exceptions.FailedToUpdate().WithOrigin(result.Error)
	}

	return relations, nil
}

func (r *UsersToStationsRepository) UpdateOne(
	stationId uuid.UUID,
	userId uuid.UUID,
	permission cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) (*schemas.UsersToStations, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

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
		return nil, r.exceptions.FailedToUpdate().WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, r.exceptions.NotFound()
	}

	return r.GetOne(
		stationId,
		userId,
		WithDB(parsedOptions.DB),
	)
}

func (r *UsersToStationsRepository) DeleteOne(
	stationId uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	result := parsedOptions.DB.
		Where("station_id = ? AND user_id = ?", stationId, userId).
		Delete(&schemas.UsersToStations{})
	if result.Error != nil {
		return r.exceptions.FailedToDelete().WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 {
		return r.exceptions.NotFound()
	}

	return nil
}

func (r *UsersToStationsRepository) DeleteMany(
	stationId uuid.UUID,
	userIds []uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.
		Where("station_id = ? AND user_id IN ?", stationId, userIds).
		Delete(&schemas.UsersToStations{})
	if result.Error != nil {
		return r.exceptions.FailedToDelete().WithOrigin(result.Error)
	}
	if result.RowsAffected != int64(len(userIds)) {
		return r.exceptions.NotFound()
	}

	return nil
}

func (r *UsersToStationsRepository) DeleteManyByStationIdsAndUserId(
	stationIds []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.
		Where("station_id IN ? AND user_id = ?", stationIds, userId).
		Delete(&schemas.UsersToStations{})
	if result.Error != nil {
		return r.exceptions.FailedToDelete().WithOrigin(result.Error)
	}
	if result.RowsAffected != int64(len(stationIds)) {
		return r.exceptions.NotFound()
	}

	return nil
}
