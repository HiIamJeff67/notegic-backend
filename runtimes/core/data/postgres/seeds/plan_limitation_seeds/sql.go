package planlimitationseeds

import (
	_ "embed"
)

//go:embed 0000_plan_limitation_seed.up.sql
var PlanLimitationSeedingDefaultDataSQL_0000_UP string

//go:embed 0000_plan_limitation_seed.down.sql
var PlanLimitationSeedingDefaultDataSQL_0000_DOWN string
