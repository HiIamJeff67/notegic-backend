package routinetask

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	routineexecution "github.com/HiIamJeff67/notegic-backend/internal/durablejob/routinetask/execution"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type RoutineTaskResultWriter struct {
	db       *gorm.DB
	executor routineexecution.RoutineTaskExecutionServiceInterface
}

func NewRoutineTaskResultWriter(
	db *gorm.DB,
	executor routineexecution.RoutineTaskExecutionServiceInterface,
) *RoutineTaskResultWriter {
	return &RoutineTaskResultWriter{db: db, executor: executor}
}

func (w *RoutineTaskResultWriter) Write(ctx context.Context, result RoutineTaskResult) error {
	switch result.Kind {
	case RoutineTaskResultKind_Completed:
		request, ok := result.Data.(cdurablejob.MarkCompletedRoutineTasksRequestDto)
		if !ok {
			return fmt.Errorf("invalid completed RoutineTask result payload %T", result.Data)
		}
		if exception := w.executor.ApplyPreparedRoutineTasks(ctx, uuid.New(), &request); exception != nil {
			return exception
		}
		return nil
	case RoutineTaskResultKind_Failed:
		request, ok := result.Data.(cdurablejob.MarkFailedRoutineTasksRequestDto)
		if !ok {
			return fmt.Errorf("invalid failed RoutineTask result payload %T", result.Data)
		}
		return w.markFailed(ctx, &request)
	default:
		return fmt.Errorf("unsupported RoutineTask result kind %q", result.Kind)
	}
}

func (w *RoutineTaskResultWriter) markFailed(
	ctx context.Context,
	request *cdurablejob.MarkFailedRoutineTasksRequestDto,
) error {
	if request == nil || len(request.Tasks) == 0 || w.db == nil {
		return cexceptions.New("InvalidDto", "RoutineTask", "MarkFailedRoutineTasks", "The failed routine task result is invalid", http.StatusBadRequest)
	}
	tx := w.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return tx.Error
	}
	now := time.Now().UTC()
	for _, task := range request.Tasks {
		reason := task.ErrorReason
		result := tx.Model(&schemas.RoutineTask{}).
			Where("id = ? AND status = ?", task.RoutineTaskId, enums.RoutineTaskStatus_Running).
			Updates(map[string]any{
				"status":          enums.RoutineTaskStatus_Idle,
				"attempts":        0,
				"actual_ended_at": task.FailedAt,
				"updated_at":      now,
			})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
		if result.RowsAffected == 0 {
			var status enums.RoutineTaskStatus
			if err := tx.Model(&schemas.RoutineTask{}).Select("status").Where("id = ?", task.RoutineTaskId).Scan(&status).Error; err != nil {
				tx.Rollback()
				return err
			}
			if status != enums.RoutineTaskStatus_Idle {
				tx.Rollback()
				return fmt.Errorf("routine task %s is not claimable for failure finalization", task.RoutineTaskId)
			}
		}
		result = tx.Model(&schemas.RoutineTaskRecord{}).
			Where("id = ? AND status = ?", task.RoutineTaskRecordId, enums.RoutineTaskRecordStatus_Running).
			Updates(map[string]any{
				"status":          enums.RoutineTaskRecordStatus_Failed,
				"actual_ended_at": task.FailedAt,
				"error_code":      task.ErrorCode,
				"error_reason":    reason,
				"updated_at":      now,
			})
		if result.Error != nil {
			tx.Rollback()
			return result.Error
		}
		if result.RowsAffected == 0 {
			var status enums.RoutineTaskRecordStatus
			if err := tx.Model(&schemas.RoutineTaskRecord{}).Select("status").Where("id = ?", task.RoutineTaskRecordId).Scan(&status).Error; err != nil {
				tx.Rollback()
				return err
			}
			if status != enums.RoutineTaskRecordStatus_Failed {
				tx.Rollback()
				return fmt.Errorf("routine task record %s is not claimable for failure finalization", task.RoutineTaskRecordId)
			}
		}
	}
	return tx.Commit().Error
}
