package repositories

import (
	"fmt"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"
	"net/http"
	"strings"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"

	corescopes "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/scopes"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type RoutineTagRepositoryInterface interface {
	GetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTagRelation, opts ...options.RepositoryOptions) (*schemas.RoutineTag, *cexceptions.Exception)
	GetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.RoutineTagRelation, opts ...options.RepositoryOptions) ([]schemas.RoutineTag, *cexceptions.Exception)
	GetAllByUserId(userId uuid.UUID, preloads []schemas.RoutineTagRelation, opts ...options.RepositoryOptions) ([]schemas.RoutineTag, *cexceptions.Exception)
	CreateOne(userId uuid.UUID, input inputs.CreateRoutineTagInput, opts ...options.RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	CreateMany(userId uuid.UUID, input []inputs.CreateRoutineTagInput, opts ...options.RepositoryOptions) ([]uuid.UUID, *cexceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateRoutineTagInput, opts ...options.RepositoryOptions) (*schemas.RoutineTag, *cexceptions.Exception)
	UpdateManyByIds(userId uuid.UUID, input []inputs.UpdateRoutineTagByIdInput, opts ...options.RepositoryOptions) *cexceptions.Exception
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
}

type RoutineTagRepository struct {
	routineTagScope corescopes.RoutineTagScopeInterface
}

func NewRoutineTagRepository(routineTagScope corescopes.RoutineTagScopeInterface) RoutineTagRepositoryInterface {
	return &RoutineTagRepository{
		routineTagScope: routineTagScope,
	}
}

func (r *RoutineTagRepository) GetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineTagRelation,
	opts ...options.RepositoryOptions,
) (*schemas.RoutineTag, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var routineTag schemas.RoutineTag
	result := parsedOptions.DB.
		Model(&schemas.RoutineTag{}).
		Where(`"RoutineTagTable".id = ? AND "RoutineTagTable".owner_id = ?`, id, userId).
		Scopes(r.routineTagScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&routineTag)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTagException().NotFound().WithOrigin(result.Error)},
		{First: routineTag.Id == uuid.Nil, Second: apiexceptions.NewRoutineTagException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &routineTag, nil
}

func (r *RoutineTagRepository) GetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.RoutineTagRelation,
	opts ...options.RepositoryOptions,
) ([]schemas.RoutineTag, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var routineTags []schemas.RoutineTag
	result := parsedOptions.DB.
		Model(&schemas.RoutineTag{}).
		Where(`"RoutineTagTable".id IN ? AND "RoutineTagTable".owner_id = ?`, ids, userId).
		Scopes(r.routineTagScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&routineTags)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTagException().NotFound().WithOrigin(result.Error)},
		{First: len(routineTags) == 0, Second: apiexceptions.NewRoutineTagException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return routineTags, nil
}

func (r *RoutineTagRepository) GetAllByUserId(
	userId uuid.UUID,
	preloads []schemas.RoutineTagRelation,
	opts ...options.RepositoryOptions,
) ([]schemas.RoutineTag, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var routineTags []schemas.RoutineTag
	result := parsedOptions.DB.
		Model(&schemas.RoutineTag{}).
		Select(`"RoutineTagTable".*`).
		Where(`"RoutineTagTable".owner_id = ?`, userId).
		Scopes(r.routineTagScope.IncludePreloads(preloads)).
		Order(`"RoutineTagTable".created_at ASC`).
		Order(`"RoutineTagTable".id ASC`).
		Find(&routineTags)
	if result.Error != nil {
		return nil, apiexceptions.NewRoutineTagException().NotFound().WithOrigin(result.Error)
	}

	return routineTags, nil
}

func (r *RoutineTagRepository) CreateOne(
	userId uuid.UUID,
	input inputs.CreateRoutineTagInput,
	opts ...options.RepositoryOptions,
) (*uuid.UUID, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	newRoutineTag := schemas.RoutineTag{
		Id:      uuid.New(),
		OwnerId: userId,
		Color:   "#FFFFFF",
	}
	if err := copier.Copy(&newRoutineTag, &input); err != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewRoutineTagException().InvalidInput().WithOrigin(err)
	}

	result := parsedOptions.DB.
		Model(&schemas.RoutineTag{}).
		Create(&newRoutineTag)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTagException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTagException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewRoutineTagException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &newRoutineTag.Id, nil
}

