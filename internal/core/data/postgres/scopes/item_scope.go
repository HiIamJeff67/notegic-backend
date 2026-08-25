package scopes

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type ItemScopeInterface interface {
	PassPermissionCheck(id uuid.UUID, itemType enums.ItemType, userId uuid.UUID, permissions []enums.AccessControlPermission) func(db *gorm.DB) *gorm.DB
	PassPermissionChecks(itemIdentities []types.Pair[uuid.UUID, enums.ItemType], userId uuid.UUID, permissions []enums.AccessControlPermission) func(db *gorm.DB) *gorm.DB
	FilterOnlyDeleted(onlyDeleted types.Ternary) func(db *gorm.DB) *gorm.DB
	IncludePreloads(preloads []schemas.ItemRelation) func(db *gorm.DB) *gorm.DB
}

type ItemScope struct{}

func NewItemScope() ItemScopeInterface {
	return &ItemScope{}
}

func (sc *ItemScope) PassPermissionCheck(id uuid.UUID, itemType enums.ItemType, userId uuid.UUID, permissions []enums.AccessControlPermission) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if permissions == nil {
			return db.Where(`"ItemTable".id = ? AND "ItemTable".type = ?`, id, itemType)
		}

		// ItemTable is a projection with root_shelf_id, so permission can be checked without joining the concrete item tables.
		subQuery := db.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToShelves{}).
			Select("1").
			Where("root_shelf_id = \"ItemTable\".root_shelf_id AND user_id = ? AND permission IN ?", userId, permissions)
		return db.Where("\"ItemTable\".id = ? AND \"ItemTable\".type = ? AND EXISTS (?)", id, itemType, subQuery)
	}
}

func (sc *ItemScope) PassPermissionChecks(itemIdentities []types.Pair[uuid.UUID, enums.ItemType], userId uuid.UUID, permissions []enums.AccessControlPermission) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		// ItemTable is a projection with root_shelf_id, so permission can be checked without joining the concrete item tables.
		values := make([][]any, len(itemIdentities))
		for index, itemIdentity := range itemIdentities {
			values[index] = []any{itemIdentity.First, itemIdentity.Second}
		}
		if permissions == nil {
			return db.Where(`("ItemTable".id, "ItemTable".type) IN ?`, values)
		}

		subQuery := db.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToShelves{}).
			Select("1").
			Where("root_shelf_id = \"ItemTable\".root_shelf_id AND user_id = ? AND permission IN ?", userId, permissions)
		return db.Where("(\"ItemTable\".id, \"ItemTable\".type) IN ? AND EXISTS (?)", values, subQuery)
	}
}

func (sc *ItemScope) FilterOnlyDeleted(onlyDeleted types.Ternary) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		switch onlyDeleted {
		case types.Ternary_Positive:
			return db.Where("\"ItemTable\".deleted_at IS NOT NULL")
		case types.Ternary_Negative:
			return db.Where("\"ItemTable\".deleted_at IS NULL")
		default:
			return db
		}
	}
}

func (sc *ItemScope) IncludePreloads(preloads []schemas.ItemRelation) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, preload := range preloads {
			db = db.Preload(string(preload))
		}
		return db
	}
}
