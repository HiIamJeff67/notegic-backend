package routinetask

import (
	"testing"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

func TestNewHandlerManagerRegistersEveryPurposePolicy(t *testing.T) {
	manager := NewHandlerManager(1)

	for _, purpose := range []cenums.RoutineTaskPurpose{
		cenums.RoutineTaskPurpose_CreateRootShelf,
		cenums.RoutineTaskPurpose_UpdateRootShelf,
		cenums.RoutineTaskPurpose_ResetRootShelf,
		cenums.RoutineTaskPurpose_CreateSubShelf,
		cenums.RoutineTaskPurpose_UpdateSubShelf,
		cenums.RoutineTaskPurpose_ResetSubShelf,
		cenums.RoutineTaskPurpose_CreateBlockPack,
		cenums.RoutineTaskPurpose_UpdateBlockPack,
		cenums.RoutineTaskPurpose_ResetBlockPack,
		cenums.RoutineTaskPurpose_AppendBlock,
		cenums.RoutineTaskPurpose_UpdateBlock,
		cenums.RoutineTaskPurpose_ResetBlock,
		cenums.RoutineTaskPurpose_CreateRoutine,
		cenums.RoutineTaskPurpose_UpdateRoutine,
	} {
		registry, exists := manager.registries[purpose]
		if !exists || registry.HandlerFunc == nil {
			t.Fatalf("missing routine task handler policy for %s", purpose)
		}
	}
}
