package repositories

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	array "github.com/HiIamJeff67/notegic-backend/shared/lib/array"
	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"
	exceptions "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/exceptions"
	inputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"
)

type SubShelfRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.SubShelfRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) (*schemas.SubShelf, *cexceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.SubShelfRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]schemas.SubShelf, *cexceptions.Exception)
	GetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.SubShelfRelation, opts ...RepositoryOptions) (*schemas.SubShelf, *cexceptions.Exception)
	GetAllByRootShelfId(rootShelfId uuid.UUID, userId uuid.UUID, preloads []schemas.SubShelfRelation, opts ...RepositoryOptions) ([]schemas.SubShelf, *cexceptions.Exception)
	CreateOneByRootShelfId(rootShelfId uuid.UUID, userId uuid.UUID, input inputs.CreateSubShelfInput, opts ...RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	CreateManyByRootShelfIds(userId uuid.UUID, input []inputs.CreateSubShelfByRootShelfIdInput, opts ...RepositoryOptions) ([]uuid.UUID, *cexceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateSubShelfInput, opts ...RepositoryOptions) (*schemas.SubShelf, *cexceptions.Exception)
	UpdateManyByIds(userId uuid.UUID, input []inputs.UpdateSubShelfByIdInput, opts ...RepositoryOptions) *cexceptions.Exception
	RestoreSoftDeletedOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) (*schemas.SubShelf, *cexceptions.Exception)
	RestoreSoftDeletedManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) ([]schemas.SubShelf, *cexceptions.Exception)
	SoftDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	SoftDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...RepositoryOptions) *cexceptions.Exception

	/* ============================== System Only Method ============================== */

	BulkCheckPermissionsAndGetManyByIds(inputs []inputs.BulkCheckSubShelfPermissionInput, preloads []schemas.SubShelfRelation, allowedPermissions []cenums.AccessControlPermission, opts ...RepositoryOptions) ([]bool, []schemas.SubShelf, *cexceptions.Exception)
	BulkCreateMany(inputs []inputs.BulkCreateSubShelfInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkUpdateMany(inputs []inputs.BulkUpdateSubShelfInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkDeleteMany(inputs []inputs.BulkDeleteSubShelfInput, opts ...RepositoryOptions) ([]bool, *cexceptions.Exception)
}

type SubShelfRepository struct {
	db *gorm.DB
	BulkSubShelfRepository
	subShelfScope scopes.SubShelfScopeInterface
	exceptions    exceptions.ShelfException
}

func NewSubShelfRepository(
	db *gorm.DB,
	subShelfScope scopes.SubShelfScopeInterface,
) SubShelfRepositoryInterface {
	return &SubShelfRepository{
		db:                     db,
		BulkSubShelfRepository: *NewBulkSubShelfRepository(db, subShelfScope),
		subShelfScope:          subShelfScope,
		exceptions:             exceptions.NewShelfException(),
	}
}

func (r *SubShelfRepository) HasPermission(
	id uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) bool {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var marker int
	result := parsedOptions.DB.
		Model(&schemas.SubShelf{}).
		Select("1").
		Scopes(r.subShelfScope.PassPermissionCheck(id, userId, allowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Limit(1).
		Scan(&marker)
	if err := result.Error; err != nil {
		return false
	}

	return marker == 1
}

func (r *SubShelfRepository) HavePermissions(
	ids []uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) bool {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var permittedIds []uuid.UUID
	result := parsedOptions.DB.
		Model(&schemas.SubShelf{}).
		Select(`DISTINCT "SubShelfTable".id`).
		Scopes(r.subShelfScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&permittedIds)
	if err := result.Error; err != nil {
		return false
	}

	return array.GetDistinctCount(ids) == array.GetDistinctCount(permittedIds)
}

func (r *SubShelfRepository) CheckPermissionAndGetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.SubShelfRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) (*schemas.SubShelf, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	subShelf := schemas.SubShelf{}
	query := parsedOptions.DB.
		Model(&schemas.SubShelf{}).
		Where(`"SubShelfTable".id = ?`, id)
	if allowedPermissions != nil && len(allowedPermissions) > 0 {
		query = query.Scopes(
			r.subShelfScope.PassPermissionCheck(id, userId, allowedPermissions),
		)
	}

	result := query.
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.subShelfScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&subShelf)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: subShelf.Id == uuid.Nil, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &subShelf, nil
}

func (r *SubShelfRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.SubShelfRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]schemas.SubShelf, *cexceptions.Exception) {
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	subShelves := []schemas.SubShelf{}
	result := parsedOptions.DB.
		Model(&schemas.SubShelf{}).
		Scopes(r.subShelfScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.subShelfScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&subShelves)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: len(subShelves) == 0, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return subShelves, nil
}

