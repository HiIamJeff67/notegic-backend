package routinetask

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	cdurablejobroutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	usersql "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/sqls/user"
)

type Claimer struct {
	validator *validator.Validate
	db        *gorm.DB
}

func NewClaimer(db *gorm.DB, validatorInstance *validator.Validate) *Claimer {
	if validatorInstance == nil {
		validatorInstance = validator.New()
	}
	return &Claimer{
		validator: validatorInstance,
		db:        db,
	}
}

func (c *Claimer) ClaimRoutineTasks(
	ctx context.Context,
	request cdurablejob.ClaimRoutineTasksRequestDto,
) (*cdurablejob.ClaimRoutineTasksResponseDto, *cexceptions.Exception) {
	if c.db == nil || request.RequestId == uuid.Nil || request.WorkerId == uuid.Nil ||
		request.BatchSize < 1 || request.BatchSize > 1000 {
		return nil, cexceptions.New(
			"InvalidDto",
			"RoutineTask",
			"Claim",
			"The routine task claim request is invalid",
			http.StatusBadRequest,
		)
	}
	if err := c.validator.Struct(request); err != nil {
		return nil, cexceptions.New(
			"InvalidDto",
			"RoutineTask",
			"Claim",
			"The routine task claim request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	tx := c.db.WithContext(ctx).Begin()
	if err := tx.Error; err != nil {
		return nil, cexceptions.New(
			"FailedToBeginTransaction",
			"RoutineTask",
			"Claim",
			"Failed to start the routine task claim transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	now := time.Now().UTC()
	var claimableRoutineTasks []sschemas.RoutineTask
	result := tx.
		Model(&sschemas.RoutineTask{}).
		Select("id, routine_id, actor_user_id, title, purpose, payload, cost_unit, priority, status, attempts, max_attempts, period, next_scheduled_at, scheduled_at, actual_started_at").
		Where("status = ?", cenums.RoutineTaskStatus_Idle).
		Where("scheduled_at <= ?", now).
		Where("attempts < max_attempts").
		Order("priority DESC, scheduled_at ASC, id ASC").
		Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
		Limit(request.BatchSize).
		Find(&claimableRoutineTasks)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"ClaimFailed",
			"RoutineTask",
			"Claim",
			"Failed to claim routine tasks",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	if len(claimableRoutineTasks) > 0 {
		actorUserIds := make([]uuid.UUID, 0, len(claimableRoutineTasks))
		seenUserIds := make(map[uuid.UUID]struct{}, len(claimableRoutineTasks))
		for _, task := range claimableRoutineTasks {
			if _, exists := seenUserIds[task.ActorUserId]; !exists {
				seenUserIds[task.ActorUserId] = struct{}{}
				actorUserIds = append(actorUserIds, task.ActorUserId)
			}
		}

		var existingQuotaUserIds []uuid.UUID
		result = tx.Model(&sschemas.UserQuota{}).
			Where("user_id IN ?", actorUserIds).
			Pluck("user_id", &existingQuotaUserIds)
		if result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "UserQuota", "Claim", "Failed to find user quotas", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}

		var users []sschemas.UserView
		userQuery := tx.Model(&sschemas.UserView{}).Select("id, created_at").Where("id IN ?", actorUserIds)
		if len(existingQuotaUserIds) > 0 {
			userQuery = userQuery.Where("id NOT IN ?", existingQuotaUserIds)
		}
		if result = userQuery.Find(&users); result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "UserQuota", "Claim", "Failed to find users without quotas", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
		if len(users) > 0 {
			quotas := make([]sschemas.UserQuota, len(users))
			for index, user := range users {
				quotas[index] = sschemas.UserQuota{
					UserId:         user.Id,
					CostUnitUsed:   0,
					CycleStartedAt: user.CreatedAt,
					NextResetAt:    user.CreatedAt.AddDate(0, 0, 30),
					UpdatedAt:      now,
					CreatedAt:      now,
				}
			}
			if result = tx.Model(&sschemas.UserQuota{}).Clauses(clause.OnConflict{DoNothing: true}).CreateInBatches(&quotas, request.BatchSize); result.Error != nil {
				tx.Rollback()
				return nil, cexceptions.New("ClaimFailed", "UserQuota", "Claim", "Failed to initialize user quotas", http.StatusInternalServerError, true).WithOrigin(result.Error)
			}
		}

		valuePlaceholders := make([]string, 0, len(claimableRoutineTasks))
		valueArgs := make([]any, 0, len(claimableRoutineTasks)*5)
		for _, task := range claimableRoutineTasks {
			valuePlaceholders = append(valuePlaceholders, "(?::uuid, ?::uuid, ?::bigint, ?::integer, ?::timestamptz)")
			valueArgs = append(valueArgs, task.Id, task.ActorUserId, task.CostUnit, task.Priority, task.ScheduledAt)
		}
		var consumedRoutineTaskIds []uuid.UUID
		result = tx.Model(&sschemas.UserQuota{}).Raw(
			fmt.Sprintf(usersql.ConsumeUserQuotaSQL, strings.Join(valuePlaceholders, ",")),
			valueArgs...,
		).Scan(&consumedRoutineTaskIds)
		if result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "UserQuota", "Claim", "Failed to consume routine task quota", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}

		consumedIds := make(map[uuid.UUID]struct{}, len(consumedRoutineTaskIds))
		for _, id := range consumedRoutineTaskIds {
			consumedIds[id] = struct{}{}
		}
		filteredTasks := make([]sschemas.RoutineTask, 0, len(claimableRoutineTasks))
		for _, task := range claimableRoutineTasks {
			if _, ok := consumedIds[task.Id]; ok {
				filteredTasks = append(filteredTasks, task)
			}
		}
		claimableRoutineTasks = filteredTasks
	}

	assignments := make([]cdurablejobroutinetasktypes.RoutineTaskAssignment, 0, len(claimableRoutineTasks))
	if len(claimableRoutineTasks) > 0 {
		claimedIds := make([]uuid.UUID, len(claimableRoutineTasks))
		recordScheduledAtByTaskId := make(map[uuid.UUID]time.Time, len(claimableRoutineTasks))
		for index, task := range claimableRoutineTasks {
			claimedIds[index] = task.Id
			recordScheduledAtByTaskId[task.Id] = task.ScheduledAt
		}

		result = tx.Model(&sschemas.RoutineTask{}).
			Where("id IN ?", claimedIds).
			Updates(map[string]any{
				"status":            cenums.RoutineTaskStatus_Running,
				"attempts":          gorm.Expr("attempts + 1"),
				"scheduled_at":      gorm.Expr(`CASE period WHEN ? THEN GREATEST(scheduled_at, next_scheduled_at) + INTERVAL '1 day' WHEN ? THEN GREATEST(scheduled_at, next_scheduled_at) + INTERVAL '7 days' WHEN ? THEN GREATEST(scheduled_at, next_scheduled_at) + INTERVAL '30 days' ELSE GREATEST(scheduled_at, next_scheduled_at) END`, cenums.RoutinePeriod_Daily, cenums.RoutinePeriod_Weekly, cenums.RoutinePeriod_Monthly),
				"next_scheduled_at": gorm.Expr(`CASE period WHEN ? THEN GREATEST(scheduled_at, next_scheduled_at) + INTERVAL '1 day' WHEN ? THEN GREATEST(scheduled_at, next_scheduled_at) + INTERVAL '7 days' WHEN ? THEN GREATEST(scheduled_at, next_scheduled_at) + INTERVAL '30 days' ELSE GREATEST(scheduled_at, next_scheduled_at) END`, cenums.RoutinePeriod_Daily, cenums.RoutinePeriod_Weekly, cenums.RoutinePeriod_Monthly),
				"actual_started_at": now,
				"actual_ended_at":   nil,
				"updated_at":        now,
			})
		if result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "RoutineTask", "Claim", "Failed to update claimed routine tasks", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}

		var claimedTasks []sschemas.RoutineTask
		if result = tx.Model(&sschemas.RoutineTask{}).Where("id IN ?", claimedIds).Find(&claimedTasks); result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "RoutineTask", "Claim", "Failed to retrieve claimed routine tasks", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
		actorUserIds := make([]uuid.UUID, 0, len(claimedTasks))
		seenActorUserIds := make(map[uuid.UUID]struct{}, len(claimedTasks))
		for _, task := range claimedTasks {
			if _, exists := seenActorUserIds[task.ActorUserId]; !exists {
				seenActorUserIds[task.ActorUserId] = struct{}{}
				actorUserIds = append(actorUserIds, task.ActorUserId)
			}
		}
		var actorUsers []sschemas.UserView
		if result = tx.Model(&sschemas.UserView{}).Select("id, public_id").Where("id IN ?", actorUserIds).Find(&actorUsers); result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "RoutineTask", "Claim", "Failed to retrieve routine task users", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
		actorUserPublicIds := make(map[uuid.UUID]uuid.UUID, len(actorUsers))
		for _, user := range actorUsers {
			actorUserPublicIds[user.Id] = user.PublicId
		}
		records := make([]sschemas.RoutineTaskRecord, len(claimedTasks))
		for index, task := range claimedTasks {
			records[index] = sschemas.RoutineTaskRecord{
				Id: uuid.New(), RoutineTaskId: task.Id, Purpose: task.Purpose, Status: cenums.RoutineTaskRecordStatus_Running,
				CostUnit: task.CostUnit, TotalAttempts: int64(task.Attempts), ScheduledAt: recordScheduledAtByTaskId[task.Id], ActualStartedAt: task.ActualStartedAt,
			}
		}
		if result = tx.CreateInBatches(&records, request.BatchSize); result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "RoutineTaskRecord", "Claim", "Failed to create routine task records", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
		recordIdByTaskId := make(map[uuid.UUID]uuid.UUID, len(records))
		for _, record := range records {
			recordIdByTaskId[record.RoutineTaskId] = record.Id
		}
		for _, task := range claimedTasks {
			startedAt := now
			if task.ActualStartedAt != nil {
				startedAt = *task.ActualStartedAt
			}
			patternValues := map[string]string{}
			var payload struct {
				Pattern cdurablejobroutinetasktypes.RoutineTaskPattern `json:"pattern"`
			}
			if err := json.Unmarshal(task.Payload, &payload); err != nil {
				tx.Rollback()
				return nil, cexceptions.New("InvalidRoutineTaskPayload", "RoutineTask", "Claim", "Routine task payload is invalid", http.StatusBadRequest).WithOrigin(err)
			}
			for key, binding := range payload.Pattern {
				switch binding.Source {
				case "scheduledAt":
					scheduledAt := recordScheduledAtByTaskId[task.Id]
					if binding.Timezone != nil && *binding.Timezone != "" {
						location, err := time.LoadLocation(*binding.Timezone)
						if err != nil {
							continue
						}
						scheduledAt = scheduledAt.In(location)
					}
					format := time.RFC3339
					if binding.Format != nil && *binding.Format != "" {
						format = *binding.Format
					}
					patternValues[key] = scheduledAt.Format(format)
				case "recordId":
					patternValues[key] = recordIdByTaskId[task.Id].String()
				case "shortRecordId":
					recordId := recordIdByTaskId[task.Id].String()
					if len(recordId) > 8 {
						recordId = recordId[:8]
					}
					patternValues[key] = recordId
				case "routineTaskId":
					patternValues[key] = task.Id.String()
				}
			}
			assignments = append(assignments, cdurablejobroutinetasktypes.RoutineTaskAssignment{
				RoutineTaskId: task.Id, RoutineTaskRecordId: recordIdByTaskId[task.Id], RoutineId: task.RoutineId,
				ActorUserId: task.ActorUserId, ActorUserPublicId: actorUserPublicIds[task.ActorUserId], Title: task.Title,
				Purpose: task.Purpose, Payload: json.RawMessage(task.Payload), CostUnit: task.CostUnit,
				Priority: task.Priority, Attempt: task.Attempts, ScheduledAt: recordScheduledAtByTaskId[task.Id], StartedAt: startedAt,
				PatternValues: patternValues,
			})
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New("FailedToCommitTransaction", "RoutineTask", "Claim", "Failed to commit the routine task claim transaction", http.StatusInternalServerError, true).WithOrigin(err)
	}
	return &cdurablejob.ClaimRoutineTasksResponseDto{
		RequestId:   request.RequestId,
		WorkerId:    request.WorkerId,
		Assignments: assignments,
	}, nil
}
