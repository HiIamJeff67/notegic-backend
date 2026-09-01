package routinetask_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"

	cdurablejob "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	routinetaskservice "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask"
	routinetaskvalidation "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/validations"
	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"
)

func TestApplyPreparedRoutineTasksCreatesNestedSubShelvesWithPlannedIdentity(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	routineRecordId := uuid.New()
	rootShelfId := uuid.New()
	rootTaskId := uuid.New()
	childTaskId := uuid.New()
	materialTaskId := uuid.New()
	rootRecordId := uuid.New()
	childRecordId := uuid.New()
	materialRecordId := uuid.New()
	rootSubShelfId := uuid.New()
	childSubShelfId := uuid.New()
	materialId := uuid.New()
	rootFakeId := croutinetasktypes.NewRoutineTaskFakeId()
	childFakeId := croutinetasktypes.NewRoutineTaskFakeId()
	now := time.Now().UTC().Truncate(time.Microsecond)

	seedExecutionUserAndPermission(t, db, actorUserId, rootShelfId, now)
	seedExecutionRoutine(t, db, routineId, routineRecordId, now, 3)
	rootPayload := croutinetasktypes.CreateSubShelfRoutineTaskPayload{
		FakeId:      rootFakeId,
		RootShelfId: rootShelfId,
		Name:        "Root",
		Path:        []uuid.UUID{},
	}
	childPayload := croutinetasktypes.CreateSubShelfRoutineTaskPayload{
		FakeId:         childFakeId,
		RootShelfId:    rootShelfId,
		PrevSubShelfId: referencePointer(rootFakeId),
		Name:           "Child",
		Path:           []uuid.UUID{rootSubShelfId},
	}
	contentType := cenums.MaterialContentType_PlainText
	materialPayload := croutinetasktypes.CreateMaterialRoutineTaskPayload{
		ParentSubShelfId: childFakeId,
		Name:             "Material",
		ContentKey:       "material-" + materialTaskId.String(),
		ContentType:      &contentType,
	}
	seedExecutionTask(t, db, routineId, rootTaskId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, rootPayload, now)
	seedExecutionTask(t, db, routineId, childTaskId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, childPayload, now)
	seedExecutionTask(t, db, routineId, materialTaskId, actorUserId, cenums.RoutineTaskPurpose_CreateMaterial, materialPayload, now)
	seedExecutionTaskRecords(t, db, routineRecordId, rootTaskId, rootRecordId, childTaskId, childRecordId, now)
	seedExecutionTaskRecord(t, db, routineRecordId, materialTaskId, materialRecordId, cenums.RoutineTaskPurpose_CreateMaterial, now)
	plan := croutinetasktypes.RoutineTaskPlan{
		RoutineId: routineId,
		Facts: map[string]uuid.UUID{
			string(rootFakeId):  rootSubShelfId,
			string(childFakeId): childSubShelfId,
		},
		PrecreatedSubShelves: map[string]croutinetasktypes.PrecreatedSubShelf{
			string(rootFakeId): {
				TaskId:      rootTaskId,
				FakeId:      rootFakeId,
				RealId:      rootSubShelfId,
				RootShelfId: rootShelfId,
				Name:        "Root",
				Path:        []uuid.UUID{},
			},
			string(childFakeId): {
				TaskId:           childTaskId,
				FakeId:           childFakeId,
				RealId:           childSubShelfId,
				RootShelfId:      rootShelfId,
				ParentSubShelfId: referencePointer(rootFakeId),
				Name:             "Child",
				Path:             []uuid.UUID{rootSubShelfId},
			},
		},
		PrecreatedSubShelfOrder: []string{string(rootFakeId), string(childFakeId)},
		ContainerObjectTaskIds:  []uuid.UUID{rootTaskId, childTaskId},
		CoreObjectTaskIds:       []uuid.UUID{materialTaskId},
		PlannedObjectIds: map[string]uuid.UUID{
			materialTaskId.String(): materialId,
		},
	}
	storeExecutionPlan(t, db, routineRecordId, plan, now)

	service := routinetaskservice.NewRoutineTaskExecutionService(routinetaskvalidation.New(), db, nil, nil)
	request := &cdurablejob.MarkCompletedRoutineTasksRequestDto{
		WorkerId: uuid.New(),
		Tasks: []croutinetasktypes.CompletedRoutineTask{
			completedExecutionTask(rootTaskId, rootRecordId, routineRecordId, routineId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, rootPayload, now),
			completedExecutionTask(childTaskId, childRecordId, routineRecordId, routineId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, childPayload, now),
			completedExecutionTask(materialTaskId, materialRecordId, routineRecordId, routineId, actorUserId, cenums.RoutineTaskPurpose_CreateMaterial, materialPayload, now),
		},
	}
	if exception := service.ApplyPreparedRoutineTasks(t.Context(), uuid.New(), request); exception != nil {
		t.Fatalf("apply nested sub shelf tasks: %v", exception)
	}

	rebuiltService := routinetaskservice.NewRoutineTaskExecutionService(routinetaskvalidation.New(), db, nil, nil)
	if exception := rebuiltService.ApplyPreparedRoutineTasks(t.Context(), uuid.New(), request); exception != nil {
		t.Fatalf("replay nested sub shelf tasks: %v", exception)
	}

	var shelves []struct {
		Id             uuid.UUID        `gorm:"column:id"`
		PrevSubShelfId *uuid.UUID       `gorm:"column:prev_sub_shelf_id"`
		Path           stypes.UUIDArray `gorm:"column:path"`
	}
	if result := db.Table(`"SubShelfTable"`).Select("id, prev_sub_shelf_id, path").Order("id").Find(&shelves); result.Error != nil {
		t.Fatalf("read created sub shelves: %v", result.Error)
	}
	if len(shelves) != 2 {
		t.Fatalf("created sub shelves = %d, want 2", len(shelves))
	}
	var childFound bool
	for _, shelf := range shelves {
		if shelf.Id != childSubShelfId {
			continue
		}
		childFound = true
		if shelf.PrevSubShelfId == nil || *shelf.PrevSubShelfId != rootSubShelfId {
			t.Fatalf("child parent = %v, want %s", shelf.PrevSubShelfId, rootSubShelfId)
		}
		if len(shelf.Path) != 1 || shelf.Path[0] != rootSubShelfId {
			t.Fatalf("child path = %v, want [%s]", shelf.Path, rootSubShelfId)
		}
	}
	if !childFound {
		t.Fatalf("created shelves do not contain planned child id %s", childSubShelfId)
	}

	var material struct {
		Id               uuid.UUID `gorm:"column:id"`
		ParentSubShelfId uuid.UUID `gorm:"column:parent_sub_shelf_id"`
	}
	if result := db.Table(`"MaterialTable"`).Select("id, parent_sub_shelf_id").Where("id = ?", materialId).First(&material); result.Error != nil {
		t.Fatalf("read created material: %v", result.Error)
	}
	if material.ParentSubShelfId != childSubShelfId {
		t.Fatalf("material parent = %s, want %s", material.ParentSubShelfId, childSubShelfId)
	}
	var materialCount int64
	if result := db.Table(`"MaterialTable"`).Count(&materialCount); result.Error != nil {
		t.Fatalf("count created materials after replay: %v", result.Error)
	} else if materialCount != 1 {
		t.Fatalf("created materials after replay = %d, want 1", materialCount)
	}

	assertExecutionTaskRecordStatuses(t, db, routineRecordId, cenums.RoutineTaskRecordStatus_Success, 3)
}