func (r *SubShelfRepository) GetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.SubShelfRelation,
	opts ...RepositoryOptions,
) (*schemas.SubShelf, *cexceptions.Exception) {
	return r.CheckPermissionAndGetOneById(
		id,
		userId,
		preloads,
		ParseRepositoryOptions(
			append([]RepositoryOptions{
				WithDB(r.db),
			}, opts...)...,
		).AllowedPermissions,
		opts...,
	)
}

func (r *SubShelfRepository) GetAllByRootShelfId(
	rootShelfId uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.SubShelfRelation,
	opts ...RepositoryOptions,
) ([]schemas.SubShelf, *cexceptions.Exception) {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	subShelves := []schemas.SubShelf{}
	query := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Where("root_shelf_id = ?", rootShelfId)
	if parsedOptions.HasAllowedPermissions() {
		subQuery := parsedOptions.DB.Model(&schemas.UsersToShelves{}).
			Select("1").
			Where(`root_shelf_id = "SubShelfTable".root_shelf_id AND user_id = ? AND permission IN ?`,
				userId, parsedOptions.AllowedPermissions,
			)
		query = query.Where("EXISTS (?)", subQuery)
	}
	if len(preloads) > 0 {
		for _, preload := range preloads {
			query = query.Preload(string(preload))
		}
	}

	result := query.Find(&subShelves)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.NotFound().WithOrigin(result.Error)},
		{First: len(subShelves) == 0, Second: r.exceptions.NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return subShelves, nil
}

func (r *SubShelfRepository) CreateOneByRootShelfId(
	rootShelfId uuid.UUID,
	userId uuid.UUID,
	input inputs.CreateSubShelfInput,
	opts ...RepositoryOptions,
) (*uuid.UUID, *cexceptions.Exception) {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, WithTransactionDB(parsedOptions.DB))
		opts = append(opts, WithLockingStrength(LockingStrengthNoKeyUpdate))
	}

	var newSubShelf schemas.SubShelf
	if input.PrevSubShelfId != nil {
		prevSubShelf, exception := r.CheckPermissionAndGetOneById(
			*input.PrevSubShelfId,
			userId,
			nil,
			parsedOptions.AllowedPermissions,
			opts...,
		)
		if exception = cexceptions.Cover(exception, []cexceptions.Pair{
			{First: prevSubShelf.RootShelfId != rootShelfId, Second: r.exceptions.InvalidDto("the given prev sub shelf is not one of the children of the given root shelf")},
		}); exception != nil {
			parsedOptions.DB.Rollback()
			return nil, exception
		}
		prevSubShelf.Path = append(prevSubShelf.Path, prevSubShelf.Id)
		newSubShelf.Path = prevSubShelf.Path
	} else {
		rootShelfRepository := NewRootShelfRepository(r.db, scopes.NewRootShelfScope())

		if !rootShelfRepository.HasPermission(
			rootShelfId,
			userId,
			parsedOptions.AllowedPermissions,
			opts...,
		) {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.NoPermission("create sub shelf by the given root shelf")
		}
	}

	if err := copier.Copy(&newSubShelf, &input); err != nil {
		parsedOptions.DB.Rollback()
		return nil, r.exceptions.InvalidInput().WithOrigin(err)
	}
	if newSubShelf.Id == uuid.Nil {
		newSubShelf.Id = uuid.New()
	}
	newSubShelf.RootShelfId = rootShelfId

	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Create(&newSubShelf)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToCreate().WithOrigin(result.Error)},
		{First: newSubShelf.Id == uuid.Nil, Second: r.exceptions.FailedToCreate()},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &newSubShelf.Id, nil
}

