package builders

import (
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/google/uuid"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

func TestDeterministicPlanBuilderBuildsFactsAndTopologicalSubShelfPlan(t *testing.T) {
	routineId := uuid.New()
	rootShelfId := uuid.New()
	rootTaskId := uuid.New()
	nestedTaskId := uuid.New()
	blockPackTaskId := uuid.New()
	materialTaskId := uuid.New()
	rootFakeId := croutinetasktypes.RoutineTaskObjectReference("f_11111111111111111111111111111111")
	nestedFakeId := croutinetasktypes.RoutineTaskObjectReference("f_22222222222222222222222222222222")

	rootPayload, err := json.Marshal(croutinetasktypes.CreateSubShelfRoutineTaskPayload{
		FakeId:      rootFakeId,
		RootShelfId: rootShelfId,
		Name:        "Root",
	})
	if err != nil {
		t.Fatalf("marshal root payload: %v", err)
	}
	nestedPayload, err := json.Marshal(croutinetasktypes.CreateSubShelfRoutineTaskPayload{
		FakeId:         nestedFakeId,
		RootShelfId:    rootShelfId,
		PrevSubShelfId: &rootFakeId,
		Name:           "Nested",
	})
	if err != nil {
		t.Fatalf("marshal nested payload: %v", err)
	}
	blockPackPayload, err := json.Marshal(croutinetasktypes.CreateBlockPackRoutineTaskPayload{
		TargetSubShelfId: nestedFakeId,
		Template: croutinetasktypes.CreateBlockPackRoutineTaskTemplate{
			Name: "Block pack",
			Blocks: []croutinetasktypes.CreateBlockPackRoutineTaskTemplateBlock{
				{ClientId: "block-1"},
			},
		},
	})
	if err != nil {
		t.Fatalf("marshal block pack payload: %v", err)
	}
	materialPayload, err := json.Marshal(croutinetasktypes.CreateMaterialRoutineTaskPayload{
		ParentSubShelfId: rootFakeId,
		Name:             "Material",
		ContentKey:       "material-key",
	})
	if err != nil {
		t.Fatalf("marshal material payload: %v", err)
	}

	plan, err := (&DeterministicPlanBuilder{}).Build(
		routineId,
		[]sschemas.RoutineTask{
			{Id: rootTaskId, RoutineId: routineId, Purpose: cenums.RoutineTaskPurpose_CreateSubShelf, Payload: rootPayload},
			{Id: nestedTaskId, RoutineId: routineId, Purpose: cenums.RoutineTaskPurpose_CreateSubShelf, Payload: nestedPayload},
			{Id: blockPackTaskId, RoutineId: routineId, Purpose: cenums.RoutineTaskPurpose_CreateBlockPack, Payload: blockPackPayload},
			{Id: materialTaskId, RoutineId: routineId, Purpose: cenums.RoutineTaskPurpose_CreateMaterial, Payload: materialPayload},
		},
		[]sschemas.RoutineTaskDependency{
			{RoutineTaskId: nestedTaskId, PreviousRoutineTaskId: rootTaskId},
			{RoutineTaskId: blockPackTaskId, PreviousRoutineTaskId: nestedTaskId},
			{RoutineTaskId: materialTaskId, PreviousRoutineTaskId: rootTaskId},
		},
		nil,
	)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}

	rootRealId := plan.Facts[string(rootFakeId)]
	nestedRealId := plan.Facts[string(nestedFakeId)]
	if rootRealId == uuid.Nil || nestedRealId == uuid.Nil || rootRealId == nestedRealId {
		t.Fatalf("facts = %#v, want distinct planned UUIDs", plan.Facts)
	}
	if got, want := plan.PrecreatedSubShelfOrder, []string{string(rootFakeId), string(nestedFakeId)}; !equalStrings(got, want) {
		t.Fatalf("sub shelf order = %v, want %v", got, want)
	}
	if got := plan.PrecreatedSubShelves[string(nestedFakeId)].Path; len(got) != 1 || got[0] != rootRealId {
		t.Fatalf("nested path = %v, want [%s]", got, rootRealId)
	}
	if len(plan.ContainerObjectTaskIds) != 2 || plan.ContainerObjectTaskIds[0] != rootTaskId || plan.ContainerObjectTaskIds[1] != nestedTaskId {
		t.Fatalf("container object task ids = %v", plan.ContainerObjectTaskIds)
	}
	if len(plan.CoreObjectTaskIds) != 2 || plan.CoreObjectTaskIds[0] != blockPackTaskId || plan.CoreObjectTaskIds[1] != materialTaskId {
		t.Fatalf("core object task ids = %v", plan.CoreObjectTaskIds)
	}
	if plan.PlannedObjectIds[blockPackTaskId.String()] == uuid.Nil || plan.PlannedObjectIds[materialTaskId.String()] == uuid.Nil {
		t.Fatalf("planned object ids = %#v, want IDs for both core create tasks", plan.PlannedObjectIds)
	}
}

