package eventscontract

import cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

const CoreLifecycleTopic cevent.Topic = "notegic.core.lifecycle.v1"

const CoreYjsMaintenanceHintTopic cevent.Topic = "notegic.core.yjs-maintenance-hint.v1"

const (
	AggregateType_RootShelf   cevent.AggregateType = "RootShelf"
	AggregateType_SubShelf    cevent.AggregateType = "SubShelf"
	AggregateType_BlockPack   cevent.AggregateType = "BlockPack"
	AggregateType_RoutineTask cevent.AggregateType = "RoutineTask"
	AggregateType_User        cevent.AggregateType = "User"
)

const (
	EventType_BlockPackAccessRevoked     cevent.EventType = "BlockPackAccessRevoked"
	EventType_BlockPackRoomPolicyChanged cevent.EventType = "BlockPackRoomPolicyChanged"
	EventType_RootShelfPermissionRevoked cevent.EventType = "RootShelfPermissionRevoked"
	EventType_RootShelfPermissionChanged cevent.EventType = "RootShelfPermissionChanged"
	EventType_RootShelfDeleted           cevent.EventType = "RootShelfDeleted"
	EventType_BlockPackChanged           cevent.EventType = "BlockPackChanged"
	EventType_BlockPackDeleted           cevent.EventType = "BlockPackDeleted"
	EventType_UserSessionsRevoked        cevent.EventType = "UserSessionsRevoked"
	EventType_YjsMaintenanceHint         cevent.EventType = "YjsMaintenanceHint"
	EventType_RoutineTaskCompleted       cevent.EventType = "RoutineTaskCompleted"
)