func (r *SubShelfRepository) CreateManyByRootShelfIds(
	userId uuid.UUID,
	input []inputs.CreateSubShelfByRootShelfIdInput,
	opts ...RepositoryOptions,
) ([]uuid.UUID, *cexceptions.Exception) {
	if len(input) == 0 {
		return nil, r.exceptions.NoChanges()
	}

	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, WithTransactionDB(parsedOptions.DB))
		opts = append(opts, WithLockingStrength(LockingStrengthNoKeyUpdate))
	}

	isPrevSubShelfExist := make(map[uuid.UUID]bool)
	isRootShelfExist := make(map[uuid.UUID]bool)
	prevSubShelfIds := make([]uuid.UUID, len(input))
	rootShelfIds := make([]uuid.UUID, len(input))
	for index, in := range input {
		if in.PrevSubShelfId != nil {
			if isPrevSubShelfExist[*in.PrevSubShelfId] {
				prevSubShelfIds[index] = *in.PrevSubShelfId
			}
			isPrevSubShelfExist[*in.PrevSubShelfId] = true
		}
		if isRootShelfExist[in.RootShelfId] {
			rootShelfIds[index] = in.RootShelfId
		}
		isRootShelfExist[in.RootShelfId] = true
	}

	validPrevSubShelves, exception := r.CheckPermissionsAndGetManyByIds(
		prevSubShelfIds,
		userId,
		nil,
		parsedOptions.AllowedPermissions,
		opts...,
	)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}
	isPrevSubShelfValid := make(map[uuid.UUID]*uuid.UUID)
	for _, validPrevSubShelf := range validPrevSubShelves {
		isPrevSubShelfValid[validPrevSubShelf.Id] = &validPrevSubShelf.RootShelfId
	}

	rootShelfRepository := NewRootShelfRepository(r.db, scopes.NewRootShelfScope())

	validRootShelves, _, exception := rootShelfRepository.CheckPermissionsAndGetManyByIds(
		rootShelfIds,
		userId,
		nil,
		parsedOptions.AllowedPermissions,
		opts...,
	)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}
	isRootShelfValid := make(map[uuid.UUID]bool)
	for _, validRootShelf := range validRootShelves {
		isRootShelfValid[validRootShelf.Id] = true
	}

	var newSubShelves []schemas.SubShelf
	for _, in := range input {
		if !isRootShelfValid[in.RootShelfId] ||
			(in.PrevSubShelfId != nil && (isPrevSubShelfValid[*in.PrevSubShelfId] != &in.RootShelfId)) {
			continue
		}
		var newSubShelf schemas.SubShelf
		if err := copier.Copy(&newSubShelf, &in); err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.InvalidInput().WithOrigin(err)
		}
		if newSubShelf.Id == uuid.Nil {
			newSubShelf.Id = uuid.New()
		}
		newSubShelf.RootShelfId = in.RootShelfId
		newSubShelves = append(newSubShelves, newSubShelf)
	}

	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		CreateInBatches(&newSubShelves, parsedOptions.BatchSize)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	newSubShelfIds := make([]uuid.UUID, len(newSubShelves))
	for index, newSubShelf := range newSubShelves {
		newSubShelfIds[index] = newSubShelf.Id
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return newSubShelfIds, nil
}