func TestDeterministicPlanBuilderBuildsMultipleRootsAndArbitraryDepth(t *testing.T) {
	routineId := uuid.New()
	rootShelfId := uuid.New()
	taskIds := make([]uuid.UUID, 5)
	fakeIds := make([]croutinetasktypes.RoutineTaskObjectReference, 5)
	payloads := make([][]byte, 5)
	parents := []int{-1, -1, 0, 2, 3}
	for index := range taskIds {
		taskIds[index] = uuid.New()
		fakeIds[index] = []croutinetasktypes.RoutineTaskObjectReference{
			"f_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			"f_bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
			"f_cccccccccccccccccccccccccccccccc",
			"f_dddddddddddddddddddddddddddddddd",
			"f_eeeeeeeeeeeeeeeeeeeeeeeeeeeeeeee",
		}[index]
		payload := croutinetasktypes.CreateSubShelfRoutineTaskPayload{
			FakeId:      fakeIds[index],
			RootShelfId: rootShelfId,
			Name:        fmt.Sprintf("Shelf %d", index),
		}
		if parents[index] >= 0 {
			payload.PrevSubShelfId = &fakeIds[parents[index]]
		}
		var err error
		payloads[index], err = json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal shelf payload %d: %v", index, err)
		}
	}

	tasks := make([]sschemas.RoutineTask, len(taskIds))
	for index, taskId := range taskIds {
		tasks[index] = sschemas.RoutineTask{
			Id:        taskId,
			RoutineId: routineId,
			Purpose:   cenums.RoutineTaskPurpose_CreateSubShelf,
			Payload:   payloads[index],
		}
	}
	dependencies := []sschemas.RoutineTaskDependency{
		{RoutineTaskId: taskIds[2], PreviousRoutineTaskId: taskIds[0]},
		{RoutineTaskId: taskIds[3], PreviousRoutineTaskId: taskIds[2]},
		{RoutineTaskId: taskIds[4], PreviousRoutineTaskId: taskIds[3]},
	}

	plan, err := (&DeterministicPlanBuilder{}).Build(routineId, tasks, dependencies, nil)
	if err != nil {
		t.Fatalf("build arbitrary-depth plan: %v", err)
	}
	if got, want := plan.PrecreatedSubShelfOrder, []string{
		string(fakeIds[0]),
		string(fakeIds[1]),
		string(fakeIds[2]),
		string(fakeIds[3]),
		string(fakeIds[4]),
	}; !equalStrings(got, want) {
		t.Fatalf("sub shelf order = %v, want %v", got, want)
	}
	if path := plan.PrecreatedSubShelves[string(fakeIds[4])].Path; len(path) != 3 || path[0] != plan.Facts[string(fakeIds[0])] || path[1] != plan.Facts[string(fakeIds[2])] || path[2] != plan.Facts[string(fakeIds[3])] {
		t.Fatalf("deep shelf path = %v, want the three real ancestor ids", path)
	}
}

