package eventbuilders

import (
	"time"

	"github.com/google/uuid"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
)

type YjsMaintenanceHintEventBuilder struct{}

func NewYjsMaintenanceHintEventBuilder() *YjsMaintenanceHintEventBuilder {
	return &YjsMaintenanceHintEventBuilder{}
}

func (b *YjsMaintenanceHintEventBuilder) Build(
	hint coreevents.YjsMaintenanceHintData,
	correlationId string,
	occurredAt time.Time,
) cevent.EventEnvelope[coreevents.YjsMaintenanceHintData] {
	return cevent.EventEnvelope[coreevents.YjsMaintenanceHintData]{
		SchemaVersion: cevent.Version,
		EventId:       uuid.New(),
		EventType:     coreevents.EventType_YjsMaintenanceHint,
		AggregateType: coreevents.AggregateType_BlockPack,
		AggregateId:   hint.BlockPackId,
		KafkaKey:      hint.BlockPackId.String(),
		OccurredAt:    occurredAt,
		CorrelationId: correlationId,
		Data:          hint,
	}
}
