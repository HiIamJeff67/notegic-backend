package shelfitemcascadingtriggersql

import (
	_ "embed"
)

const (
	CascadingSoftDeleteRootShelfTriggerFunctionName = "trigger_function_cascading_soft_delete_root_shelf"
	CascadingSoftDeleteSubShelfTriggerFunctionName  = "trigger_function_cascading_soft_delete_sub_shelf"
	CascadingRestoreRootShelfTriggerFunctionName    = "trigger_function_cascading_restore_soft_deleted_root_shelf"
	CascadingRestoreSubShelfTriggerFunctionName     = "trigger_function_cascading_restore_soft_deleted_sub_shelf"
	CascadingMoveSubShelfTriggerFunctionName        = "trigger_function_cascading_move_sub_shelf"
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