func TestApplyPreparedRoutineTasksRollsBackAllObjectsWhenPermissionFails(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	routineRecordId := uuid.New()
	permittedRootShelfId := uuid.New()
	deniedRootShelfId := uuid.New()
	permittedTaskId := uuid.New()
	deniedTaskId := uuid.New()
	permittedRecordId := uuid.New()
	deniedRecordId := uuid.New()
	permittedSubShelfId := uuid.New()
	deniedSubShelfId := uuid.New()
	permittedFakeId := croutinetasktypes.NewRoutineTaskFakeId()
	deniedFakeId := croutinetasktypes.NewRoutineTaskFakeId()
	now := time.Now().UTC().Truncate(time.Microsecond)

	seedExecutionUserAndPermission(t, db, actorUserId, permittedRootShelfId, now)
	seedExecutionRoutine(t, db, routineId, routineRecordId, now, 2)
	permittedPayload := croutinetasktypes.CreateSubShelfRoutineTaskPayload{
		FakeId:      permittedFakeId,
		RootShelfId: permittedRootShelfId,
		Name:        "Permitted",
		Path:        []uuid.UUID{},
	}
	deniedPayload := croutinetasktypes.CreateSubShelfRoutineTaskPayload{
		FakeId:      deniedFakeId,
		RootShelfId: deniedRootShelfId,
		Name:        "Denied",
		Path:        []uuid.UUID{},
	}
	seedExecutionTask(t, db, routineId, permittedTaskId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, permittedPayload, now)
	seedExecutionTask(t, db, routineId, deniedTaskId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, deniedPayload, now)
	seedExecutionTaskRecords(t, db, routineRecordId, permittedTaskId, permittedRecordId, deniedTaskId, deniedRecordId, now)
	storeExecutionPlan(t, db, routineRecordId, croutinetasktypes.RoutineTaskPlan{
		RoutineId: routineId,
		Facts: map[string]uuid.UUID{
			string(permittedFakeId): permittedSubShelfId,
			string(deniedFakeId):    deniedSubShelfId,
		},
		PrecreatedSubShelves: map[string]croutinetasktypes.PrecreatedSubShelf{
			string(permittedFakeId): {
				TaskId:      permittedTaskId,
				FakeId:      permittedFakeId,
				RealId:      permittedSubShelfId,
				RootShelfId: permittedRootShelfId,
				Name:        "Permitted",
				Path:        []uuid.UUID{},
			},
			string(deniedFakeId): {
				TaskId:      deniedTaskId,
				FakeId:      deniedFakeId,
				RealId:      deniedSubShelfId,
				RootShelfId: deniedRootShelfId,
				Name:        "Denied",
				Path:        []uuid.UUID{},
			},
		},
		PrecreatedSubShelfOrder: []string{string(permittedFakeId), string(deniedFakeId)},
		ContainerObjectTaskIds:  []uuid.UUID{permittedTaskId, deniedTaskId},
		CoreObjectTaskIds:       []uuid.UUID{},
		PlannedObjectIds:        map[string]uuid.UUID{},
	}, now)

	service := routinetaskservice.NewRoutineTaskExecutionService(routinetaskvalidation.New(), db, nil, nil)
	request := &cdurablejob.MarkCompletedRoutineTasksRequestDto{
		WorkerId: uuid.New(),
		Tasks: []croutinetasktypes.CompletedRoutineTask{
			completedExecutionTask(permittedTaskId, permittedRecordId, routineRecordId, routineId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, permittedPayload, now),
			completedExecutionTask(deniedTaskId, deniedRecordId, routineRecordId, routineId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, deniedPayload, now),
		},
	}
	if exception := service.ApplyPreparedRoutineTasks(t.Context(), uuid.New(), request); exception == nil {
		t.Fatal("apply tasks succeeded, want permission failure")
	}

	var shelfCount int64
	if result := db.Table(`"SubShelfTable"`).Count(&shelfCount); result.Error != nil {
		t.Fatalf("count rolled back sub shelves: %v", result.Error)
	} else if shelfCount != 0 {
		t.Fatalf("rolled back sub shelves = %d, want 0", shelfCount)
	}
	assertExecutionTaskRecordStatuses(t, db, routineRecordId, cenums.RoutineTaskRecordStatus_Running, 2)
}

