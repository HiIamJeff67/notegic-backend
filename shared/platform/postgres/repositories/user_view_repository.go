package repositories

import (
	"net/http"

	"gorm.io/gorm"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type UserViewRepositoryInterface interface {
	GetOneByPublicId(input inputs.GetUserViewByPublicIdInput, opts RepositoryOptionFields) (*schemas.UserView, *cexceptions.Exception)
}

type UserViewRepository struct{}

func NewUserViewRepository() UserViewRepositoryInterface {
	return &UserViewRepository{}
}

func (r *UserViewRepository) GetOneByPublicId(
	input inputs.GetUserViewByPublicIdInput,
	parsedOptions RepositoryOptionFields,
) (*schemas.UserView, *cexceptions.Exception) {
	if parsedOptions.DB == nil {
		return nil, cexceptions.New(
			"DatabaseUnavailable",
			"UserView",
			"GetOneByPublicId",
			"A database connection is required",
			http.StatusInternalServerError,
			true,
		)
	}

	userView := schemas.UserView{}
	result := parsedOptions.DB.Model(&schemas.UserView{}).
		Scopes(scopes.UserViewByPublicId(input.PublicId)).
		First(&userView)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, cexceptions.New("NotFound", "UserView", "GetOneByPublicId", "User view not found", http.StatusNotFound)
		}
		return nil, cexceptions.New(
			"FailedToGet",
			"UserView",
			"GetOneByPublicId",
			"Failed to get user view",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	return &userView, nil
}
