package postgres

import platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

const (
	TableName_UserTable        platformpostgres.TableName = "UserTable"
	TableName_UserAccountTable platformpostgres.TableName = "UserAccountTable"
	TableName_UserQuotaTable   platformpostgres.TableName = "UserQuotaTable"
	TableName_UserInfoTable    platformpostgres.TableName = "UserInfoTable"
	TableName_UserSettingTable platformpostgres.TableName = "UserSettingTable"

	TableName_BadgeTable         platformpostgres.TableName = "BadgeTable"
	TableName_UsersToBadgesTable platformpostgres.TableName = "UsersToBadgesTable"

	TableName_ThemeTable platformpostgres.TableName = "ThemeTable"

	TableName_UsersToShelvesTable       platformpostgres.TableName = "UsersToShelvesTable"
	TableName_RootShelfTable            platformpostgres.TableName = "RootShelfTable"
	TableName_SubShelfTable             platformpostgres.TableName = "SubShelfTable"
	TableName_MaterialTable             platformpostgres.TableName = "MaterialTable"
	TableName_BlockPackTable            platformpostgres.TableName = "BlockPackTable"
	TableName_BlockPackYjsDocumentTable platformpostgres.TableName = "BlockPackYjsDocumentTable"
	TableName_BlockPackYjsUpdateTable   platformpostgres.TableName = "BlockPackYjsUpdateTable"
	TableName_BlockTable                platformpostgres.TableName = "BlockTable"
	TableName_ItemTable                 platformpostgres.TableName = "ItemTable"

	TableName_RoutinesToItemsTable   platformpostgres.TableName = "RoutinesToItemsTable"
	TableName_UsersToStationsTable   platformpostgres.TableName = "UsersToStationsTable"
	TableName_StationTable           platformpostgres.TableName = "StationTable"
	TableName_RoutineTable           platformpostgres.TableName = "RoutineTable"
	TableName_RoutineDependencyTable platformpostgres.TableName = "RoutineDependencyTable"
	TableName_RoutineTaskTable       platformpostgres.TableName = "RoutineTaskTable"
	TableName_RoutineTaskRecordTable platformpostgres.TableName = "RoutineTaskRecordTable"
	TableName_RoutineTagTable        platformpostgres.TableName = "RoutineTagTable"
	TableName_RoutinesToTagsTable    platformpostgres.TableName = "RoutinesToTagsTable"

	TableName_UsersToBillingPlansTable platformpostgres.TableName = "UsersToBillingPlansTable"

	TableName_PlanLimitationTable platformpostgres.TableName = "PlanLimitationTable"
	TableName_BillingPlanTable    platformpostgres.TableName = "BillingPlanTable"
)

var _validTableNames = map[string]platformpostgres.TableName{
	"UserTable":        TableName_UserTable,
	"UserAccountTable": TableName_UserAccountTable,
	"UserQuotaTable":   TableName_UserQuotaTable,
	"UserInfoTable":    TableName_UserInfoTable,
	"UserSettingTable": TableName_UserSettingTable,

	"BadgeTable":         TableName_BadgeTable,
	"UsersToBadgesTable": TableName_UsersToBadgesTable,

	"ThemeTable": TableName_ThemeTable,

	"UsersToShelvesTable":       TableName_UsersToShelvesTable,
	"RootShelfTable":            TableName_RootShelfTable,
	"SubShelfTable":             TableName_SubShelfTable,
	"MaterialTable":             TableName_MaterialTable,
	"BlockPackTable":            TableName_BlockPackTable,
	"BlockPackYjsDocumentTable": TableName_BlockPackYjsDocumentTable,
	"BlockPackYjsUpdateTable":   TableName_BlockPackYjsUpdateTable,
	"BlockTable":                TableName_BlockTable,
	"ItemTable":                 TableName_ItemTable,

	"RoutinesToItemsTable":   TableName_RoutinesToItemsTable,
	"UsersToStationsTable":   TableName_UsersToStationsTable,
	"StationTable":           TableName_StationTable,
	"RoutineTable":           TableName_RoutineTable,
	"RoutineDependencyTable": TableName_RoutineDependencyTable,
	"RoutineTaskTable":       TableName_RoutineTaskTable,
	"RoutineTaskRecordTable": TableName_RoutineTaskRecordTable,
	"RoutineTagTable":        TableName_RoutineTagTable,
	"RoutinesToTagsTable":    TableName_RoutinesToTagsTable,

	"UsersToBillingPlansTable": TableName_UsersToBillingPlansTable,

	"PlanLimitationTable": TableName_PlanLimitationTable,
	"BillingPlanTable":    TableName_BillingPlanTable,
}

func ConvertToTableName(tableName string) (platformpostgres.TableName, bool) {
	validTableName, ok := _validTableNames[tableName]
	return validTableName, ok
}
