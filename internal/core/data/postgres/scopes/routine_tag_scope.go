package scopes

import (
	"gorm.io/gorm"

	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
)

type RoutineTagScopeInterface interface {
	IncludePreloads(preloads []schemas.RoutineTagRelation) func(db *gorm.DB) *gorm.DB
}

type RoutineTagScope struct{}

func NewRoutineTagScope() RoutineTagScopeInterface {
	return &RoutineTagScope{}
}

func (sc *RoutineTagScope) IncludePreloads(preloads []schemas.RoutineTagRelation) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, preload := range preloads {
			db = db.Preload(string(preload))
		}
		return db
	}
}
