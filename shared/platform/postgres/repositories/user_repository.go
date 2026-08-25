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

type UserRepositoryInterface interface {
	GetOneById(id uuid.UUID, preloads []schemas.UserRelation, opts ...RepositoryOptions) (*schemas.User, *cexceptions.Exception)
	GetOneByPublicId(publicId uuid.UUID, preloads []schemas.UserRelation, opts ...RepositoryOptions) (*schemas.User, *cexceptions.Exception)
	GetOneByName(name string, preloads []schemas.UserRelation, opts ...RepositoryOptions) (*schemas.User, *cexceptions.Exception)
	GetOneByEmail(email string, preloads []schemas.UserRelation, opts ...RepositoryOptions) (*schemas.User, *cexceptions.Exception)
	GetAll(opts ...RepositoryOptions) ([]schemas.User, *cexceptions.Exception)
	CreateOne(input inputs.CreateUserInput, opts ...RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	UpdateOneById(id uuid.UUID, input inputs.PartialUpdateUserInput, opts ...RepositoryOptions) (*schemas.User, *cexceptions.Exception)
}

type UserRepository struct {
	db         *gorm.DB
	exceptions exceptions.UserException
}

func NewUserRepository(db *gorm.DB) UserRepositoryInterface {
	return &UserRepository{
		db: db, exceptions: exceptions.NewUserException()}
}

func (r *UserRepository) GetOneById(
	id uuid.UUID,
	preloads []schemas.UserRelation,
	opts ...RepositoryOptions,
) (*schemas.User, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	user := schemas.User{}

	db := parsedOptions.DB.Model(&schemas.User{})
	if len(preloads) > 0 {
		for _, preload := range preloads {
			db = db.Preload(string(preload))
		}
	}

	result := db.Where("id = ?", id).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&user)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: user.Id == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &user, nil
}

func (r *UserRepository) GetOneByPublicId(
	publicId uuid.UUID,
	preloads []schemas.UserRelation,
	opts ...RepositoryOptions,
) (*schemas.User, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	user := schemas.User{}
	query := parsedOptions.DB.Model(&schemas.User{})
	for _, preload := range preloads {
		query = query.Preload(string(preload))
	}

	result := query.
		Where("public_id = ?", publicId).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&user)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: user.Id == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &user, nil
}

func (r *UserRepository) GetOneByName(
	name string,
	preloads []schemas.UserRelation,
	opts ...RepositoryOptions,
) (*schemas.User, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	user := schemas.User{}

	db := parsedOptions.DB.Model(&schemas.User{})
	if len(preloads) > 0 {
		for _, preload := range preloads {
			db = db.Preload(string(preload))
		}
	}

	result := db.Where("name = ?", name).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&user)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: user.Id == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &user, nil
}

func (r *UserRepository) GetOneByEmail(
	email string,
	preloads []schemas.UserRelation,
	opts ...RepositoryOptions,
) (*schemas.User, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	user := schemas.User{}

	query := parsedOptions.DB.Model(&schemas.User{})
	if len(preloads) > 0 {
		for _, preload := range preloads {
			query = query.Preload(string(preload))
		}
	}

	result := query.Where("email = ?", email).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&user)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: user.Id == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &user, nil
}

func (r *UserRepository) GetAll(
	opts ...RepositoryOptions,
) ([]schemas.User, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	users := []schemas.User{}

	result := parsedOptions.DB.Preload("UserInfo").
		Preload("UserAccount").
		Preload("UserSetting").
		Preload("Badges").
		Preload("Themes").
		Find(&users)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: len(users) == 0, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}
	return users, nil
}

func (r *UserRepository) CreateOne(
	input inputs.CreateUserInput,
	opts ...RepositoryOptions,
) (*uuid.UUID, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	// note that the create operation in gorm will NOT return anything
	// but the default value we set in gorm field in the above struct will be returned if we specified it in the "returning"
	var newUser schemas.User
	if err := copier.Copy(&newUser, &input); err != nil {
		return nil, r.exceptions.FailedToCreate().WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.User{}).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Create(&newUser)
	if err := result.Error; err != nil {
		// instead of using exceptions.Cover(), we can just get the error string and switch on it to return the corresponded exceptions
		// this approach is faster and more straight forward
		switch err.Error() {
		case "ERROR: duplicate key value violates unique constraint \"uni_UserTable_name\" (SQLSTATE 23505)":
			return nil, r.exceptions.DuplicateName(input.Name)
		case "ERROR: duplicate key value violates unique constraint \"uni_UserTable_email\" (SQLSTATE 23505)":
			return nil, r.exceptions.DuplicateEmail(input.Email)
		default:
			return nil, r.exceptions.FailedToCreate() // .WithOrigin(err) <- don't show the database error to outside
		}
	}
	if result.RowsAffected == 0 {
		// check the remaining condition here,
		// since there's only 1 more condition to check,
		// there's no need to use exceptions.Cover() to map all the it
		return nil, r.exceptions.NoChanges()
	}

	return &newUser.Id, nil
}

func (r *UserRepository) UpdateOneById(
	id uuid.UUID,
	input inputs.PartialUpdateUserInput,
	opts ...RepositoryOptions,
) (*schemas.User, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	existingUser, exception := r.GetOneById(
		id,
		nil,
		opts...,
	)
	if exception != nil || existingUser == nil {
		return nil, exception
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingUser)
	if err != nil {
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true)
	}

	result := parsedOptions.DB.Model(&schemas.User{}).
		Where("id = ?", id).
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
