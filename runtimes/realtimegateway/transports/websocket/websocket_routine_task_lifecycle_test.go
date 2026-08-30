package websocket

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/google/uuid"

	realtimeleasecache "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/data/redis/realtimelease"
	realtimetypes "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/types"
)

func TestBroadcastRoutineTaskLifecycleEventTargetsOnlyTheActor(t *testing.T) {
	actorUserPublicId := uuid.New()
	unrelatedUserPublicId := uuid.New()
	actorConnector := &Connector{
		Id:           uuid.New(),
		UserPublicId: actorUserPublicId,
		outbound: &outboundQueue{
			started:      true,
			wake:         make(chan struct{}, 1),
			controlQueue: make([][]byte, 0),
			channelQueues: make(map[uint32]*struct {
				messages []struct {
					payload   []byte
					frameType realtimetypes.BinaryFrameType
				}
				queuedBytes int64
			}),
		},
	}
	unrelatedConnector := &Connector{
		Id:           uuid.New(),
		UserPublicId: unrelatedUserPublicId,
		outbound: &outboundQueue{
			started:      true,
			wake:         make(chan struct{}, 1),
			controlQueue: make([][]byte, 0),
			channelQueues: make(map[uint32]*struct {
				messages []struct {
					payload   []byte
					frameType realtimetypes.BinaryFrameType
				}
				queuedBytes int64
			}),
		},
	}
	adapter := &WebSocketAdapter{
		connectors: map[uuid.UUID]*Connector{
			actorConnector.Id:     actorConnector,
			unrelatedConnector.Id: unrelatedConnector,
		},
	}

	event := realtimeleasecache.RoutineTaskLifecycleEvent{
		EventId:             uuid.New(),
		RoutineTaskId:       uuid.New(),
		RoutineTaskRecordId: uuid.New(),
		RoutineId:           uuid.New(),
		ActorUserPublicId:   actorUserPublicId,
		Purpose:             "CreateBlockPack",
		Status:              "running",
		Attempt:             1,
		OccurredAt:          time.Now().UTC(),
	}
	adapter.broadcastRoutineTaskLifecycleEvent(event)

	if len(actorConnector.outbound.controlQueue) != 1 {
		t.Fatalf("actor queue length = %d, want 1", len(actorConnector.outbound.controlQueue))
	}
	if len(unrelatedConnector.outbound.controlQueue) != 0 {
		t.Fatalf("unrelated queue length = %d, want 0", len(unrelatedConnector.outbound.controlQueue))
	}

	frame := realtimetypes.RoutineTaskLifecycleFrame{}
	if err := json.Unmarshal(actorConnector.outbound.controlQueue[0], &frame); err != nil {
		t.Fatalf("decode RoutineTask lifecycle frame: %v", err)
	}
	if frame.Type != realtimetypes.FrameType_RoutineTaskLifecycle ||
		frame.RoutineTaskId != event.RoutineTaskId || frame.Status != event.Status {
		t.Fatalf("unexpected RoutineTask lifecycle frame: %#v", frame)
	}
}
