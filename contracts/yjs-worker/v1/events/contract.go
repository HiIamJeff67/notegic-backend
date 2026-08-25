package adapterseventscontract

import (
	"github.com/google/uuid"

	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
)

const (
	YjsWorkerCoreCommandTopic            cevent.Topic = "notegic.adapters.core.command.v1"
	CoreYjsWorkerReplyTopic              cevent.Topic = "notegic.core.adapters.reply.v1"
	YjsWorkerCoreMaintenanceCommandTopic cevent.Topic = "notegic.core.adapters.maintenance-command.v1"
	CoreYjsWorkerMaintenanceResultTopic  cevent.Topic = "notegic.adapters.core.maintenance-result.v1"
)

const (
	EventType_YjsWorkerCommand          cevent.EventType = "YjsWorkerCommand"
	EventType_YjsWorkerCommandCompleted cevent.EventType = "YjsWorkerCommandCompleted"
	EventType_YjsMaintenanceCommand     cevent.EventType = "YjsMaintenanceCommand"
	EventType_YjsMaintenanceCompleted   cevent.EventType = "YjsMaintenanceCompleted"
)

const AggregateType_BlockPack cevent.AggregateType = "BlockPack"

type YjsMaintenanceOperation string

const (
	YjsMaintenanceOperation_Compact YjsMaintenanceOperation = "compact"
	YjsMaintenanceOperation_Project YjsMaintenanceOperation = "project"
)

type YjsMaintenanceCommandData struct {
	RequestId      uuid.UUID               `json:"requestId"`
	BlockPackId    uuid.UUID               `json:"blockPackId"`
	DocumentId     uuid.UUID               `json:"documentId"`
	Operation      YjsMaintenanceOperation `json:"operation"`
	TargetSequence int64                   `json:"targetSequence"`
	CorrelationId  string                  `json:"correlationId"`
}

type YjsMaintenanceResultData struct {
	RequestId              uuid.UUID               `json:"requestId"`
	BlockPackId            uuid.UUID               `json:"blockPackId"`
	DocumentId             uuid.UUID               `json:"documentId"`
	Operation              YjsMaintenanceOperation `json:"operation"`
	TargetSequence         int64                   `json:"targetSequence"`
	Success                bool                    `json:"success"`
	CompactedUntilSequence int64                   `json:"compactedUntilSequence"`
	ProjectedUntilSequence int64                   `json:"projectedUntilSequence"`
	Error                  string                  `json:"error,omitempty"`
}
