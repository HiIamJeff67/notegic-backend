package realtimegatewayproducers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
)

type RoutineTaskCompletionProducer struct {
	producer *skafka.Producer
}

func NewRoutineTaskCompletionProducer(
	producer *skafka.Producer,
) *RoutineTaskCompletionProducer {
	return &RoutineTaskCompletionProducer{
		producer: producer,
	}
}

func (p *RoutineTaskCompletionProducer) ProduceRoutineTaskCompleted(
	ctx context.Context,
	completedTasks []croutinetasktypes.CompletedRoutineTask,
	workerId uuid.UUID,
) error {
	for _, completedTask := range completedTasks {
		payload, err := json.Marshal(cevent.EventEnvelope[coreevents.RoutineTaskCompletedData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.NewSHA1(uuid.NameSpaceURL, []byte(completedTask.RoutineTaskRecordId.String()+":completed")),
			EventType:     coreevents.EventType_RoutineTaskCompleted,
			AggregateType: coreevents.AggregateType_RoutineTask,
			AggregateId:   completedTask.RoutineTaskId,
			KafkaKey:      completedTask.RoutineTaskId.String(),
			OccurredAt:    completedTask.CompletedAt,
			CorrelationId: workerId.String(),
			Data: coreevents.RoutineTaskCompletedData{
				RoutineTaskId:       completedTask.RoutineTaskId,
				RoutineTaskRecordId: completedTask.RoutineTaskRecordId,
				RoutineRecordId:     completedTask.RoutineRecordId,
				RoutineId:           completedTask.PreparedTask.RoutineId,
				ActorUserPublicId:   completedTask.PreparedTask.ActorUserPublicId,
				Purpose:             completedTask.PreparedTask.Purpose,
				WorkerId:            workerId,
				Attempt:             completedTask.PreparedTask.Attempt,
				CompletedAt:         completedTask.CompletedAt,
				ExecutionResult:     completedTask.ExecutionResult,
			},
		})
		if err != nil {
			return err
		}
		if err := p.producer.Produce(
			ctx,
			coreevents.CoreLifecycleTopic.String(),
			completedTask.RoutineTaskId.String(),
			payload,
		); err != nil {
			return err
		}
	}

	return nil
}
