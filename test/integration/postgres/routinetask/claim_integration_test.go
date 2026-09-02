package routinetask_test

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	cdurablejobroutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	durablejobconfig "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/configs"
	routineexecution "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask"
	routinetaskrecoverers "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/recovery/recoverers"
	routinetaskworker "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/workers/routinetask"
)

type queryCounterLogger struct {
	logger.Interface
	count atomic.Int64
}

type noopRoutineTaskExecutionService struct{}

func (noopRoutineTaskExecutionService) ApplyPreparedRoutineTasks(
	context.Context,
	uuid.UUID,
	*cdurablejob.MarkCompletedRoutineTasksRequestDto,
) *cexceptions.Exception {
	return nil
}

func (noopRoutineTaskExecutionService) ApplyFailedRoutineTasks(
	context.Context,
	uuid.UUID,
	*cdurablejob.MarkFailedRoutineTasksRequestDto,
) *cexceptions.Exception {
	return nil
}

func (noopRoutineTaskExecutionService) ApplyResult(
	context.Context,
	uuid.UUID,
	cdurablejobroutinetasktypes.Result,
) *cexceptions.Exception {
	return nil
}

func (l *queryCounterLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	l.count.Add(1)
	l.Interface.Trace(ctx, begin, fc, err)
}

func TestClaimRoutinesInitializesManyTaskRecordsWithoutPerTaskQueries(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	seedRoutineWithManyTasks(t, db, actorUserId, routineId, 90)

	queryLogger := &queryCounterLogger{
		Interface: logger.New(log.New(io.Discard, "", 0), logger.Config{LogLevel: logger.Info}),
	}
	claimer := routinetaskworker.NewClaimer(db.Session(&gorm.Session{Logger: queryLogger}), nil)
	response, exception := claimer.ClaimRoutines(
		t.Context(),
		cdurablejob.ClaimRoutinesRequestDto{
			RequestId: uuid.New(),
			WorkerId:  uuid.New(),
			BatchSize: 1,
		},
	)
	if exception != nil {
		t.Fatalf("claim routine with many tasks: %v", exception)
	}
	if response == nil || len(response.RoutineAssignments) != 1 || len(response.RoutineAssignments[0].RoutineTasks) != 90 {
		t.Fatalf("claimed routines = %#v, want one routine with 90 tasks", response)
	}
	if queries := queryLogger.count.Load(); queries >= 40 {
		t.Fatalf("claim query count = %d, want fewer than 40 set-based queries", queries)
	}
}

func TestRoutineTaskPipelineUsesBoundedQueriesAcrossPhases(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	seedRoutineWithManyTasks(t, db, actorUserId, routineId, 90)

	queryLogger := &queryCounterLogger{
		Interface: logger.New(log.New(io.Discard, "", 0), logger.Config{LogLevel: logger.Info}),
	}
	claimer := routinetaskworker.NewClaimer(db.Session(&gorm.Session{Logger: queryLogger}), nil)
	response, exception := claimer.ClaimRoutines(
		t.Context(),
		cdurablejob.ClaimRoutinesRequestDto{
			RequestId: uuid.New(),
			WorkerId:  uuid.New(),
			BatchSize: 1,
		},
	)
	if exception != nil {
		t.Fatalf("claim routine for full pipeline query count: %v", exception)
	}
	if response == nil || len(response.RoutineAssignments) != 1 || len(response.RoutineAssignments[0].RoutineTasks) != 90 {
		t.Fatalf("claimed full pipeline routines = %#v, want one routine with 90 tasks", response)
	}

	engine := routinetaskworker.NewEngine(
		durablejobconfig.Config{},
		claimer,
		routineexecution.NewPlanService(db, nil),
		noopRoutineTaskExecutionService{},
		nil,
		nil,
		90,
	)
	defer engine.Stop()
	if err := engine.HandleRoutineAssignments(t.Context(), response.RoutineAssignments); err != nil {
		t.Fatalf("handle full routine task pipeline: %v", err)
	}
	if queries := queryLogger.count.Load(); queries >= 100 {
		t.Fatalf("full routine task pipeline query count = %d, want fewer than 100 set-based queries", queries)
	}
}

func TestClaimRoutinesSkipsLockedRoutineWithoutWaiting(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	seedRoutineWithManyTasks(t, db, actorUserId, routineId, 1)

	tx := db.Begin()
	var routine sschemas.Routine
	if result := tx.Model(&sschemas.Routine{}).
		Where("id = ?", routineId).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		First(&routine); result.Error != nil {
		tx.Rollback()
		t.Fatalf("lock routine for claim: %v", result.Error)
	}

	type claimResult struct {
		response *cdurablejob.ClaimRoutinesResponseDto
		hasError bool
	}
	resultChannel := make(chan claimResult, 1)
	claimer := routinetaskworker.NewClaimer(db, nil)
	go func() {
		response, exception := claimer.ClaimRoutines(
			t.Context(),
			cdurablejob.ClaimRoutinesRequestDto{
				RequestId: uuid.New(),
				WorkerId:  uuid.New(),
				BatchSize: 1,
			},
		)
		resultChannel <- claimResult{
			response: response,
			hasError: exception != nil,
		}
	}()

	select {
	case result := <-resultChannel:
		tx.Rollback()
		if result.hasError {
			t.Fatal("claim locked routine returned an exception")
		}
		if result.response == nil || len(result.response.RoutineAssignments) != 0 {
			t.Fatalf("claim locked routine response = %#v, want no routines", result.response)
		}
	case <-time.After(500 * time.Millisecond):
		tx.Rollback()
		t.Fatal("claim waited on a locked routine instead of using SKIP LOCKED")
	}
}

