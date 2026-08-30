package realtimegatewayproducers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

func TestRoutineTaskLifecycleProducerReturnsErrorWhenKafkaIsUnavailable(t *testing.T) {
	producer := NewRoutineTaskLifecycleProducer(nil)
	err := producer.ProduceRoutineTaskRunning(
		context.Background(),
		croutinetasktypes.RoutineTaskAssignment{
			RoutineTaskId:       uuid.New(),
			RoutineTaskRecordId: uuid.New(),
			RoutineId:           uuid.New(),
			ActorUserId:         uuid.New(),
			ActorUserPublicId:   uuid.New(),
			Purpose:             cenums.RoutineTaskPurpose_CreateBlockPack,
			Payload:             json.RawMessage(`{}`),
			Attempt:             1,
			StartedAt:           time.Now().UTC(),
		},
	)
	if err == nil {
		t.Fatal("expected an unavailable Kafka producer error")
	}
}
