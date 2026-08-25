package routinetask

import (
	"testing"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

func TestNewHandlerManagerRegistersEveryPurposePolicy(t *testing.T) {
	manager := NewHandlerManager(1)

	for _, purpose := range []enums.RoutineTaskPurpose{
		enums.RoutineTaskPurpose_CreateRootShelf,
		enums.RoutineTaskPurpose_UpdateRootShelf,
		enums.RoutineTaskPurpose_ResetRootShelf,
		enums.RoutineTaskPurpose_CreateSubShelf,
		enums.RoutineTaskPurpose_UpdateSubShelf,
		enums.RoutineTaskPurpose_ResetSubShelf,
		enums.RoutineTaskPurpose_CreateBlockPack,
		enums.RoutineTaskPurpose_UpdateBlockPack,
		enums.RoutineTaskPurpose_ResetBlockPack,
		enums.RoutineTaskPurpose_AppendBlock,
		enums.RoutineTaskPurpose_UpdateBlock,
		enums.RoutineTaskPurpose_ResetBlock,
		enums.RoutineTaskPurpose_CreateRoutine,
		enums.RoutineTaskPurpose_UpdateRoutine,
	} {
		registry, exists := manager.registries[purpose]
		if !exists || registry.HandlerFunc == nil {
			t.Fatalf("missing routine task handler policy for %s", purpose)
		}
	}
}
