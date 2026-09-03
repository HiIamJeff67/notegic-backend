package scopes

import (
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func Locking(lockingStrength *string, lockingOptions ...string) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if lockingStrength == nil {
			return db
		}

		options := ""
		if len(lockingOptions) > 0 {
			options = lockingOptions[0]
		}

		return db.Clauses(clause.Locking{
			Strength: *lockingStrength,
			Options:  options,
			Table:    clause.Table{Name: clause.CurrentTable},
		})
	}
}
