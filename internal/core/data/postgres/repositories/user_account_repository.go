package repositories

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm/clause"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"

	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/scopes"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
)

type UserAccountRepositoryInterface interface {
	GetOneByUserId(userId uuid.UUID, opts ...options.RepositoryOptions) (*schemas.UserAccount, *exceptions.Exception)
	CreateOneByUserId(userId uuid.UUID, input inputs.CreateUserAccountInput, opts ...options.RepositoryOptions) (*uuid.UUID, *exceptions.Exception)
	UpdateOneByUserId(userId uuid.UUID, input inputs.PartialUpdateUserAccountInput, opts ...options.RepositoryOptions) (*schemas.UserAccount, *exceptions.Exception)
}

type UserAccountRepository struct{}

func NewUserAccountRepository() UserAccountRepositoryInterface {
	return &UserAccountRepository{}
}

func (r *UserAccountRepository) GetOneByUserId(
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) (*schemas.UserAccount, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var userAccount schemas.UserAccount
	result := parsedOptions.DB.Model(&schemas.UserAccount{}).
		Where("user_id = ?", userId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&userAccount)
	if err := result.Error; err != nil {
		return nil, apiexceptions.NewUserAccountException().NotFound().WithOrigin(err)
	}

	return &userAccount, nil
}

func (r *UserAccountRepository) CreateOneByUserId(
	userId uuid.UUID,
	input inputs.CreateUserAccountInput,
	opts ...options.RepositoryOptions,
) (*uuid.UUID, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var newUserAccount schemas.UserAccount
	newUserAccount.UserId = userId

	if err := copier.Copy(&newUserAccount, &input); err != nil {
		return nil, apiexceptions.NewUserAccountException().FailedToCreate().WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.UserAccount{}).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Create(&newUserAccount)
	if err := result.Error; err != nil {
		return nil, apiexceptions.NewUserAccountException().FailedToCreate().WithOrigin(err)
	}

	return &newUserAccount.Id, nil
}

func (r *UserAccountRepository) UpdateOneByUserId(
	userId uuid.UUID,
	input inputs.PartialUpdateUserAccountInput,
	opts ...options.RepositoryOptions,
) (*schemas.UserAccount, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	existingUserAccount, exception := r.GetOneByUserId(
		userId,
		opts...,
	)
	if exception = exceptions.Cover(exception, []exceptions.Pair{
		{First: existingUserAccount == nil, Second: apiexceptions.NewUserAccountException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingUserAccount)
	if err != nil {
		return nil, exceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true)
	}

	result := parsedOptions.DB.Model(&schemas.UserAccount{}).
		Where("user_id = ?", userId).
		Select("*").
		Updates(&updates)
	if err := result.Error; err != nil {
		return nil, apiexceptions.NewUserAccountException().FailedToUpdate().WithOrigin(err)
	}
	if result.RowsAffected == 0 {
		return nil, apiexceptions.NewUserAccountException().NoChanges()
	}

	return &updates, nil
}

// We do not allow to just delete the userAccount,
// instead, the userAccount is only deleted by deleting the user
// func DeleteUserAccount(userId uuid.UUID) (deletedUserAccount User, err error) {}
