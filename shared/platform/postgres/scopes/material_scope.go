package scopes

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type MaterialScopeInterface interface {
	PassPermissionCheck(id uuid.UUID, userId uuid.UUID, permissions []enums.AccessControlPermission) func(db *gorm.DB) *gorm.DB
	PassPermissionChecks(ids []uuid.UUID, userId uuid.UUID, permissions []enums.AccessControlPermission) func(db *gorm.DB) *gorm.DB
	FilterOnlyDeleted(onlyDeleted types.Ternary) func(db *gorm.DB) *gorm.DB
	IncludePreloads(preloads []schemas.MaterialRelation) func(db *gorm.DB) *gorm.DB
}

type MaterialScope struct{}

func NewMaterialScope() MaterialScopeInterface {
	return &MaterialScope{}
}

func (sc *MaterialScope) PassPermissionCheck(id uuid.UUID, userId uuid.UUID, permissions []enums.AccessControlPermission) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if permissions == nil {
			return db.Where(`"MaterialTable".id = ?`, id)
		}

		// Use gorm.DB.Session to build a fresh statement for the subquery to avoid inheriting outer query clauses (especially in UPDATE/DELETE).
		subQuery := db.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToShelves{}).
			Select("1").
			Joins("INNER JOIN \"SubShelfTable\" ss ON ss.root_shelf_id = \"UsersToShelvesTable\".root_shelf_id").
			Where("ss.id = \"MaterialTable\".parent_sub_shelf_id").
			Where("user_id = ? AND permission IN ?", userId, permissions)
		return db.Where("\"MaterialTable\".id = ? AND EXISTS (?)", id, subQuery)
	}
}

func (sc *MaterialScope) PassPermissionChecks(ids []uuid.UUID, userId uuid.UUID, permissions []enums.AccessControlPermission) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if permissions == nil {
			return db.Where(`"MaterialTable".id IN ?`, ids)
		}

		// Use gorm.DB.Session to build a fresh statement for the subquery to avoid inheriting outer query clauses (especially in UPDATE/DELETE).
		subQuery := db.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToShelves{}).
			Select("1").
			Joins("INNER JOIN \"SubShelfTable\" ss ON ss.root_shelf_id = \"UsersToShelvesTable\".root_shelf_id").
			Where("ss.id = \"MaterialTable\".parent_sub_shelf_id").
			Where("user_id = ? AND permission IN ?", userId, permissions)
		return db.Where("\"MaterialTable\".id IN ? AND EXISTS (?)", ids, subQuery)
	}
}

func (sc *MaterialScope) FilterOnlyDeleted(onlyDeleted types.Ternary) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		switch onlyDeleted {
		case types.Ternary_Positive:
			return db.Where("\"MaterialTable\".deleted_at IS NOT NULL")
		case types.Ternary_Negative:
			return db.Where("\"MaterialTable\".deleted_at IS NULL")
		default:
			return db
		}
	}
}

func (sc *MaterialScope) IncludePreloads(preloads []schemas.MaterialRelation) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, preload := range preloads {
			db = db.Preload(string(preload))
		}
		return db
	}
}