func TestClaimRoutinesWaitsForAvailableQuotaAndRetriesTheSameRoutineRecord(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	seedRoutineWithManyTasks(t, db, actorUserId, routineId, 2)
	now := time.Now().UTC().Truncate(time.Microsecond)
	if result := db.Exec(
		`INSERT INTO "UserQuotaTable" (user_id, routine_task_cost_unit_used, cycle_started_at, next_reset_at, updated_at, created_at) VALUES (?, 100, ?, ?, ?, ?)`,
		actorUserId,
		now,
		now.Add(30*24*time.Hour),
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed exhausted user quota: %v", result.Error)
	}

	claimer := routinetaskworker.NewClaimer(db, nil)
	response, exception := claimer.ClaimRoutines(
		t.Context(),
		cdurablejob.ClaimRoutinesRequestDto{RequestId: uuid.New(), WorkerId: uuid.New(), BatchSize: 1},
	)
	if exception != nil {
		t.Fatalf("claim with exhausted quota: %v", exception)
	}
	if response == nil || len(response.RoutineAssignments) != 0 {
		t.Fatalf("claim with exhausted quota = %#v, want no routines", response)
	}

	var routineRecord sschemas.RoutineRecord
	if result := db.Model(&sschemas.RoutineRecord{}).
		Select("id").
		Where("routine_id = ?", routineId).
		First(&routineRecord); result.Error != nil {
		t.Fatalf("read pending routine record: %v", result.Error)
	}
	routineRecordId := routineRecord.Id
	var readyTaskCount int64
	if result := db.Table(`"RoutineTaskRecordTable"`).Where("routine_record_id = ? AND status = ?", routineRecordId, cenums.RoutineTaskRecordStatus_Ready).Count(&readyTaskCount); result.Error != nil {
		t.Fatalf("count quota-blocked task records: %v", result.Error)
	} else if readyTaskCount != 2 {
		t.Fatalf("quota-blocked ready task records = %d, want 2", readyTaskCount)
	}

	if result := db.Model(&struct{}{}).Table(`"UserQuotaTable"`).Where("user_id = ?", actorUserId).Update("routine_task_cost_unit_used", 98); result.Error != nil {
		t.Fatalf("restore two quota units: %v", result.Error)
	}
	response, exception = claimer.ClaimRoutines(
		t.Context(),
		cdurablejob.ClaimRoutinesRequestDto{RequestId: uuid.New(), WorkerId: uuid.New(), BatchSize: 1},
	)
	if exception != nil {
		t.Fatalf("retry claim after quota restore: %v", exception)
	}
	if response == nil || len(response.RoutineAssignments) != 1 || len(response.RoutineAssignments[0].RoutineTasks) != 2 {
		t.Fatalf("retry claim after quota restore = %#v, want the same routine with two tasks", response)
	}
	if response.RoutineAssignments[0].RoutineRecordId != routineRecordId {
		t.Fatalf("retried routine record id = %s, want %s", response.RoutineAssignments[0].RoutineRecordId, routineRecordId)
	}
}

func TestClaimRoutinesDoesNotOverconsumeQuotaAfterStaleRecovery(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	seedRoutineWithManyTasks(t, db, actorUserId, routineId, 1)
	if result := db.Model(&sschemas.RoutineTask{}).
		Where("routine_id = ?", routineId).
		Update("max_attempts", 2); result.Error != nil {
		t.Fatalf("allow routine task recovery retry: %v", result.Error)
	}
	now := time.Now().UTC().Truncate(time.Microsecond)
	if result := db.Exec(
		`INSERT INTO "UserQuotaTable" (user_id, routine_task_cost_unit_used, cycle_started_at, next_reset_at, updated_at, created_at) VALUES (?, 99, ?, ?, ?, ?)`,
		actorUserId,
		now,
		now.Add(30*24*time.Hour),
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed nearly exhausted user quota: %v", result.Error)
	}

	claimer := routinetaskworker.NewClaimer(db, nil)
	firstResponse, exception := claimer.ClaimRoutines(
		t.Context(),
		cdurablejob.ClaimRoutinesRequestDto{RequestId: uuid.New(), WorkerId: uuid.New(), BatchSize: 1},
	)
	if exception != nil {
		t.Fatalf("claim before stale recovery: %v", exception)
	}
	if firstResponse == nil || len(firstResponse.RoutineAssignments) != 1 || len(firstResponse.RoutineAssignments[0].RoutineTasks) != 1 {
		t.Fatalf("first claim = %#v, want one routine with one task", firstResponse)
	}
	routineRecordId := firstResponse.RoutineAssignments[0].RoutineRecordId
	staleBefore := time.Now().UTC()
	if result := db.Model(&sschemas.RoutineTaskRecord{}).
		Where("routine_record_id = ?", routineRecordId).
		Updates(map[string]any{
			"actual_started_at": staleBefore.Add(-time.Hour),
			"updated_at":        staleBefore.Add(-time.Hour),
		}); result.Error != nil {
		t.Fatalf("mark routine task stale for quota recovery: %v", result.Error)
	}
	recoveredCount, err := routinetaskrecoverers.NewStaleRecordRecoverer(db).RecoverStaleRoutineTaskRecords(
		t.Context(),
		staleBefore,
	)
	if err != nil {
		t.Fatalf("recover stale routine task for quota recovery: %v", err)
	}
	if recoveredCount != 1 {
		t.Fatalf("recovered routine task count = %d, want 1", recoveredCount)
	}

	secondResponse, exception := claimer.ClaimRoutines(
		t.Context(),
		cdurablejob.ClaimRoutinesRequestDto{RequestId: uuid.New(), WorkerId: uuid.New(), BatchSize: 1},
	)
	if exception != nil {
		t.Fatalf("claim after stale recovery with exhausted quota: %v", exception)
	}
	if secondResponse == nil || len(secondResponse.RoutineAssignments) != 0 {
		t.Fatalf("claim after stale recovery with exhausted quota = %#v, want no routines", secondResponse)
	}

	var quotaUsed int64
	if result := db.Table(`"UserQuotaTable"`).Select("routine_task_cost_unit_used").Where("user_id = ?", actorUserId).Scan(&quotaUsed); result.Error != nil {
		t.Fatalf("read quota after rejected retry: %v", result.Error)
	} else if quotaUsed != 100 {
		t.Fatalf("quota after rejected retry = %d, want 100", quotaUsed)
	}
	assertExecutionTaskRecordStatuses(t, db, routineRecordId, cenums.RoutineTaskRecordStatus_Ready, 1)
}

