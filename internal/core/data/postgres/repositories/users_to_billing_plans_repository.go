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

type UsersToBillingPlansRepositoryInterface interface {
	GetOnyById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) (*schemas.UsersToBillingPlans, *exceptions.Exception)
	GetAllByUserId(userId uuid.UUID, opts ...options.RepositoryOptions) ([]schemas.UsersToBillingPlans, *exceptions.Exception)
	CreateOne(userId uuid.UUID, input inputs.CreateUsersToBillingPlansInput, opts ...options.RepositoryOptions) (*uuid.UUID, *exceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateUsersToBillingPlansInput, opts ...options.RepositoryOptions) (*schemas.UsersToBillingPlans, *exceptions.Exception)
	DeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *exceptions.Exception
	DeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *exceptions.Exception
}

type UsersToBillingPlansRepository struct{}

func NewUsersToBillingPlansRepository() UsersToBillingPlansRepositoryInterface {
	return &UsersToBillingPlansRepository{}
}

func (r *UsersToBillingPlansRepository) GetOnyById(
	id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions,
) (*schemas.UsersToBillingPlans, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var usersToBillingPlans schemas.UsersToBillingPlans
	result := parsedOptions.DB.Model(&schemas.UsersToBillingPlans{}).
		Where("id = ? and user_id = ?", id, userId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&usersToBillingPlans)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewUsersToBillingPlansException().NotFound().WithOrigin(result.Error)},
		{First: usersToBillingPlans.Id == uuid.Nil, Second: apiexceptions.NewUsersToBillingPlansException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &usersToBillingPlans, nil
}

func (r *UsersToBillingPlansRepository) GetAllByUserId(
	userId uuid.UUID, opts ...options.RepositoryOptions,
) ([]schemas.UsersToBillingPlans, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var usersToBillingPlans []schemas.UsersToBillingPlans
	result := parsedOptions.DB.Model(&schemas.UsersToBillingPlans{}).
		Where("user_id = ?", userId).
		Find(&usersToBillingPlans)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewUsersToBillingPlansException().NotFound().WithOrigin(result.Error)},
		{First: len(usersToBillingPlans) == 0, Second: apiexceptions.NewUsersToBillingPlansException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return usersToBillingPlans, nil
}

func (r *UsersToBillingPlansRepository) CreateOne(
	userId uuid.UUID,
	input inputs.CreateUsersToBillingPlansInput,
	opts ...options.RepositoryOptions,
) (*uuid.UUID, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var newUsersToBillingPlans schemas.UsersToBillingPlans
	newUsersToBillingPlans.UserId = userId

	if err := copier.Copy(&newUsersToBillingPlans, &input); err != nil {
		return nil, apiexceptions.NewUsersToBillingPlansException().FailedToCreate().WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.UsersToBillingPlans{}).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Create(&newUsersToBillingPlans)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewUsersToBillingPlansException().FailedToCreate().WithOrigin(result.Error)},
		{First: newUsersToBillingPlans.Id == uuid.Nil, Second: apiexceptions.NewUsersToBillingPlansException().FailedToCreate()},
	}); exception != nil {
		return nil, exception
	}

	return &newUsersToBillingPlans.Id, nil
}

func (r *UsersToBillingPlansRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateUsersToBillingPlansInput,
	opts ...options.RepositoryOptions,
) (*schemas.UsersToBillingPlans, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	existingUsersToBillingPlans, exception := r.GetOnyById(
		id,
		userId,
		opts...,
	)
	if exception := exceptions.Cover(exception, []exceptions.Pair{
		{First: existingUsersToBillingPlans == nil, Second: apiexceptions.NewUsersToBillingPlansException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingUsersToBillingPlans)
	if err != nil {
		return nil, exceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true)
	}

	result := parsedOptions.DB.Model(&schemas.UsersToBillingPlans{}).
		Where("id = ? and user_id = ?", id, userId).
		Select("*").
		Updates(&updates)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewUsersToBillingPlansException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewUsersToBillingPlansException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &updates, nil
}

func (r *UsersToBillingPlansRepository) DeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.UsersToBillingPlans{}).
		Where("id = ? and user_id = ?", id, userId).
		Delete(&schemas.UsersToBillingPlans{})
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewUsersToBillingPlansException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewUsersToBillingPlansException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *UsersToBillingPlansRepository) DeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *exceptions.Exception {
	if len(ids) == 0 {
		return apiexceptions.NewUsersToBillingPlansException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.UsersToBillingPlans{}).
		Where("ids IN ? and user_id = ?", ids, userId).
		Delete(&schemas.UsersToBillingPlans{})
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewUsersToBillingPlansException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewUsersToBillingPlansException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}