func TestDeterministicPlanBuilderReusesExistingFact(t *testing.T) {
	routineId := uuid.New()
	fakeId := croutinetasktypes.RoutineTaskObjectReference("f_33333333333333333333333333333333")
	existingRealId := uuid.New()
	payload, err := json.Marshal(croutinetasktypes.CreateSubShelfRoutineTaskPayload{
		FakeId:      fakeId,
		RootShelfId: uuid.New(),
		Name:        "Existing plan",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	plan, err := (&DeterministicPlanBuilder{}).Build(
		routineId,
		[]sschemas.RoutineTask{{
			Id:        uuid.New(),
			RoutineId: routineId,
			Purpose:   cenums.RoutineTaskPurpose_CreateSubShelf,
			Payload:   payload,
		}},
		nil,
		&croutinetasktypes.RoutineTaskPlan{
			RoutineId: routineId,
			Facts: map[string]uuid.UUID{
				string(fakeId): existingRealId,
			},
		},
	)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if plan.Facts[string(fakeId)] != existingRealId {
		t.Fatalf("fact = %s, want existing real id %s", plan.Facts[string(fakeId)], existingRealId)
	}
}

func TestDeterministicPlanBuilderReusesPlannedCoreObjectId(t *testing.T) {
	routineId := uuid.New()
	taskId := uuid.New()
	plannedObjectId := uuid.New()
	payload, err := json.Marshal(croutinetasktypes.CreateMaterialRoutineTaskPayload{
		ParentSubShelfId: croutinetasktypes.RoutineTaskObjectReference(uuid.New().String()),
		Name:             "Existing material",
	})
	if err != nil {
		t.Fatalf("marshal material payload: %v", err)
	}

	plan, err := (&DeterministicPlanBuilder{}).Build(
		routineId,
		[]sschemas.RoutineTask{{
			Id:        taskId,
			RoutineId: routineId,
			Purpose:   cenums.RoutineTaskPurpose_CreateMaterial,
			Payload:   payload,
		}},
		nil,
		&croutinetasktypes.RoutineTaskPlan{
			RoutineId: routineId,
			PlannedObjectIds: map[string]uuid.UUID{
				taskId.String(): plannedObjectId,
			},
		},
	)
	if err != nil {
		t.Fatalf("build plan: %v", err)
	}
	if plan.PlannedObjectIds[taskId.String()] != plannedObjectId {
		t.Fatalf("planned object id = %s, want %s", plan.PlannedObjectIds[taskId.String()], plannedObjectId)
	}
}

func TestDeterministicPlanBuilderRejectsUnknownFakeParent(t *testing.T) {
	routineId := uuid.New()
	fakeId := croutinetasktypes.RoutineTaskObjectReference("f_44444444444444444444444444444444")
	unknownParent := croutinetasktypes.RoutineTaskObjectReference("f_55555555555555555555555555555555")
	payload, err := json.Marshal(croutinetasktypes.CreateSubShelfRoutineTaskPayload{
		FakeId:         fakeId,
		RootShelfId:    uuid.New(),
		PrevSubShelfId: &unknownParent,
		Name:           "Invalid",
	})
	if err != nil {
		t.Fatalf("marshal payload: %v", err)
	}

	_, err = (&DeterministicPlanBuilder{}).Build(
		routineId,
		[]sschemas.RoutineTask{{
			Id:        uuid.New(),
			RoutineId: routineId,
			Purpose:   cenums.RoutineTaskPurpose_CreateSubShelf,
			Payload:   payload,
		}},
		nil,
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "unknown fake parent") {
		t.Fatalf("error = %v, want unknown fake parent", err)
	}
}

func TestDeterministicPlanBuilderRejectsCycles(t *testing.T) {
	routineId := uuid.New()
	firstFakeId := croutinetasktypes.RoutineTaskObjectReference("f_66666666666666666666666666666666")
	secondFakeId := croutinetasktypes.RoutineTaskObjectReference("f_77777777777777777777777777777777")
	firstTaskId := uuid.New()
	secondTaskId := uuid.New()
	rootShelfId := uuid.New()
	firstPayload, err := json.Marshal(croutinetasktypes.CreateSubShelfRoutineTaskPayload{
		FakeId:         firstFakeId,
		RootShelfId:    rootShelfId,
		PrevSubShelfId: &secondFakeId,
		Name:           "First",
	})
	if err != nil {
		t.Fatalf("marshal first payload: %v", err)
	}
	secondPayload, err := json.Marshal(croutinetasktypes.CreateSubShelfRoutineTaskPayload{
		FakeId:         secondFakeId,
		RootShelfId:    rootShelfId,
		PrevSubShelfId: &firstFakeId,
		Name:           "Second",
	})
	if err != nil {
		t.Fatalf("marshal second payload: %v", err)
	}

	_, err = (&DeterministicPlanBuilder{}).Build(
		routineId,
		[]sschemas.RoutineTask{
			{Id: firstTaskId, RoutineId: routineId, Purpose: cenums.RoutineTaskPurpose_CreateSubShelf, Payload: firstPayload},
			{Id: secondTaskId, RoutineId: routineId, Purpose: cenums.RoutineTaskPurpose_CreateSubShelf, Payload: secondPayload},
		},
		[]sschemas.RoutineTaskDependency{{
			RoutineTaskId:         firstTaskId,
			PreviousRoutineTaskId: secondTaskId,
		}, {
			RoutineTaskId:         secondTaskId,
			PreviousRoutineTaskId: firstTaskId,
		}},
		nil,
	)
	if err == nil || !strings.Contains(err.Error(), "dependency graph contains a cycle") {
		t.Fatalf("error = %v, want task dependency cycle", err)
	}
}

func equalStrings(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}
