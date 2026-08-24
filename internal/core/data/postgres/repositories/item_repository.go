package repositories

import (
	"github.com/google/uuid"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	array "github.com/HiIamJeff67/notegic-backend/shared/lib/array"

	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	scopes "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/scopes"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
)

type ItemRepositoryInterface interface {
	HasPermission(id uuid.UUID, itemType enums.ItemType, userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	HavePermissions(itemIdentities []types.Pair[uuid.UUID, enums.ItemType], userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) bool
	CheckPermissionAndGetOneById(id uuid.UUID, itemType enums.ItemType, userId uuid.UUID, preloads []schemas.ItemRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) (*schemas.Item, *exceptions.Exception)
	CheckPermissionsAndGetManyByIds(itemIdentities []types.Pair[uuid.UUID, enums.ItemType], userId uuid.UUID, preloads []schemas.ItemRelation, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]schemas.Item, *exceptions.Exception)
	GetPermittedIdentities(itemIdentities []types.Pair[uuid.UUID, enums.ItemType], userId uuid.UUID, allowedPermissions []enums.AccessControlPermission, opts ...options.RepositoryOptions) ([]types.Pair[uuid.UUID, enums.ItemType], *exceptions.Exception)
}

type ItemRepository struct {
	itemScope scopes.ItemScopeInterface
}

func NewItemRepository(
	itemScope scopes.ItemScopeInterface,
) ItemRepositoryInterface {
	return &ItemRepository{
		itemScope: itemScope,
	}
}

func (r *ItemRepository) HasPermission(
	id uuid.UUID,
	itemType enums.ItemType,
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var marker int
	result := parsedOptions.DB.
		Model(&schemas.Item{}).
		Select("1").
		Scopes(r.itemScope.PassPermissionCheck(id, itemType, userId, allowedPermissions)).
		Scopes(r.itemScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Limit(1).
		Scan(&marker)
	if err := result.Error; err != nil {
		return false
	}

	return marker == 1
}

func (r *ItemRepository) HavePermissions(
	itemIdentities []types.Pair[uuid.UUID, enums.ItemType],
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) bool {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var permittedItems []schemas.Item
	result := parsedOptions.DB.
		Model(&schemas.Item{}).
		Select(`DISTINCT "ItemTable".id, "ItemTable".type`).
		Scopes(r.itemScope.PassPermissionChecks(itemIdentities, userId, allowedPermissions)).
		Scopes(r.itemScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&permittedItems)
	if err := result.Error; err != nil {
		return false
	}

	permittedItemIdentities := make([]types.Pair[uuid.UUID, enums.ItemType], len(permittedItems))
	for index, permittedItem := range permittedItems {
		permittedItemIdentities[index] = types.Pair[uuid.UUID, enums.ItemType]{
			First:  permittedItem.Id,
			Second: permittedItem.Type,
		}
	}

	return array.GetDistinctCount(itemIdentities) == array.GetDistinctCount(permittedItemIdentities)
}

func (r *ItemRepository) CheckPermissionAndGetOneById(
	id uuid.UUID,
	itemType enums.ItemType,
	userId uuid.UUID,
	preloads []schemas.ItemRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) (*schemas.Item, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var item schemas.Item
	query := parsedOptions.DB.
		Model(&schemas.Item{}).
		Where(`"ItemTable".id = ? AND "ItemTable".type = ?`, id, itemType)
	if allowedPermissions != nil && len(allowedPermissions) > 0 {
		query = query.Scopes(
			r.itemScope.PassPermissionCheck(id, itemType, userId, allowedPermissions),
		)
	}

	result := query.
		Scopes(r.itemScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.itemScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		First(&item)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewItemException().NotFound().WithOrigin(result.Error)},
		{First: item.Id == uuid.Nil, Second: apiexceptions.NewItemException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return &item, nil
}

func (r *ItemRepository) CheckPermissionsAndGetManyByIds(
	itemIdentities []types.Pair[uuid.UUID, enums.ItemType],
	userId uuid.UUID,
	preloads []schemas.ItemRelation,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]schemas.Item, *exceptions.Exception) {
	parsedOptions := options.ParseRepositoryOptions(opts...)

	var items []schemas.Item
	result := parsedOptions.DB.
		Model(&schemas.Item{}).
		Scopes(r.itemScope.PassPermissionChecks(itemIdentities, userId, allowedPermissions)).
		Scopes(r.itemScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Scopes(r.itemScope.IncludePreloads(preloads)).
		Scopes(scopes.Locking(parsedOptions.LockingStrength)).
		Find(&items)
	if exception := exceptions.Cover(nil, []exceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewItemException().NotFound().WithOrigin(result.Error)},
		{First: len(items) == 0, Second: apiexceptions.NewItemException().NotFound()},
	}); exception != nil {
		return nil, exception
	}

	return items, nil
}

func (r *ItemRepository) GetPermittedIdentities(
	itemIdentities []types.Pair[uuid.UUID, enums.ItemType],
	userId uuid.UUID,
	allowedPermissions []enums.AccessControlPermission,
	opts ...options.RepositoryOptions,
) ([]types.Pair[uuid.UUID, enums.ItemType], *exceptions.Exception) {
	if len(itemIdentities) == 0 {
		return []types.Pair[uuid.UUID, enums.ItemType]{}, nil
	}

	parsedOptions := options.ParseRepositoryOptions(opts...)

	var permittedItems []schemas.Item
	result := parsedOptions.DB.
		Model(&schemas.Item{}).
		Select(`DISTINCT "ItemTable".id, "ItemTable".type`).
		Scopes(r.itemScope.PassPermissionChecks(itemIdentities, userId, allowedPermissions)).
		Scopes(r.itemScope.FilterOnlyDeleted(parsedOptions.OnlyDeleted)).
		Find(&permittedItems)
	if result.Error != nil {
		return nil, apiexceptions.NewItemException().NotFound().WithOrigin(result.Error)
	}

	permittedItemIdentities := make([]types.Pair[uuid.UUID, enums.ItemType], len(permittedItems))
	for index, permittedItem := range permittedItems {
		permittedItemIdentities[index] = types.Pair[uuid.UUID, enums.ItemType]{
			First:  permittedItem.Id,
			Second: permittedItem.Type,
		}
	}

	return permittedItemIdentities, nil
}
