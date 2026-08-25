package repositories

import (
	"fmt"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"gorm.io/gorm/clause"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	array "github.com/HiIamJeff67/notegic-backend/shared/lib/array"
	partialupdate "github.com/HiIamJeff67/notegic-backend/shared/lib/partialupdate"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type SubShelfRepositoryInterface interface {
	HasPermission(id uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	HavePermissions(ids []uuid.UUID, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.SubShelfRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) (*schemas.SubShelf, *cexceptions.Exception)
	CheckPermissionsAndGetManyByIds(ids []uuid.UUID, userId uuid.UUID, preloads []schemas.SubShelfRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]schemas.SubShelf, *cexceptions.Exception)
	GetOneById(id uuid.UUID, userId uuid.UUID, preloads []schemas.SubShelfRelation, opts ...options.RepositoryOptions) (*schemas.SubShelf, *cexceptions.Exception)
	GetAllByRootShelfId(rootShelfId uuid.UUID, userId uuid.UUID, preloads []schemas.SubShelfRelation, opts ...options.RepositoryOptions) ([]schemas.SubShelf, *cexceptions.Exception)
	CreateOneByRootShelfId(rootShelfId uuid.UUID, userId uuid.UUID, input inputs.CreateSubShelfInput, opts ...options.RepositoryOptions) (*uuid.UUID, *cexceptions.Exception)
	CreateManyByRootShelfIds(userId uuid.UUID, input []inputs.CreateSubShelfByRootShelfIdInput, opts ...options.RepositoryOptions) ([]uuid.UUID, *cexceptions.Exception)
	UpdateOneById(id uuid.UUID, userId uuid.UUID, input inputs.PartialUpdateSubShelfInput, opts ...options.RepositoryOptions) (*schemas.SubShelf, *cexceptions.Exception)
	UpdateManyByIds(userId uuid.UUID, input []inputs.UpdateSubShelfByIdInput, opts ...options.RepositoryOptions) *cexceptions.Exception
	RestoreSoftDeletedOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) (*schemas.SubShelf, *cexceptions.Exception)
	RestoreSoftDeletedManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) ([]schemas.SubShelf, *cexceptions.Exception)
	SoftDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	SoftDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	HardDeleteOneById(id uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception
	HardDeleteManyByIds(ids []uuid.UUID, userId uuid.UUID, opts ...options.RepositoryOptions) *cexceptions.Exception

	/* ============================== System Only Method ============================== */

	BulkCheckPermissionsAndGetManyByIds(inputs []inputs.BulkCheckSubShelfPermissionInput, preloads []schemas.SubShelfRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]bool, []schemas.SubShelf, *cexceptions.Exception)
	BulkCreateMany(inputs []inputs.BulkCreateSubShelfInput, opts ...options.RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkUpdateMany(inputs []inputs.BulkUpdateSubShelfInput, opts ...options.RepositoryOptions) ([]bool, *cexceptions.Exception)
	BulkDeleteMany(inputs []inputs.BulkDeleteSubShelfInput, opts ...options.RepositoryOptions) ([]bool, *cexceptions.Exception)
}

type SubShelfRepository struct {
	subShelfScope scopes.SubShelfScopeInterface
}

func NewSubShelfRepository(subShelfScope scopes.SubShelfScopeInterface) SubShelfRepositoryInterface {
	return &SubShelfRepository{
		subShelfScope: subShelfScope,
	}
}

func (r *SubShelfRepository) HasPermission(
	id uuid.UUID,
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

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
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

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
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) (*schemas.SubShelf, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().NotFound().WithOrigin(result.Error)},
		{First: subShelf.Id == uuid.Nil, Second: apiexceptions.NewShelfException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &subShelf, nil
}

func (r *SubShelfRepository) CheckPermissionsAndGetManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.SubShelfRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]schemas.SubShelf, *cexceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	subShelves := []schemas.SubShelf{}
	result := parsedOptions.DB.
		Model(&schemas.SubShelf{}).
		Scopes(r.subShelfScope.PassPermissionChecks(ids, userId, allowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.subShelfScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&subShelves)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().NotFound().WithOrigin(result.Error)},
		{First: len(subShelves) == 0, Second: apiexceptions.NewShelfException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return subShelves, nil
}

