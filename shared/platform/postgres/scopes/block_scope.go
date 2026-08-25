package scopes

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type BlockScopeInterface interface {
	PassPermissionCheck(id uuid.UUID, userId uuid.UUID, permissions []cenums.AccessControlPermission) func(db *gorm.DB) *gorm.DB
	PassPermissionChecks(ids []uuid.UUID, userId uuid.UUID, permissions []cenums.AccessControlPermission) func(db *gorm.DB) *gorm.DB
	IncludePreloads(preloads []schemas.BlockRelation) func(db *gorm.DB) *gorm.DB
}

type BlockScope struct{}

func NewBlockScope() BlockScopeInterface {
	return &BlockScope{}
}

func (sc *BlockScope) PassPermissionCheck(id uuid.UUID, userId uuid.UUID, permissions []cenums.AccessControlPermission) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if permissions == nil {
			return db.Where(`"BlockTable".id = ?`, id)
		}

		// Use gorm.DB.Session to build a fresh statement for the subquery to avoid inheriting outer query clauses (especially in UPDATE/DELETE).
		subQuery := db.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToShelves{}).
			Select("1").
			Joins("INNER JOIN \"SubShelfTable\" ss ON ss.root_shelf_id = \"UsersToShelvesTable\".root_shelf_id").
			Joins("INNER JOIN \"BlockPackTable\" bp ON bp.parent_sub_shelf_id = ss.id").
			Where("bp.id = \"BlockTable\".block_pack_id").
			Where("bp.deleted_at IS NULL").
			Where("user_id = ? AND permission IN ?", userId, permissions)
		return db.Where("\"BlockTable\".id = ? AND EXISTS (?)", id, subQuery)
	}
}

func (sc *BlockScope) PassPermissionChecks(ids []uuid.UUID, userId uuid.UUID, permissions []cenums.AccessControlPermission) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if permissions == nil {
			return db.Where(`"BlockTable".id IN ?`, ids)
		}

		// Use gorm.DB.Session to build a fresh statement for the subquery to avoid inheriting outer query clauses (especially in UPDATE/DELETE).
		subQuery := db.Session(&gorm.Session{NewDB: true}).
			Model(&schemas.UsersToShelves{}).
			Select("1").
			Joins("INNER JOIN \"SubShelfTable\" ss ON ss.root_shelf_id = \"UsersToShelvesTable\".root_shelf_id").
			Joins("INNER JOIN \"BlockPackTable\" bp ON bp.parent_sub_shelf_id = ss.id").
			Where("bp.id = \"BlockTable\".block_pack_id").
			Where("bp.deleted_at IS NULL").
			Where("user_id = ? AND permission IN ?", userId, permissions)
		return db.Where("\"BlockTable\".id IN ? AND EXISTS (?)", ids, subQuery)
	}
}

func (sc *BlockScope) IncludePreloads(preloads []schemas.BlockRelation) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, preload := range preloads {
			db = db.Preload(string(preload))
		}
		return db
	}
}
