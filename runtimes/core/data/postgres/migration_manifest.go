package postgres

import (
	ctypes "github.com/HiIamJeff67/notegic-backend/contracts/types"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	spostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"

	constraints "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/constraints"
	triggers "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/triggers"
	views "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/views"
)

// DatabaseMigrationManifest describes the schemas owned and migrated by Core.
// Database permissions are intentionally configured separately.
var DatabaseMigrationManifest = spostgres.MigrationManifest{
	Runtime: ctypes.Runtime_Core,
	Enums: map[string][]string{
		new(cenums.AccessControlPermission).Name():    cenums.AllAccessControlPermissionStrings,
		new(cenums.BadgeType).Name():                  cenums.AllBadgeTypeStrings,
		new(cenums.BillingIntervalUnit).Name():        cenums.AllBillingIntervalUnitStrings,
		new(cenums.BillingPlanName).Name():            cenums.AllBillingPlanNameStrings,
		new(cenums.BillingPlanStatus).Name():          cenums.AllBillingPlanStatusStrings,
		new(cenums.BlockType).Name():                  cenums.AllBlockTypeStrings,
		new(cenums.CountryCode).Name():                cenums.AllCountryCodeStrings,
		new(cenums.Country).Name():                    cenums.AllCountryStrings,
		new(cenums.ItemType).Name():                   cenums.AllItemTypeStrings,
		new(cenums.Language).Name():                   cenums.AllLanguageStrings,
		new(cenums.UserSettingDensity).Name():         cenums.AllUserSettingDensityStrings,
		new(cenums.UserSettingStartSurface).Name():    cenums.AllUserSettingStartSurfaceStrings,
		new(cenums.MaterialContentType).Name():        cenums.AllMaterialContentTypeStrings,
		new(cenums.RoutinePeriod).Name():              cenums.AllRoutinePeriodStrings,
		new(cenums.RoutinePhase).Name():               cenums.AllRoutinePhaseStrings,
		new(cenums.RoutineRecordStatus).Name():        cenums.AllRoutineRecordStatusStrings,
		new(cenums.RoutineStatus).Name():              cenums.AllRoutineStatusStrings,
		new(cenums.RoutineTaskPurpose).Name():         cenums.AllRoutineTaskPurposeStrings,
		new(cenums.RoutineTaskRecordErrorCode).Name(): cenums.AllRoutineTaskRecordErrorCodeStrings,
		new(cenums.RoutineTaskRecordStatus).Name():    cenums.AllRoutineTaskRecordStatusStrings,
		new(cenums.SupportedIcon).Name():              cenums.AllSupportedIconStrings,
		new(cenums.SupportedCurrencyCode).Name():      cenums.AllSupportedCurrencyCodeStrings,
		new(cenums.UserGender).Name():                 cenums.AllUserGenderStrings,
		new(cenums.UserPlan).Name():                   cenums.AllUserPlanStrings,
		new(cenums.UserRole).Name():                   cenums.AllUserRoleStrings,
		new(cenums.UserStatus).Name():                 cenums.AllUserStatusStrings,
		new(cenums.UsersToBillingPlansStatus).Name():  cenums.AllUsersToBillingPlansStatusStrings,
	},
	Views:       views.MigratingViewSQLs,
	Triggers:    triggers.MigratingTriggerSQLs,
	Constraints: constraints.MigratingConstraintSQLs,
	Tables: []any{
		&sschemas.User{},
		&sschemas.UserInfo{},
		&sschemas.UserAccount{},
		&sschemas.UserSetting{},
		&sschemas.APIKey{},
		&sschemas.UsersToBadges{},
		&sschemas.Badge{},
		&sschemas.Theme{},
		&sschemas.UsersToShelves{},
		&sschemas.RootShelf{},
		&sschemas.SubShelf{},
		&sschemas.Material{},
		&sschemas.BlockPack{},
		&sschemas.BlockPackYjsDocument{},
		&sschemas.BlockPackYjsUpdate{},
		&sschemas.Block{},
		&sschemas.Item{},
		&sschemas.Station{},
		&sschemas.Routine{},
		&sschemas.RoutineTag{},
		&sschemas.UsersToStations{},
		&sschemas.RoutinesToItems{},
		&sschemas.RoutinesToTags{},
		&sschemas.InboxEvent{},
		&sschemas.OutboxEvent{},
		&sschemas.UsersToBillingPlans{},
		&sschemas.PlanLimitation{},
		&sschemas.BillingPlan{},
	},
}
