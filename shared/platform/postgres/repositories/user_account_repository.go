package repositories

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"
	exceptions "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type UserAccountRepositoryInterface interface {
	GetOneByUserId(userId uuid.UUID, opts ...RepositoryOptions) (*schemas.UserAccount, *cexceptions.Exception)
	CreateOneByUserId(userId uuid.UUID, input inputs.CreateUserAccountInput, opts ...RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	UpdateOneByUserId(userId uuid.UUID, input inputs.PartialUpdateUserAccountInput, opts ...RepositoryOptions) (*schemas.UserAccount, *cexceptions.Exception)
}

type UserAccountRepository struct {
	db         *gorm.DB
	exceptions exceptions.UserAccountException
}

func NewUserAccountRepository(db *gorm.DB) UserAccountRepositoryInterface {
	return &UserAccountRepository{
		db: db, exceptions: exceptions.NewUserAccountException()}
}

func (r *UserAccountRepository) GetOneByUserId(
	userId uuid.UUID,
	opts ...RepositoryOptions,
) (*schemas.UserAccount, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var userAccount schemas.UserAccount
	result := parsedOptions.DB.Model(&schemas.UserAccount{}).
		Where("user_id = ?", userId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&userAccount)
	if err := result.Error; err != nil {
		return nil, r.exceptions.NotFound().WithOrigin(err)
	}

	return &userAccount, nil
}

func (r *UserAccountRepository) CreateOneByUserId(
	userId uuid.UUID,
	input inputs.CreateUserAccountInput,
	opts ...RepositoryOptions,
) (*uuid.UUID, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var newUserAccount schemas.UserAccount
	newUserAccount.UserId = userId

	if err := copier.Copy(&newUserAccount, &input); err != nil {
		return nil, r.exceptions.FailedToCreate().WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.UserAccount{}).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Create(&newUserAccount)
	if err := result.Error; err != nil {
		return nil, r.exceptions.FailedToCreate().WithOrigin(err)
	}

	return &newUserAccount.Id, nil
}

func (r *UserAccountRepository) UpdateOneByUserId(
	userId uuid.UUID,
	input inputs.PartialUpdateUserAccountInput,
	opts ...RepositoryOptions,
) (*schemas.UserAccount, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	existingUserAccount, exception := r.GetOneByUserId(
		userId,
		opts...,
	)
	if exception = cexceptions.Cover(exception, []cexceptions.Pair{
		{First: existingUserAccount == nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingUserAccount)
	if err != nil {
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true)
	}

	result := parsedOptions.DB.Model(&schemas.UserAccount{}).
		Where("user_id = ?", userId).
		Select("*").
		Updates(&updates)
	if err := result.Error; err != nil {
		return nil, r.exceptions.FailedToUpdate().WithOrigin(err)
	}
	if result.RowsAffected == 0 {
		return nil, r.exceptions.NoChanges()
	}

	return &updates, nil
}

// We do not allow to just delete the userAccount,
// instead, the userAccount is only deleted by deleting the user
// func DeleteUserAccount(userId uuid.UUID) (deletedUserAccount User, err error) {}
