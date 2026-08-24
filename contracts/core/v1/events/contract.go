package eventscontract

import eventcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

const CoreLifecycleTopic eventcontract.Topic = "notegic.core.lifecycle.v1"

const CoreDurableJobYjsMaintenanceHintTopic eventcontract.Topic = "notegic.core.durablejob.yjs-maintenance-hint.v1"

const (
	AggregateType_RootShelf   eventcontract.AggregateType = "RootShelf"
	AggregateType_SubShelf    eventcontract.AggregateType = "SubShelf"
	AggregateType_BlockPack   eventcontract.AggregateType = "BlockPack"
	AggregateType_RoutineTask eventcontract.AggregateType = "RoutineTask"
	AggregateType_User        eventcontract.AggregateType = "User"
)

const (
	EventType_BlockPackAccessRevoked     eventcontract.EventType = "BlockPackAccessRevoked"
	EventType_BlockPackRoomPolicyChanged eventcontract.EventType = "BlockPackRoomPolicyChanged"
	EventType_RootShelfPermissionRevoked eventcontract.EventType = "RootShelfPermissionRevoked"
	EventType_RootShelfPermissionChanged eventcontract.EventType = "RootShelfPermissionChanged"
	EventType_RootShelfDeleted           eventcontract.EventType = "RootShelfDeleted"
	EventType_BlockPackChanged           eventcontract.EventType = "BlockPackChanged"
	EventType_BlockPackDeleted           eventcontract.EventType = "BlockPackDeleted"
	EventType_UserSessionsRevoked        eventcontract.EventType = "UserSessionsRevoked"
	EventType_YjsMaintenanceHint         eventcontract.EventType = "YjsMaintenanceHint"
	EventType_RoutineTaskCompleted       eventcontract.EventType = "RoutineTaskCompleted"
)
