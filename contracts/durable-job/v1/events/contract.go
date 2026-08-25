package durablejobeventscontract

import cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

const (
	DurableJobRealtimeGatewayRoutineTaskLifecycleTopic cevent.Topic = "notegic.durablejob.realtime-gateway.routine-task-lifecycle.v1"
)

const (
	AggregateType_BlockPack        cevent.AggregateType = "BlockPack"
	AggregateType_DurableJobWorker cevent.AggregateType = "DurableJobWorker"
	AggregateType_RoutineTask      cevent.AggregateType = "RoutineTask"
)

const (
	EventType_RoutineTaskClaimRequested cevent.EventType = "RoutineTaskClaimRequested"
	EventType_RoutineTasksAssigned      cevent.EventType = "RoutineTasksAssigned"
	EventType_RoutineTaskRunning        cevent.EventType = "RoutineTaskRunning"
)
