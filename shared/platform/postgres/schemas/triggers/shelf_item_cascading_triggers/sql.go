package shelfitemcascadingtriggersql

import (
	_ "embed"
)

var (
	//go:embed cascading_soft_delete_root_shelf_trigger.sql
	CascadingSoftDeleteRootShelfTriggerSQL string

	//go:embed cascading_soft_delete_sub_shelf_trigger.sql
	CascadingSoftDeleteSubShelfTriggerSQL string

	//go:embed cascading_restore_soft_deleted_root_shelf_trigger.sql
	CascadingRestoreRootShelfTriggerSQL string

	//go:embed cascading_restore_soft_deleted_sub_shelf_trigger.sql
	CascadingRestoreSubShelfTriggerSQL string

	//go:embed cascading_move_sub_shelf_trigger.sql
	CascadingMoveSubShelfTriggerSQL string
)
