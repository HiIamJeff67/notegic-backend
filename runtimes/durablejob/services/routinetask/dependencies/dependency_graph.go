package dependencies

import (
	"fmt"

	"github.com/google/uuid"
)

type Edge struct {
	TaskId         uuid.UUID
	PreviousTaskId uuid.UUID
}

func Validate(taskIds []uuid.UUID, edges []Edge) error {
	taskSet := make(map[uuid.UUID]struct{}, len(taskIds))
	for _, taskId := range taskIds {
		if taskId == uuid.Nil {
			return fmt.Errorf("routine task dependency contains an empty task id")
		}
		if _, exists := taskSet[taskId]; exists {
			return fmt.Errorf("routine task %q is duplicated", taskId)
		}
		taskSet[taskId] = struct{}{}
	}

	previousTaskIdsByTaskId := make(map[uuid.UUID][]uuid.UUID, len(taskIds))
	seenEdges := make(map[Edge]struct{}, len(edges))
	for _, edge := range edges {
		if edge.TaskId == uuid.Nil || edge.PreviousTaskId == uuid.Nil {
			return fmt.Errorf("routine task dependency contains an empty task id")
		}
		if _, exists := taskSet[edge.TaskId]; !exists {
			return fmt.Errorf("routine task dependency child %q is not in the routine", edge.TaskId)
		}
		if _, exists := taskSet[edge.PreviousTaskId]; !exists {
			return fmt.Errorf("routine task dependency parent %q is not in the routine", edge.PreviousTaskId)
		}
		if edge.TaskId == edge.PreviousTaskId {
			return fmt.Errorf("routine task %q depends on itself", edge.TaskId)
		}
		if _, exists := seenEdges[edge]; exists {
			return fmt.Errorf("routine task dependency %q -> %q is duplicated", edge.TaskId, edge.PreviousTaskId)
		}
		seenEdges[edge] = struct{}{}
		previousTaskIdsByTaskId[edge.TaskId] = append(previousTaskIdsByTaskId[edge.TaskId], edge.PreviousTaskId)
	}

	visitStateByTaskId := make(map[uuid.UUID]uint8, len(taskIds))
	var visit func(uuid.UUID) error
	visit = func(taskId uuid.UUID) error {
		switch visitStateByTaskId[taskId] {
		case 1:
			return fmt.Errorf("routine task dependency graph contains a cycle")
		case 2:
			return nil
		}
		visitStateByTaskId[taskId] = 1
		for _, previousTaskId := range previousTaskIdsByTaskId[taskId] {
			if err := visit(previousTaskId); err != nil {
				return err
			}
		}
		visitStateByTaskId[taskId] = 2
		return nil
	}

	for _, taskId := range taskIds {
		if err := visit(taskId); err != nil {
			return err
		}
	}
	return nil
}
