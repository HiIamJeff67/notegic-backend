package scopes

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type SubShelfScopeInterface interface {
	PassPermissionCheck(id uuid.UUID, userId uuid.UUID, permissions []enums.AccessControlPermission) func(db *gorm.DB) *gorm.DB
	PassPermissionChecks(ids []uuid.UUID, userId uuid.UUID, permissions []enums.AccessControlPermission) func(db *gorm.DB) *gorm.DB
	FilterOnlyDeleted(onlyDeleted types.Ternary) func(db *gorm.DB) *gorm.DB
	IncludePreloads(preloads []schemas.SubShelfRelation) func(db *gorm.DB) *gorm.DB
}

type SubShelfScope struct{}

func NewSubShelfScope() SubShelfScopeInterface {
	return &SubShelfScope{}
}

func (sc *SubShelfScope) PassPermissionCheck(id uuid.UUID, userId uuid.UUID, permissions []enums.AccessControlPermission) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if permissions == nil {
			return db.Where("id = ?", id)
		}

		// Use gorm.DB.Session to build a fresh statement for the subquery to avoid inheriting outer query clauses (especially in UPDATE/DELETE).
		subQuery := db.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToShelves{}).
			Select("1").
			Where("root_shelf_id = \"SubShelfTable\".root_shelf_id AND user_id = ? AND permission IN ?", userId, permissions)
		return db.Where("id = ? AND EXISTS (?)", id, subQuery)
	}
}

func (sc *SubShelfScope) PassPermissionChecks(ids []uuid.UUID, userId uuid.UUID, permissions []enums.AccessControlPermission) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if permissions == nil {
			return db.Where(`"SubShelfTable".id IN ?`, ids)
		}

		// Use gorm.DB.Session to build a fresh statement for the subquery to avoid inheriting outer query clauses (especially in UPDATE/DELETE).
		subQuery := db.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToShelves{}).
			Select("1").
			Where("root_shelf_id = \"SubShelfTable\".root_shelf_id AND user_id = ? AND permission IN ?", userId, permissions)
		return db.Where("id IN ? AND EXISTS (?)", ids, subQuery)
	}
}

func (sc *SubShelfScope) FilterOnlyDeleted(onlyDeleted types.Ternary) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		switch onlyDeleted {
		case types.Ternary_Positive:
			return db.Where("\"SubShelfTable\".deleted_at IS NOT NULL")
		case types.Ternary_Negative:
			return db.Where("\"SubShelfTable\".deleted_at IS NULL")
		default:
			return db
		}
	}
}

func (sc *SubShelfScope) IncludePreloads(preloads []schemas.SubShelfRelation) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, preload := range preloads {
			db = db.Preload(string(preload))
		}
		return db
	}
}
