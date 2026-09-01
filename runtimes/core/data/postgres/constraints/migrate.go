package constraints

import (
	blockconstraints "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/constraints/block_constraints"
	routineconstraints "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/constraints/routine_constraints"
	subshelfconstraints "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/constraints/sub_shelf_constraints"
	userstobillingplansconstraints "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/constraints/users_to_billing_plans_constraints"
)

var MigratingConstraintSQLs = []string{
	userstobillingplansconstraints.UserIdBillingPlanIdPartialStatusIndexSQL,
	blockconstraints.BlockSiblingPointerConstraintsSQL,
	routineconstraints.RoutineScheduledTimeMinutePrecisionCheckSQL,
	routineconstraints.RoutineScheduledTimeInPeriodCheckSQL,
	subshelfconstraints.SubShelfPreviousReferenceConstraintSQL,
}