func TestApplyPreparedRoutineTasksRollsBackOnDeferredParentForeignKeyFailure(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	routineRecordId := uuid.New()
	rootShelfId := uuid.New()
	taskId := uuid.New()
	recordId := uuid.New()
	subShelfId := uuid.New()
	missingParentId := uuid.New()
	fakeId := croutinetasktypes.NewRoutineTaskFakeId()
	now := time.Now().UTC().Truncate(time.Microsecond)

	seedExecutionUserAndPermission(t, db, actorUserId, rootShelfId, now)
	seedExecutionRoutine(t, db, routineId, routineRecordId, now, 1)
	payload := croutinetasktypes.CreateSubShelfRoutineTaskPayload{
		FakeId:         fakeId,
		RootShelfId:    rootShelfId,
		PrevSubShelfId: referencePointer(croutinetasktypes.RoutineTaskObjectReference(missingParentId.String())),
		Path:           []uuid.UUID{},
		Name:           "Invalid parent",
	}
	seedExecutionTask(t, db, routineId, taskId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, payload, now)
	seedExecutionTaskRecord(t, db, routineRecordId, taskId, recordId, cenums.RoutineTaskPurpose_CreateSubShelf, now)
	storeExecutionPlan(t, db, routineRecordId, croutinetasktypes.RoutineTaskPlan{
		RoutineId: routineId,
		Facts: map[string]uuid.UUID{
			string(fakeId): subShelfId,
		},
		PrecreatedSubShelves: map[string]croutinetasktypes.PrecreatedSubShelf{
			string(fakeId): {
				TaskId:           taskId,
				FakeId:           fakeId,
				RealId:           subShelfId,
				RootShelfId:      rootShelfId,
				ParentSubShelfId: referencePointer(croutinetasktypes.RoutineTaskObjectReference(missingParentId.String())),
				Name:             "Invalid parent",
				Path:             []uuid.UUID{},
			},
		},
		PrecreatedSubShelfOrder: []string{string(fakeId)},
		ContainerObjectTaskIds:  []uuid.UUID{taskId},
		CoreObjectTaskIds:       []uuid.UUID{},
		PlannedObjectIds:        map[string]uuid.UUID{},
	}, now)

	service := routinetaskservice.NewRoutineTaskExecutionService(routinetaskvalidation.New(), db, nil, nil)
	request := &cdurablejob.MarkCompletedRoutineTasksRequestDto{
		WorkerId: uuid.New(),
		Tasks: []croutinetasktypes.CompletedRoutineTask{
			completedExecutionTask(taskId, recordId, routineRecordId, routineId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, payload, now),
		},
	}
	if exception := service.ApplyPreparedRoutineTasks(t.Context(), uuid.New(), request); exception == nil {
		t.Fatal("apply task succeeded, want deferred parent foreign key failure")
	}

	var shelfCount int64
	if result := db.Table(`"SubShelfTable"`).Count(&shelfCount); result.Error != nil {
		t.Fatalf("count rolled back deferred-fk sub shelf: %v", result.Error)
	} else if shelfCount != 0 {
		t.Fatalf("deferred-fk rolled back sub shelves = %d, want 0", shelfCount)
	}
	assertExecutionTaskRecordStatuses(t, db, routineRecordId, cenums.RoutineTaskRecordStatus_Running, 1)
}

