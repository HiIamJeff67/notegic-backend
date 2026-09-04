package constraints

import (
	blockconstraints "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/constraints/block_constraints"
	routineconstraints "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/constraints/routine_constraints"
	shelfconstraints "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/constraints/shelf_constraints"
	subshelfconstraints "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/constraints/sub_shelf_constraints"
	userconstraints "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/constraints/user_constraints"
	userstobillingplansconstraints "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/constraints/users_to_billing_plans_constraints"
)

var MigratingConstraintSQLs = []string{
	userstobillingplansconstraints.UserIdBillingPlanIdPartialStatusIndexSQL,
	userconstraints.UserForeignKeysSQL,
	shelfconstraints.ShelfForeignKeysSQL,
	blockconstraints.BlockSiblingPointerConstraintsSQL,
	blockconstraints.BlockForeignKeysSQL,
	routineconstraints.RoutineScheduledTimeMinutePrecisionCheckSQL,
	routineconstraints.RoutineScheduledTimeInPeriodCheckSQL,
	routineconstraints.RoutineForeignKeysSQL,
	subshelfconstraints.SubShelfPreviousReferenceConstraintSQL,
}
