package routines

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/routine-task-dependencies"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/core/exceptions"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

func TestValidateRoutineTaskDependencyBatchRejectsCycleAcrossNewEdges(t *testing.T) {
	firstTaskId := uuid.New()
	secondTaskId := uuid.New()
	routineTasks := []sschemas.RoutineTask{
		{
			Id: firstTaskId,
		},
		{
			Id: secondTaskId,
		},
	}
	inputs := []coretypes.CreatableRoutineTaskDependency{
		{
			RoutineTaskId:         firstTaskId,
			PreviousRoutineTaskId: secondTaskId,
		},
		{
			RoutineTaskId:         secondTaskId,
			PreviousRoutineTaskId: firstTaskId,
		},
	}

	if exception := validateRoutineTaskDependencyBatch(routineTasks, inputs, nil, apiexceptions.NewRoutineTaskDependencyException()); exception == nil {
		t.Fatal("expected cycle validation to fail")
	}
}

func TestValidateRoutineTaskDependencyBatchRejectsDuplicateDependency(t *testing.T) {
	firstTaskId := uuid.New()
	secondTaskId := uuid.New()
	routineTasks := []sschemas.RoutineTask{
		{
			Id: firstTaskId,
		},
		{
			Id: secondTaskId,
		},
	}
	input := coretypes.CreatableRoutineTaskDependency{
		RoutineTaskId:         firstTaskId,
		PreviousRoutineTaskId: secondTaskId,
	}

	exception := validateRoutineTaskDependencyBatch(
		routineTasks,
		[]coretypes.CreatableRoutineTaskDependency{input, input},
		nil,
		apiexceptions.NewRoutineTaskDependencyException(),
	)
	if exception == nil {
		t.Fatal("expected duplicate validation to fail")
	}
	if exception.Reason != "DependencyAlreadyExists" || exception.HTTPStatusCode() != http.StatusConflict {
		t.Fatalf("duplicate validation exception = %v, want conflict", exception)
	}
}

func TestValidateRoutineTaskDependencyBatchAllowsIndependentEdges(t *testing.T) {
	firstTaskId := uuid.New()
	secondTaskId := uuid.New()
	thirdTaskId := uuid.New()
	routineTasks := []sschemas.RoutineTask{
		{
			Id: firstTaskId,
		},
		{
			Id: secondTaskId,
		},
		{
			Id: thirdTaskId,
		},
	}
	inputs := []coretypes.CreatableRoutineTaskDependency{
		{
			RoutineTaskId:         secondTaskId,
			PreviousRoutineTaskId: firstTaskId,
		},
		{
			RoutineTaskId:         thirdTaskId,
			PreviousRoutineTaskId: firstTaskId,
		},
	}

	if exception := validateRoutineTaskDependencyBatch(routineTasks, inputs, nil, apiexceptions.NewRoutineTaskDependencyException()); exception != nil {
		t.Fatalf("expected independent edges to be valid: %v", exception)
	}
}

func TestValidateRoutineTaskDependencyBatchAllowsEmptyGraph(t *testing.T) {
	if exception := validateRoutineTaskDependencyBatch(nil, nil, nil, apiexceptions.NewRoutineTaskDependencyException()); exception != nil {
		t.Fatalf("empty dependency graph must be valid: %v", exception)
	}
}

func TestValidateRoutineTaskDependencyBatchAllowsMultipleRootsAndBranches(t *testing.T) {
	rootTaskId := uuid.New()
	secondRootTaskId := uuid.New()
	firstBranchTaskId := uuid.New()
	secondBranchTaskId := uuid.New()
	firstLeafTaskId := uuid.New()
	secondLeafTaskId := uuid.New()
	routineTasks := []sschemas.RoutineTask{
		{Id: rootTaskId},
		{Id: secondRootTaskId},
		{Id: firstBranchTaskId},
		{Id: secondBranchTaskId},
		{Id: firstLeafTaskId},
		{Id: secondLeafTaskId},
	}
	inputs := []coretypes.CreatableRoutineTaskDependency{
		{
			RoutineTaskId:         firstBranchTaskId,
			PreviousRoutineTaskId: rootTaskId,
		},
		{
			RoutineTaskId:         secondBranchTaskId,
			PreviousRoutineTaskId: rootTaskId,
		},
		{
			RoutineTaskId:         firstLeafTaskId,
			PreviousRoutineTaskId: firstBranchTaskId,
		},
		{
			RoutineTaskId:         secondLeafTaskId,
			PreviousRoutineTaskId: secondRootTaskId,
		},
	}

	if exception := validateRoutineTaskDependencyBatch(routineTasks, inputs, nil, apiexceptions.NewRoutineTaskDependencyException()); exception != nil {
		t.Fatalf("multiple roots and branches should be valid: %v", exception)
	}
}

func TestValidateRoutineTaskDependencyBatchRejectsSelfEdgeAndCrossRoutineTask(t *testing.T) {
	routineTaskId := uuid.New()
	foreignRoutineTaskId := uuid.New()
	routineTasks := []sschemas.RoutineTask{{Id: routineTaskId}}

	selfEdgeException := validateRoutineTaskDependencyBatch(
		routineTasks,
		[]coretypes.CreatableRoutineTaskDependency{
			{
				RoutineTaskId:         routineTaskId,
				PreviousRoutineTaskId: routineTaskId,
			},
		},
		nil,
		apiexceptions.NewRoutineTaskDependencyException(),
	)
	if selfEdgeException == nil || selfEdgeException.Reason != "InvalidInput" {
		t.Fatalf("self-edge exception = %v, want InvalidInput", selfEdgeException)
	}

	crossRoutineException := validateRoutineTaskDependencyBatch(
		routineTasks,
		[]coretypes.CreatableRoutineTaskDependency{
			{
				RoutineTaskId:         routineTaskId,
				PreviousRoutineTaskId: foreignRoutineTaskId,
			},
		},
		nil,
		apiexceptions.NewRoutineTaskDependencyException(),
	)
	if crossRoutineException == nil || crossRoutineException.Reason != "InvalidInput" {
		t.Fatalf("cross-routine exception = %v, want InvalidInput", crossRoutineException)
	}
}