func TestClaimRoutinesReusesSnapshotIdentityAfterStaleRecovery(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	seedRoutineWithManyTasks(t, db, actorUserId, routineId, 2)
	if result := db.Model(&sschemas.RoutineTask{}).
		Where("routine_id = ?", routineId).
		Update("max_attempts", 2); result.Error != nil {
		t.Fatalf("allow routine task retry: %v", result.Error)
	}

	claimer := routinetaskworker.NewClaimer(db, nil)
	firstResponse, exception := claimer.ClaimRoutines(
		t.Context(),
		cdurablejob.ClaimRoutinesRequestDto{RequestId: uuid.New(), WorkerId: uuid.New(), BatchSize: 1},
	)
	if exception != nil {
		t.Fatalf("first claim before stale recovery: %v", exception)
	}
	if firstResponse == nil || len(firstResponse.RoutineAssignments) != 1 || len(firstResponse.RoutineAssignments[0].RoutineTasks) != 2 {
		t.Fatalf("first claim = %#v, want one routine with two tasks", firstResponse)
	}
	routineRecordId := firstResponse.RoutineAssignments[0].RoutineRecordId
	firstTaskRecordIds := make(map[uuid.UUID]struct{}, 2)
	for _, task := range firstResponse.RoutineAssignments[0].RoutineTasks {
		firstTaskRecordIds[task.RoutineTaskRecordId] = struct{}{}
	}
	var firstSnapshot string
	if result := db.Model(&sschemas.RoutineRecord{}).
		Select("snapshot::text").
		Where("id = ?", routineRecordId).
		Scan(&firstSnapshot); result.Error != nil {
		t.Fatalf("read first routine snapshot: %v", result.Error)
	}

	staleBefore := time.Now().UTC()
	if result := db.Model(&sschemas.RoutineTaskRecord{}).
		Where("routine_record_id = ?", routineRecordId).
		Updates(map[string]any{
			"actual_started_at": staleBefore.Add(-time.Hour),
			"updated_at":        staleBefore.Add(-time.Hour),
		}); result.Error != nil {
		t.Fatalf("mark routine tasks stale: %v", result.Error)
	}
	recoveredCount, err := routinetaskrecoverers.NewStaleRecordRecoverer(db).RecoverStaleRoutineTaskRecords(
		t.Context(),
		staleBefore,
	)
	if err != nil {
		t.Fatalf("recover stale routine tasks: %v", err)
	}
	if recoveredCount != 2 {
		t.Fatalf("recovered stale routine tasks = %d, want 2", recoveredCount)
	}

	secondResponse, exception := claimer.ClaimRoutines(
		t.Context(),
		cdurablejob.ClaimRoutinesRequestDto{RequestId: uuid.New(), WorkerId: uuid.New(), BatchSize: 1},
	)
	if exception != nil {
		t.Fatalf("second claim after stale recovery: %v", exception)
	}
	if secondResponse == nil || len(secondResponse.RoutineAssignments) != 1 || len(secondResponse.RoutineAssignments[0].RoutineTasks) != 2 {
		t.Fatalf("second claim = %#v, want one routine with two tasks", secondResponse)
	}
	if secondResponse.RoutineAssignments[0].RoutineRecordId != routineRecordId {
		t.Fatalf("second routine record id = %s, want %s", secondResponse.RoutineAssignments[0].RoutineRecordId, routineRecordId)
	}
	for _, task := range secondResponse.RoutineAssignments[0].RoutineTasks {
		if _, exists := firstTaskRecordIds[task.RoutineTaskRecordId]; !exists {
			t.Fatalf("second claim returned new task record id %s", task.RoutineTaskRecordId)
		}
	}
	var secondSnapshot string
	if result := db.Model(&sschemas.RoutineRecord{}).
		Select("snapshot::text").
		Where("id = ?", routineRecordId).
		Scan(&secondSnapshot); result.Error != nil {
		t.Fatalf("read second routine snapshot: %v", result.Error)
	}
	if firstSnapshot != secondSnapshot {
		t.Fatalf("routine snapshot changed across stale recovery retry")
	}
}

func TestClaimRoutinesClaimsAllReadyTasksForEachRoutine(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	firstRoutineId := uuid.New()
	secondRoutineId := uuid.New()
	seedRoutineTaskClaimData(t, db, actorUserId, firstRoutineId, secondRoutineId)

	claimer := routinetaskworker.NewClaimer(db, nil)
	response, exception := claimer.ClaimRoutines(
		t.Context(),
		cdurablejob.ClaimRoutinesRequestDto{
			RequestId: uuid.New(),
			WorkerId:  uuid.New(),
			BatchSize: 1,
		},
	)
	if exception != nil {
		t.Fatalf("claim routines: %v", exception)
	}
	if response == nil || len(response.RoutineAssignments) != 1 {
		t.Fatalf("claimed routines = %#v, want one routine", response)
	}
	claimedRoutineId := response.RoutineAssignments[0].RoutineId
	if claimedRoutineId != firstRoutineId && claimedRoutineId != secondRoutineId {
		t.Fatalf("claimed routine id = %s, want one of the seeded routines", claimedRoutineId)
	}
	if len(response.RoutineAssignments[0].RoutineTasks) != 2 {
		t.Fatalf("claimed routine tasks = %d, want all two ready tasks", len(response.RoutineAssignments[0].RoutineTasks))
	}

	var runningTaskRecordCount int64
	if result := db.Model(&struct{}{}).Table(`"RoutineTaskRecordTable"`).Where("status = ?", cenums.RoutineTaskRecordStatus_Running).Count(&runningTaskRecordCount); result.Error != nil {
		t.Fatalf("count running task records: %v", result.Error)
	} else if runningTaskRecordCount != 2 {
		t.Fatalf("running task records = %d, want 2", runningTaskRecordCount)
	}

	var secondRoutineRecordCount int64
	remainingRoutineId := secondRoutineId
	if claimedRoutineId == secondRoutineId {
		remainingRoutineId = firstRoutineId
	}
	if result := db.Table(`"RoutineRecordTable"`).Where("routine_id = ?", remainingRoutineId).Count(&secondRoutineRecordCount); result.Error != nil {
		t.Fatalf("count remaining routine records: %v", result.Error)
	} else if secondRoutineRecordCount != 0 {
		t.Fatalf("remaining routine records = %d, want 0 because batch size is routine-scoped", secondRoutineRecordCount)
	}
}

