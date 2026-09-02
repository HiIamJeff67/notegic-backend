package builders

import (
	"encoding/json"
	"fmt"
	"sort"

	"github.com/google/uuid"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	routinetaskdependencies "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/services/routinetask/dependencies"
)

type DeterministicPlanBuilder struct{}

func (b *DeterministicPlanBuilder) Build(
	routineId uuid.UUID,
	tasks []sschemas.RoutineTask,
	dependencies []sschemas.RoutineTaskDependency,
	existingPlan *croutinetasktypes.RoutineTaskPlan,
) (*croutinetasktypes.RoutineTaskPlan, error) {
	if routineId == uuid.Nil {
		return nil, fmt.Errorf("routine task plan routine id is required")
	}
	if len(tasks) == 0 {
		return nil, fmt.Errorf("routine task plan must contain at least one task")
	}

	taskById := make(map[uuid.UUID]sschemas.RoutineTask, len(tasks))
	taskOrder := make([]uuid.UUID, len(tasks))
	for index, task := range tasks {
		if task.Id == uuid.Nil || task.RoutineId != routineId {
			return nil, fmt.Errorf("routine task %q does not belong to routine %q", task.Id, routineId)
		}
		if _, exists := taskById[task.Id]; exists {
			return nil, fmt.Errorf("routine task %q is duplicated", task.Id)
		}
		taskById[task.Id] = task
		taskOrder[index] = task.Id
	}

	nextTaskIdsByPreviousId := make(map[uuid.UUID][]uuid.UUID, len(tasks))
	taskDependencyCount := make(map[uuid.UUID]int, len(tasks))
	seenDependencies := make(map[[2]uuid.UUID]struct{}, len(dependencies))
	dependencyEdges := make([]routinetaskdependencies.Edge, 0, len(dependencies))
	for _, dependency := range dependencies {
		if dependency.RoutineTaskId == uuid.Nil || dependency.PreviousRoutineTaskId == uuid.Nil {
			return nil, fmt.Errorf("routine task dependency contains an empty task id")
		}
		if _, exists := taskById[dependency.RoutineTaskId]; !exists {
			return nil, fmt.Errorf("routine task dependency child %q is not in the routine", dependency.RoutineTaskId)
		}
		if _, exists := taskById[dependency.PreviousRoutineTaskId]; !exists {
			return nil, fmt.Errorf("routine task dependency parent %q is not in the routine", dependency.PreviousRoutineTaskId)
		}
		if dependency.RoutineTaskId == dependency.PreviousRoutineTaskId {
			return nil, fmt.Errorf("routine task %q depends on itself", dependency.RoutineTaskId)
		}
		dependencyKey := [2]uuid.UUID{dependency.RoutineTaskId, dependency.PreviousRoutineTaskId}
		if _, exists := seenDependencies[dependencyKey]; exists {
			return nil, fmt.Errorf("routine task dependency %q -> %q is duplicated", dependency.RoutineTaskId, dependency.PreviousRoutineTaskId)
		}
		seenDependencies[dependencyKey] = struct{}{}
		dependencyEdges = append(dependencyEdges, routinetaskdependencies.Edge{
			TaskId:         dependency.RoutineTaskId,
			PreviousTaskId: dependency.PreviousRoutineTaskId,
		})
		taskDependencyCount[dependency.RoutineTaskId]++
		nextTaskIdsByPreviousId[dependency.PreviousRoutineTaskId] = append(
			nextTaskIdsByPreviousId[dependency.PreviousRoutineTaskId],
			dependency.RoutineTaskId,
		)
	}
	if err := routinetaskdependencies.Validate(taskOrder, dependencyEdges); err != nil {
		return nil, err
	}

	readyTaskIds := make([]uuid.UUID, 0, len(tasks))
	for _, taskId := range taskOrder {
		if taskDependencyCount[taskId] == 0 {
			readyTaskIds = append(readyTaskIds, taskId)
		}
	}
	taskAncestorsById := make(map[uuid.UUID]map[uuid.UUID]struct{}, len(tasks))
	for len(readyTaskIds) > 0 {
		taskId := readyTaskIds[0]
		readyTaskIds = readyTaskIds[1:]
		for _, nextTaskId := range nextTaskIdsByPreviousId[taskId] {
			if taskAncestorsById[nextTaskId] == nil {
				taskAncestorsById[nextTaskId] = make(map[uuid.UUID]struct{})
			}
			taskAncestorsById[nextTaskId][taskId] = struct{}{}
			for ancestorTaskId := range taskAncestorsById[taskId] {
				taskAncestorsById[nextTaskId][ancestorTaskId] = struct{}{}
			}
			taskDependencyCount[nextTaskId]--
			if taskDependencyCount[nextTaskId] == 0 {
				readyTaskIds = append(readyTaskIds, nextTaskId)
			}
		}
	}
	existingFacts := map[string]uuid.UUID(nil)
	existingPlannedObjectIds := map[string]uuid.UUID(nil)
	if existingPlan != nil {
		if existingPlan.RoutineId != routineId {
			return nil, fmt.Errorf("existing routine task plan belongs to routine %q, not %q", existingPlan.RoutineId, routineId)
		}
		existingFacts = existingPlan.Facts
		existingPlannedObjectIds = existingPlan.PlannedObjectIds
	}
	facts := make(map[string]uuid.UUID, len(existingFacts))
	for fakeId, realId := range existingFacts {
		reference := croutinetasktypes.RoutineTaskObjectReference(fakeId)
		if !reference.IsFakeId() || realId == uuid.Nil {
			return nil, fmt.Errorf("existing fact %q is invalid", fakeId)
		}
		facts[fakeId] = realId
	}
	plannedObjectIds := make(map[string]uuid.UUID, len(existingPlannedObjectIds))
	for taskId, realId := range existingPlannedObjectIds {
		if realId == uuid.Nil {
			return nil, fmt.Errorf("existing planned object id for task %q is invalid", taskId)
		}
		parsedTaskId, err := uuid.Parse(taskId)
		if err != nil {
			return nil, fmt.Errorf("existing planned object task id %q is invalid: %w", taskId, err)
		}
		if _, exists := taskById[parsedTaskId]; !exists {
			return nil, fmt.Errorf("existing planned object task id %q is not in the routine", taskId)
		}
		plannedObjectIds[taskId] = realId
	}

	plan := &croutinetasktypes.RoutineTaskPlan{
		RoutineId:              routineId,
		Facts:                  facts,
		PrecreatedSubShelves:   make(map[string]croutinetasktypes.PrecreatedSubShelf),
		ContainerObjectTaskIds: make([]uuid.UUID, 0),
		CoreObjectTaskIds:      make([]uuid.UUID, 0),
		PlannedObjectIds:       plannedObjectIds,
	}
	for _, task := range tasks {
		objectKind, isObjectKind := task.Purpose.ObjectKind()
		if !isObjectKind {
			return nil, fmt.Errorf("routine task %q has unsupported purpose %q", task.Id, task.Purpose)
		}
		switch objectKind {
		case cenums.RoutineTaskObjectKind_Container:
			plan.ContainerObjectTaskIds = append(plan.ContainerObjectTaskIds, task.Id)
		case cenums.RoutineTaskObjectKind_Core:
			plan.CoreObjectTaskIds = append(plan.CoreObjectTaskIds, task.Id)
		}
		if task.Purpose != cenums.RoutineTaskPurpose_CreateSubShelf {
			continue
		}

		var payload croutinetasktypes.CreateSubShelfRoutineTaskPayload
		if err := json.Unmarshal(task.Payload, &payload); err != nil {
			return nil, fmt.Errorf("decode create sub shelf payload for task %q: %w", task.Id, err)
		}
		if payload.RootShelfId == uuid.Nil || payload.Name == "" {
			return nil, fmt.Errorf("create sub shelf task %q has an invalid payload", task.Id)
		}
		fakeId := string(payload.FakeId)
		if !payload.FakeId.IsFakeId() {
			return nil, fmt.Errorf("create sub shelf task %q has an invalid fake id", task.Id)
		}
		if _, exists := plan.PrecreatedSubShelves[fakeId]; exists {
			return nil, fmt.Errorf("fake id %q is duplicated", fakeId)
		}
		realId, exists := facts[fakeId]
		if !exists {
			realId = uuid.New()
		}
		facts[fakeId] = realId
		plan.Facts[fakeId] = realId
		path := []uuid.UUID(nil)
		if payload.PrevSubShelfId == nil {
			path = []uuid.UUID{}
		}
		plan.PrecreatedSubShelves[fakeId] = croutinetasktypes.PrecreatedSubShelf{
			TaskId:           task.Id,
			FakeId:           payload.FakeId,
			RealId:           realId,
			RootShelfId:      payload.RootShelfId,
			ParentSubShelfId: payload.PrevSubShelfId,
			Name:             payload.Name,
			Path:             path,
		}
	}

	parentTaskByFakeId := make(map[string]string, len(plan.PrecreatedSubShelves))
	taskIdByFakeId := make(map[string]uuid.UUID, len(plan.PrecreatedSubShelves))
	for fakeId, precreatedSubShelf := range plan.PrecreatedSubShelves {
		taskIdByFakeId[fakeId] = precreatedSubShelf.TaskId
	}
	for fakeId, precreatedSubShelf := range plan.PrecreatedSubShelves {
		if precreatedSubShelf.ParentSubShelfId == nil {
			continue
		}
		if err := precreatedSubShelf.ParentSubShelfId.Validate(); err != nil {
			return nil, fmt.Errorf("sub shelf %q has an invalid parent reference: %w", fakeId, err)
		}
		if !precreatedSubShelf.ParentSubShelfId.IsFakeId() {
			continue
		}
		parentFakeId := string(*precreatedSubShelf.ParentSubShelfId)
		if _, exists := plan.PrecreatedSubShelves[parentFakeId]; !exists {
			return nil, fmt.Errorf("sub shelf %q references unknown fake parent %q", fakeId, parentFakeId)
		}
		parentTaskId := taskIdByFakeId[parentFakeId]
		if _, dependsOnParent := taskAncestorsById[precreatedSubShelf.TaskId][parentTaskId]; !dependsOnParent {
			return nil, fmt.Errorf("sub shelf task %q does not depend on its fake parent task %q", precreatedSubShelf.TaskId, parentTaskId)
		}
		parentTaskByFakeId[fakeId] = parentFakeId
	}

	precreatedOrder := make([]string, 0, len(plan.PrecreatedSubShelves))
	for fakeId := range plan.PrecreatedSubShelves {
		precreatedOrder = append(precreatedOrder, fakeId)
	}
	sort.Strings(precreatedOrder)
	parentCountByFakeId := make(map[string]int, len(precreatedOrder))
	childrenByParentFakeId := make(map[string][]string, len(precreatedOrder))
	for fakeId, parentFakeId := range parentTaskByFakeId {
		parentCountByFakeId[fakeId] = 1
		childrenByParentFakeId[parentFakeId] = append(childrenByParentFakeId[parentFakeId], fakeId)
	}
	for parentFakeId := range childrenByParentFakeId {
		sort.Strings(childrenByParentFakeId[parentFakeId])
	}
	readyFakeIds := make([]string, 0, len(precreatedOrder))
	for _, fakeId := range precreatedOrder {
		if parentCountByFakeId[fakeId] == 0 {
			readyFakeIds = append(readyFakeIds, fakeId)
		}
	}
	precreatedOrder = precreatedOrder[:0]
	for len(readyFakeIds) > 0 {
		fakeId := readyFakeIds[0]
		readyFakeIds = readyFakeIds[1:]
		precreatedOrder = append(precreatedOrder, fakeId)
		parentFakeId, hasParent := parentTaskByFakeId[fakeId]
		if hasParent {
			parent := plan.PrecreatedSubShelves[parentFakeId]
			current := plan.PrecreatedSubShelves[fakeId]
			current.Path = append(append([]uuid.UUID{}, parent.Path...), parent.RealId)
			plan.PrecreatedSubShelves[fakeId] = current
		}
		for _, childFakeId := range childrenByParentFakeId[fakeId] {
			parentCountByFakeId[childFakeId]--
			if parentCountByFakeId[childFakeId] == 0 {
				readyFakeIds = append(readyFakeIds, childFakeId)
			}
		}
	}
	if len(precreatedOrder) != len(plan.PrecreatedSubShelves) {
		return nil, fmt.Errorf("sub shelf parent graph contains a cycle")
	}
	plan.PrecreatedSubShelfOrder = precreatedOrder

	for _, task := range tasks {
		if task.Purpose == cenums.RoutineTaskPurpose_CreateMaterial {
			var payload croutinetasktypes.CreateMaterialRoutineTaskPayload
			if err := json.Unmarshal(task.Payload, &payload); err != nil {
				return nil, fmt.Errorf("decode create material payload for task %q: %w", task.Id, err)
			}
			if err := payload.ParentSubShelfId.Validate(); err != nil {
				return nil, fmt.Errorf("create material task %q has an invalid parent reference: %w", task.Id, err)
			}
			if payload.ParentSubShelfId.IsFakeId() {
				parentFakeId := string(payload.ParentSubShelfId)
				if _, exists := plan.Facts[parentFakeId]; !exists {
					return nil, fmt.Errorf("create material task %q references unknown fake parent %q", task.Id, payload.ParentSubShelfId)
				}
				parentTaskId := taskIdByFakeId[parentFakeId]
				if _, dependsOnParent := taskAncestorsById[task.Id][parentTaskId]; !dependsOnParent {
					return nil, fmt.Errorf("create material task %q does not depend on its fake parent task %q", task.Id, parentTaskId)
				}
			}
			if _, exists := plan.PlannedObjectIds[task.Id.String()]; !exists {
				plan.PlannedObjectIds[task.Id.String()] = uuid.New()
			}
		} else if task.Purpose == cenums.RoutineTaskPurpose_CreateBlockPack {
			var payload croutinetasktypes.CreateBlockPackRoutineTaskPayload
			if err := json.Unmarshal(task.Payload, &payload); err != nil {
				return nil, fmt.Errorf("decode create block pack payload for task %q: %w", task.Id, err)
			}
			if err := payload.TargetSubShelfId.Validate(); err != nil {
				return nil, fmt.Errorf("create block pack task %q has an invalid parent reference: %w", task.Id, err)
			}
			if payload.TargetSubShelfId.IsFakeId() {
				parentFakeId := string(payload.TargetSubShelfId)
				if _, exists := plan.Facts[parentFakeId]; !exists {
					return nil, fmt.Errorf("create block pack task %q references unknown fake parent %q", task.Id, payload.TargetSubShelfId)
				}
				parentTaskId := taskIdByFakeId[parentFakeId]
				if _, dependsOnParent := taskAncestorsById[task.Id][parentTaskId]; !dependsOnParent {
					return nil, fmt.Errorf("create block pack task %q does not depend on its fake parent task %q", task.Id, parentTaskId)
				}
			}
			if _, exists := plan.PlannedObjectIds[task.Id.String()]; !exists {
				plan.PlannedObjectIds[task.Id.String()] = uuid.New()
			}
		}
	}

	return plan, nil
}
