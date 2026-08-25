package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptions "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type BadgeRepositoryInterface interface {
	GetOneById(id uuid.UUID, preloads []schemas.BadgeRelation, opts ...RepositoryOptions) (*schemas.Badge, *cexceptions.Exception)
}

type BadgeRepository struct {
	db         *gorm.DB
	exceptions exceptions.BadgeException
}

func NewBadgeRepository(db *gorm.DB) BadgeRepositoryInterface {
	return &BadgeRepository{
		db: db, exceptions: exceptions.NewBadgeException()}
}

func (r *BadgeRepository) GetOneById(
	id uuid.UUID,
	preloads []schemas.BadgeRelation,
	opts ...RepositoryOptions,
) (*schemas.Badge, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	badge := schemas.Badge{}

	query := parsedOptions.DB.Model(&schemas.Badge{})
	if len(preloads) > 0 {
		for _, preload := range preloads {
			query = query.Preload(string(preload))
		}
	}

	result := query.Where("id = ?", id).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&badge)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: badge.Id == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &badge, nil
}
