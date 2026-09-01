package triggers

import accountingtrigger "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/triggers/accounting_triggers"

var RoutineTaskTriggerSQLs = []string{
	accountingtrigger.AccountingInsertedRoutineTaskTriggerSQL,
	accountingtrigger.AccountingDeletedRoutineTaskTriggerSQL,
	accountingtrigger.AccountingUpdatedRoutineTaskTriggerSQL,
}
