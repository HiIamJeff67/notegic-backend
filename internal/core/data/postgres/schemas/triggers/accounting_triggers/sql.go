package accountingtriggersql

import (
	_ "embed"
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

	//go:embed accounting_inserted_routine_task_trigger.sql
	AccountingInsertedRoutineTaskTriggerSQL string

	//go:embed accounting_deleted_routine_task_trigger.sql
	AccountingDeletedRoutineTaskTriggerSQL string

	//go:embed accounting_updated_routine_task_trigger.sql
	AccountingUpdatedRoutineTaskTriggerSQL string

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
