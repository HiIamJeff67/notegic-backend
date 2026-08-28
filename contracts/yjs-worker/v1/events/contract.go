package adapterseventscontract

import cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

const (
	YjsWorkerCoreCommandTopic cevent.Topic = "notegic.adapters.core.command.v1"
	CoreYjsWorkerReplyTopic   cevent.Topic = "notegic.core.adapters.reply.v1"
)

const (
	EventType_YjsWorkerCommand          cevent.EventType = "YjsWorkerCommand"
	EventType_YjsWorkerCommandCompleted cevent.EventType = "YjsWorkerCommandCompleted"
)

const AggregateType_BlockPack cevent.AggregateType = "BlockPack"
