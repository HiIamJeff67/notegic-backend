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

type UsersToShelvesRepositoryInterface interface {
	GetOne(rootShelfId uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) (*schemas.UsersToShelves, *cexceptions.Exception)
	GetMany(rootShelfId uuid.UUID, userIds []uuid.UUID, opts ...RepositoryOptions) ([]schemas.UsersToShelves, *cexceptions.Exception)
	GetManyByRootShelfIdsAndUserId(rootShelfIds []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) ([]schemas.UsersToShelves, *cexceptions.Exception)
	CreateOne(rootShelfId uuid.UUID, userId uuid.UUID, permission cenums.AccessControlPermission, opts ...RepositoryOptions) (*schemas.UsersToShelves, *cexceptions.Exception)
	UpsertMany(rootShelfId uuid.UUID, userIds []uuid.UUID, permissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]schemas.UsersToShelves, *cexceptions.Exception)
	UpdateOne(rootShelfId uuid.UUID, userId uuid.UUID, permission cenums.AccessControlPermission, opts ...RepositoryOptions) (*schemas.UsersToShelves, *cexceptions.Exception)
	DeleteOne(rootShelfId uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	DeleteMany(rootShelfId uuid.UUID, userIds []uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	DeleteManyByRootShelfIdsAndUserId(rootShelfIds []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
}

type UsersToShelvesRepository struct {
	db         *gorm.DB
	exceptions exceptions.ShelfException
}

func NewUsersToShelvesRepository(db *gorm.DB) UsersToShelvesRepositoryInterface {
	return &UsersToShelvesRepository{
		db: db, exceptions: exceptions.NewShelfException()}
}

func (r *UsersToShelvesRepository) GetOne(
	rootShelfId uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) (*schemas.UsersToShelves, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var relation schemas.UsersToShelves
	result := parsedOptions.DB.
		Model(&schemas.UsersToShelves{}).
		Preload(string(schemas.UsersToShelvesRelation_User)).
		Where("root_shelf_id = ? AND user_id = ?", rootShelfId, userId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&relation)
	if result.Error != nil {
		return nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	return &relation, nil
}

func (r *UsersToShelvesRepository) GetMany(
	rootShelfId uuid.UUID,
	userIds []uuid.UUID,
	opts ...RepositoryOptions,
) ([]schemas.UsersToShelves, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var relations []schemas.UsersToShelves
	result := parsedOptions.DB.
		Model(&schemas.UsersToShelves{}).
		Where("root_shelf_id = ? AND user_id IN ?", rootShelfId, userIds).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&relations)
	if result.Error != nil {
		return nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	return relations, nil
}

func (r *UsersToShelvesRepository) GetManyByRootShelfIdsAndUserId(
	rootShelfIds []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) ([]schemas.UsersToShelves, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var relations []schemas.UsersToShelves
	result := parsedOptions.DB.
		Model(&schemas.UsersToShelves{}).
		Where("root_shelf_id IN ? AND user_id = ?", rootShelfIds, userId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&relations)
	if result.Error != nil {
		return nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	return relations, nil
}

func (r *UsersToShelvesRepository) CreateOne(
	rootShelfId uuid.UUID,
	userId uuid.UUID,
	permission cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) (*schemas.UsersToShelves, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	relation := schemas.UsersToShelves{
		RootShelfId: rootShelfId,
		UserId:      userId,
		Permission:  permission,
	}
	result := parsedOptions.DB.Create(&relation)
	if result.Error != nil {
		return nil, r.exceptions.FailedToCreate().WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 {
		return nil, r.exceptions.NoChanges()
	}

	return r.GetOne(
		rootShelfId,
		userId,
		WithDB(parsedOptions.DB),
	)
}

func (r *UsersToShelvesRepository) UpsertMany(
	rootShelfId uuid.UUID,
	userIds []uuid.UUID,
	permissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]schemas.UsersToShelves, *cexceptions.Exception) {
	if len(userIds) != len(permissions) {
		return nil, r.exceptions.InvalidInput("userIds and permissions must have equal lengths")
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	relations := make([]schemas.UsersToShelves, len(userIds))
	for index, userId := range userIds {
		relations[index] = schemas.UsersToShelves{
			RootShelfId: rootShelfId,
			UserId:      userId,
			Permission:  permissions[index],
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
						Name: "root_shelf_id",
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

func (r *UsersToShelvesRepository) UpdateOne(
	rootShelfId uuid.UUID,
	userId uuid.UUID,
	permission cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) (*schemas.UsersToShelves, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	relation := schemas.UsersToShelves{
		RootShelfId: rootShelfId,
		UserId:      userId,
		Permission:  permission,
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
		rootShelfId,
		userId,
		WithDB(parsedOptions.DB),
	)
}

func (r *UsersToShelvesRepository) DeleteOne(
	rootShelfId uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	result := parsedOptions.DB.
		Where("root_shelf_id = ? AND user_id = ?", rootShelfId, userId).
		Delete(&schemas.UsersToShelves{})
	if result.Error != nil {
		return r.exceptions.FailedToDelete().WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 {
		return r.exceptions.NotFound()
	}

	return nil
}

func (r *UsersToShelvesRepository) DeleteMany(
	rootShelfId uuid.UUID,
	userIds []uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.
		Where("root_shelf_id = ? AND user_id IN ?", rootShelfId, userIds).
		Delete(&schemas.UsersToShelves{})
	if result.Error != nil {
		return r.exceptions.FailedToDelete().WithOrigin(result.Error)
	}
	if result.RowsAffected != int64(len(userIds)) {
		return r.exceptions.NotFound()
	}

	return nil
}

func (r *UsersToShelvesRepository) DeleteManyByRootShelfIdsAndUserId(
	rootShelfIds []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.
		Where("root_shelf_id IN ? AND user_id = ?", rootShelfIds, userId).
		Delete(&schemas.UsersToShelves{})
	if result.Error != nil {
		return r.exceptions.FailedToDelete().WithOrigin(result.Error)
	}
	if result.RowsAffected != int64(len(rootShelfIds)) {
		return r.exceptions.NotFound()
	}

	return nil
}
