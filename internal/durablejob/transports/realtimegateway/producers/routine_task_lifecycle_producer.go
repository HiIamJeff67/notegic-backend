package realtimegatewayproducers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	cdurablejobevents "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/events"
	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
)

type RoutineTaskLifecycleProducer struct {
	producer *skafka.Producer
}

func NewRoutineTaskLifecycleProducer(
	producer *skafka.Producer,
) *RoutineTaskLifecycleProducer {
	return &RoutineTaskLifecycleProducer{
		producer: producer,
	}
}

func (p *RoutineTaskLifecycleProducer) ProduceRoutineTaskRunning(
	ctx context.Context,
	assignment croutinetasktypes.RoutineTaskAssignment,
) error {
	now := time.Now().UTC()
	payload, err := json.Marshal(cevent.EventEnvelope[cdurablejobevents.RoutineTaskRunningData]{
		SchemaVersion: cevent.Version,
		EventId:       uuid.New(),
		EventType:     cdurablejobevents.EventType_RoutineTaskRunning,
		AggregateType: cdurablejobevents.AggregateType_RoutineTask,
		AggregateId:   assignment.RoutineTaskId,
		KafkaKey:      assignment.RoutineTaskId.String(),
		OccurredAt:    now,
		CorrelationId: assignment.RoutineTaskRecordId.String(),
		Data: cdurablejobevents.RoutineTaskRunningData{
			RoutineTaskId:       assignment.RoutineTaskId,
			RoutineTaskRecordId: assignment.RoutineTaskRecordId,
			RoutineId:           assignment.RoutineId,
			ActorUserPublicId:   assignment.ActorUserPublicId,
			Purpose:             assignment.Purpose,
			Attempt:             assignment.Attempt,
			StartedAt:           now,
		},
	})
	if err != nil {
		return err
	}

	return p.producer.Produce(
		ctx,
		cdurablejobevents.DurableJobRealtimeGatewayRoutineTaskLifecycleTopic.String(),
		assignment.RoutineTaskId.String(),
		payload,
	)
}
