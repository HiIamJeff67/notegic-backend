package eventbuilders

import (
	"time"

	"github.com/google/uuid"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
)

type RoutineTaskCompletionEventBuilder struct{}

func NewRoutineTaskCompletionEventBuilder() *RoutineTaskCompletionEventBuilder {
	return &RoutineTaskCompletionEventBuilder{}
}

func (b *RoutineTaskCompletionEventBuilder) Build(
	completedTask croutinetasktypes.CompletedRoutineTask,
	workerId uuid.UUID,
	occurredAt time.Time,
) cevent.EventEnvelope[coreevents.RoutineTaskCompletedData] {
	return cevent.EventEnvelope[coreevents.RoutineTaskCompletedData]{
		SchemaVersion: cevent.Version,
		EventId:       uuid.New(),
		EventType:     coreevents.EventType_RoutineTaskCompleted,
		AggregateType: coreevents.AggregateType_RoutineTask,
		AggregateId:   completedTask.RoutineTaskId,
		KafkaKey:      completedTask.RoutineTaskId.String(),
		OccurredAt:    occurredAt,
		CorrelationId: workerId.String(),
		Data: coreevents.RoutineTaskCompletedData{
			RoutineTaskId:       completedTask.RoutineTaskId,
			RoutineTaskRecordId: completedTask.RoutineTaskRecordId,
			RoutineId:           completedTask.PreparedTask.RoutineId,
			ActorUserPublicId:   completedTask.PreparedTask.ActorUserPublicId,
			Purpose:             completedTask.PreparedTask.Purpose,
			WorkerId:            workerId,
			Attempt:             completedTask.PreparedTask.Attempt,
			CompletedAt:         completedTask.CompletedAt,
		},
	}
}