func (r *RoutineTagRepository) CreateMany(
	userId uuid.UUID,
	input []inputs.CreateRoutineTagInput,
	opts ...options.RepositoryOptions,
) ([]uuid.UUID, *cexceptions.Exception) {
	if len(input) == 0 {
		return nil, apiexceptions.NewRoutineTagException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	newRoutineTags := make([]schemas.RoutineTag, 0, len(input))
	for _, in := range input {
		newRoutineTag := schemas.RoutineTag{
			Id:      uuid.New(),
			OwnerId: userId,
			Color:   "#FFFFFF",
		}
		if err := copier.Copy(&newRoutineTag, &in); err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewRoutineTagException().InvalidInput().WithOrigin(err)
		}
		if newRoutineTag.Id == uuid.Nil {
			newRoutineTag.Id = uuid.New()
		}
		if newRoutineTag.Color == "" {
			newRoutineTag.Color = "#FFFFFF"
		}
		newRoutineTags = append(newRoutineTags, newRoutineTag)
	}

	if len(newRoutineTags) == 0 {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewRoutineTagException().NoChanges()
	}

	result := parsedOptions.DB.
		Model(&schemas.RoutineTag{}).
		CreateInBatches(&newRoutineTags, parsedOptions.BatchSize)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTagException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTagException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	newRoutineTagIds := make([]uuid.UUID, len(newRoutineTags))
	for index, newRoutineTag := range newRoutineTags {
		newRoutineTagIds[index] = newRoutineTag.Id
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewRoutineTagException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return newRoutineTagIds, nil
}

func (r *RoutineTagRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateRoutineTagInput,
	opts ...options.RepositoryOptions,
) (*schemas.RoutineTag, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	existingRoutineTag, exception := r.GetOneById(id, userId, nil, opts...)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingRoutineTag)
	if err != nil {
		parsedOptions.DB.Rollback()
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true).WithOrigin(err)
	}

	result := parsedOptions.DB.
		Model(&schemas.RoutineTag{}).
		Where(`"RoutineTagTable".id = ?`, id).
		Select("*").
		Updates(&updates)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTagException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTagException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}
	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewRoutineTagException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &updates, nil
}

func (r *RoutineTagRepository) UpdateManyByIds(
	userId uuid.UUID,
	input []inputs.UpdateRoutineTagByIdInput,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	if len(input) == 0 {
		return apiexceptions.NewRoutineTagException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	}

	ids := make([]uuid.UUID, len(input))
	for index, in := range input {
		ids[index] = in.Id
	}
	validRoutineTags, exception := r.GetManyByIds(ids, userId, nil, opts...)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return apiexceptions.NewRoutineTagException().NotFound()
	}

	isRoutineTagValid := make(map[uuid.UUID]bool, len(validRoutineTags))
	for _, validRoutineTag := range validRoutineTags {
		isRoutineTagValid[validRoutineTag.Id] = true
	}

	var valuePlaceholders []string
	var valueArgs []interface{}
	for _, in := range input {
		if !isRoutineTagValid[in.Id] {
			continue
		}

		setIconNull := partialupdate.CheckSetNull(in.PartialUpdateInput.SetNull, "Icon")

		valuePlaceholders = append(valuePlaceholders, `(?::uuid, ?::text, ?::text, ?::"SupportedIcon", ?::boolean)`)
		valueArgs = append(valueArgs,
			in.Id,
			in.PartialUpdateInput.Values.Name,
			in.PartialUpdateInput.Values.Color,
			in.PartialUpdateInput.Values.Icon,
			setIconNull,
		)
	}

	if len(valuePlaceholders) == 0 {
		parsedOptions.DB.Rollback()
		return apiexceptions.NewRoutineTagException().NoChanges()
	}

	sql := fmt.Sprintf(`
		UPDATE "RoutineTagTable" AS rt
		SET
			name = COALESCE(v.name::text, rt.name),
			color = COALESCE(v.color::text, rt.color),
			icon = CASE
				WHEN v.set_icon_null::boolean THEN NULL
				ELSE COALESCE(v.icon::"SupportedIcon", rt.icon)
			END,
			updated_at = NOW()
		FROM (VALUES %s) AS v(id, name, color, icon, set_icon_null)
		WHERE rt.id = v.id::uuid
	`, strings.Join(valuePlaceholders, ","))
	result := parsedOptions.DB.Exec(sql, valueArgs...)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTagException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTagException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return apiexceptions.NewRoutineTagException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return nil
}

func (r *RoutineTagRepository) HardDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.
		Model(&schemas.RoutineTag{}).
		Where(`"RoutineTagTable".id = ? AND "RoutineTagTable".owner_id = ?`, id, userId).
		Delete(&schemas.RoutineTag{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTagException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTagException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *RoutineTagRepository) HardDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	if len(ids) == 0 {
		return apiexceptions.NewRoutineTagException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.
		Model(&schemas.RoutineTag{}).
		Where(`"RoutineTagTable".id IN ? AND "RoutineTagTable".owner_id = ?`, ids, userId).
		Delete(&schemas.RoutineTag{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewRoutineTagException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewRoutineTagException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}
