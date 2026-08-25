package repositories

import (
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"
	"net/http"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm/clause"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"

	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type UserSettingRepositoryInterface interface {
	GetOneByUserId(userId uuid.UUID, opts ...options.RepositoryOptions) (*schemas.UserSetting, *cexceptions.Exception)
	CreateOneByUserId(userId uuid.UUID, input inputs.CreateUserSettingInput, opts ...options.RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	UpdateOneByUserId(userId uuid.UUID, input inputs.PartialUpdateUserSettingInput, opts ...options.RepositoryOptions) (*schemas.UserSetting, *cexceptions.Exception)
}

type UserSettingRepository struct{}

func NewUserSettingRepository() UserSettingRepositoryInterface {
	return &UserSettingRepository{}
}

func (r *UserSettingRepository) GetOneByUserId(
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) (*schemas.UserSetting, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var userSetting schemas.UserSetting
	result := parsedOptions.DB.Model(&schemas.UserSetting{}).
		Where("user_id = ?", userId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&userSetting)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
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
) (*uuid.UUID, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var newUserSetting schemas.UserSetting
	newUserSetting.UserId = userId
	if err := copier.Copy(&newUserSetting, &input); err != nil {
		return nil, apiexceptions.NewUserSettingException().FailedToCreate().WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.UserSetting{}).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Create(&newUserSetting)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
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
) (*schemas.UserSetting, *cexceptions.Exception) {
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
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true)
	}

	result := parsedOptions.DB.Model(&schemas.UserSetting{}).
		Where("user_id = ?").
		Select("*").
		Updates(&updates)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewUserSettingException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewUserSettingException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &updates, nil
}
