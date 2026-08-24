package websocket

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	realtimeleasecache "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/data/redis/realtimelease"
	realtimetypes "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/types"
)

func TestBroadcastNotificationTargetsOnlyTheRecipient(t *testing.T) {
	recipientUserPublicId := uuid.New()
	unrelatedUserPublicId := uuid.New()
	recipientConnector := &Connector{
		Id:           uuid.New(),
		UserPublicId: recipientUserPublicId,
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
			recipientConnector.Id: recipientConnector,
			unrelatedConnector.Id: unrelatedConnector,
		},
	}

	event := realtimeleasecache.NotificationEvent{
		EventId:               uuid.New(),
		NotificationId:        uuid.New(),
		RecipientUserPublicId: recipientUserPublicId,
		Type:                  "news",
		Priority:              "normal",
		TemplateKey:           "news",
		TemplateVersion:       1,
		Payload:               json.RawMessage(`{"title":"Release update"}`),
	}
	adapter.broadcastNotification(event)

	if len(recipientConnector.outbound.controlQueue) != 1 {
		t.Fatalf("recipient queue length = %d, want 1", len(recipientConnector.outbound.controlQueue))
	}
	if len(unrelatedConnector.outbound.controlQueue) != 0 {
		t.Fatalf("unrelated queue length = %d, want 0", len(unrelatedConnector.outbound.controlQueue))
	}

	frame := realtimetypes.NotificationFrame{}
	if err := json.Unmarshal(recipientConnector.outbound.controlQueue[0], &frame); err != nil {
		t.Fatalf("decode notification frame: %v", err)
	}
	if frame.Type != realtimetypes.FrameType_Notification || frame.NotificationId != event.NotificationId {
		t.Fatalf("unexpected notification frame: %#v", frame)
	}
	if string(frame.Payload) != string(event.Payload) {
		t.Fatalf("notification payload = %s, want %s", frame.Payload, event.Payload)
	}
}
