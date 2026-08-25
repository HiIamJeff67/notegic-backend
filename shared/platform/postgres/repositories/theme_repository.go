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

type ThemeRepositoryInterface interface {
	GetOneById(id uuid.UUID, preloads []schemas.ThemeRelation, opts ...RepositoryOptions) (*schemas.Theme, *cexceptions.Exception)
	GetAll(opts ...RepositoryOptions) ([]schemas.Theme, *cexceptions.Exception)
	CreateOneByAuthorId(authorId uuid.UUID, input inputs.CreateThemeInput, opts ...RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	UpdateOneById(id uuid.UUID, authorId uuid.UUID, input inputs.PartialUpdateThemeInput, opts ...RepositoryOptions) (*schemas.Theme, *cexceptions.Exception)
	DeleteOneById(id uuid.UUID, authorId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
}

type ThemeRepository struct {
	db         *gorm.DB
	exceptions exceptions.ThemeException
}

func NewThemeRepository(db *gorm.DB) ThemeRepositoryInterface {
	return &ThemeRepository{
		db: db, exceptions: exceptions.NewThemeException()}
}

func (r *ThemeRepository) GetOneById(
	id uuid.UUID,
	preloads []schemas.ThemeRelation,
	opts ...RepositoryOptions,
) (*schemas.Theme, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var theme schemas.Theme

	query := parsedOptions.DB.Model(&schemas.Theme{})
	if len(preloads) > 0 {
		for _, preload := range preloads {
			query = query.Preload(string(preload))
		}
	}

	result := query.Where("id = ?", id).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&theme)
	if err := result.Error; err != nil {
		return nil, r.exceptions.NotFound().WithOrigin(err)
	}

	return &theme, nil
}

func (r *ThemeRepository) GetAll(
	opts ...RepositoryOptions,
) ([]schemas.Theme, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var themes []schemas.Theme
	result := parsedOptions.DB.Model(&schemas.Theme{}).
		Find(&themes)
	if err := result.Error; err != nil {
		return nil, r.exceptions.NotFound().WithOrigin(err)
	}

	return themes, nil
}

func (r *ThemeRepository) CreateOneByAuthorId(
	authorId uuid.UUID,
	input inputs.CreateThemeInput,
	opts ...RepositoryOptions,
) (*uuid.UUID, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var newTheme schemas.Theme
	newTheme.AuthorId = authorId

	if err := copier.Copy(&newTheme, &input); err != nil {
		return nil, r.exceptions.FailedToCreate().WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.Theme{}).
		Clauses(clause.Returning{Columns: []clause.Column{{Name: "id"}}}).
		Create(&newTheme)
	if err := result.Error; err != nil {
		return nil, r.exceptions.FailedToCreate().WithOrigin(err)
	}

	return &newTheme.Id, nil
}

func (r *ThemeRepository) UpdateOneById(
	id uuid.UUID,
	authorId uuid.UUID,
	input inputs.PartialUpdateThemeInput,
	opts ...RepositoryOptions,
) (*schemas.Theme, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	existingTheme, exception := r.GetOneById(
		id,
		nil,
		opts...,
	)
	if exception != nil || existingTheme == nil {
		return nil, exception
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingTheme)
	if err != nil {
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true)
	}

	result := parsedOptions.DB.Model(&schemas.Theme{}).
		Where("id = ? AND author_id = ?", id, authorId).
		Select("*").
		Updates(&updates)
	if err := result.Error; err != nil {
		return nil, r.exceptions.FailedToUpdate().WithOrigin(err)
	}
	if result.RowsAffected == 0 { // check if we do update it or not
		return nil, r.exceptions.NoChanges()
	}

	return &updates, nil
}

func (r *ThemeRepository) DeleteOneById(
	id uuid.UUID,
	authorId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	// * If you need to use the functionality of RETURNING from PostgreSQL
	// var deletedTheme schemas.Theme

	// result := r.db.Model(&schemas.Theme{}).
	// 	Where("id = ? AND author_id = ?", id, authorId).
	// 	Clauses(clause.Returning{}).
	// 	Delete(&deletedTheme)

	result := parsedOptions.DB.Model(&schemas.Theme{}).
		Where("id = ? AND author_id = ?", id, authorId).
		Delete(&schemas.Theme{})
	if err := result.Error; err != nil {
		return r.exceptions.FailedToDelete().WithOrigin(err)
	}
	if result.RowsAffected == 0 {
		return r.exceptions.NotFound()
	}

	return nil
}
