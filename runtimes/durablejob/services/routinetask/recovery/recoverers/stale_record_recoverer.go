package recoverers

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	routinetasksql "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/sqls/routinetask"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
)

type StaleRecordRecovererInterface interface {
	RecoverStaleRoutineTaskRecords(context.Context, time.Time) (int64, error)
}

type StaleRecordRecoverer struct {
	db *gorm.DB
}

func NewStaleRecordRecoverer(db *gorm.DB) *StaleRecordRecoverer {
	return &StaleRecordRecoverer{db: db}
}

func (r *StaleRecordRecoverer) RecoverStaleRoutineTaskRecords(
	ctx context.Context,
	staleBefore time.Time,
) (int64, error) {
	if r.db == nil {
		return 0, fmt.Errorf("DurableJob routine task recovery database is not configured")
	}

	tx := r.db.WithContext(ctx).Begin()
	lockingStrength := "UPDATE"

	var staleRecords []struct {
		Id              uuid.UUID `gorm:"column:id"`
		RoutineRecordId uuid.UUID `gorm:"column:routine_record_id"`
		Attempts        int32     `gorm:"column:attempts"`
		MaxAttempts     int32     `gorm:"column:max_attempts"`
	}
	result := tx.Model(&sschemas.RoutineTaskRecord{}).
		Select(`"RoutineTaskRecordTable".id, "RoutineTaskRecordTable".routine_record_id,
			"RoutineTaskRecordTable".attempts, routine_task.max_attempts`).
		Joins(`INNER JOIN "RoutineTaskTable" routine_task ON routine_task.id = "RoutineTaskRecordTable".routine_task_id`).
		Where(`"RoutineTaskRecordTable".status = ?`, cenums.RoutineTaskRecordStatus_Running).
		Where(`"RoutineTaskRecordTable".actual_started_at IS NOT NULL`).
		Where(`"RoutineTaskRecordTable".actual_started_at <= ?`, staleBefore).
		Scopes(sscopes.Locking(&lockingStrength, "SKIP LOCKED")).
		Find(&staleRecords)
	if result.Error != nil {
		tx.Rollback()
		return 0, fmt.Errorf("find stale routine task records: %w", result.Error)
	}
	if len(staleRecords) == 0 {
		if err := tx.Commit().Error; err != nil {
			return 0, fmt.Errorf("commit empty routine task recovery transaction: %w", err)
		}
		return 0, nil
	}

	retryIds := make([]uuid.UUID, 0, len(staleRecords))
	failedIds := make([]uuid.UUID, 0, len(staleRecords))
	routineRecordIds := make([]uuid.UUID, 0, len(staleRecords))
	seenRoutineRecordIds := make(map[uuid.UUID]struct{}, len(staleRecords))
	for _, record := range staleRecords {
		if record.Attempts < record.MaxAttempts {
			retryIds = append(retryIds, record.Id)
		} else {
			failedIds = append(failedIds, record.Id)
		}
		if _, exists := seenRoutineRecordIds[record.RoutineRecordId]; !exists {
			seenRoutineRecordIds[record.RoutineRecordId] = struct{}{}
			routineRecordIds = append(routineRecordIds, record.RoutineRecordId)
		}
	}

	now := time.Now().UTC()
	if len(retryIds) > 0 {
		result = tx.Model(&sschemas.RoutineTaskRecord{}).
			Where("id IN ? AND status = ?", retryIds, cenums.RoutineTaskRecordStatus_Running).
			Updates(map[string]any{
				"status":            cenums.RoutineTaskRecordStatus_Ready,
				"actual_started_at": nil,
				"updated_at":        now,
			})
		if result.Error != nil {
			tx.Rollback()
			return 0, fmt.Errorf("requeue stale routine task records: %w", result.Error)
		}
	}
	if len(failedIds) > 0 {
		result = tx.Model(&sschemas.RoutineTaskRecord{}).
			Where("id IN ? AND status = ?", failedIds, cenums.RoutineTaskRecordStatus_Running).
			Updates(map[string]any{
				"status":          cenums.RoutineTaskRecordStatus_Failed,
				"error_code":      cenums.RoutineTaskRecordErrorCode_Timeout,
				"error_reason":    "routine task execution lease expired",
				"actual_ended_at": now,
				"updated_at":      now,
			})
		if result.Error != nil {
			tx.Rollback()
			return 0, fmt.Errorf("fail exhausted stale routine task records: %w", result.Error)
		}

		result = tx.Exec(
			routinetasksql.BlockRoutineTaskRecordDependenciesSQL,
			failedIds,
			cenums.RoutineTaskRecordStatus_Blocked,
			now,
			cenums.RoutineTaskRecordStatus_Waiting,
			cenums.RoutineTaskRecordStatus_Ready,
		)
		if result.Error != nil {
			tx.Rollback()
			return 0, fmt.Errorf("block dependent stale routine tasks: %w", result.Error)
		}
		failedRoutineRecordIdsQuery := tx.Model(&sschemas.RoutineTaskRecord{}).
			Select("routine_record_id").
			Where("id IN ?", failedIds)
		result = tx.Model(&sschemas.RoutineTaskRecord{}).
			Where("routine_record_id IN (?)", failedRoutineRecordIdsQuery).
			Where("status IN ?", []cenums.RoutineTaskRecordStatus{
				cenums.RoutineTaskRecordStatus_Waiting,
				cenums.RoutineTaskRecordStatus_Ready,
			}).
			Where(`EXISTS (
				SELECT 1
				FROM "RoutineTaskRecordTable" barrier_record
				INNER JOIN "RoutineTaskTable" barrier_task
					ON barrier_task.id = barrier_record.routine_task_id
				WHERE barrier_record.routine_record_id = "RoutineTaskRecordTable".routine_record_id
					AND barrier_task.purpose IN (?, ?, ?)
					AND barrier_record.status IN (?, ?)
			)`, cenums.RoutineTaskPurpose_CreateSubShelf, cenums.RoutineTaskPurpose_CreateBlockPack, cenums.RoutineTaskPurpose_CreateMaterial, cenums.RoutineTaskRecordStatus_Failed, cenums.RoutineTaskRecordStatus_Blocked).
			Updates(map[string]any{
				"status":     cenums.RoutineTaskRecordStatus_Blocked,
				"updated_at": now,
			})
		if result.Error != nil {
			tx.Rollback()
			return 0, fmt.Errorf("block tasks behind deterministic creation failure: %w", result.Error)
		}
	}

	result = tx.Exec(
		routinetasksql.UpdateRoutineRecordAggregateSQL,
		cenums.RoutineRecordStatus_Running,
		cenums.RoutineRecordStatus_Blocked,
		cenums.RoutineRecordStatus_Failed,
		cenums.RoutineRecordStatus_Success,
		now,
		now,
		routineRecordIds,
	)
	if result.Error != nil {
		tx.Rollback()
		return 0, fmt.Errorf("update recovered routine record aggregates: %w", result.Error)
	}

	routineIdsToFinalize := tx.Model(&sschemas.RoutineRecord{}).
		Select("routine_id").
		Where("id IN ? AND status IN ?", routineRecordIds, []cenums.RoutineRecordStatus{
			cenums.RoutineRecordStatus_Success,
			cenums.RoutineRecordStatus_Failed,
			cenums.RoutineRecordStatus_Blocked,
		})
	result = tx.Model(&sschemas.Routine{}).
		Where("id IN (?)", routineIdsToFinalize).
		Updates(map[string]any{
			"status": gorm.Expr(
				`CASE WHEN period IS NULL THEN ?::"RoutineStatus" ELSE ?::"RoutineStatus" END`,
				cenums.RoutineStatus_Completed,
				cenums.RoutineStatus_Scheduled,
			),
			"updated_at": now,
		})
	if result.Error != nil {
		tx.Rollback()
		return 0, fmt.Errorf("finalize recovered routines: %w", result.Error)
	}
	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("commit routine task recovery transaction: %w", err)
	}

	return int64(len(staleRecords)), nil
}
