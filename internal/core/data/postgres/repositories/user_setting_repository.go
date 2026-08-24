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

type UserSettingRepositoryInterface interface {
	GetOneByUserId(userId uuid.UUID, opts ...options.RepositoryOptions) (*schemas.UserSetting, *exceptions.Exception)
	CreateOneByUserId(userId uuid.UUID, input inputs.CreateUserSettingInput, opts ...options.RepositoryOptions) (*uuid.UUID, *exceptions.Exception)
	UpdateOneByUserId(userId uuid.UUID, input inputs.PartialUpdateUserSettingInput, opts ...options.RepositoryOptions) (*schemas.UserSetting, *exceptions.Exception)
}

type UserSettingRepository struct{}

func NewUserSettingRepository() UserSettingRepositoryInterface {
	return &UserSettingRepository{}
}

func (r *UserSettingRepository) GetOneByUserId(
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) (*schemas.UserSetting, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var userSetting schemas.UserSetting
	result := parsedOptions.DB.Model(&schemas.UserSetting{}).
		Where("user_id = ?", userId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&userSetting)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewUserSettingException().NotFound().WithOrigin(result.Error)},
		{First: userSetting.Id == uuid.Nil, Second: apiexceptions.NewUserSettingException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &userSetting, nil
}

func (r *UserSettingRepository) CreateOneByUserId(
	userId uuid.UUID,
	input inputs.CreateUserSettingInput,
	opts ...options.RepositoryOptions,
) (*uuid.UUID, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var newUserSetting schemas.UserSetting
	newUserSetting.UserId = userId
	if err := copier.Copy(&newUserSetting, &input); err != nil {
		return nil, apiexceptions.NewUserSettingException().FailedToCreate().WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.UserSetting{}).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Create(&newUserSetting)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewUserSettingException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewUserSettingException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &newUserSetting.Id, nil
}

func (r *UserSettingRepository) UpdateOneByUserId(
	userId uuid.UUID,
	input inputs.PartialUpdateUserSettingInput,
	opts ...options.RepositoryOptions,
) (*schemas.UserSetting, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	existingUserSetting, exception := r.GetOneByUserId(
		userId,
		opts...,
	)
	if exception != nil || existingUserSetting == nil {
		return nil, exception
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingUserSetting)
	if err != nil {
		return nil, exceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true)
	}

	result := parsedOptions.DB.Model(&schemas.UserSetting{}).
		Where("user_id = ?").
		Select("*").
		Updates(&updates)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewUserSettingException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewUserSettingException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &updates, nil
}