func (r *SubShelfRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateSubShelfInput,
	opts ...RepositoryOptions,
) (*schemas.SubShelf, *cexceptions.Exception) {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if !parsedOptions.IsTransactionStarted {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, WithTransactionDB(parsedOptions.DB))
		opts = append(opts, WithLockingStrength(LockingStrengthNoKeyUpdate))
	}

	existingSubShelf, exception := r.CheckPermissionAndGetOneById(
		id,
		userId,
		nil,
		parsedOptions.AllowedPermissions,
		opts...,
	)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	updates, err := partialupdate.PartialUpdatePreprocess(input.Values, input.SetNull, *existingSubShelf)
	if err != nil {
		parsedOptions.DB.Rollback()
		return nil, cexceptions.New("FailedToPreprocessPartialUpdate", "Repository", "Update", "Failed to preprocess partial update", http.StatusInternalServerError, true).WithOrigin(err)
	}

	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Where("id = ? AND deleted_at IS NULL", id).
		Select("*").
		Updates(&updates)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &updates, nil
}

func (r *SubShelfRepository) UpdateManyByIds(
	userId uuid.UUID,
	input []inputs.UpdateSubShelfByIdInput,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, WithTransactionDB(parsedOptions.DB))
		opts = append(opts, WithLockingStrength(LockingStrengthNoKeyUpdate))
	}

	isSubShelfValid := make(map[uuid.UUID]bool)
	if parsedOptions.HasAllowedPermissions() {
		subShelfIds := make([]uuid.UUID, len(input))
		for index, in := range input {
			subShelfIds[index] = in.Id
		}

		validSubShelves, exception := r.CheckPermissionsAndGetManyByIds(
			subShelfIds,
			userId,
			nil,
			parsedOptions.AllowedPermissions,
			opts...,
		)
		if exception != nil {
			parsedOptions.DB.Rollback()
			return exception
		}

		for _, validSubShelf := range validSubShelves {
			isSubShelfValid[validSubShelf.Id] = true
		}
	}

	var valuePlaceholders []string
	var valueArgs []interface{}
	for _, in := range input {
		if parsedOptions.HasAllowedPermissions() && !isSubShelfValid[in.Id] {
			continue
		}

		valuePlaceholders = append(valuePlaceholders, "(?::uuid, ?::text)")
		valueArgs = append(valueArgs,
			in.Id,
			in.PartialUpdateInput.Values.Name,
		)
	}

	sql := fmt.Sprintf(`
		UPDATE "SubShelfTable" AS s
		SET
			name = COALESCE(v.name::text, s.name)
		FROM (VALUES %s) AS v(id, name)
		WHERE s.id = v.id::uuid AND s.deleted_at IS NULL
	`, strings.Join(valuePlaceholders, ","))
	result := parsedOptions.DB.Exec(sql, valueArgs...)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return r.exceptions.FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return nil
}

