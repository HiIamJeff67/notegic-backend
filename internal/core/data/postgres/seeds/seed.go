package seeds

import (
	_ "embed"

	billingplanseeds "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/seeds/billing_plan_seeds"
	planlimitationseeds "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/seeds/plan_limitation_seeds"
)

var SeedingDefaultDataSQLs = []string{
	planlimitationseeds.PlanLimitationSeedingDefaultDataSQL_0000_UP,
	billingplanseeds.BillingPlanSeedingDefaultDataSQL_0000_UP,
}
