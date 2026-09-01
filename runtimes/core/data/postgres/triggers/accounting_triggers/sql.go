package accountingtriggersql

import (
	_ "embed"
)

const (
	AccountingMutatedBlockPackTriggerFunctionName   = "trigger_function_accounting_mutated_block_pack"
	AccountingInsertedBlockTriggerFunctionName      = "trigger_function_accounting_inserted_block"
	AccountingDeletedBlockTriggerFunctionName       = "trigger_function_accounting_deleted_block"
	AccountingMutatedRootShelfTriggerFunctionName   = "trigger_function_accounting_mutated_root_shelf"
	AccountingMutatedSubShelfTriggerFunctionName    = "trigger_function_accounting_mutated_sub_shelf"
	AccountingMutatedMaterialTriggerFunctionName    = "trigger_function_accounting_mutated_material"
	AccountingInsertedRoutineTagTriggerFunctionName = "trigger_function_accounting_inserted_routine_tag"
	AccountingDeletedRoutineTagTriggerFunctionName  = "trigger_function_accounting_deleted_routine_tag"
	AccountingInsertedRoutineTriggerFunctionName    = "trigger_function_accounting_inserted_routine"
	AccountingDeletedRoutineTriggerFunctionName     = "trigger_function_accounting_deleted_routine"
	AccountingInsertedStationTriggerFunctionName    = "trigger_function_accounting_inserted_station"
	AccountingDeletedStationTriggerFunctionName     = "trigger_function_accounting_deleted_station"
	AccountingMutatedStationTriggerFunctionName     = "trigger_function_accounting_mutated_station"
)

var (
	//go:embed accounting_mutated_block_pack_trigger.sql
	AccountingMutatedBlockPackTriggerSQL string

	//go:embed accounting_inserted_block_trigger.sql
	AccountingInsertedBlockTriggerSQL string

	//go:embed accounting_deleted_block_trigger.sql
	AccountingDeletedBlockTriggerSQL string

	//go:embed accounting_mutated_root_shelf_trigger.sql
	AccountingMutatedRootShelfTriggerSQL string

	//go:embed accounting_mutated_sub_shelf_trigger.sql
	AccountingMutatedSubShelfTriggerSQL string

	//go:embed accounting_mutated_material_trigger.sql
	AccountingMutatedMaterialTriggerSQL string

	//go:embed accounting_inserted_routine_tag_trigger.sql
	AccountingInsertedRoutineTagTriggerSQL string

	//go:embed accounting_deleted_routine_tag_trigger.sql
	AccountingDeletedRoutineTagTriggerSQL string

	//go:embed accounting_inserted_routine_trigger.sql
	AccountingInsertedRoutineTriggerSQL string

	//go:embed accounting_deleted_routine_trigger.sql
	AccountingDeletedRoutineTriggerSQL string

	//go:embed accounting_inserted_station_trigger.sql
	AccountingInsertedStationTriggerSQL string

	//go:embed accounting_deleted_station_trigger.sql
	AccountingDeletedStationTriggerSQL string

	//go:embed accounting_mutated_station_trigger.sql
	AccountingMutatedStationTriggerSQL string
)
