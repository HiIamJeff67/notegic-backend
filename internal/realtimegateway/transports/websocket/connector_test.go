package websocket

import (
	"encoding/json"
	"testing"

	"github.com/google/uuid"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"

	realtimetypes "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/types"
)

func TestConnectorPrioritizesControlFramesOverChannelData(t *testing.T) {
	connector := newOutboundTestConnector()

	if err := connector.writeBinary(realtimetypes.BinaryFrame{
		Version:            byte(sconstants.RealtimeProtocolVersion),
		Type:               realtimetypes.BinaryFrameType_YjsDocument,
		ConnectorChannelId: 1,
		Payload:            []byte{1},
	}); err != nil {
		t.Fatalf("failed to enqueue binary frame: %v", err)
	}
	if err := connector.writeJSON(realtimetypes.ErrorFrame{
		Version: sconstants.RealtimeProtocolVersion,
		Type:    realtimetypes.FrameType_Error,
		Code:    realtimetypes.ErrorCode_ChannelBackpressure,
		Message: "resubscribe",
	}); err != nil {
		t.Fatalf("failed to enqueue control frame: %v", err)
	}

	if len(connector.outbound.controlQueue) != 1 {
		t.Fatalf("expected one queued control frame, got %d", len(connector.outbound.controlQueue))
	}

	var errorFrame realtimetypes.ErrorFrame
	if err := json.Unmarshal(connector.outbound.controlQueue[0], &errorFrame); err != nil {
		t.Fatalf("failed to decode control frame: %v", err)
	}
	if errorFrame.Code != realtimetypes.ErrorCode_ChannelBackpressure {
		t.Fatalf("unexpected control frame: %#v", errorFrame)
	}

	if len(connector.outbound.channelQueues[1].messages) != 1 {
		t.Fatal("expected binary frame to remain queued behind the control frame")
	}
}

func TestConnectorCoalescesQueuedAwarenessPerChannel(t *testing.T) {
	connector := newOutboundTestConnector()

	for _, payload := range [][]byte{{1}, {2}} {
		if err := connector.writeBinary(realtimetypes.BinaryFrame{
			Version:            byte(sconstants.RealtimeProtocolVersion),
			Type:               realtimetypes.BinaryFrameType_Awareness,
			ConnectorChannelId: 1,
			Payload:            payload,
		}); err != nil {
			t.Fatalf("failed to enqueue awareness frame: %v", err)
		}
	}

	queue := connector.outbound.channelQueues[1]
	if queue == nil || len(queue.messages) != 1 {
		t.Fatalf("expected one coalesced awareness frame, got %#v", queue)
	}

	var frame realtimetypes.BinaryFrame
	if err := frame.UnmarshalBytes(queue.messages[0].payload); err != nil {
		t.Fatalf("failed to decode awareness frame: %v", err)
	}
	if string(frame.Payload) != string([]byte{2}) {
		t.Fatalf("expected newest awareness payload, got %v", frame.Payload)
	}
}

func TestConnectorLimitsYjsQueueWithoutDroppingQueuedUpdates(t *testing.T) {
	connector := newOutboundTestConnector()
	frame := realtimetypes.BinaryFrame{
		Version:            byte(sconstants.RealtimeProtocolVersion),
		Type:               realtimetypes.BinaryFrameType_YjsDocument,
		ConnectorChannelId: 1,
		Payload:            []byte{1},
	}

	for index := 0; index < sconstants.RealtimeMaxOutboundFramesPerChannel; index++ {
		if err := connector.writeBinary(frame); err != nil {
			t.Fatalf("failed to enqueue update %d: %v", index, err)
		}
	}

	if err := connector.writeBinary(frame); err == nil {
		t.Fatal("expected queue overflow")
	}
	if len(connector.outbound.channelQueues[1].messages) != sconstants.RealtimeMaxOutboundFramesPerChannel {
		t.Fatalf("queued updates were dropped: %d", len(connector.outbound.channelQueues[1].messages))
	}
}

func TestConnectorInitializesOutboundQueue(t *testing.T) {
	connector := &Connector{
		channels: make(map[uint32]realtimetypes.Channel),
		outbound: newOutboundQueue(nil),
	}

	if connector.outbound == nil || connector.outbound.channelQueues == nil {
		t.Fatal("expected connector to initialize its outbound queue")
	}
}

func TestConnectorUnsubscribeClearsOutboundChannelQueue(t *testing.T) {
	connector := newOutboundTestConnector()
	connector.channels[1] = realtimetypes.Channel{
		Type: realtimetypes.ChannelType_BlockPack,
		Id:   uuid.New(),
	}

	if err := connector.writeBinary(realtimetypes.BinaryFrame{
		Version:            byte(sconstants.RealtimeProtocolVersion),
		Type:               realtimetypes.BinaryFrameType_YjsDocument,
		ConnectorChannelId: 1,
		Payload:            []byte{1},
	}); err != nil {
		t.Fatalf("failed to enqueue binary frame: %v", err)
	}

	if _, exists := connector.unsubscribe(1); !exists {
		t.Fatal("expected subscribed channel to be removed")
	}

	if _, exists := connector.outbound.channelQueues[1]; exists {
		t.Fatal("expected unsubscribe to clear the outbound channel queue")
	}
}

func TestConnectorWriterStartsAndStops(t *testing.T) {
	connector := &Connector{
		outbound: newOutboundQueue(nil),
	}

	connector.startWriter()
	if !connector.outbound.started {
		t.Fatal("expected writer to start")
	}

	connector.stopWriter()

	select {
	case <-connector.outbound.stopped:
	default:
		t.Fatal("expected writer to stop")
	}
}

func newOutboundTestConnector() *Connector {
	outbound := newOutboundQueue(nil)
	outbound.wake = make(chan struct{}, 1)
	outbound.done = make(chan struct{})
	outbound.stopped = make(chan struct{})
	outbound.started = true

	return &Connector{
		channels: make(map[uint32]realtimetypes.Channel),
		outbound: outbound,
	}
}
