package realtimegatewayproducers

import (
	"context"
	"encoding/json"

	"github.com/google/uuid"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"

	eventbuilders "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/transports/realtimegateway/eventbuilders"
)

type RoutineTaskCompletionProducer struct {
	producer     *skafka.Producer
	eventBuilder *eventbuilders.RoutineTaskCompletionEventBuilder
}

func NewRoutineTaskCompletionProducer(
	producer *skafka.Producer,
) *RoutineTaskCompletionProducer {
	return &RoutineTaskCompletionProducer{
		producer:     producer,
		eventBuilder: eventbuilders.NewRoutineTaskCompletionEventBuilder(),
	}
}

func (p *RoutineTaskCompletionProducer) ProduceRoutineTaskCompleted(
	ctx context.Context,
	completedTasks []croutinetasktypes.CompletedRoutineTask,
	workerId uuid.UUID,
) error {
	for _, completedTask := range completedTasks {
		payload, err := json.Marshal(p.eventBuilder.Build(
			completedTask,
			workerId,
			completedTask.CompletedAt,
		))
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