func TestApplyPreparedRoutineTasksRollsBackOnTriggerFailure(t *testing.T) {
	db := openRoutineTaskClaimIntegrationDB(t)
	actorUserId := uuid.New()
	routineId := uuid.New()
	routineRecordId := uuid.New()
	firstRootShelfId := uuid.New()
	secondRootShelfId := uuid.New()
	firstTaskId := uuid.New()
	secondTaskId := uuid.New()
	firstRecordId := uuid.New()
	secondRecordId := uuid.New()
	firstSubShelfId := uuid.New()
	secondSubShelfId := uuid.New()
	firstFakeId := croutinetasktypes.NewRoutineTaskFakeId()
	secondFakeId := croutinetasktypes.NewRoutineTaskFakeId()
	now := time.Now().UTC().Truncate(time.Microsecond)

	seedExecutionUserAndPermission(t, db, actorUserId, firstRootShelfId, now)
	if result := db.Exec(
		`INSERT INTO "UsersToShelvesTable" (user_id, root_shelf_id, permission, updated_at, created_at) VALUES (?, ?, ?, ?, ?)`,
		actorUserId,
		secondRootShelfId,
		cenums.AccessControlPermission_Owner,
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed second execution shelf permission: %v", result.Error)
	}
	seedExecutionRoutine(t, db, routineId, routineRecordId, now, 2)
	firstPayload := croutinetasktypes.CreateSubShelfRoutineTaskPayload{
		FakeId:      firstFakeId,
		RootShelfId: firstRootShelfId,
		Name:        "Allowed",
		Path:        []uuid.UUID{},
	}
	secondPayload := croutinetasktypes.CreateSubShelfRoutineTaskPayload{
		FakeId:      secondFakeId,
		RootShelfId: secondRootShelfId,
		Name:        "TriggerFailure",
		Path:        []uuid.UUID{},
	}
	seedExecutionTask(t, db, routineId, firstTaskId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, firstPayload, now)
	seedExecutionTask(t, db, routineId, secondTaskId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, secondPayload, now)
	seedExecutionTaskRecords(t, db, routineRecordId, firstTaskId, firstRecordId, secondTaskId, secondRecordId, now)
	storeExecutionPlan(t, db, routineRecordId, croutinetasktypes.RoutineTaskPlan{
		RoutineId: routineId,
		Facts: map[string]uuid.UUID{
			string(firstFakeId):  firstSubShelfId,
			string(secondFakeId): secondSubShelfId,
		},
		PrecreatedSubShelves: map[string]croutinetasktypes.PrecreatedSubShelf{
			string(firstFakeId): {
				TaskId:      firstTaskId,
				FakeId:      firstFakeId,
				RealId:      firstSubShelfId,
				RootShelfId: firstRootShelfId,
				Name:        "Allowed",
				Path:        []uuid.UUID{},
			},
			string(secondFakeId): {
				TaskId:      secondTaskId,
				FakeId:      secondFakeId,
				RealId:      secondSubShelfId,
				RootShelfId: secondRootShelfId,
				Name:        "TriggerFailure",
				Path:        []uuid.UUID{},
			},
		},
		PrecreatedSubShelfOrder: []string{string(firstFakeId), string(secondFakeId)},
		ContainerObjectTaskIds:  []uuid.UUID{firstTaskId, secondTaskId},
		CoreObjectTaskIds:       []uuid.UUID{},
		PlannedObjectIds:        map[string]uuid.UUID{},
	}, now)
	if result := db.Exec(`
		CREATE FUNCTION fail_routine_task_sub_shelf_insert() RETURNS trigger AS $$
		BEGIN
			IF NEW.name = 'TriggerFailure' THEN
				RAISE EXCEPTION 'routine task trigger failure';
			END IF;
			RETURN NEW;
		END;
		$$ LANGUAGE plpgsql`); result.Error != nil {
		t.Fatalf("create sub shelf failure trigger function: %v", result.Error)
	}
	if result := db.Exec(`
		CREATE TRIGGER fail_routine_task_sub_shelf_insert
		BEFORE INSERT ON "SubShelfTable"
		FOR EACH ROW EXECUTE FUNCTION fail_routine_task_sub_shelf_insert()`); result.Error != nil {
		t.Fatalf("create sub shelf failure trigger: %v", result.Error)
	}

	service := routinetaskservice.NewRoutineTaskExecutionService(routinetaskvalidation.New(), db, nil, nil)
	request := &cdurablejob.MarkCompletedRoutineTasksRequestDto{
		WorkerId: uuid.New(),
		Tasks: []croutinetasktypes.CompletedRoutineTask{
			completedExecutionTask(firstTaskId, firstRecordId, routineRecordId, routineId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, firstPayload, now),
			completedExecutionTask(secondTaskId, secondRecordId, routineRecordId, routineId, actorUserId, cenums.RoutineTaskPurpose_CreateSubShelf, secondPayload, now),
		},
	}
	if exception := service.ApplyPreparedRoutineTasks(t.Context(), uuid.New(), request); exception == nil {
		t.Fatal("apply tasks succeeded, want trigger failure")
	}

	var shelfCount int64
	if result := db.Table(`"SubShelfTable"`).Count(&shelfCount); result.Error != nil {
		t.Fatalf("count rolled back trigger-failure sub shelves: %v", result.Error)
	} else if shelfCount != 0 {
		t.Fatalf("trigger-failure rolled back sub shelves = %d, want 0", shelfCount)
	}
	assertExecutionTaskRecordStatuses(t, db, routineRecordId, cenums.RoutineTaskRecordStatus_Running, 2)
}

