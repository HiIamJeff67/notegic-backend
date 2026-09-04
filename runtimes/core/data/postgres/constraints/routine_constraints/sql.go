package routineconstraints

import (
	_ "embed"
)

var (
	//go:embed routine_scheduled_time_minute_precision_check.sql
	RoutineScheduledTimeMinutePrecisionCheckSQL string

	//go:embed routine_scheduled_time_in_period_check.sql
	RoutineScheduledTimeInPeriodCheckSQL string

	//go:embed routine_foreign_keys.sql
	RoutineForeignKeysSQL string
)
