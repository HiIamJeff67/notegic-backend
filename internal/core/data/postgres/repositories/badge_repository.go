package repositories

import (
	"github.com/google/uuid"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/scopes"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
)

type BadgeRepositoryInterface interface {
	GetOneById(id uuid.UUID, preloads []schemas.BadgeRelation, opts ...options.RepositoryOptions) (*schemas.Badge, *exceptions.Exception)
}

type BadgeRepository struct{}

func NewBadgeRepository() BadgeRepositoryInterface {
	return &BadgeRepository{}
}

func (r *BadgeRepository) GetOneById(
	id uuid.UUID,
	preloads []schemas.BadgeRelation,
	opts ...options.RepositoryOptions,
) (*schemas.Badge, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

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
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBadgeException().NotFound().WithOrigin(result.Error)},
		{First: badge.Id == uuid.Nil, Second: apiexceptions.NewBadgeException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &badge, nil
}