func completedExecutionTask(
	taskId uuid.UUID,
	recordId uuid.UUID,
	routineRecordId uuid.UUID,
	routineId uuid.UUID,
	actorUserId uuid.UUID,
	purpose cenums.RoutineTaskPurpose,
	payload any,
	now time.Time,
) croutinetasktypes.CompletedRoutineTask {
	payloadBytes, _ := json.Marshal(payload)
	return croutinetasktypes.CompletedRoutineTask{
		RoutineTaskId:       taskId,
		RoutineTaskRecordId: recordId,
		RoutineRecordId:     routineRecordId,
		CompletedAt:         now,
		PreparedTask: &croutinetasktypes.PreparedRoutineTask{
			RoutineTaskId:       taskId,
			RoutineTaskRecordId: recordId,
			RoutineRecordId:     routineRecordId,
			RoutineId:           routineId,
			ActorUserId:         actorUserId,
			ActorUserPublicId:   uuid.New(),
			Attempt:             1,
			Purpose:             purpose,
			Payload:             payloadBytes,
			PreparedAt:          now,
		},
	}
}

func referencePointer(reference croutinetasktypes.RoutineTaskObjectReference) *croutinetasktypes.RoutineTaskObjectReference {
	return &reference
}

func seedExecutionUserAndPermission(t *testing.T, db *gorm.DB, userId, rootShelfId uuid.UUID, now time.Time) {
	t.Helper()
	if result := db.Exec(
		`INSERT INTO "UserView" (id, public_id, plan, status, created_at) VALUES (?, ?, ?, ?, ?)`,
		userId,
		uuid.New(),
		cenums.UserPlan_Free,
		cenums.UserStatus_Online,
		now,
	); result.Error != nil {
		t.Fatalf("seed execution user: %v", result.Error)
	}
	if result := db.Exec(
		`INSERT INTO "UsersToShelvesTable" (user_id, root_shelf_id, permission, updated_at, created_at) VALUES (?, ?, ?, ?, ?)`,
		userId,
		rootShelfId,
		cenums.AccessControlPermission_Owner,
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed execution shelf permission: %v", result.Error)
	}
}

