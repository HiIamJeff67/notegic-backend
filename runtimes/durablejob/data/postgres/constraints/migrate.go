package constraints

import (
	routinetaskdependencyconstraints "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/constraints/routine_task_dependency_constraints"
	routinetaskrecordconstraints "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/constraints/routine_task_record_constraints"
	userquotaconstraints "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/data/postgres/constraints/user_quota_constraints"
)

var MigratingConstraintSQLs = []string{
	userquotaconstraints.UserQuotaConstraintSQL,
	routinetaskdependencyconstraints.RoutineTaskDependencyForeignKeysSQL,
	routinetaskrecordconstraints.RoutineTaskRecordForeignKeysSQL,
}
