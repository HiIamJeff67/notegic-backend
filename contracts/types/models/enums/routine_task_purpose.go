package enums

type RoutineTaskPurpose string

const (
	RoutineTaskPurpose_CreateRootShelf RoutineTaskPurpose = "CreateRootShelf" // create a root shelf with nothing inside of it
	RoutineTaskPurpose_UpdateRootShelf RoutineTaskPurpose = "UpdateRootShelf" // update the columns of the given root shelf
	RoutineTaskPurpose_ResetRootShelf  RoutineTaskPurpose = "ResetRootShelf"  // reset the children of the root shelf
	RoutineTaskPurpose_CreateSubShelf  RoutineTaskPurpose = "CreateSubShelf"  // create a sub shelf with nothing inside of it
	RoutineTaskPurpose_UpdateSubShelf  RoutineTaskPurpose = "UpdateSubShelf"  // update the columns of the given sub shelf
	RoutineTaskPurpose_ResetSubShelf   RoutineTaskPurpose = "ResetSubShelf"   // reset the children of the given sub shelf
	RoutineTaskPurpose_CreateBlockPack RoutineTaskPurpose = "CreateBlockPack" // create a block pack with the given content within the routine task payload
	RoutineTaskPurpose_UpdateBlockPack RoutineTaskPurpose = "UpdateBlockPack" // update blocks in the block pack
	RoutineTaskPurpose_ResetBlockPack  RoutineTaskPurpose = "ResetBlockPack"  // reset the block pack to an empty block pack
	RoutineTaskPurpose_AppendBlock     RoutineTaskPurpose = "AppendBlock"     // create a block at the end of the given block pack with the given props and content within the routine task payload
	RoutineTaskPurpose_UpdateBlock     RoutineTaskPurpose = "UpdateBlock"     // update a block with the given props and content within the routine task payload
	RoutineTaskPurpose_ResetBlock      RoutineTaskPurpose = "ResetBlock"      // reset the block to a paragraph with empty props and content
	RoutineTaskPurpose_CreateRoutine   RoutineTaskPurpose = "CreateRoutine"   // create a routine with no links
	RoutineTaskPurpose_UpdateRoutine   RoutineTaskPurpose = "UpdateRoutine"   // update the columns of the given routine, excluded links to it
)