func seedExecutionRoutine(t *testing.T, db *gorm.DB, routineId, routineRecordId uuid.UUID, now time.Time, taskCount int) {
	t.Helper()
	if result := db.Exec(
		`INSERT INTO "RoutineTable" (id, title, description, is_pinned, scheduled_start_at, scheduled_end_at, period, timezone, status, definition_version, updated_at, created_at) VALUES (?, 'Execution routine', '', false, ?, ?, NULL, 'UTC', ?, 1, ?, ?)`,
		routineId,
		now,
		now.Add(time.Hour),
		cenums.RoutineStatus_InProgress,
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed execution routine: %v", result.Error)
	}
	if result := db.Exec(
		`INSERT INTO "RoutineRecordTable" (id, routine_id, definition_version, status, scheduled_at, total_task_count, success_task_count, failed_task_count, blocked_task_count, running_task_count, waiting_task_count, snapshot, updated_at, created_at) VALUES (?, ?, 1, ?, ?, ?, 0, 0, ?, 0, 0, '{}'::jsonb, ?, ?)`,
		routineRecordId,
		routineId,
		cenums.RoutineRecordStatus_Running,
		now,
		taskCount,
		taskCount,
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed execution routine record: %v", result.Error)
	}
}

func seedExecutionTask(
	t *testing.T,
	db *gorm.DB,
	routineId, taskId, actorUserId uuid.UUID,
	purpose cenums.RoutineTaskPurpose,
	payload any,
	now time.Time,
) {
	t.Helper()
	payloadBytes, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("encode execution task payload: %v", err)
	}
	if result := db.Exec(
		`INSERT INTO "RoutineTaskTable" (id, routine_id, actor_user_id, title, purpose, payload, cost_unit, priority, max_attempts, updated_at, created_at) VALUES (?, ?, ?, 'Create sub shelf', ?, ?::jsonb, 1, 0, 1, ?, ?)`,
		taskId,
		routineId,
		actorUserId,
		purpose,
		payloadBytes,
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed execution task: %v", result.Error)
	}
}

