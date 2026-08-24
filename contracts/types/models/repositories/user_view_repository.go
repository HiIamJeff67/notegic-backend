package repositories

import (
	"net/http"

	"gorm.io/gorm"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	models "github.com/HiIamJeff67/notegic-backend/contracts/types/models"
	inputs "github.com/HiIamJeff67/notegic-backend/contracts/types/models/inputs"
	scopes "github.com/HiIamJeff67/notegic-backend/contracts/types/models/scopes"
)

type UserViewRepositoryInterface interface {
	GetOneByPublicId(input inputs.GetUserViewByPublicIdInput, opts RepositoryOptionFields) (*models.UserView, *exceptions.Exception)
}

type UserViewRepository struct{}

func NewUserViewRepository() UserViewRepositoryInterface {
	return &UserViewRepository{}
}

func (r *UserViewRepository) GetOneByPublicId(
	input inputs.GetUserViewByPublicIdInput,
	parsedOptions RepositoryOptionFields,
) (*models.UserView, *exceptions.Exception) {
	if parsedOptions.DB == nil {
		return nil, exceptions.New(
			"DatabaseUnavailable",
			"UserView",
			"GetOneByPublicId",
			"A database connection is required",
			http.StatusInternalServerError,
			true,
		)
	}

	userView := models.UserView{}
	result := parsedOptions.DB.Model(&models.UserView{}).
		Scopes(scopes.UserViewByPublicId(input.PublicId)).
		First(&userView)
	if result.Error != nil {
		if result.Error == gorm.ErrRecordNotFound {
			return nil, exceptions.New("NotFound", "UserView", "GetOneByPublicId", "User view not found", http.StatusNotFound)
		}
		return nil, exceptions.New(
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
