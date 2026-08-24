package scopes

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Locking(lockingStrength *string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if lockingStrength == nil {
			return db
		}

		return db.Clauses(clause.Locking{
			Strength: *lockingStrength,
			Table:    clause.Table{Name: clause.CurrentTable},
		})
	}
}
