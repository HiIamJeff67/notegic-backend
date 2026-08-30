package postgres

import (
	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	billingplanseeds "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/seeds/billing_plan_seeds"
	planlimitationseeds "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/seeds/plan_limitation_seeds"
)

// DatabaseSeedManifest describes the default data owned and seeded by Core.
var DatabaseSeedManifest = spostgres.SeedManifest{
	Runtime: ctypes.Runtime_Core,
	SQLs: []string{
		planlimitationseeds.PlanLimitationSeedingDefaultDataSQL_0000_UP,
		billingplanseeds.BillingPlanSeedingDefaultDataSQL_0000_UP,
	},
}
