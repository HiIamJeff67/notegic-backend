package postgres

type TableName string

func (tableName TableName) String() string {
	return string(tableName)
}

const (
	TableName_UserTable                 TableName = "UserTable"
	TableName_UserAccountTable          TableName = "UserAccountTable"
	TableName_UserQuotaTable            TableName = "UserQuotaTable"
	TableName_UserInfoTable             TableName = "UserInfoTable"
	TableName_UserSettingTable          TableName = "UserSettingTable"
	TableName_UserView                  TableName = "UserView"
	TableName_UserProjection            TableName = "UserProjection"
	TableName_InboxEventTable           TableName = "InboxEventTable"
	TableName_OutboxEventTable          TableName = "OutboxEventTable"
	TableName_NotificationTable         TableName = "NotificationTable"
	TableName_APIKeyTable               TableName = "APIKeyTable"
	TableName_BadgeTable                TableName = "BadgeTable"
	TableName_UsersToBadgesTable        TableName = "UsersToBadgesTable"
	TableName_ThemeTable                TableName = "ThemeTable"
	TableName_UsersToShelvesTable       TableName = "UsersToShelvesTable"
	TableName_RootShelfTable            TableName = "RootShelfTable"
	TableName_SubShelfTable             TableName = "SubShelfTable"
	TableName_MaterialTable             TableName = "MaterialTable"
	TableName_BlockPackTable            TableName = "BlockPackTable"
	TableName_BlockPackYjsDocumentTable TableName = "BlockPackYjsDocumentTable"
	TableName_BlockPackYjsUpdateTable   TableName = "BlockPackYjsUpdateTable"
	TableName_BlockTable                TableName = "BlockTable"
	TableName_ItemTable                 TableName = "ItemTable"
	TableName_RoutinesToItemsTable      TableName = "RoutinesToItemsTable"
	TableName_UsersToStationsTable      TableName = "UsersToStationsTable"
	TableName_StationTable              TableName = "StationTable"
	TableName_RoutineTable              TableName = "RoutineTable"
	TableName_RoutineDependencyTable    TableName = "RoutineDependencyTable"
	TableName_RoutineRecordTable        TableName = "RoutineRecordTable"
	TableName_RoutineTaskTable          TableName = "RoutineTaskTable"
	TableName_RoutineTaskRecordTable    TableName = "RoutineTaskRecordTable"
	TableName_RoutineTagTable           TableName = "RoutineTagTable"
	TableName_RoutinesToTagsTable       TableName = "RoutinesToTagsTable"
	TableName_UsersToBillingPlansTable  TableName = "UsersToBillingPlansTable"
	TableName_PlanLimitationTable       TableName = "PlanLimitationTable"
	TableName_BillingPlanTable          TableName = "BillingPlanTable"
)
