package constraints

import (
	blockconstraints "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/constraints/block_constraints"
	routineconstraints "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/constraints/routine_constraints"
	userstobillingplansconstraints "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/constraints/users_to_billing_plans_constraints"
)

var MigratingConstraintSQLs = []string{
	userstobillingplansconstraints.UserIdBillingPlanIdPartialStatusIndexSQL,
	blockconstraints.BlockSiblingPointerConstraintsSQL,
	routineconstraints.RoutineScheduledTimeMinutePrecisionCheckSQL,
	routineconstraints.RoutineScheduledTimeInPeriodCheckSQL,
}
