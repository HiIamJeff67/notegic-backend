package repositories

import (
	"gorm.io/gorm"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	exceptions "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type UserViewRepositoryInterface interface {
	GetOneByPublicId(input inputs.GetUserViewByPublicIdInput, opts RepositoryOptionFields) (*schemas.UserView, *cexceptions.Exception)
}

type UserViewRepository struct {
	exceptions exceptions.UserViewException
}

func NewUserViewRepository(repositoryExceptions ...exceptions.UserViewException) UserViewRepositoryInterface {
	repositoryException := exceptions.NewUserViewException()
	if len(repositoryExceptions) > 0 {
		repositoryException = repositoryExceptions[0]
	}

	return &UserViewRepository{exceptions: repositoryException}
}

func (r *UserViewRepository) GetOneByPublicId(
	input inputs.GetUserViewByPublicIdInput,
	parsedOptions RepositoryOptionFields,
) (*schemas.UserView, *cexceptions.Exception) {
	if parsedOptions.DB == nil {
		return nil, r.exceptions.DatabaseUnavailable()
	}

	userView := schemas.UserView{}
	result := parsedOptions.DB.Model(&schemas.UserView{}).
		Scopes(scopes.UserViewByPublicId(input.PublicId)).
		First(&userView)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, r.exceptions.NotFoundByPublicId()
		}
		return nil, r.exceptions.FailedToGetByPublicId().WithOrigin(result.Error)
	}

	return &userView, nil
}
