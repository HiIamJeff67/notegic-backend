package postgres

import (
	types "github.com/HiIamJeff67/notegic-backend/contracts/types"
	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
	platformschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	constraints "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/constraints"
	triggers "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/triggers"
	views "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas/views"
)

// DatabaseMigrationManifest describes the schemas owned and migrated by Core.
// Database permissions are intentionally configured separately.
var DatabaseMigrationManifest = platformpostgres.MigrationManifest{
	Runtime: types.Runtime_Core,
	Enums: map[string][]string{
		new(enums.AccessControlPermission).Name():    enums.AllAccessControlPermissionStrings,
		new(enums.BadgeType).Name():                  enums.AllBadgeTypeStrings,
		new(enums.BillingIntervalUnit).Name():        enums.AllBillingIntervalUnitStrings,
		new(enums.BillingPlanName).Name():            enums.AllBillingPlanNameStrings,
		new(enums.BillingPlanStatus).Name():          enums.AllBillingPlanStatusStrings,
		new(enums.BlockType).Name():                  enums.AllBlockTypeStrings,
		new(enums.CountryCode).Name():                enums.AllCountryCodeStrings,
		new(enums.Country).Name():                    enums.AllCountryStrings,
		new(enums.ItemType).Name():                   enums.AllItemTypeStrings,
		new(enums.Language).Name():                   enums.AllLanguageStrings,
		new(enums.UserSettingDensity).Name():         enums.AllUserSettingDensityStrings,
		new(enums.UserSettingStartSurface).Name():    enums.AllUserSettingStartSurfaceStrings,
		new(enums.MaterialContentType).Name():        enums.AllMaterialContentTypeStrings,
		new(enums.RoutinePeriod).Name():              enums.AllRoutinePeriodStrings,
		new(enums.RoutineStatus).Name():              enums.AllRoutineStatusStrings,
		new(enums.RoutineTaskPurpose).Name():         enums.AllRoutineTaskPurposeStrings,
		new(enums.RoutineTaskStatus).Name():          enums.AllRoutineTaskStatusStrings,
		new(enums.RoutineTaskRecordErrorCode).Name(): enums.AllRoutineTaskRecordErrorCodeStrings,
		new(enums.RoutineTaskRecordStatus).Name():    enums.AllRoutineTaskRecordStatusStrings,
		new(enums.SupportedIcon).Name():              enums.AllSupportedIconStrings,
		new(enums.SupportedCurrencyCode).Name():      enums.AllSupportedCurrencyCodeStrings,
		new(enums.UserGender).Name():                 enums.AllUserGenderStrings,
		new(enums.UserPlan).Name():                   enums.AllUserPlanStrings,
		new(enums.UserRole).Name():                   enums.AllUserRoleStrings,
		new(enums.UserStatus).Name():                 enums.AllUserStatusStrings,
		new(enums.UsersToBillingPlansStatus).Name():  enums.AllUsersToBillingPlansStatusStrings,
	},
	Views:       views.MigratingViewSQLs,
	Triggers:    triggers.MigratingTriggerSQLs,
	Constraints: constraints.MigratingConstraintSQLs,
	Tables: []any{
		&platformschemas.User{},
		&platformschemas.UserInfo{},
		&platformschemas.UserAccount{},
		&platformschemas.UserSetting{},
		&platformschemas.APIKey{},
		&platformschemas.UsersToBadges{},
		&platformschemas.Badge{},
		&platformschemas.Theme{},
		&platformschemas.UsersToShelves{},
		&platformschemas.RootShelf{},
		&platformschemas.SubShelf{},
		&platformschemas.Material{},
		&platformschemas.BlockPack{},
		&platformschemas.BlockPackYjsDocument{},
		&platformschemas.BlockPackYjsUpdate{},
		&platformschemas.Block{},
		&platformschemas.Item{},
		&platformschemas.Station{},
		&platformschemas.Routine{},
		&platformschemas.RoutineTag{},
		&platformschemas.UsersToStations{},
		&platformschemas.RoutinesToItems{},
		&platformschemas.RoutinesToTags{},
		&platformschemas.InboxEvent{},
		&platformschemas.OutboxEvent{},
		&platformschemas.UsersToBillingPlans{},
		&platformschemas.PlanLimitation{},
		&platformschemas.BillingPlan{},
	},
}
