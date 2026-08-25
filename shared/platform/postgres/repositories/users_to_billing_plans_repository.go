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

type UsersToBillingPlansRepositoryInterface interface {
	GetOnyById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) (*schemas.UsersToBillingPlans, *cexceptions.Exception)
	GetAllByUserId(userId uuid.UUID, opts ...RepositoryOptions) ([]schemas.UsersToBillingPlans, *cexceptions.Exception)
	CreateOne(userId uuid.UUID, input inputs.CreateUsersToBillingPlansInput, opts ...RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateUsersToBillingPlansInput, opts ...RepositoryOptions) (*schemas.UsersToBillingPlans, *cexceptions.Exception)
	DeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	DeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
}

type UsersToBillingPlansRepository struct {
	db         *gorm.DB
	exceptions exceptions.UsersToBillingPlansException
}

func NewUsersToBillingPlansRepository(db *gorm.DB) UsersToBillingPlansRepositoryInterface {
	return &UsersToBillingPlansRepository{
		db: db, exceptions: exceptions.NewUsersToBillingPlansException()}
}

func (r *UsersToBillingPlansRepository) GetOnyById(
	id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions,
) (*schemas.UsersToBillingPlans, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var usersToBillingPlans schemas.UsersToBillingPlans
	result := parsedOptions.DB.Model(&schemas.UsersToBillingPlans{}).
		Where("id = ? and user_id = ?", id, userId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&usersToBillingPlans)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: usersToBillingPlans.Id == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &usersToBillingPlans, nil
}

func (r *UsersToBillingPlansRepository) GetAllByUserId(
	userId uuid.UUID, opts ...RepositoryOptions,
) ([]schemas.UsersToBillingPlans, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var usersToBillingPlans []schemas.UsersToBillingPlans
	result := parsedOptions.DB.Model(&schemas.UsersToBillingPlans{}).
		Where("user_id = ?", userId).
		Find(&usersToBillingPlans)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: len(usersToBillingPlans) == 0, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return usersToBillingPlans, nil
}

func (r *UsersToBillingPlansRepository) CreateOne(
	userId uuid.UUID,
	input inputs.CreateUsersToBillingPlansInput,
	opts ...RepositoryOptions,
) (*uuid.UUID, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var newUsersToBillingPlans schemas.UsersToBillingPlans
	newUsersToBillingPlans.UserId = userId

	if err := copier.Copy(&newUsersToBillingPlans, &input); err != nil {
		return nil, r.exceptions.FailedToCreate().WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.UsersToBillingPlans{}).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Create(&newUsersToBillingPlans)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToCreate().WithOrigin(result.Error)},
		{First: newUsersToBillingPlans.Id == uuid.Nil, Second: r.exceptions.FailedToCreate()},
	}); exception != nil {
		return nil, exception
	}

	return &newUsersToBillingPlans.Id, nil
}

func (r *UsersToBillingPlansRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateUsersToBillingPlansInput,
	opts ...RepositoryOptions,
) (*schemas.UsersToBillingPlans, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	existingUsersToBillingPlans, exception := r.GetOnyById(
		id,
		userId,
		opts...,
	)
	if exception := cexceptions.Cover(exception, []cexceptions.Pair{
		{First: existingUsersToBillingPlans == nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingUsersToBillingPlans)
	if err != nil {
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true)
	}

	result := parsedOptions.DB.Model(&schemas.UsersToBillingPlans{}).
		Where("id = ? and user_id = ?", id, userId).
		Select("*").
		Updates(&updates)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &updates, nil
}

func (r *UsersToBillingPlansRepository) DeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.Model(&schemas.UsersToBillingPlans{}).
		Where("id = ? and user_id = ?", id, userId).
		Delete(&schemas.UsersToBillingPlans{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *UsersToBillingPlansRepository) DeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	if len(ids) == 0 {
		return r.exceptions.NoChanges()
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.Model(&schemas.UsersToBillingPlans{}).
		Where("ids IN ? and user_id = ?", ids, userId).
		Delete(&schemas.UsersToBillingPlans{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}
