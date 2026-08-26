package repositories

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptions "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type UserProjectionRepositoryInterface interface {
	CreateIfNotExists(input inputs.CreateUserProjectionInput, opts ...RepositoryOptions) *cexceptions.Exception
}

type UserProjectionRepository struct {
	db         *gorm.DB
	exceptions exceptions.UserProjectionException
}

func NewUserProjectionRepository(db *gorm.DB) UserProjectionRepositoryInterface {
	return &UserProjectionRepository{
		db:         db,
		exceptions: exceptions.NewUserProjectionException(),
	}
}

func (r *UserProjectionRepository) CreateIfNotExists(
	input inputs.CreateUserProjectionInput,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)
	if parsedOptions.DB == nil || input.PublicId == uuid.Nil {
		return r.exceptions.FailedToCreate()
	}

	projection := schemas.UserProjection{
		Id:       uuid.New(),
		PublicId: input.PublicId,
		Plan:     input.Plan,
		Status:   input.Status,
	}
	result := parsedOptions.DB.Model(&schemas.UserProjection{}).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&projection)
	if result.Error != nil {
		return r.exceptions.FailedToCreate().WithOrigin(result.Error)
	}

	return nil
}