func (r *SubShelfRepository) GetOneById(
	id uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.SubShelfRelation,
	opts ...options.RepositoryOptions,
) (*schemas.SubShelf, *cexceptions.Exception) {
	return r.CheckPermissionAndGetOneById(
		id,
		userId,
		preloads,
		options.ParseRepositoryOptions(opts...).AllowedPermissions,
		opts...,
	)
}

func (r *SubShelfRepository) GetAllByRootShelfId(
	rootShelfId uuid.UUID,
	userId uuid.UUID,
	preloads []schemas.SubShelfRelation,
	opts ...options.RepositoryOptions,
) ([]schemas.SubShelf, *cexceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().NotFound().WithOrigin(result.Error)},
		{First: len(subShelves) == 0, Second: apiexceptions.NewShelfException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return subShelves, nil
}

func (r *SubShelfRepository) CreateOneByRootShelfId(
	rootShelfId uuid.UUID,
	userId uuid.UUID,
	input inputs.CreateSubShelfInput,
	opts ...options.RepositoryOptions,
) (*uuid.UUID, *cexceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
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
			{First: prevSubShelf.RootShelfId != rootShelfId, Second: apiexceptions.NewShelfException().InvalidDto("the given prev sub shelf is not one of the children of the given root shelf")},
		}); exception != nil {
			parsedOptions.DB.Rollback()
			return nil, exception
		}
		prevSubShelf.Path = append(prevSubShelf.Path, prevSubShelf.Id)
		newSubShelf.Path = prevSubShelf.Path
	} else {
		rootShelfRepository := NewRootShelfRepository(scopes.NewRootShelfScope())

		if !rootShelfRepository.HasPermission(
			rootShelfId,
			userId,
			parsedOptions.AllowedPermissions,
			opts...,
		) {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().NoPermission("create sub shelf by the given root shelf")
		}
	}

	if err := copier.Copy(&newSubShelf, &input); err != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewShelfException().InvalidInput().WithOrigin(err)
	}
	if newSubShelf.Id == uuid.Nil {
		newSubShelf.Id = uuid.New()
	}
	newSubShelf.RootShelfId = rootShelfId

	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Create(&newSubShelf)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToCreate().WithOrigin(result.Error)},
		{First: newSubShelf.Id == uuid.Nil, Second: apiexceptions.NewShelfException().FailedToCreate()},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &newSubShelf.Id, nil
}

func (r *SubShelfRepository) CreateManyByRootShelfIds(
	userId uuid.UUID,
	input []inputs.CreateSubShelfByRootShelfIdInput,
	opts ...options.RepositoryOptions,
) ([]uuid.UUID, *cexceptions.Exception) {
	if len(input) == 0 {
		return nil, apiexceptions.NewShelfException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
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

	rootShelfRepository := NewRootShelfRepository(scopes.NewRootShelfScope())

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
			return nil, apiexceptions.NewShelfException().InvalidInput().WithOrigin(err)
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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToCreate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
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
			return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return newSubShelfIds, nil
}

func (r *SubShelfRepository) UpdateOneById(
	id uuid.UUID,
	userId uuid.UUID,
	input inputs.PartialUpdateSubShelfInput,
	opts ...options.RepositoryOptions,
) (*schemas.SubShelf, *cexceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if !parsedOptions.IsTransactionStarted {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return &updates, nil
}

func (r *SubShelfRepository) UpdateManyByIds(
	userId uuid.UUID,
	input []inputs.UpdateSubShelfByIdInput,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
		opts = append(opts, options.WithTransactionDB(parsedOptions.DB))
		opts = append(opts, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
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
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		parsedOptions.DB.Rollback()
		return exception
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	return nil
}

func (r *SubShelfRepository) RestoreSoftDeletedOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) (*schemas.SubShelf, *cexceptions.Exception) {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var restoredSubShelf schemas.SubShelf
	result := parsedOptions.DB.Model(&restoredSubShelf).
		Scopes(r.subShelfScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Clauses(clause.Returning{}).
		Where(`"SubShelfTable".id = ?`, id).
		Updates(map[string]interface{}{"deleted_at": nil}) // force to assign null value
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)},
		{First: restoredSubShelf.Id == uuid.Nil, Second: apiexceptions.NewShelfException().FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return &restoredSubShelf, nil
}

func (r *SubShelfRepository) RestoreSoftDeletedManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) ([]schemas.SubShelf, *cexceptions.Exception) {
	if len(ids) == 0 {
		return nil, apiexceptions.NewShelfException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var restoredSubShelves []schemas.SubShelf
	result := parsedOptions.DB.Model(&restoredSubShelves).
		Scopes(r.subShelfScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Clauses(clause.Returning{}).
		Where(`"SubShelfTable".id IN ?`, ids).
		Updates(map[string]interface{}{"deleted_at": nil}) // force to assign null value
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)},
		{First: len(restoredSubShelves) == 0, Second: apiexceptions.NewShelfException().FailedToUpdate()},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		return nil, exception
	}

	return restoredSubShelves, nil
}

