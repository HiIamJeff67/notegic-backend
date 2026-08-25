package repositories

import (
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"
	"net/http"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm/clause"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	platformschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"

	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type UserAccountRepositoryInterface interface {
	GetOneByUserId(userId uuid.UUID, opts ...options.RepositoryOptions) (*platformschemas.UserAccount, *cexceptions.Exception)
	CreateOneByUserId(userId uuid.UUID, input inputs.CreateUserAccountInput, opts ...options.RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	UpdateOneByUserId(userId uuid.UUID, input inputs.PartialUpdateUserAccountInput, opts ...options.RepositoryOptions) (*platformschemas.UserAccount, *cexceptions.Exception)
}

type UserAccountRepository struct{}

func NewUserAccountRepository() UserAccountRepositoryInterface {
	return &UserAccountRepository{}
}

func (r *UserAccountRepository) GetOneByUserId(
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) (*platformschemas.UserAccount, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var userAccount platformschemas.UserAccount
	result := parsedOptions.DB.Model(&platformschemas.UserAccount{}).
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
) (*uuid.UUID, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var newUserAccount platformschemas.UserAccount
	newUserAccount.UserId = userId

	if err := copier.Copy(&newUserAccount, &input); err != nil {
		return nil, apiexceptions.NewUserAccountException().FailedToCreate().WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&platformschemas.UserAccount{}).
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
) (*platformschemas.UserAccount, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	existingUserAccount, exception := r.GetOneByUserId(
		userId,
		opts...,
	)
	if exception = cexceptions.Cover(exception, []cexceptions.Pair{
		{First: existingUserAccount == nil, Second: apiexceptions.NewUserAccountException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingUserAccount)
	if err != nil {
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true)
	}

	result := parsedOptions.DB.Model(&platformschemas.UserAccount{}).
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