func (r *SubShelfRepository) RestoreSoftDeletedOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) (*schemas.SubShelf, *cexceptions.Exception) {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var restoredSubShelf schemas.SubShelf
	result := parsedOptions.DB.Model(&restoredSubShelf).
		Scopes(r.subShelfScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Clauses(clause.Returning{}).
		Where(`"SubShelfTable".id = ?`, id).
		Updates(map[string]interface{}{"deleted_at": nil}) // force to assign null value
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: restoredSubShelf.Id == uuid.Nil, Second: r.exceptions.FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &restoredSubShelf, nil
}

func (r *SubShelfRepository) RestoreSoftDeletedManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) ([]schemas.SubShelf, *cexceptions.Exception) {
	if len(ids) == 0 {
		return nil, r.exceptions.NoChanges()
	}

	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	var restoredSubShelves []schemas.SubShelf
	result := parsedOptions.DB.Model(&restoredSubShelves).
		Scopes(r.subShelfScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Clauses(clause.Returning{}).
		Where(`"SubShelfTable".id IN ?`, ids).
		Updates(map[string]interface{}{"deleted_at": nil}) // force to assign null value
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: len(restoredSubShelves) == 0, Second: r.exceptions.FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return restoredSubShelves, nil
}

func (r *SubShelfRepository) SoftDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Scopes(r.subShelfScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"SubShelfTable".id = ?`, id).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *SubShelfRepository) SoftDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	if len(ids) == 0 {
		return r.exceptions.NoChanges()
	}

	opts = append(opts, WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Scopes(r.subShelfScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"SubShelfTable".id IN ?`, ids).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *SubShelfRepository) HardDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Scopes(r.subShelfScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"SubShelfTable".id = ?`, id).
		Delete(&schemas.SubShelf{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *SubShelfRepository) HardDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...RepositoryOptions,
) *cexceptions.Exception {
	if len(ids) == 0 {
		return r.exceptions.NoChanges()
	}

	opts = append(opts, WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Scopes(r.subShelfScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"SubShelfTable".id IN ?`, ids).
		Delete(&schemas.SubShelf{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: r.exceptions.FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: r.exceptions.NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

/* ============================== System Only Method ============================== */

func (r *SubShelfRepository) BulkCheckPermissionsAndGetManyByIds(
	inputs []inputs.BulkCheckSubShelfPermissionInput,
	preloads []schemas.SubShelfRelation,
	allowedPermissions []cenums.AccessControlPermission,
	opts ...RepositoryOptions,
) ([]bool, []schemas.SubShelf, *cexceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, []schemas.SubShelf{}, nil
	}

	parsedOptions := ParseRepositoryOptions(
		append([]RepositoryOptions{
			WithDB(r.db),
		}, opts...)...,
	)

	successes := make([]bool, len(inputs))
	ids := make([]uuid.UUID, 0, len(inputs))
	userIds := make([]uuid.UUID, 0, len(inputs))
	for _, in := range inputs {
		ids = append(ids, in.Id)
		userIds = append(userIds, in.UserId)
	}

	validIdSet := make(map[uuid.UUID]bool, len(ids))
	validTargetByUserId := make(map[[2]uuid.UUID]bool)
	if allowedPermissions != nil {
		var validTargets []struct {
			Id     uuid.UUID `gorm:"column:id"`
			UserId uuid.UUID `gorm:"column:user_id"`
		}
		result := parsedOptions.DB.Model(&schemas.SubShelf{}).
			Select(`"SubShelfTable".id, uts.user_id`).
			Joins(`INNER JOIN "UsersToShelvesTable" AS uts ON uts.root_shelf_id = "SubShelfTable".root_shelf_id`).
			Where(`"SubShelfTable".id IN ?`, ids).
			Where("uts.user_id IN ? AND uts.permission IN ?", userIds, allowedPermissions).
			Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
			Scan(&validTargets)
		if result.Error != nil {
			return nil, nil, r.exceptions.NotFound().WithOrigin(result.Error)
		}

		for _, validTarget := range validTargets {
			validTargetByUserId[[2]uuid.UUID{validTarget.Id, validTarget.UserId}] = true
			validIdSet[validTarget.Id] = true
		}
	} else {
		var validIds []uuid.UUID
		result := parsedOptions.DB.Model(&schemas.SubShelf{}).
			Select(`"SubShelfTable".id`).
			Where(`"SubShelfTable".id IN ?`, ids).
			Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
			Scan(&validIds)
		if result.Error != nil {
			return nil, nil, r.exceptions.NotFound().WithOrigin(result.Error)
		}

		for _, validId := range validIds {
			validIdSet[validId] = true
		}
	}

	validIds := make([]uuid.UUID, 0, len(validIdSet))
	for validId := range validIdSet {
		validIds = append(validIds, validId)
	}
	if len(validIds) == 0 {
		return successes, []schemas.SubShelf{}, nil
	}

	var subShelves []schemas.SubShelf
	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Where(`"SubShelfTable".id IN ?`, validIds).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.subShelfScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&subShelves)
	if result.Error != nil {
		return nil, nil, r.exceptions.NotFound().WithOrigin(result.Error)
	}

	foundIdSet := make(map[uuid.UUID]bool, len(subShelves))
	for _, subShelf := range subShelves {
		foundIdSet[subShelf.Id] = true
	}
	for index, in := range inputs {
		if validIdSet[in.Id] &&
			foundIdSet[in.Id] &&
			(allowedPermissions == nil || validTargetByUserId[[2]uuid.UUID{in.Id, in.UserId}]) {
			successes[index] = true
		}
	}

	return successes, subShelves, nil
}
