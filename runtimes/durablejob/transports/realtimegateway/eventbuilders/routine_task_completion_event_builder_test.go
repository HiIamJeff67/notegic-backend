package eventbuilders

import (
	"testing"
	"time"

	"github.com/google/uuid"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
)

func TestRoutineTaskCompletionEventBuilderBuildsCoreOwnedLifecycleEvent(t *testing.T) {
	taskId := uuid.New()
	recordId := uuid.New()
	routineId := uuid.New()
	actorUserPublicId := uuid.New()
	workerId := uuid.New()
	completedAt := time.Date(2026, 8, 6, 12, 0, 0, 0, time.UTC)
	event := NewRoutineTaskCompletionEventBuilder().Build(
		croutinetasktypes.CompletedRoutineTask{
			RoutineTaskId:       taskId,
			RoutineTaskRecordId: recordId,
			CompletedAt:         completedAt,
			PreparedTask: &croutinetasktypes.PreparedRoutineTask{
				RoutineId:         routineId,
				ActorUserPublicId: actorUserPublicId,
				Purpose:           "CreateBlockPack",
				Attempt:           2,
			},
		},
		workerId,
		completedAt,
	)
	retriedEvent := NewRoutineTaskCompletionEventBuilder().Build(
		croutinetasktypes.CompletedRoutineTask{
			RoutineTaskId:       taskId,
			RoutineTaskRecordId: recordId,
			CompletedAt:         completedAt,
			PreparedTask: &croutinetasktypes.PreparedRoutineTask{
				RoutineId:         routineId,
				ActorUserPublicId: actorUserPublicId,
				Purpose:           "CreateBlockPack",
				Attempt:           2,
			},
		},
		workerId,
		completedAt,
	)

	if event.EventType != coreevents.EventType_RoutineTaskCompleted ||
		event.AggregateType != coreevents.AggregateType_RoutineTask ||
		event.AggregateId != taskId || event.KafkaKey != taskId.String() {
		t.Fatalf("unexpected event envelope: %#v", event)
	}
	if event.Data.WorkerId != workerId || event.Data.Attempt != 2 ||
		event.Data.RoutineId != routineId || event.Data.ActorUserPublicId != actorUserPublicId ||
		event.Data.Purpose != "CreateBlockPack" {
		t.Fatalf("unexpected event data: %#v", event.Data)
	}
	if event.EventId != retriedEvent.EventId {
		t.Fatalf("completion event id is not retry-stable: %s != %s", event.EventId, retriedEvent.EventId)
	}
}
