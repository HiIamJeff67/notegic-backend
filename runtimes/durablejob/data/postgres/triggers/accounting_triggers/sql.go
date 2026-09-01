package accountingtriggersql

import (
	_ "embed"
)

const (
	AccountingInsertedRoutineTaskTriggerFunctionName = "trigger_function_accounting_inserted_routine_task"
	AccountingDeletedRoutineTaskTriggerFunctionName  = "trigger_function_accounting_deleted_routine_task"
	AccountingUpdatedRoutineTaskTriggerFunctionName  = "trigger_function_accounting_updated_routine_task"
)

var (
	//go:embed accounting_inserted_routine_task_trigger.sql
	AccountingInsertedRoutineTaskTriggerSQL string

	//go:embed accounting_deleted_routine_task_trigger.sql
	AccountingDeletedRoutineTaskTriggerSQL string

	//go:embed accounting_updated_routine_task_trigger.sql
	AccountingUpdatedRoutineTaskTriggerSQL string
)