func (r *SubShelfRepository) SoftDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Scopes(r.subShelfScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"SubShelfTable".id = ?`, id).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *SubShelfRepository) SoftDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	if len(ids) == 0 {
		return apiexceptions.NewShelfException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Negative))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Scopes(r.subShelfScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"SubShelfTable".id IN ?`, ids).
		Update("deleted_at", time.Now())
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *SubShelfRepository) HardDeleteOneById(
	id uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Scopes(r.subShelfScope.PassPermissionCheck(id, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"SubShelfTable".id = ?`, id).
		Delete(&schemas.SubShelf{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

func (r *SubShelfRepository) HardDeleteManyByIds(
	ids []uuid.UUID,
	userId uuid.UUID,
	opts ...options.RepositoryOptions,
) *cexceptions.Exception {
	if len(ids) == 0 {
		return apiexceptions.NewShelfException().NoChanges()
	}

	opts = append(opts, options.WithOnlyDeleted(types.Ternary_Positive))
	parsedOptions := options.ParseRepositoryOptions(opts...)

	result := parsedOptions.DB.Model(&schemas.SubShelf{}).
		Scopes(r.subShelfScope.PassPermissionChecks(ids, userId, parsedOptions.AllowedPermissions)).
		Scopes(r.subShelfScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Where(`"SubShelfTable".id IN ?`, ids).
		Delete(&schemas.SubShelf{})
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToDelete().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		return exception
	}

	return nil
}

/* ============================== System Only Method ============================== */

func (r *SubShelfRepository) BulkCheckPermissionsAndGetManyByIds(
	inputs []inputs.BulkCheckSubShelfPermissionInput,
	preloads []schemas.SubShelfRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]bool, []schemas.SubShelf, *cexceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, []schemas.SubShelf{}, nil
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

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
			return nil, nil, apiexceptions.NewShelfException().NotFound().WithOrigin(result.Error)
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
			return nil, nil, apiexceptions.NewShelfException().NotFound().WithOrigin(result.Error)
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
		return nil, nil, apiexceptions.NewShelfException().NotFound().WithOrigin(result.Error)
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

func (r *SubShelfRepository) BulkCreateMany(
	inputs []inputs.BulkCreateSubShelfInput,
	opts ...options.RepositoryOptions,
) ([]bool, *cexceptions.Exception) {
	if len(inputs) == 0 {
		return []bool{}, apiexceptions.NewShelfException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	successes := make([]bool, len(inputs))
	rootShelfIds := make([]uuid.UUID, 0, len(inputs))
	userIds := make([]uuid.UUID, 0, len(inputs))
	prevSubShelfIds := make([]uuid.UUID, 0, len(inputs))
	for _, in := range inputs {
		rootShelfIds = append(rootShelfIds, in.RootShelfId)
		userIds = append(userIds, in.UserId)
		if in.PrevSubShelfId != nil && *in.PrevSubShelfId != uuid.Nil {
			prevSubShelfIds = append(prevSubShelfIds, *in.PrevSubShelfId)
		}
	}

	var usersToShelves []schemas.UsersToShelves
	result := parsedOptions.DB.Model(&schemas.UsersToShelves{}).
		Where("root_shelf_id IN ? AND user_id IN ? AND permission IN ?", rootShelfIds, userIds, parsedOptions.AllowedPermissions).
		Find(&usersToShelves)
	if result.Error != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewShelfException().FailedToCreate().WithOrigin(result.Error)
	}

	validRootShelfByUserId := make(map[[2]uuid.UUID]bool, len(usersToShelves))
	for _, usersToShelf := range usersToShelves {
		validRootShelfByUserId[[2]uuid.UUID{usersToShelf.RootShelfId, usersToShelf.UserId}] = true
	}

	prevSubShelfById := make(map[uuid.UUID]schemas.SubShelf)
	if len(prevSubShelfIds) > 0 {
		var prevSubShelves []schemas.SubShelf
		result = parsedOptions.DB.Model(&schemas.SubShelf{}).
			Where("id IN ? AND deleted_at IS NULL", prevSubShelfIds).
			Find(&prevSubShelves)
		if result.Error != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToCreate().WithOrigin(result.Error)
		}
		for _, prevSubShelf := range prevSubShelves {
			prevSubShelfById[prevSubShelf.Id] = prevSubShelf
		}
	}

	newSubShelves := make([]schemas.SubShelf, 0, len(inputs))
	successIndexes := make([]int, 0, len(inputs))
	for index, in := range inputs {
		if !validRootShelfByUserId[[2]uuid.UUID{in.RootShelfId, in.UserId}] {
			continue
		}

		newSubShelfId := uuid.New()
		if in.Id != nil && *in.Id != uuid.Nil {
			newSubShelfId = *in.Id
		}

		newSubShelf := schemas.SubShelf{
			Id:             newSubShelfId,
			RootShelfId:    in.RootShelfId,
			PrevSubShelfId: in.PrevSubShelfId,
			Name:           in.Name,
			Path:           types.UUIDArray{},
		}

		if in.PrevSubShelfId != nil && *in.PrevSubShelfId != uuid.Nil {
			prevSubShelf, exist := prevSubShelfById[*in.PrevSubShelfId]
			if !exist || prevSubShelf.RootShelfId != in.RootShelfId {
				continue
			}
			newSubShelf.Path = append(prevSubShelf.Path, prevSubShelf.Id)
		}

		newSubShelves = append(newSubShelves, newSubShelf)
		successIndexes = append(successIndexes, index)
	}

	if len(newSubShelves) == 0 {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return successes, nil
	}

	result = parsedOptions.DB.Model(&schemas.SubShelf{}).
		CreateInBatches(&newSubShelves, parsedOptions.BatchSize)
	if result.Error != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewShelfException().FailedToCreate().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	for _, successIndex := range successIndexes {
		successes[successIndex] = true
	}

	return successes, nil
}

func (r *SubShelfRepository) BulkUpdateMany(
	bulkInputs []inputs.BulkUpdateSubShelfInput,
	opts ...options.RepositoryOptions,
) ([]bool, *cexceptions.Exception) {
	if len(bulkInputs) == 0 {
		return []bool{}, apiexceptions.NewShelfException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	checkInputs := make([]inputs.BulkCheckSubShelfPermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckSubShelfPermissionInput{
			UserId: in.UserId,
			Id:     in.Id,
		}
	}
	checkOptions := append(opts, options.WithTransactionDB(parsedOptions.DB))
	checkOptions = append(checkOptions, options.WithOnlyDeleted(types.Ternary_Negative))
	checkOptions = append(checkOptions, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	successes, _, exception := r.BulkCheckPermissionsAndGetManyByIds(
		checkInputs,
		nil,
		parsedOptions.AllowedPermissions,
		checkOptions...,
	)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	valuePlaceholders := make([]string, 0, len(bulkInputs))
	valueArgs := make([]interface{}, 0, len(bulkInputs)*3)
	for index, in := range bulkInputs {
		if !successes[index] {
			continue
		}

		valuePlaceholders = append(valuePlaceholders, "(?::int, ?::uuid, ?::text)")
		valueArgs = append(valueArgs,
			index,
			in.Id,
			in.PartialUpdateInput.Values.Name,
		)
	}
	if len(valuePlaceholders) == 0 {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return successes, nil
	}

	sql := fmt.Sprintf(`
		WITH payload(idx, id, name) AS (
			VALUES %s
		),
		updated AS (
			UPDATE "SubShelfTable" AS ss
			SET
				name = COALESCE(v.name::text, ss.name),
				updated_at = NOW()
			FROM payload AS v
			WHERE ss.id = v.id::uuid
				AND ss.deleted_at IS NULL
			RETURNING ss.id
		)
		SELECT v.idx
		FROM payload AS v
		INNER JOIN updated AS u ON u.id = v.id::uuid
	`, strings.Join(valuePlaceholders, ","))

	var updatedIndexes []struct {
		Index int `gorm:"column:idx"`
	}
	result := parsedOptions.DB.Raw(sql, valueArgs...).Scan(&updatedIndexes)
	if result.Error != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	successes = make([]bool, len(bulkInputs))
	for _, updatedIndex := range updatedIndexes {
		if updatedIndex.Index >= 0 && updatedIndex.Index < len(successes) {
			successes[updatedIndex.Index] = true
		}
	}

	return successes, nil
}

func (r *SubShelfRepository) BulkDeleteMany(
	bulkInputs []inputs.BulkDeleteSubShelfInput,
	opts ...options.RepositoryOptions,
) ([]bool, *cexceptions.Exception) {
	if len(bulkInputs) == 0 {
		return []bool{}, apiexceptions.NewShelfException().NoChanges()
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	shouldStartTransaction := !parsedOptions.IsTransactionStarted
	if shouldStartTransaction {
		parsedOptions.DB = parsedOptions.DB.Begin()
	}

	checkInputs := make([]inputs.BulkCheckSubShelfPermissionInput, len(bulkInputs))
	for index, in := range bulkInputs {
		checkInputs[index] = inputs.BulkCheckSubShelfPermissionInput{
			UserId: in.UserId,
			Id:     in.Id,
		}
	}
	checkOptions := append(opts, options.WithTransactionDB(parsedOptions.DB))
	checkOptions = append(checkOptions, options.WithOnlyDeleted(types.Ternary_Negative))
	checkOptions = append(checkOptions, options.WithLockingStrength(options.LockingStrengthNoKeyUpdate))
	successes, _, exception := r.BulkCheckPermissionsAndGetManyByIds(
		checkInputs,
		nil,
		parsedOptions.AllowedPermissions,
		checkOptions...,
	)
	if exception != nil {
		parsedOptions.DB.Rollback()
		return nil, exception
	}

	validIds := make([]uuid.UUID, 0, len(bulkInputs))
	for index, in := range bulkInputs {
		if successes[index] {
			validIds = append(validIds, in.Id)
		}
	}
	if len(validIds) == 0 {
		if shouldStartTransaction {
			parsedOptions.DB.Rollback()
		}
		return successes, nil
	}

	var deletedSubShelves []schemas.SubShelf
	result := parsedOptions.DB.Model(&deletedSubShelves).
		Clauses(clause.Returning{}).
		Where("id IN ? AND deleted_at IS NULL", validIds).
		Updates(map[string]interface{}{"deleted_at": time.Now(), "updated_at": time.Now()})
	if result.Error != nil {
		parsedOptions.DB.Rollback()
		return nil, apiexceptions.NewShelfException().FailedToDelete().WithOrigin(result.Error)
	}

	if shouldStartTransaction {
		if err := parsedOptions.DB.Commit().Error; err != nil {
			parsedOptions.DB.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
		}
	}

	deletedIdSet := make(map[uuid.UUID]bool, len(deletedSubShelves))
	for _, deletedSubShelf := range deletedSubShelves {
		deletedIdSet[deletedSubShelf.Id] = true
	}
	for index, in := range bulkInputs {
		if successes[index] && deletedIdSet[in.Id] {
			successes[index] = true
		} else {
			successes[index] = false
		}
	}

	return successes, nil
}
