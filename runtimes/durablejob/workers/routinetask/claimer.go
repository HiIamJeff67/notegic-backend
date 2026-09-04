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

	routinetasksql "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/sqls/routinetask"
	usersql "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/sqls/user"
	routinetaskdependencies "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/dependencies"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
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

func (c *Claimer) ClaimRoutines(
	ctx context.Context,
	request cdurablejob.ClaimRoutinesRequestDto,
) (*cdurablejob.ClaimRoutinesResponseDto, *cexceptions.Exception) {
	if c.db == nil || request.RequestId == uuid.Nil || request.WorkerId == uuid.Nil ||
		request.BatchSize < 1 || request.BatchSize > 1000 {
		return nil, cexceptions.New("InvalidDto", "RoutineTask", "Claim", "The routine task claim request is invalid", http.StatusBadRequest)
	}
	if err := c.validator.Struct(request); err != nil {
		return nil, cexceptions.New("InvalidDto", "RoutineTask", "Claim", "The routine task claim request is invalid", http.StatusBadRequest).WithOrigin(err)
	}

	tx := c.db.WithContext(ctx).Begin()
	lockingStrength := "UPDATE"

	now := time.Now().UTC()
	var dueRoutines []sschemas.Routine
	result := tx.Model(&sschemas.Routine{}).
		Select("id, title, description, is_pinned, scheduled_start_at, scheduled_end_at, period, timezone, status, definition_version").
		Where("deleted_at IS NULL").
		Where("status = ?", cenums.RoutineStatus_Scheduled).
		Where("scheduled_start_at <= ?", now).
		Order("scheduled_start_at ASC, id ASC").
		Scopes(sscopes.Locking(&lockingStrength, "SKIP LOCKED")).
		Limit(request.BatchSize).
		Find(&dueRoutines)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New("ClaimFailed", "Routine", "Claim", "Failed to find due routines", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}

	routineIds := make([]uuid.UUID, len(dueRoutines))
	routineRecordIds := make([]uuid.UUID, len(dueRoutines))
	routineRecordByRoutineId := make(map[uuid.UUID]sschemas.RoutineRecord, len(dueRoutines))
	routineSnapshots := make(map[uuid.UUID]map[string]any, len(dueRoutines))
	recordScheduledAt := make([]time.Time, len(dueRoutines))
	records := make([]sschemas.RoutineRecord, len(dueRoutines))
	for index, routine := range dueRoutines {
		routineIds[index] = routine.Id
		recordScheduledAt[index] = routine.ScheduledStartAt
		routineSnapshots[routine.Id] = map[string]any{
			"id":                routine.Id,
			"definitionVersion": routine.DefinitionVersion,
			"title":             routine.Title,
			"description":       routine.Description,
			"isPinned":          routine.IsPinned,
			"scheduledStartAt":  routine.ScheduledStartAt,
			"scheduledEndAt":    routine.ScheduledEndAt,
			"period":            routine.Period,
			"timezone":          routine.Timezone,
			"routineTasks":      []any{},
		}
		records[index] = sschemas.RoutineRecord{
			Id:                uuid.New(),
			RoutineId:         routine.Id,
			DefinitionVersion: routine.DefinitionVersion,
			Status:            cenums.RoutineRecordStatus_Pending,
			ScheduledAt:       routine.ScheduledStartAt,
			Snapshot:          []byte("{}"),
		}
	}

	if len(dueRoutines) > 0 {
		result = tx.Model(&sschemas.RoutineRecord{}).
			Clauses(clause.OnConflict{DoNothing: true}).
			CreateInBatches(&records, request.BatchSize)
		if result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "RoutineRecord", "Claim", "Failed to create routine records", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
		result = tx.Model(&sschemas.Routine{}).
			Where("id IN ?", routineIds).
			Updates(map[string]any{
				"status": cenums.RoutineStatus_InProgress,
				"phase":  cenums.RoutinePhase_Claimed,
				"scheduled_start_at": gorm.Expr(`CASE period
				WHEN ? THEN scheduled_start_at + INTERVAL '1 day'
				WHEN ? THEN scheduled_start_at + INTERVAL '7 days'
				WHEN ? THEN scheduled_start_at + INTERVAL '30 days'
				ELSE scheduled_start_at END`, cenums.RoutinePeriod_Daily, cenums.RoutinePeriod_Weekly, cenums.RoutinePeriod_Monthly),
				"scheduled_end_at": gorm.Expr(`CASE period
				WHEN ? THEN scheduled_end_at + INTERVAL '1 day'
				WHEN ? THEN scheduled_end_at + INTERVAL '7 days'
				WHEN ? THEN scheduled_end_at + INTERVAL '30 days'
				ELSE scheduled_end_at END`, cenums.RoutinePeriod_Daily, cenums.RoutinePeriod_Weekly, cenums.RoutinePeriod_Monthly),
				"updated_at": now,
			})
		if result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "Routine", "Claim", "Failed to advance routine schedules", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}

		var storedRecords []sschemas.RoutineRecord
		result = tx.Model(&sschemas.RoutineRecord{}).
			Where("routine_id IN ? AND scheduled_at IN ?", routineIds, recordScheduledAt).
			Find(&storedRecords)
		if result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "RoutineRecord", "Claim", "Failed to retrieve routine records", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
		routineRecordByRoutineId = make(map[uuid.UUID]sschemas.RoutineRecord, len(storedRecords))
		for _, routine := range dueRoutines {
			for _, record := range storedRecords {
				if record.RoutineId == routine.Id &&
					record.DefinitionVersion == routine.DefinitionVersion &&
					record.ScheduledAt.Equal(routine.ScheduledStartAt) {
					routineRecordByRoutineId[routine.Id] = record
					break
				}
			}
			if _, exists := routineRecordByRoutineId[routine.Id]; !exists {
				tx.Rollback()
				return nil, cexceptions.New("ClaimFailed", "RoutineRecord", "Claim", "Failed to resolve routine records", http.StatusInternalServerError, true)
			}
		}
		for index, routine := range dueRoutines {
			routineRecordIds[index] = routineRecordByRoutineId[routine.Id].Id
		}

		var routineTasks []sschemas.RoutineTask
		result = tx.Model(&sschemas.RoutineTask{}).
			Where("routine_id IN ?", routineIds).
			Find(&routineTasks)
		if result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "RoutineTask", "Claim", "Failed to retrieve routine tasks", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
		result = tx.Model(&sschemas.RoutineTask{}).
			Where("routine_id IN ?", routineIds).
			Update("phase", cenums.RoutinePhase_Claimed)
		if result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "RoutineTask", "Claim", "Failed to update claimed routine task phases", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
		if len(routineTasks) > 0 {
			taskIds := make([]uuid.UUID, len(routineTasks))
			for index, task := range routineTasks {
				taskIds[index] = task.Id
			}
			var dependencies []sschemas.RoutineTaskDependency
			result = tx.Model(&sschemas.RoutineTaskDependency{}).
				Where("routine_task_id IN ?", taskIds).
				Find(&dependencies)
			if result.Error != nil {
				tx.Rollback()
				return nil, cexceptions.New("ClaimFailed", "RoutineTaskDependency", "Claim", "Failed to retrieve routine task dependencies", http.StatusInternalServerError, true).WithOrigin(result.Error)
			}
			previousCountByTaskId := make(map[uuid.UUID]int, len(taskIds))
			previousTaskIdsByTaskId := make(map[uuid.UUID][]uuid.UUID, len(taskIds))
			taskPurposeById := make(map[uuid.UUID]cenums.RoutineTaskPurpose, len(routineTasks))
			dependencyEdges := make([]routinetaskdependencies.Edge, len(dependencies))
			for _, task := range routineTasks {
				taskPurposeById[task.Id] = task.Purpose
			}
			for index, dependency := range dependencies {
				previousCountByTaskId[dependency.RoutineTaskId]++
				previousTaskIdsByTaskId[dependency.RoutineTaskId] = append(
					previousTaskIdsByTaskId[dependency.RoutineTaskId],
					dependency.PreviousRoutineTaskId,
				)
				dependencyEdges[index] = routinetaskdependencies.Edge{
					TaskId:         dependency.RoutineTaskId,
					PreviousTaskId: dependency.PreviousRoutineTaskId,
				}
			}
			dependencyGraphError := routinetaskdependencies.Validate(taskIds, dependencyEdges)
			taskRecords := make([]sschemas.RoutineTaskRecord, len(routineTasks))
			taskRecordIndex := 0
			for _, task := range routineTasks {
				record := routineRecordByRoutineId[task.RoutineId]
				status := cenums.RoutineTaskRecordStatus_Ready
				if previousCountByTaskId[task.Id] > 0 {
					status = cenums.RoutineTaskRecordStatus_Waiting
				}
				if dependencyGraphError == nil &&
					(task.Purpose == cenums.RoutineTaskPurpose_CreateSubShelf ||
						task.Purpose == cenums.RoutineTaskPurpose_CreateBlockPack ||
						task.Purpose == cenums.RoutineTaskPurpose_CreateMaterial) {
					status = cenums.RoutineTaskRecordStatus_Ready
					for _, previousTaskId := range previousTaskIdsByTaskId[task.Id] {
						previousPurpose := taskPurposeById[previousTaskId]
						if previousPurpose != cenums.RoutineTaskPurpose_CreateSubShelf &&
							previousPurpose != cenums.RoutineTaskPurpose_CreateBlockPack &&
							previousPurpose != cenums.RoutineTaskPurpose_CreateMaterial {
							status = cenums.RoutineTaskRecordStatus_Waiting
							break
						}
					}
				}
				taskRecords[taskRecordIndex] = sschemas.RoutineTaskRecord{
					Id:              uuid.New(),
					RoutineRecordId: record.Id,
					RoutineTaskId:   task.Id,
					Purpose:         task.Purpose,
					Status:          status,
					CostUnit:        task.CostUnit,
					PayloadSnapshot: task.Payload,
					ResultSnapshot:  []byte("{}"),
				}
				routineSnapshots[task.RoutineId]["routineTasks"] = append(
					routineSnapshots[task.RoutineId]["routineTasks"].([]any),
					map[string]any{
						"id":                     task.Id,
						"title":                  task.Title,
						"purpose":                task.Purpose,
						"payload":                json.RawMessage(task.Payload),
						"costUnit":               task.CostUnit,
						"priority":               task.Priority,
						"maxAttempts":            task.MaxAttempts,
						"previousRoutineTaskIds": previousTaskIdsByTaskId[task.Id],
					},
				)
				taskRecordIndex++
			}
			if taskRecordIndex > 0 {
				taskRecords = taskRecords[:taskRecordIndex]
				result = tx.Model(&sschemas.RoutineTaskRecord{}).
					Clauses(clause.OnConflict{DoNothing: true}).
					CreateInBatches(&taskRecords, len(taskRecords))
				if result.Error != nil {
					tx.Rollback()
					return nil, cexceptions.New("ClaimFailed", "RoutineTaskRecord", "Claim", "Failed to create routine task records", http.StatusInternalServerError, true).WithOrigin(result.Error)
				}
			}
			result = tx.Exec(routinetasksql.UpdateRoutineRecordInitialAggregateSQL, now, routineRecordIds)
			if result.Error != nil {
				tx.Rollback()
				return nil, cexceptions.New("ClaimFailed", "RoutineRecord", "Claim", "Failed to update routine record aggregates", http.StatusInternalServerError, true).WithOrigin(result.Error)
			}
		}
		snapshotPlaceholders := make([]string, 0, len(routineSnapshots))
		snapshotArgs := make([]any, 0, len(routineSnapshots)*2)
		for routineId, snapshot := range routineSnapshots {
			snapshotData, err := json.Marshal(snapshot)
			if err != nil {
				tx.Rollback()
				return nil, cexceptions.New("ClaimFailed", "RoutineRecord", "Claim", "Failed to encode routine record snapshots", http.StatusInternalServerError, true).WithOrigin(err)
			}
			snapshotPlaceholders = append(snapshotPlaceholders, "(?::uuid, ?::jsonb)")
			snapshotArgs = append(snapshotArgs, routineRecordByRoutineId[routineId].Id, snapshotData)
		}
		if len(snapshotPlaceholders) > 0 {
			result = tx.Exec(fmt.Sprintf(routinetasksql.UpdateRoutineRecordSnapshotSQL, strings.Join(snapshotPlaceholders, ",")), append([]any{now}, snapshotArgs...)...)
			if result.Error != nil {
				tx.Rollback()
				return nil, cexceptions.New("ClaimFailed", "RoutineRecord", "Claim", "Failed to store routine record snapshots", http.StatusInternalServerError, true).WithOrigin(result.Error)
			}
		}
	}
	if len(routineRecordIds) > 0 {
		result = tx.Model(&sschemas.RoutineRecord{}).
			Where("id IN ? AND status = ? AND total_task_count = 0", routineRecordIds, cenums.RoutineRecordStatus_Pending).
			Updates(map[string]any{
				"status":          cenums.RoutineRecordStatus_Success,
				"actual_ended_at": now,
				"updated_at":      now,
			})
		if result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "RoutineRecord", "Claim", "Failed to finalize empty routine records", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
		routineIdsToFinalize := tx.Model(&sschemas.RoutineRecord{}).
			Select("routine_id").
			Where("id IN ? AND status = ?", routineRecordIds, cenums.RoutineRecordStatus_Success)
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
			return nil, cexceptions.New("ClaimFailed", "Routine", "Claim", "Failed to finalize routine schedules", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
	}

	var claimableRoutineRecords []sschemas.RoutineRecord
	result = tx.Model(&sschemas.RoutineRecord{}).
		Where(`"RoutineRecordTable".status IN ?`, []cenums.RoutineRecordStatus{
			cenums.RoutineRecordStatus_Pending,
			cenums.RoutineRecordStatus_Running,
		}).
		Where(`EXISTS (
			SELECT 1
			FROM "RoutineTaskRecordTable" ready_record
			INNER JOIN "RoutineTaskTable" ready_task
				ON ready_task.id = ready_record.routine_task_id
			WHERE ready_record.routine_record_id = "RoutineRecordTable".id
				AND ready_record.status = ?
				AND ready_record.attempts < ready_task.max_attempts
				AND (
					(
						ready_task.purpose IN (?, ?, ?)
						AND (
							ready_task.purpose <> ?
							OR NOT EXISTS (
								SELECT 1
								FROM "RoutineTaskRecordTable" shelf_record
								INNER JOIN "RoutineTaskTable" shelf_task
									ON shelf_task.id = shelf_record.routine_task_id
								WHERE shelf_record.routine_record_id = ready_record.routine_record_id
									AND shelf_task.purpose = ?
									AND shelf_record.status <> ?
							)
						)
					)
					OR (
						ready_task.purpose NOT IN (?, ?, ?)
						AND NOT EXISTS (
							SELECT 1
							FROM "RoutineTaskRecordTable" deterministic_record
							INNER JOIN "RoutineTaskTable" deterministic_task
								ON deterministic_task.id = deterministic_record.routine_task_id
							WHERE deterministic_record.routine_record_id = ready_record.routine_record_id
								AND deterministic_task.purpose IN (?, ?, ?)
								AND deterministic_record.status <> ?
						)
					)
				)
		)`,
			cenums.RoutineTaskRecordStatus_Ready,
			cenums.RoutineTaskPurpose_CreateSubShelf,
			cenums.RoutineTaskPurpose_CreateBlockPack,
			cenums.RoutineTaskPurpose_CreateMaterial,
			cenums.RoutineTaskPurpose_CreateBlockPack,
			cenums.RoutineTaskPurpose_CreateSubShelf,
			cenums.RoutineTaskRecordStatus_Success,
			cenums.RoutineTaskPurpose_CreateSubShelf,
			cenums.RoutineTaskPurpose_CreateBlockPack,
			cenums.RoutineTaskPurpose_CreateMaterial,
			cenums.RoutineTaskPurpose_CreateSubShelf,
			cenums.RoutineTaskPurpose_CreateBlockPack,
			cenums.RoutineTaskPurpose_CreateMaterial,
			cenums.RoutineTaskRecordStatus_Success,
		).
		Order(`"RoutineRecordTable".created_at ASC, "RoutineRecordTable".id ASC`).
		Scopes(sscopes.Locking(&lockingStrength, "SKIP LOCKED")).
		Limit(request.BatchSize).
		Find(&claimableRoutineRecords)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New("ClaimFailed", "RoutineRecord", "Claim", "Failed to find claimable routines", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	claimableRoutineRecordIds := make([]uuid.UUID, len(claimableRoutineRecords))
	for index, record := range claimableRoutineRecords {
		claimableRoutineRecordIds[index] = record.Id
	}

	var readyRecords []sschemas.RoutineTaskRecord
	if len(claimableRoutineRecordIds) > 0 {
		result = tx.Model(&sschemas.RoutineTaskRecord{}).
			Select(`"RoutineTaskRecordTable".*`).
			Joins(`INNER JOIN "RoutineTaskTable" routine_task ON routine_task.id = "RoutineTaskRecordTable".routine_task_id`).
			Where(`"RoutineTaskRecordTable".routine_record_id IN ?`, claimableRoutineRecordIds).
			Where(`"RoutineTaskRecordTable".status = ?`, cenums.RoutineTaskRecordStatus_Ready).
			Where(`"RoutineTaskRecordTable".attempts < routine_task.max_attempts`).
			Where(`(
				(
					routine_task.purpose IN (?, ?, ?)
					AND (
						routine_task.purpose <> ?
						OR NOT EXISTS (
							SELECT 1
							FROM "RoutineTaskRecordTable" shelf_record
							INNER JOIN "RoutineTaskTable" shelf_task
								ON shelf_task.id = shelf_record.routine_task_id
							WHERE shelf_record.routine_record_id = "RoutineTaskRecordTable".routine_record_id
								AND shelf_task.purpose = ?
								AND shelf_record.status <> ?
						)
					)
				)
				OR (
					routine_task.purpose NOT IN (?, ?, ?)
					AND NOT EXISTS (
						SELECT 1
						FROM "RoutineTaskRecordTable" deterministic_record
						INNER JOIN "RoutineTaskTable" deterministic_task
							ON deterministic_task.id = deterministic_record.routine_task_id
						WHERE deterministic_record.routine_record_id = "RoutineTaskRecordTable".routine_record_id
							AND deterministic_task.purpose IN (?, ?, ?)
							AND deterministic_record.status <> ?
						)
				)
			)`,
				cenums.RoutineTaskPurpose_CreateSubShelf,
				cenums.RoutineTaskPurpose_CreateBlockPack,
				cenums.RoutineTaskPurpose_CreateMaterial,
				cenums.RoutineTaskPurpose_CreateBlockPack,
				cenums.RoutineTaskPurpose_CreateSubShelf,
				cenums.RoutineTaskRecordStatus_Success,
				cenums.RoutineTaskPurpose_CreateSubShelf,
				cenums.RoutineTaskPurpose_CreateBlockPack,
				cenums.RoutineTaskPurpose_CreateMaterial,
				cenums.RoutineTaskPurpose_CreateSubShelf,
				cenums.RoutineTaskPurpose_CreateBlockPack,
				cenums.RoutineTaskPurpose_CreateMaterial,
				cenums.RoutineTaskRecordStatus_Success,
			).
			Order(`routine_task.priority DESC, "RoutineTaskRecordTable".created_at ASC, "RoutineTaskRecordTable".id ASC`).
			Scopes(sscopes.Locking(&lockingStrength, "SKIP LOCKED")).
			Find(&readyRecords)
		if result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "RoutineTaskRecord", "Claim", "Failed to find routine tasks under claimed routines", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
	}
	if len(readyRecords) == 0 {
		routineAssignments := make([]cdurablejobroutinetasktypes.RoutineAssignment, 0, len(dueRoutines))
		for _, routine := range dueRoutines {
			record, exists := routineRecordByRoutineId[routine.Id]
			tasks, hasTasks := routineSnapshots[routine.Id]["routineTasks"].([]any)
			if !exists || !hasTasks || len(tasks) == 0 ||
				(record.Status != cenums.RoutineRecordStatus_Pending && record.Status != cenums.RoutineRecordStatus_Running) {
				continue
			}
			routineAssignments = append(routineAssignments, cdurablejobroutinetasktypes.RoutineAssignment{
				RoutineId:         routine.Id,
				RoutineRecordId:   record.Id,
				DefinitionVersion: routine.DefinitionVersion,
				ScheduledAt:       record.ScheduledAt,
				RoutineTasks:      []cdurablejobroutinetasktypes.RoutineTaskAssignment{},
			})
		}
		if err := tx.Commit().Error; err != nil {
			return nil, cexceptions.New("FailedToCommitTransaction", "RoutineTask", "Claim", "Failed to commit the routine task claim transaction", http.StatusInternalServerError, true).WithOrigin(err)
		}
		return &cdurablejob.ClaimRoutinesResponseDto{
			RequestId:          request.RequestId,
			WorkerId:           request.WorkerId,
			RoutineAssignments: routineAssignments,
		}, nil
	}

	taskIds := make([]uuid.UUID, len(readyRecords))
	for index, record := range readyRecords {
		taskIds[index] = record.RoutineTaskId
	}
	var tasks []sschemas.RoutineTask
	result = tx.Model(&sschemas.RoutineTask{}).
		Where("id IN ?", taskIds).
		Find(&tasks)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New("ClaimFailed", "RoutineTask", "Claim", "Failed to retrieve ready routine tasks", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	taskById := make(map[uuid.UUID]sschemas.RoutineTask, len(tasks))
	for _, task := range tasks {
		taskById[task.Id] = task
	}
	recordRoutineIds := make([]uuid.UUID, 0, len(readyRecords))
	for _, record := range readyRecords {
		recordRoutineIds = append(recordRoutineIds, record.RoutineRecordId)
	}
	var routineRecords []sschemas.RoutineRecord
	result = tx.Model(&sschemas.RoutineRecord{}).
		Where("id IN ?", recordRoutineIds).
		Find(&routineRecords)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New("ClaimFailed", "RoutineRecord", "Claim", "Failed to retrieve ready routine record schedules", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	routineIdByRecordId := make(map[uuid.UUID]uuid.UUID, len(routineRecords))
	definitionVersionByRecordId := make(map[uuid.UUID]int64, len(routineRecords))
	scheduledAtByRecordId := make(map[uuid.UUID]time.Time, len(routineRecords))
	for _, record := range routineRecords {
		routineIdByRecordId[record.Id] = record.RoutineId
		definitionVersionByRecordId[record.Id] = record.DefinitionVersion
		scheduledAtByRecordId[record.Id] = record.ScheduledAt
	}

	actorUserIds := make([]uuid.UUID, 0, len(tasks))
	seenActorUserIds := make(map[uuid.UUID]struct{}, len(tasks))
	for _, task := range tasks {
		if _, exists := seenActorUserIds[task.ActorUserId]; !exists {
			seenActorUserIds[task.ActorUserId] = struct{}{}
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
	userQuery := tx.Model(&sschemas.UserView{}).
		Select("id, created_at").
		Where("id IN ?", actorUserIds)
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
				CycleStartedAt: user.CreatedAt,
				NextResetAt:    user.CreatedAt.AddDate(0, 0, 30),
				UpdatedAt:      now,
				CreatedAt:      now,
			}
		}
		if result = tx.Model(&sschemas.UserQuota{}).
			Clauses(clause.OnConflict{DoNothing: true}).
			CreateInBatches(&quotas, len(quotas)); result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New("ClaimFailed", "UserQuota", "Claim", "Failed to initialize user quotas", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
	}
	var actorUsers []sschemas.UserView
	result = tx.Model(&sschemas.UserView{}).
		Select("id, public_id").
		Where("id IN ?", actorUserIds).
		Find(&actorUsers)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New("ClaimFailed", "RoutineTask", "Claim", "Failed to retrieve routine task users", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	actorUserPublicIds := make(map[uuid.UUID]uuid.UUID, len(actorUsers))
	for _, user := range actorUsers {
		actorUserPublicIds[user.Id] = user.PublicId
	}

	valuePlaceholders := make([]string, 0, len(readyRecords))
	valueArgs := make([]any, 0, len(readyRecords)*5)
	for _, record := range readyRecords {
		task := taskById[record.RoutineTaskId]
		valuePlaceholders = append(valuePlaceholders, "(?::uuid, ?::uuid, ?::bigint, ?::integer, ?::timestamptz)")
		valueArgs = append(valueArgs, record.Id, task.ActorUserId, task.CostUnit, task.Priority, scheduledAtByRecordId[record.RoutineRecordId])
	}
	var consumedRecordIds []uuid.UUID
	result = tx.Model(&sschemas.UserQuota{}).
		Raw(
			fmt.Sprintf(usersql.ConsumeUserQuotaSQL, strings.Join(valuePlaceholders, ",")),
			valueArgs...,
		).
		Scan(&consumedRecordIds)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New("ClaimFailed", "UserQuota", "Claim", "Failed to consume routine task quota", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	if len(consumedRecordIds) == 0 {
		if err := tx.Commit().Error; err != nil {
			return nil, cexceptions.New("FailedToCommitTransaction", "RoutineTask", "Claim", "Failed to commit the routine task claim transaction", http.StatusInternalServerError, true).WithOrigin(err)
		}
		return &cdurablejob.ClaimRoutinesResponseDto{
			RequestId:          request.RequestId,
			WorkerId:           request.WorkerId,
			RoutineAssignments: []cdurablejobroutinetasktypes.RoutineAssignment{},
		}, nil
	}

	result = tx.Model(&sschemas.RoutineTaskRecord{}).
		Where("id IN ? AND status = ?", consumedRecordIds, cenums.RoutineTaskRecordStatus_Ready).
		Updates(map[string]any{
			"status":            cenums.RoutineTaskRecordStatus_Running,
			"attempts":          gorm.Expr("attempts + 1"),
			"actual_started_at": now,
			"updated_at":        now,
		})
	if result.Error != nil || result.RowsAffected != int64(len(consumedRecordIds)) {
		tx.Rollback()
		if result.Error != nil {
			return nil, cexceptions.New("ClaimFailed", "RoutineTaskRecord", "Claim", "Failed to claim routine task records", http.StatusInternalServerError, true).WithOrigin(result.Error)
		}
		return nil, cexceptions.New("ClaimFailed", "RoutineTaskRecord", "Claim", "Routine task record claim count does not match the quota claim", http.StatusConflict, true)
	}
	result = tx.Model(&sschemas.RoutineRecord{}).
		Where("id IN (SELECT routine_record_id FROM \"RoutineTaskRecordTable\" WHERE id IN ?) AND status = ?", consumedRecordIds, cenums.RoutineRecordStatus_Pending).
		Updates(map[string]any{
			"status":            cenums.RoutineRecordStatus_Running,
			"actual_started_at": now,
			"updated_at":        now,
		})
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New("ClaimFailed", "RoutineRecord", "Claim", "Failed to start routine records", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	routineIdsForPhase := tx.Model(&sschemas.RoutineRecord{}).
		Select("routine_id").
		Where("id IN ?", consumedRecordIds)
	result = tx.Model(&sschemas.Routine{}).
		Where("id IN (?)", routineIdsForPhase).
		Update("phase", cenums.RoutinePhase_Claimed)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New("ClaimFailed", "Routine", "Claim", "Failed to update claimed routine phases", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	result = tx.Model(&sschemas.RoutineTask{}).
		Where("routine_id IN (?)", routineIdsForPhase).
		Update("phase", cenums.RoutinePhase_Claimed)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New("ClaimFailed", "RoutineTask", "Claim", "Failed to update claimed routine task phases", http.StatusInternalServerError, true).WithOrigin(result.Error)
	}
	consumedIds := make(map[uuid.UUID]struct{}, len(consumedRecordIds))
	for _, id := range consumedRecordIds {
		consumedIds[id] = struct{}{}
	}
	routineAssignments := make([]cdurablejobroutinetasktypes.RoutineAssignment, 0, len(recordRoutineIds))
	routineAssignmentIndexByRecordId := make(map[uuid.UUID]int, len(recordRoutineIds))
	for _, record := range readyRecords {
		if _, ok := consumedIds[record.Id]; !ok {
			continue
		}
		task := taskById[record.RoutineTaskId]
		var payload struct {
			Pattern cdurablejobroutinetasktypes.RoutineTaskPattern `json:"pattern"`
		}
		if err := json.Unmarshal(task.Payload, &payload); err != nil {
			tx.Rollback()
			return nil, cexceptions.New("InvalidRoutineTaskPayload", "RoutineTask", "Claim", "Routine task payload is invalid", http.StatusBadRequest).WithOrigin(err)
		}
		patternValues := map[string]string{}
		for key, binding := range payload.Pattern {
			scheduledAt := scheduledAtByRecordId[record.RoutineRecordId]
			switch binding.Source {
			case "scheduledAt":
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
				patternValues[key] = record.Id.String()
			case "shortRecordId":
				recordId := record.Id.String()
				if len(recordId) > 8 {
					recordId = recordId[:8]
				}
				patternValues[key] = recordId
			case "routineTaskId":
				patternValues[key] = task.Id.String()
			}
		}
		routineAssignmentIndex, exists := routineAssignmentIndexByRecordId[record.RoutineRecordId]
		if !exists {
			routineAssignmentIndex = len(routineAssignments)
			routineAssignmentIndexByRecordId[record.RoutineRecordId] = routineAssignmentIndex
			routineAssignments = append(routineAssignments, cdurablejobroutinetasktypes.RoutineAssignment{
				RoutineId:         routineIdByRecordId[record.RoutineRecordId],
				RoutineRecordId:   record.RoutineRecordId,
				DefinitionVersion: definitionVersionByRecordId[record.RoutineRecordId],
				ScheduledAt:       scheduledAtByRecordId[record.RoutineRecordId],
				RoutineTasks:      make([]cdurablejobroutinetasktypes.RoutineTaskAssignment, 0),
			})
		}
		routineAssignments[routineAssignmentIndex].RoutineTasks = append(
			routineAssignments[routineAssignmentIndex].RoutineTasks,
			cdurablejobroutinetasktypes.RoutineTaskAssignment{
				RoutineTaskId:       task.Id,
				RoutineTaskRecordId: record.Id,
				RoutineRecordId:     record.RoutineRecordId,
				RoutineId:           routineIdByRecordId[record.RoutineRecordId],
				ActorUserId:         task.ActorUserId,
				ActorUserPublicId:   actorUserPublicIds[task.ActorUserId],
				Title:               task.Title,
				Purpose:             task.Purpose,
				Payload:             json.RawMessage(task.Payload),
				CostUnit:            task.CostUnit,
				Priority:            task.Priority,
				Attempt:             record.Attempts + 1,
				ScheduledAt:         scheduledAtByRecordId[record.RoutineRecordId],
				StartedAt:           now,
				PatternValues:       patternValues,
			},
		)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New("FailedToCommitTransaction", "RoutineTask", "Claim", "Failed to commit the routine task claim transaction", http.StatusInternalServerError, true).WithOrigin(err)
	}

	return &cdurablejob.ClaimRoutinesResponseDto{
		RequestId:          request.RequestId,
		WorkerId:           request.WorkerId,
		RoutineAssignments: routineAssignments,
	}, nil
}