func TestHandleRoutineAssignmentsAdvancesRoutineThroughPhasePipeline(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	seedRoutineWithManyTasks(t, db, actorUserId, routineId, 1)

	claimer := routinetaskworker.NewClaimer(db, nil)
	response, exception := claimer.ClaimRoutines(
		t.Context(),
		cdurablejob.ClaimRoutinesRequestDto{
			RequestId: uuid.New(),
			WorkerId:  uuid.New(),
			BatchSize: 1,
		},
	)
	if exception != nil {
		t.Fatalf("claim routine for phase pipeline: %v", exception)
	}
	if response == nil || len(response.RoutineAssignments) != 1 || len(response.RoutineAssignments[0].RoutineTasks) != 1 {
		t.Fatalf("claimed phase pipeline routines = %#v, want one routine with one task", response)
	}

	engine := routinetaskworker.NewEngine(
		durablejobconfig.Config{},
		claimer,
		routineexecution.NewPlanService(db, nil),
		noopRoutineTaskExecutionService{},
		nil,
		nil,
	)
	defer engine.Stop()
	if err := engine.HandleRoutineAssignments(t.Context(), response.RoutineAssignments); err != nil {
		t.Fatalf("handle routine phase pipeline: %v", err)
	}

	var routinePhase string
	if result := db.Model(&sschemas.Routine{}).
		Select("phase").
		Where("id = ?", routineId).
		Scan(&routinePhase); result.Error != nil {
		t.Fatalf("read completed routine phase: %v", result.Error)
	}
	if routinePhase != string(cenums.RoutinePhase_Analysis) {
		t.Fatalf("completed routine phase = %s, want %s", routinePhase, cenums.RoutinePhase_Analysis)
	}
}

func TestClaimRoutinesClaimsEachRoutineOnceConcurrently(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	firstRoutineId := uuid.New()
	secondRoutineId := uuid.New()
	seedRoutineTaskClaimData(t, db, actorUserId, firstRoutineId, secondRoutineId)

	claimer := routinetaskworker.NewClaimer(db, nil)
	start := make(chan struct{})
	responses := make(chan *cdurablejob.ClaimRoutinesResponseDto, 2)
	exceptions := make(chan error, 2)
	var workers sync.WaitGroup
	for index := 0; index < 2; index++ {
		workers.Add(1)
		go func() {
			defer workers.Done()
			<-start
			response, exception := claimer.ClaimRoutines(
				t.Context(),
				cdurablejob.ClaimRoutinesRequestDto{
					RequestId: uuid.New(),
					WorkerId:  uuid.New(),
					BatchSize: 1,
				},
			)
			if exception != nil {
				exceptions <- exception
				return
			}
			responses <- response
		}()
	}
	close(start)
	workers.Wait()
	close(responses)
	close(exceptions)

	for err := range exceptions {
		if err != nil {
			t.Fatalf("concurrent routine claim: %v", err)
		}
	}
	claimedRoutineIds := make(map[uuid.UUID]struct{}, 2)
	claimedTaskCount := 0
	for response := range responses {
		if response == nil || len(response.RoutineAssignments) != 1 {
			t.Fatalf("concurrent claim response = %#v, want one routine", response)
		}
		claimedRoutineIds[response.RoutineAssignments[0].RoutineId] = struct{}{}
		claimedTaskCount += len(response.RoutineAssignments[0].RoutineTasks)
	}
	if len(claimedRoutineIds) != 2 {
		t.Fatalf("concurrently claimed routine ids = %v, want two distinct routines", claimedRoutineIds)
	}
	if claimedTaskCount != 4 {
		t.Fatalf("concurrently claimed task count = %d, want 4", claimedTaskCount)
	}
}

