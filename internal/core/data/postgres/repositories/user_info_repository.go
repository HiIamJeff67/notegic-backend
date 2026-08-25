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

type UserInfoRepositoryInterface interface {
	GetOneByUserId(userId uuid.UUID, opts ...options.RepositoryOptions) (*schemas.UserInfo, *cexceptions.Exception)
	CreateOneByUserId(userId uuid.UUID, input inputs.CreateUserInfoInput, opts ...options.RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	UpdateOneByUserId(userId uuid.UUID, input inputs.PartialUpdateUserInfoInput, opts ...options.RepositoryOptions) (*schemas.UserInfo, *cexceptions.Exception)
}

type UserInfoRepository struct{}

func NewUserInfoRepository() UserInfoRepositoryInterface {
	return &UserInfoRepository{}
}

func (r *UserInfoRepository) GetOneByUserId(
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) (*schemas.UserInfo, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	userInfo := schemas.UserInfo{}
	result := parsedOptions.DB.Model(&schemas.UserInfo{}).
		Where("user_id = ?", userId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&userInfo)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewUserInfoException().NotFound().WithOrigin(result.Error)},
		{First: userInfo.Id == uuid.Nil, Second: apiexceptions.NewUserInfoException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &userInfo, nil
}

func (r *UserInfoRepository) CreateOneByUserId(
	userId uuid.UUID,
	input inputs.CreateUserInfoInput,
	opts ...options.RepositoryOptions,
) (*uuid.UUID, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var newUserInfo schemas.UserInfo
	newUserInfo.UserId = userId
	if err := copier.Copy(&newUserInfo, &input); err != nil {
		return nil, apiexceptions.NewUserInfoException().FailedToCreate().WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.UserInfo{}).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Create(&newUserInfo)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewUserInfoException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewUserInfoException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &newUserInfo.Id, nil
}

func (r *UserInfoRepository) UpdateOneByUserId(
	userId uuid.UUID,
	input inputs.PartialUpdateUserInfoInput,
	opts ...options.RepositoryOptions,
) (*schemas.UserInfo, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	existingUserInfo, exception := r.GetOneByUserId(
		userId,
		opts...,
	)
	if exception != nil || existingUserInfo == nil {
		return nil, exception
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingUserInfo)
	if err != nil {
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true)
	}

	result := parsedOptions.DB.Model(&schemas.UserInfo{}).
		Where("user_id = ?", userId).
		Select("*").
		Updates(&updates)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewUserInfoException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewUserInfoException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &updates, nil
}
