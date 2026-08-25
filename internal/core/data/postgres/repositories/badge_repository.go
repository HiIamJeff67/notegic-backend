package repositories

import (
	"github.com/google/uuid"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type BadgeRepositoryInterface interface {
	GetOneById(id uuid.UUID, preloads []schemas.BadgeRelation, opts ...options.RepositoryOptions) (*schemas.Badge, *cexceptions.Exception)
}

type BadgeRepository struct{}

func NewBadgeRepository() BadgeRepositoryInterface {
	return &BadgeRepository{}
}

func (r *BadgeRepository) GetOneById(
	id uuid.UUID,
	preloads []schemas.BadgeRelation,
	opts ...options.RepositoryOptions,
) (*schemas.Badge, *cexceptions.Exception) {
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
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewBadgeException().NotFound().WithOrigin(result.Error)},
		{First: badge.Id == uuid.Nil, Second: apiexceptions.NewBadgeException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &badge, nil
}