func seedExecutionTaskRecords(
	t *testing.T,
	db *gorm.DB,
	routineRecordId, firstTaskId, firstRecordId, secondTaskId, secondRecordId uuid.UUID,
	now time.Time,
) {
	t.Helper()
	for _, values := range [][2]uuid.UUID{{firstTaskId, firstRecordId}, {secondTaskId, secondRecordId}} {
		seedExecutionTaskRecord(t, db, routineRecordId, values[0], values[1], cenums.RoutineTaskPurpose_CreateSubShelf, now)
	}
}

func seedExecutionTaskRecord(
	t *testing.T,
	db *gorm.DB,
	routineRecordId, taskId, recordId uuid.UUID,
	purpose cenums.RoutineTaskPurpose,
	now time.Time,
) {
	t.Helper()
	if result := db.Exec(
		`INSERT INTO "RoutineTaskRecordTable" (id, routine_record_id, routine_task_id, purpose, status, cost_unit, attempts, payload_snapshot, result_snapshot, updated_at, created_at) VALUES (?, ?, ?, ?, ?, 1, 1, '{}'::jsonb, '{}'::jsonb, ?, ?)`,
		recordId,
		routineRecordId,
		taskId,
		purpose,
		cenums.RoutineTaskRecordStatus_Running,
		now,
		now,
	); result.Error != nil {
		t.Fatalf("seed execution task record: %v", result.Error)
	}
}

func storeExecutionPlan(t *testing.T, db *gorm.DB, routineRecordId uuid.UUID, plan croutinetasktypes.RoutineTaskPlan, now time.Time) {
	t.Helper()
	snapshot, err := json.Marshal(map[string]any{"routineTaskPlan": plan})
	if err != nil {
		t.Fatalf("encode execution plan: %v", err)
	}
	if result := db.Model(&struct{}{}).Table(`"RoutineRecordTable"`).Where("id = ?", routineRecordId).Updates(map[string]any{"snapshot": datatypes.JSON(snapshot), "updated_at": now}); result.Error != nil {
		t.Fatalf("store execution plan: %v", result.Error)
	}
}

func assertExecutionTaskRecordStatuses(t *testing.T, db *gorm.DB, routineRecordId uuid.UUID, expected cenums.RoutineTaskRecordStatus, expectedCount int) {
	t.Helper()
	var statuses []string
	if result := db.Table(`"RoutineTaskRecordTable"`).Select("status").Where("routine_record_id = ?", routineRecordId).Find(&statuses); result.Error != nil {
		t.Fatalf("read execution task record statuses: %v", result.Error)
	}
	if len(statuses) != expectedCount {
		t.Fatalf("execution task record statuses = %v, want %d records with %s", statuses, expectedCount, expected)
	}
	for _, status := range statuses {
		if status != string(expected) {
			t.Fatalf("execution task record status = %s, want %s", status, expected)
		}
	}
}
