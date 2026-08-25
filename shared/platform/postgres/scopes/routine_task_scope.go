package scopes

import (
	"github.com/google/uuid"
	"gorm.io/gorm"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type RoutineTaskScopeInterface interface {
	PassPermissionCheck(id uuid.UUID, userId uuid.UUID, permissions []cenums.AccessControlPermission) func(db *gorm.DB) *gorm.DB
	PassPermissionChecks(ids []uuid.UUID, userId uuid.UUID, permissions []cenums.AccessControlPermission) func(db *gorm.DB) *gorm.DB
	IncludePreloads(preloads []schemas.RoutineTaskRelation) func(db *gorm.DB) *gorm.DB
}

type RoutineTaskScope struct{}

func NewRoutineTaskScope() RoutineTaskScopeInterface {
	return &RoutineTaskScope{}
}

func (sc *RoutineTaskScope) PassPermissionCheck(id uuid.UUID, userId uuid.UUID, permissions []cenums.AccessControlPermission) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if permissions == nil {
			return db.Where(`"RoutineTaskTable".id = ?`, id)
		}

		// Use gorm.DB.Session to build a fresh statement for the subquery to avoid inheriting outer query clauses (especially in UPDATE/DELETE).
		subQuery := db.Session(&gorm.Session{NewDB: true}).
			Table(`"RoutineTable" routine`).
			Select("1").
			Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = routine.station_id`).
			Where(`routine.id = "RoutineTaskTable".routine_id`).
			Where(`routine.deleted_at IS NULL`).
			Where("uts.user_id = ? AND uts.permission IN ?", userId, permissions)
		return db.Where("\"RoutineTaskTable\".id = ? AND EXISTS (?)", id, subQuery)
	}
}

func (sc *RoutineTaskScope) PassPermissionChecks(ids []uuid.UUID, userId uuid.UUID, permissions []cenums.AccessControlPermission) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		if permissions == nil {
			return db.Where(`"RoutineTaskTable".id IN ?`, ids)
		}

		// Use gorm.DB.Session to build a fresh statement for the subquery to avoid inheriting outer query clauses (especially in UPDATE/DELETE).
		subQuery := db.Session(&gorm.Session{NewDB: true}).
			Table(`"RoutineTable" routine`).
			Select("1").
			Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = routine.station_id`).
			Where(`routine.id = "RoutineTaskTable".routine_id`).
			Where(`routine.deleted_at IS NULL`).
			Where("uts.user_id = ? AND uts.permission IN ?", userId, permissions)
		return db.Where("\"RoutineTaskTable\".id IN ? AND EXISTS (?)", ids, subQuery)
	}
}

func (sc *RoutineTaskScope) IncludePreloads(preloads []schemas.RoutineTaskRelation) func(db *gorm.DB) *gorm.DB {
	return func(db *gorm.DB) *gorm.DB {
		for _, preload := range preloads {
			db = db.Preload(string(preload))
		}
		return db
	}
}
