package durablejobconsumers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	durablejobcontract "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1"
	durablejobeventscontract "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/events"
	durablejobroutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	eventcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"

	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	coreenums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	routineservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/routines"
)

type routineTaskExecutionServiceStub struct {
	eventId uuid.UUID
	called  bool
}

func (s *routineTaskExecutionServiceStub) ValidateRoutineTaskPayload(
	coreenums.RoutineTaskPurpose,
	datatypes.JSON,
) *exceptions.Exception {
	return nil
}

func (s *routineTaskExecutionServiceStub) ResolveRoutineTaskPatterns(
	context.Context,
	[]schemas.RoutineTask,
	[]uuid.UUID,
	[]durablejobroutinetasktypes.RoutineTaskPattern,
	[]coreenums.AccessControlPermission,
) ([]map[string]string, []bool, *exceptions.Exception) {
	return nil, nil, nil
}

func (s *routineTaskExecutionServiceStub) ApplyPreparedRoutineTasks(
	_ context.Context,
	eventId uuid.UUID,
	_ *durablejobcontract.MarkCompletedRoutineTasksRequestDto,
) *exceptions.Exception {
	s.eventId = eventId
	s.called = true
	return nil
}

var _ routineservices.RoutineTaskExecutionServiceInterface = (*routineTaskExecutionServiceStub)(nil)

func TestRoutineTaskResultConsumerDelegatesCompletedResultToCoreExecutionService(t *testing.T) {
	workerId := uuid.New()
	eventId := uuid.New()
	request := durablejobcontract.MarkCompletedRoutineTasksRequestDto{WorkerId: workerId}
	data, err := json.Marshal(request)
	if err != nil {
		t.Fatalf("marshal completion request: %v", err)
	}
	event := eventcontract.EventEnvelope[json.RawMessage]{
		SchemaVersion: eventcontract.Version,
		EventId:       eventId,
		EventType:     durablejobeventscontract.EventType_RoutineTasksCompleted,
		AggregateType: durablejobeventscontract.AggregateType_DurableJobWorker,
		AggregateId:   workerId,
		KafkaKey:      workerId.String(),
		OccurredAt:    time.Now().UTC(),
		Data:          data,
	}
	stub := &routineTaskExecutionServiceStub{}
	consumer := &DurableJobRoutineTaskResultConsumer{
		routineTaskExecutionService: stub,
	}

	if err := consumer.consume(
		context.Background(),
		platformkafka.ConsumerRecord{Key: workerId.String()},
		event,
	); err != nil {
		t.Fatalf("consume completion result: %v", err)
	}
	if !stub.called || stub.eventId != eventId {
		t.Fatalf("execution service was not called with event %s", eventId)
	}
}

func TestRoutineTaskResultConsumerRejectsMismatchedWorkerEnvelope(t *testing.T) {
	workerId := uuid.New()
	event := eventcontract.EventEnvelope[json.RawMessage]{
		SchemaVersion: eventcontract.Version,
		EventId:       uuid.New(),
		EventType:     durablejobeventscontract.EventType_RoutineTasksCompleted,
		AggregateType: durablejobeventscontract.AggregateType_DurableJobWorker,
		AggregateId:   workerId,
		KafkaKey:      workerId.String(),
		OccurredAt:    time.Now().UTC(),
		Data:          json.RawMessage(`{"workerId":"` + uuid.NewString() + `"}`),
	}
	consumer := &DurableJobRoutineTaskResultConsumer{}

	err := consumer.consume(
		context.Background(),
		platformkafka.ConsumerRecord{Key: workerId.String()},
		event,
	)
	consumerError, ok := err.(*platformkafka.ConsumerError)
	if !ok || consumerError.Classification != platformkafka.ErrorClassification_SchemaIncompatible {
		t.Fatalf("error = %#v, want schema-incompatible ConsumerError", err)
	}
}
