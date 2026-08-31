package routinetask

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type RoutineTaskRecoveryServiceInterface interface {
	RecoverStaleRoutineTaskRecords(context.Context, time.Time) (int64, error)
}

type RoutineTaskRecoveryService struct {
	db *gorm.DB
}

func NewRoutineTaskRecoveryService(db *gorm.DB) *RoutineTaskRecoveryService {
	return &RoutineTaskRecoveryService{db: db}
}

func (s *RoutineTaskRecoveryService) RecoverStaleRoutineTaskRecords(
	ctx context.Context,
	staleBefore time.Time,
) (int64, error) {
	if s.db == nil {
		return 0, fmt.Errorf("DurableJob routine task recovery database is not configured")
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return 0, fmt.Errorf("begin routine task recovery transaction: %w", tx.Error)
	}

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
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
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

		result = tx.Exec(`WITH RECURSIVE blocked_tasks(routine_record_id, routine_task_id) AS (
			SELECT routine_record_id, routine_task_id
			FROM "RoutineTaskRecordTable"
			WHERE id IN ?
			UNION
			SELECT blocked_tasks.routine_record_id, dependency.routine_task_id
			FROM blocked_tasks
			INNER JOIN "RoutineDependencyTable" dependency
				ON dependency.previous_routine_task_id = blocked_tasks.routine_task_id
		)
		UPDATE "RoutineTaskRecordTable" AS routine_task_record
		SET status = ?, updated_at = ?
		FROM blocked_tasks
		WHERE routine_task_record.routine_record_id = blocked_tasks.routine_record_id
			AND routine_task_record.routine_task_id = blocked_tasks.routine_task_id
			AND routine_task_record.status IN (?, ?)`,
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
	}

	result = tx.Exec(`UPDATE "RoutineRecordTable" AS routine_record
		SET success_task_count = counts.success_task_count,
			failed_task_count = counts.failed_task_count,
			blocked_task_count = counts.blocked_task_count,
			running_task_count = counts.running_task_count,
			waiting_task_count = counts.waiting_task_count,
			status = CASE
				WHEN counts.running_task_count > 0 OR counts.waiting_task_count > 0 THEN ?::"RoutineRecordStatus"
				WHEN counts.failed_task_count > 0 OR counts.blocked_task_count > 0 THEN ?::"RoutineRecordStatus"
				ELSE ?::"RoutineRecordStatus"
			END,
			actual_ended_at = CASE
				WHEN counts.running_task_count = 0 AND counts.waiting_task_count = 0 THEN ?
				ELSE routine_record.actual_ended_at
			END,
			updated_at = ?
		FROM (
			SELECT routine_record_id,
				COUNT(*) FILTER (WHERE status = 'Success')::integer AS success_task_count,
				COUNT(*) FILTER (WHERE status = 'Failed')::integer AS failed_task_count,
				COUNT(*) FILTER (WHERE status = 'Blocked')::integer AS blocked_task_count,
				COUNT(*) FILTER (WHERE status = 'Running')::integer AS running_task_count,
				COUNT(*) FILTER (WHERE status IN ('Waiting', 'Ready'))::integer AS waiting_task_count
			FROM "RoutineTaskRecordTable"
			WHERE routine_record_id IN ?
			GROUP BY routine_record_id
		) counts
		WHERE routine_record.id = counts.routine_record_id`,
		cenums.RoutineRecordStatus_Running,
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

	result = tx.Exec(`UPDATE "RoutineTable" AS routine
		SET status = CASE WHEN routine.period IS NULL THEN ? ELSE ? END, updated_at = ?
		WHERE routine.id IN (
			SELECT routine_record.routine_id
			FROM "RoutineRecordTable" routine_record
			WHERE routine_record.id IN ?
				AND routine_record.status IN (?, ?, ?)
		)`,
		cenums.RoutineStatus_Completed,
		cenums.RoutineStatus_Scheduled,
		now,
		routineRecordIds,
		cenums.RoutineRecordStatus_Success,
		cenums.RoutineRecordStatus_Failed,
		cenums.RoutineRecordStatus_Blocked,
	)
	if result.Error != nil {
		tx.Rollback()
		return 0, fmt.Errorf("finalize recovered routines: %w", result.Error)
	}
	if err := tx.Commit().Error; err != nil {
		return 0, fmt.Errorf("commit routine task recovery transaction: %w", err)
	}

	return int64(len(staleRecords)), nil
}