func TestClaimRoutinesRetriesTerminalPlanOnlyAfterDefinitionRevision(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	firstTaskId := uuid.New()
	secondTaskId := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	if result := db.Exec(
		`INSERT INTO "UserView" (id, public_id, plan, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		actorUserId,
		uuid.New(),
		cenums.UserPlan_Free,
		cenums.UserStatus_Online,
		now,
	); result.Error != nil {
		t.Fatalf("seed terminal plan user view: %v", result.Error)
	}
	if result := db.Exec(
		`INSERT INTO "RoutineTable" (id, title, description, is_pinned, scheduled_start_at, scheduled_end_at, period, timezone, status, definition_version, updated_at, created_at) VALUES (?, 'Terminal plan routine', '', false, ?, ?, NULL, 'UTC', ?, 1, ?, ?)`,
		routineId,
		now.Add(-time.Minute),
		now.Add(time.Hour),
		cenums.RoutineStatus_Scheduled,
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed terminal plan routine: %v", result.Error)
	}
	rootShelfId := uuid.New()
	firstFakeId := "f_11111111111111111111111111111111"
	secondFakeId := "f_22222222222222222222222222222222"
	for index, taskId := range []uuid.UUID{firstTaskId, secondTaskId} {
		fakeId := firstFakeId
		if index == 1 {
			fakeId = secondFakeId
		}
		payload := fmt.Sprintf(`{"fakeId":"%s","rootShelfId":"%s","name":"Shelf %d"}`, fakeId, rootShelfId, index)
		if result := db.Exec(
			`INSERT INTO "RoutineTaskTable" (id, routine_id, actor_user_id, title, purpose, payload, cost_unit, priority, max_attempts, updated_at, created_at) VALUES (?, ?, ?, ?, ?, ?::jsonb, 1, 0, 1, ?, ?)`,
			taskId,
			routineId,
			actorUserId,
			"Invalid shelf",
			cenums.RoutineTaskPurpose_CreateSubShelf,
			payload,
			now,
			now,
		); result.Error != nil {
			t.Fatalf("seed terminal plan task %s: %v", taskId, result.Error)
		}
	}
	if result := db.Exec(
		`INSERT INTO "RoutineDependencyTable" (routine_task_id, previous_routine_task_id) VALUES (?, ?), (?, ?)`,
		firstTaskId,
		secondTaskId,
		secondTaskId,
		firstTaskId,
	); result.Error != nil {
		t.Fatalf("seed terminal plan cycle: %v", result.Error)
	}

	claimer := routinetaskworker.NewClaimer(db, nil)
	firstResponse, firstException := claimer.ClaimRoutines(
		t.Context(),
		cdurablejob.ClaimRoutinesRequestDto{RequestId: uuid.New(), WorkerId: uuid.New(), BatchSize: 1},
	)
	if firstException != nil {
		t.Fatalf("claim invalid routine plan: %v", firstException)
	}
	if firstResponse == nil || len(firstResponse.RoutineAssignments) != 1 || len(firstResponse.RoutineAssignments[0].RoutineTasks) != 0 {
		t.Fatalf("invalid plan response = %#v, want one routine without ready tasks", firstResponse)
	}
	engine := routinetaskworker.NewEngine(
		durablejobconfig.Config{},
		claimer,
		routineexecution.NewPlanService(db, nil),
		noopRoutineTaskExecutionService{},
		nil,
		nil,
	)
	defer engine.Stop()
	if err := engine.HandleRoutineAssignments(t.Context(), firstResponse.RoutineAssignments); err != nil {
		t.Fatalf("handle invalid routine plan: %v", err)
	}
	var recordStatus string
	if result := db.Table(`"RoutineRecordTable"`).Select("status").Where("routine_id = ?", routineId).Scan(&recordStatus); result.Error != nil {
		t.Fatalf("read terminal routine record status: %v", result.Error)
	} else if recordStatus != string(cenums.RoutineRecordStatus_Blocked) {
		t.Fatalf("terminal routine record status = %s, want %s", recordStatus, cenums.RoutineRecordStatus_Blocked)
	}
	var routineStatus string
	if result := db.Table(`"RoutineTable"`).Select("status").Where("id = ?", routineId).Scan(&routineStatus); result.Error != nil {
		t.Fatalf("read terminal routine status: %v", result.Error)
	} else if routineStatus != string(cenums.RoutineStatus_OverDue) {
		t.Fatalf("terminal routine status = %s, want %s", routineStatus, cenums.RoutineStatus_OverDue)
	}
	var activeTaskRecordCount int64
	if result := db.Table(`"RoutineTaskRecordTable"`).
		Where("routine_record_id IN (SELECT id FROM \"RoutineRecordTable\" WHERE routine_id = ?)", routineId).
		Where("status IN ?", []cenums.RoutineTaskRecordStatus{
			cenums.RoutineTaskRecordStatus_Waiting,
			cenums.RoutineTaskRecordStatus_Ready,
			cenums.RoutineTaskRecordStatus_Running,
		}).Count(&activeTaskRecordCount); result.Error != nil {
		t.Fatalf("read terminal routine task record statuses: %v", result.Error)
	} else if activeTaskRecordCount != 0 {
		t.Fatalf("terminal routine active task record count = %d, want 0", activeTaskRecordCount)
	}
	var routinePhase string
	if result := db.Table(`"RoutineTable"`).Select("phase").Where("id = ?", routineId).Scan(&routinePhase); result.Error != nil {
		t.Fatalf("read terminal routine phase: %v", result.Error)
	} else if routinePhase != string(cenums.RoutinePhase_Plan) {
		t.Fatalf("terminal routine phase = %s, want %s", routinePhase, cenums.RoutinePhase_Plan)
	}

	secondResponse, secondException := claimer.ClaimRoutines(
		t.Context(),
		cdurablejob.ClaimRoutinesRequestDto{RequestId: uuid.New(), WorkerId: uuid.New(), BatchSize: 1},
	)
	if secondException != nil {
		t.Fatalf("reclaim terminal routine plan: %v", secondException)
	}
	if secondResponse == nil || len(secondResponse.RoutineAssignments) != 0 {
		t.Fatalf("terminal retry response = %#v, want no routines", secondResponse)
	}

	if result := db.Exec(
		`DELETE FROM "RoutineDependencyTable" WHERE routine_task_id IN (?, ?) OR previous_routine_task_id IN (?, ?)`,
		firstTaskId,
		secondTaskId,
		firstTaskId,
		secondTaskId,
	); result.Error != nil {
		t.Fatalf("remove terminal routine cycle: %v", result.Error)
	}
	if result := db.Exec(
		`UPDATE "RoutineTable" SET status = ?, definition_version = 2, scheduled_start_at = ?, scheduled_end_at = ?, updated_at = ? WHERE id = ?`,
		cenums.RoutineStatus_Scheduled,
		now.Add(-time.Minute),
		now.Add(time.Hour),
		now,
		routineId,
	); result.Error != nil {
		t.Fatalf("revise terminal routine definition: %v", result.Error)
	}

	thirdResponse, thirdException := claimer.ClaimRoutines(
		t.Context(),
		cdurablejob.ClaimRoutinesRequestDto{RequestId: uuid.New(), WorkerId: uuid.New(), BatchSize: 1},
	)
	if thirdException != nil {
		t.Fatalf("claim revised routine plan: %v", thirdException)
	}
	if thirdResponse == nil || len(thirdResponse.RoutineAssignments) != 1 || len(thirdResponse.RoutineAssignments[0].RoutineTasks) != 2 {
		t.Fatalf("revised routine response = %#v, want one routine with two tasks", thirdResponse)
	}
	if thirdResponse.RoutineAssignments[0].DefinitionVersion != 2 {
		t.Fatalf("revised routine definition version = %d, want 2", thirdResponse.RoutineAssignments[0].DefinitionVersion)
	}
}

func TestRecoverStaleRoutineTaskRecordsRequeuesRetryableTasks(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	routineRecordId := uuid.New()
	routineTaskId := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	seedRoutineTaskRecoveryData(t, db, actorUserId, routineId, routineRecordId, routineTaskId, now, 1, 2, cenums.RoutineTaskRecordStatus_Running)

	recoveredCount, err := routinetaskrecoverers.NewStaleRecordRecoverer(db).RecoverStaleRoutineTaskRecords(
		t.Context(),
		now.Add(time.Minute),
	)
	if err != nil {
		t.Fatalf("recover stale routine task records: %v", err)
	}
	if recoveredCount != 1 {
		t.Fatalf("recovered record count = %d, want 1", recoveredCount)
	}

	var status string
	if result := db.Table(`"RoutineTaskRecordTable"`).Select("status").Where("id = ?", routineTaskId).Scan(&status); result.Error != nil {
		t.Fatalf("read requeued task status: %v", result.Error)
	} else if status != string(cenums.RoutineTaskRecordStatus_Ready) {
		t.Fatalf("requeued task status = %s, want %s", status, cenums.RoutineTaskRecordStatus_Ready)
	}
}

func TestRecoverStaleRoutineTaskRecordsBlocksDependentsAfterAttemptsExhausted(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	routineRecordId := uuid.New()
	failedTaskId := uuid.New()
	dependentTaskId := uuid.New()
	now := time.Now().UTC().Truncate(time.Microsecond)
	seedRoutineTaskRecoveryData(t, db, actorUserId, routineId, routineRecordId, failedTaskId, now, 2, 2, cenums.RoutineTaskRecordStatus_Running)
	if result := db.Exec(
		`INSERT INTO "RoutineTaskTable" (id, routine_id, actor_user_id, title, purpose, payload, cost_unit, priority, max_attempts, updated_at, created_at) VALUES (?, ?, ?, 'dependent task', ?, '{}'::jsonb, 1, 0, 1, ?, ?)`,
		dependentTaskId,
		routineId,
		actorUserId,
		cenums.RoutineTaskPurpose_GetRoutine,
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed dependent routine task: %v", result.Error)
	}
	dependentRecordId := uuid.New()
	if result := db.Exec(
		`INSERT INTO "RoutineTaskRecordTable" (id, routine_record_id, routine_task_id, purpose, status, cost_unit, attempts, payload_snapshot, result_snapshot, updated_at, created_at) VALUES (?, ?, ?, ?, ?, 1, 0, '{}'::jsonb, '{}'::jsonb, ?, ?)`,
		dependentRecordId,
		routineRecordId,
		dependentTaskId,
		cenums.RoutineTaskPurpose_GetRoutine,
		cenums.RoutineTaskRecordStatus_Waiting,
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed dependent routine task record: %v", result.Error)
	}
	if result := db.Exec(
		`INSERT INTO "RoutineDependencyTable" (routine_task_id, previous_routine_task_id) VALUES (?, ?)`,
		dependentTaskId,
		failedTaskId,
	); result.Error != nil {
		t.Fatalf("seed routine task dependency: %v", result.Error)
	}

	recoveredCount, err := routinetaskrecoverers.NewStaleRecordRecoverer(db).RecoverStaleRoutineTaskRecords(
		t.Context(),
		now.Add(time.Minute),
	)
	if err != nil {
		t.Fatalf("recover exhausted stale routine task record: %v", err)
	}
	if recoveredCount != 1 {
		t.Fatalf("recovered record count = %d, want 1", recoveredCount)
	}

	var statuses []string
	if result := db.Table(`"RoutineTaskRecordTable"`).Select("status").Where("routine_record_id = ?", routineRecordId).Order("id").Scan(&statuses); result.Error != nil {
		t.Fatalf("read recovered task statuses: %v", result.Error)
	} else {
		statusSet := make(map[string]struct{}, len(statuses))
		for _, status := range statuses {
			statusSet[status] = struct{}{}
		}
		if len(statuses) != 2 {
			t.Fatalf("recovered task statuses = %v, want one Failed and one Blocked", statuses)
		}
		if _, exists := statusSet[string(cenums.RoutineTaskRecordStatus_Failed)]; !exists {
			t.Fatalf("recovered task statuses = %v, want one Failed and one Blocked", statuses)
		}
		if _, exists := statusSet[string(cenums.RoutineTaskRecordStatus_Blocked)]; !exists {
			t.Fatalf("recovered task statuses = %v, want one Failed and one Blocked", statuses)
		}
	}
}

func openRoutineTaskClaimIntegrationDB(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv("NOTEGIC_RUN_DURABLEJOB_ROUTINETASK_INTEGRATION") != "1" {
		t.Skip("set NOTEGIC_RUN_DURABLEJOB_ROUTINETASK_INTEGRATION=1 to run DurableJob routine task integration tests")
	}

	config, err := platformpostgres.LoadConfig(
		routineTaskIntegrationEnv("POSTGRES_DURABLEJOB_ROUTINETASK_HOST", "127.0.0.1"),
		routineTaskIntegrationEnv("POSTGRES_DURABLEJOB_ROUTINETASK_USER", "notegic"),
		routineTaskIntegrationEnv("POSTGRES_DURABLEJOB_ROUTINETASK_PASSWORD", "notegic"),
		routineTaskIntegrationEnv("POSTGRES_DURABLEJOB_ROUTINETASK_NAME", "notegic_integration"),
		routineTaskIntegrationEnv("POSTGRES_DURABLEJOB_ROUTINETASK_PORT", "15432"),
	)
	if err != nil {
		t.Fatalf("load PostgreSQL routine task integration config: %v", err)
	}

	admin, err := platformpostgres.Connect(config)
	if err != nil {
		t.Fatalf("connect PostgreSQL routine task integration database: %v", err)
	}
	schemaName := "durablejob_routinetask_" + strings.ReplaceAll(uuid.NewString(), "-", "")
	if result := admin.Exec(`CREATE SCHEMA "` + schemaName + `"`); result.Error != nil {
		platformpostgres.Disconnect(admin)
		t.Fatalf("create isolated routine task integration schema: %v", result.Error)
	}

	dsn := platformpostgres.ConnectionString(config) + " options='-c search_path=" + schemaName + "'"
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		admin.Exec(`DROP SCHEMA "` + schemaName + `" CASCADE`)
		platformpostgres.Disconnect(admin)
		t.Fatalf("connect isolated routine task integration schema: %v", err)
	}
	if err := createRoutineTaskClaimIntegrationTables(db); err != nil {
		admin.Exec(`DROP SCHEMA "` + schemaName + `" CASCADE`)
		platformpostgres.Disconnect(admin)
		t.Fatalf("create routine task integration tables: %v", err)
	}

	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
		admin.Exec(`DROP SCHEMA "` + schemaName + `" CASCADE`)
		platformpostgres.Disconnect(admin)
	})
	return db
}

func createRoutineTaskClaimIntegrationTables(db *gorm.DB) error {
	statements := []string{
		`CREATE TYPE "RoutineRecordStatus" AS ENUM ('Pending', 'Running', 'Success', 'Failed', 'Blocked', 'Canceled')`,
		`CREATE TABLE "RootShelfTable" (id uuid PRIMARY KEY, owner_id uuid NOT NULL, name text NOT NULL, sub_shelf_count bigint NOT NULL DEFAULT 0, item_count bigint NOT NULL DEFAULT 0, last_analyzed_at timestamptz NOT NULL DEFAULT NOW(), deleted_at timestamptz, updated_at timestamptz NOT NULL, created_at timestamptz NOT NULL)`,
		`CREATE TABLE "UsersToShelvesTable" (user_id uuid NOT NULL, root_shelf_id uuid NOT NULL, permission text NOT NULL, updated_at timestamptz NOT NULL, created_at timestamptz NOT NULL, PRIMARY KEY (user_id, root_shelf_id))`,
		`CREATE TABLE "SubShelfTable" (id uuid PRIMARY KEY, name varchar(128) NOT NULL, root_shelf_id uuid NOT NULL, prev_sub_shelf_id uuid, path uuid[] NOT NULL DEFAULT '{}', deleted_at timestamptz, updated_at timestamptz NOT NULL, created_at timestamptz NOT NULL, CONSTRAINT sub_shelf_previous_reference FOREIGN KEY (prev_sub_shelf_id) REFERENCES "SubShelfTable" (id) DEFERRABLE INITIALLY DEFERRED)`,
		`CREATE TABLE "MaterialTable" (id uuid PRIMARY KEY, parent_sub_shelf_id uuid NOT NULL, name varchar(128) NOT NULL, size bigint NOT NULL DEFAULT 0, content_key text NOT NULL UNIQUE, content_type text NOT NULL DEFAULT 'none', parse_media_type varchar(128) NOT NULL DEFAULT '', deleted_at timestamptz, updated_at timestamptz NOT NULL, created_at timestamptz NOT NULL)`,
		`CREATE TABLE "RoutineTable" (id uuid PRIMARY KEY, title text NOT NULL, description text NOT NULL, is_pinned boolean NOT NULL, scheduled_start_at timestamptz NOT NULL, scheduled_end_at timestamptz NOT NULL, period text, timezone text NOT NULL, status text NOT NULL, phase text, definition_version bigint NOT NULL, deleted_at timestamptz, updated_at timestamptz NOT NULL, created_at timestamptz NOT NULL)`,
		`CREATE TABLE "RoutineRecordTable" (id uuid PRIMARY KEY, routine_id uuid NOT NULL, definition_version bigint NOT NULL, status "RoutineRecordStatus" NOT NULL, scheduled_at timestamptz NOT NULL, actual_started_at timestamptz, actual_ended_at timestamptz, total_task_count integer NOT NULL, success_task_count integer NOT NULL, failed_task_count integer NOT NULL, blocked_task_count integer NOT NULL, running_task_count integer NOT NULL, waiting_task_count integer NOT NULL, snapshot jsonb NOT NULL, updated_at timestamptz NOT NULL, created_at timestamptz NOT NULL, UNIQUE (routine_id, scheduled_at, definition_version))`,
		`CREATE TABLE "RoutineTaskTable" (id uuid PRIMARY KEY, routine_id uuid NOT NULL, actor_user_id uuid NOT NULL, title text NOT NULL, purpose text NOT NULL, payload jsonb NOT NULL, cost_unit bigint NOT NULL, priority integer NOT NULL, max_attempts integer NOT NULL, updated_at timestamptz NOT NULL, created_at timestamptz NOT NULL)`,
		`CREATE TABLE "RoutineDependencyTable" (routine_task_id uuid NOT NULL, previous_routine_task_id uuid NOT NULL, PRIMARY KEY (routine_task_id, previous_routine_task_id))`,
		`CREATE TABLE "RoutineTaskRecordTable" (id uuid PRIMARY KEY, routine_record_id uuid NOT NULL, routine_task_id uuid NOT NULL, purpose text NOT NULL, status text NOT NULL, error_code text, error_reason varchar(256), cost_unit bigint NOT NULL, attempts integer NOT NULL, payload_snapshot jsonb NOT NULL, result_snapshot jsonb NOT NULL, actual_started_at timestamptz, actual_ended_at timestamptz, updated_at timestamptz NOT NULL, created_at timestamptz NOT NULL, UNIQUE (routine_record_id, routine_task_id))`,
		`CREATE TABLE "UserView" (id uuid PRIMARY KEY, public_id uuid NOT NULL, plan text NOT NULL, status text NOT NULL, created_at timestamptz NOT NULL)`,
		`CREATE TABLE "UserQuotaTable" (id uuid PRIMARY KEY DEFAULT gen_random_uuid(), user_id uuid NOT NULL UNIQUE, routine_task_cost_unit_used bigint NOT NULL, cycle_started_at timestamptz NOT NULL, next_reset_at timestamptz NOT NULL, updated_at timestamptz NOT NULL, created_at timestamptz NOT NULL)`,
		`CREATE TABLE "PlanLimitationTable" (key text PRIMARY KEY, max_routine_task_cost_unit_count integer NOT NULL)`,
		`INSERT INTO "PlanLimitationTable" (key, max_routine_task_cost_unit_count) VALUES ('Free', 100)`,
	}
	for _, statement := range statements {
		if result := db.Exec(statement); result.Error != nil {
			return result.Error
		}
	}
	return nil
}

func seedRoutineTaskClaimData(t *testing.T, db *gorm.DB, actorUserId, firstRoutineId, secondRoutineId uuid.UUID) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	if result := db.Exec(
		`INSERT INTO "UserView" (id, public_id, plan, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		actorUserId,
		uuid.New(),
		cenums.UserPlan_Free,
		cenums.UserStatus_Online,
		now,
	); result.Error != nil {
		t.Fatalf("seed user view: %v", result.Error)
	}
	for _, routineId := range []uuid.UUID{firstRoutineId, secondRoutineId} {
		if result := db.Exec(
			`INSERT INTO "RoutineTable" (id, title, description, is_pinned, scheduled_start_at, scheduled_end_at, period, timezone, status, definition_version, updated_at, created_at) VALUES (?, ?, '', false, ?, ?, NULL, 'UTC', ?, 1, ?, ?)`,
			routineId,
			routineId.String(),
			now.Add(-time.Minute),
			now.Add(time.Hour),
			cenums.RoutineStatus_Scheduled,
			now,
			now,
		); result.Error != nil {
			t.Fatalf("seed routine %s: %v", routineId, result.Error)
		}
	}
	for routineIndex, routineId := range []uuid.UUID{firstRoutineId, secondRoutineId} {
		for taskIndex := 0; taskIndex < 2; taskIndex++ {
			taskId := uuid.New()
			payload := fmt.Sprintf(`{"fakeId":"f_%032d","rootShelfId":"%s","name":"Shelf %d"}`, routineIndex*2+taskIndex+1, uuid.New(), taskIndex)
			if result := db.Exec(
				`INSERT INTO "RoutineTaskTable" (id, routine_id, actor_user_id, title, purpose, payload, cost_unit, priority, max_attempts, updated_at, created_at) VALUES (?, ?, ?, ?, ?, ?::jsonb, 1, ?, 1, ?, ?)`,
				taskId,
				routineId,
				actorUserId,
				"Create shelf",
				cenums.RoutineTaskPurpose_CreateSubShelf,
				payload,
				taskIndex,
				now,
				now,
			); result.Error != nil {
				t.Fatalf("seed routine task %s: %v", taskId, result.Error)
			}
		}
	}
}

func seedRoutineWithManyTasks(t *testing.T, db *gorm.DB, actorUserId, routineId uuid.UUID, taskCount int) {
	t.Helper()
	now := time.Now().UTC().Truncate(time.Microsecond)
	if result := db.Exec(
		`INSERT INTO "UserView" (id, public_id, plan, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		actorUserId,
		uuid.New(),
		cenums.UserPlan_Free,
		cenums.UserStatus_Online,
		now,
	); result.Error != nil {
		t.Fatalf("seed many-task user view: %v", result.Error)
	}
	if result := db.Exec(
		`INSERT INTO "RoutineTable" (id, title, description, is_pinned, scheduled_start_at, scheduled_end_at, period, timezone, status, definition_version, updated_at, created_at) VALUES (?, 'Many-task routine', '', false, ?, ?, NULL, 'UTC', ?, 1, ?, ?)`,
		routineId,
		now.Add(-time.Minute),
		now.Add(time.Hour),
		cenums.RoutineStatus_Scheduled,
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed many-task routine: %v", result.Error)
	}
	for taskIndex := 0; taskIndex < taskCount; taskIndex++ {
		taskId := uuid.New()
		payload := fmt.Sprintf(`{"fakeId":"f_%032d","rootShelfId":"%s","name":"Shelf %d"}`, taskIndex+1, uuid.New(), taskIndex)
		if result := db.Exec(
			`INSERT INTO "RoutineTaskTable" (id, routine_id, actor_user_id, title, purpose, payload, cost_unit, priority, max_attempts, updated_at, created_at) VALUES (?, ?, ?, 'Create shelf', ?, ?::jsonb, 1, 0, 1, ?, ?)`,
			taskId,
			routineId,
			actorUserId,
			cenums.RoutineTaskPurpose_CreateSubShelf,
			payload,
			now,
			now,
		); result.Error != nil {
			t.Fatalf("seed many-task routine task %d: %v", taskIndex, result.Error)
		}
	}
}

func seedRoutineTaskRecoveryData(
	t *testing.T,
	db *gorm.DB,
	actorUserId uuid.UUID,
	routineId uuid.UUID,
	routineRecordId uuid.UUID,
	routineTaskId uuid.UUID,
	now time.Time,
	attempts int,
	maxAttempts int,
	status cenums.RoutineTaskRecordStatus,
) {
	t.Helper()
	if result := db.Exec(
		`INSERT INTO "UserView" (id, public_id, plan, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		actorUserId,
		uuid.New(),
		cenums.UserPlan_Free,
		cenums.UserStatus_Online,
		now,
	); result.Error != nil {
		t.Fatalf("seed recovery user view: %v", result.Error)
	}
	if result := db.Exec(
		`INSERT INTO "RoutineTable" (id, title, description, is_pinned, scheduled_start_at, scheduled_end_at, period, timezone, status, definition_version, updated_at, created_at) VALUES (?, 'Recovery routine', '', false, ?, ?, NULL, 'UTC', ?, 1, ?, ?)`,
		routineId,
		now.Add(-time.Hour),
		now.Add(time.Hour),
		cenums.RoutineStatus_InProgress,
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed recovery routine: %v", result.Error)
	}
	if result := db.Exec(
		`INSERT INTO "RoutineRecordTable" (id, routine_id, definition_version, status, scheduled_at, total_task_count, success_task_count, failed_task_count, blocked_task_count, running_task_count, waiting_task_count, snapshot, updated_at, created_at) VALUES (?, ?, 1, ?, ?, 1, 0, 0, 0, 1, 0, '{}'::jsonb, ?, ?)`,
		routineRecordId,
		routineId,
		cenums.RoutineRecordStatus_Running,
		now,
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed recovery routine record: %v", result.Error)
	}
	if result := db.Exec(
		`INSERT INTO "RoutineTaskTable" (id, routine_id, actor_user_id, title, purpose, payload, cost_unit, priority, max_attempts, updated_at, created_at) VALUES (?, ?, ?, 'stale task', ?, '{}'::jsonb, 1, 0, ?, ?, ?)`,
		routineTaskId,
		routineId,
		actorUserId,
		cenums.RoutineTaskPurpose_GetRoutine,
		maxAttempts,
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed recovery routine task: %v", result.Error)
	}
	if result := db.Exec(
		`INSERT INTO "RoutineTaskRecordTable" (id, routine_record_id, routine_task_id, purpose, status, cost_unit, attempts, payload_snapshot, result_snapshot, actual_started_at, updated_at, created_at) VALUES (?, ?, ?, ?, ?, 1, ?, '{}'::jsonb, '{}'::jsonb, ?, ?, ?)`,
		routineTaskId,
		routineRecordId,
		routineTaskId,
		cenums.RoutineTaskPurpose_GetRoutine,
		status,
		attempts,
		now.Add(-time.Hour),
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed recovery routine task record: %v", result.Error)
	}
}

func routineTaskIntegrationEnv(name, fallback string) string {
	if value := os.Getenv(name); value != "" {
		return value
	}
	return fallback
}
