package websocket

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"go.opentelemetry.io/otel/attribute"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	smetrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"

	realtimetypes "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/types"
)

type outboundQueue struct {
	connection *websocket.Conn

	mutex      sync.Mutex // protect controlQueue, channelQueues, channelOrder, channelOrderCursor, wake, done, stopped, started
	writeMutex sync.Mutex // protect websocket connection write methods, ex. writeBinary, writeJSON, writeControl, SetWriteDeadline
	wake       chan struct{}
	done       chan struct{}
	stopped    chan struct{}
	stopOnce   sync.Once
	started    bool

	controlQueue       [][]byte
	controlQueuedBytes int64
	channelQueues      map[uint32]*struct {
		messages []struct {
			payload   []byte
			frameType realtimetypes.BinaryFrameType
		}
		queuedBytes int64
	}
	channelOrder       []uint32
	channelOrderCursor int
}

func newOutboundQueue(connection *websocket.Conn) *outboundQueue {
	return &outboundQueue{
		connection:   connection,
		controlQueue: make([][]byte, 0, sconstants.RealtimeMaxOutboundControlFrames),
		channelQueues: make(map[uint32]*struct {
			messages []struct {
				payload   []byte
				frameType realtimetypes.BinaryFrameType
			}
			queuedBytes int64
		}),
		channelOrder: make([]uint32, 0, sconstants.RealtimeMaxChannelsPerConnection),
	}
}

func (q *outboundQueue) startWriter() {
	q.mutex.Lock()
	if q.started {
		q.mutex.Unlock()

		return
	}
	q.wake = make(chan struct{}, 1)
	q.done = make(chan struct{})
	q.stopped = make(chan struct{})
	q.started = true
	q.mutex.Unlock()

	go q.runWriter()
}

func (q *outboundQueue) stopWriter() {
	q.mutex.Lock()
	if !q.started {
		q.mutex.Unlock()

		return
	}
	done := q.done
	stopped := q.stopped
	q.mutex.Unlock()

	q.stopOnce.Do(func() {
		close(done)
	})

	<-stopped
}

func (q *outboundQueue) writeJSON(frame any) error {
	payload, err := json.Marshal(frame)
	if err != nil {
		return err
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if !q.started || len(q.controlQueue) >= sconstants.RealtimeMaxOutboundControlFrames {
		return errors.New("realtime control queue is full")
	}

	q.controlQueue = append(q.controlQueue, payload)
	q.controlQueuedBytes += int64(len(payload))
	smetrics.NotegicMeter.Value(context.Background(), "realtime.outbound.queue.depth", int64(len(q.controlQueue)),
		attribute.String("queueType", "control"),
	)
	smetrics.NotegicMeter.Value(context.Background(), "realtime.outbound.queue.bytes", q.controlQueuedBytes,
		attribute.String("queueType", "control"),
	)
	q.trySignalWriterLocked()

	return nil
}

func (q *outboundQueue) writeControl(messageType int, payload []byte) error {
	q.writeMutex.Lock()
	defer q.writeMutex.Unlock()

	return q.connection.WriteControl(
		messageType,
		payload,
		time.Now().Add(sconstants.RealtimeControlWriteTimeout),
	)
}

func (q *outboundQueue) writeBinary(frame realtimetypes.BinaryFrame) error {
	payload, err := frame.MarshalBytes()
	if err != nil {
		return err
	}

	q.mutex.Lock()
	defer q.mutex.Unlock()

	if !q.started {
		return errors.New("realtime outbound writer is not started")
	}

	queue, exists := q.channelQueues[frame.ConnectorChannelId]
	if !exists {
		queue = &struct {
			messages []struct {
				payload   []byte
				frameType realtimetypes.BinaryFrameType
			}
			queuedBytes int64
		}{}
		q.channelQueues[frame.ConnectorChannelId] = queue
		q.channelOrder = append(q.channelOrder, frame.ConnectorChannelId)
	}

	if frame.Type == realtimetypes.BinaryFrameType_Awareness {
		for index := len(queue.messages) - 1; index >= 0; index-- {
			message := queue.messages[index]
			if message.frameType != realtimetypes.BinaryFrameType_Awareness {
				continue
			}

			queuedBytesWithoutAwareness := queue.queuedBytes - int64(len(message.payload))
			if queuedBytesWithoutAwareness+int64(len(payload)) > sconstants.RealtimeMaxOutboundBytesPerChannel {
				return errors.New("realtime channel outbound queue is full")
			}

			queue.messages[index] = struct {
				payload   []byte
				frameType realtimetypes.BinaryFrameType
			}{
				payload:   payload,
				frameType: frame.Type,
			}
			queue.queuedBytes = queuedBytesWithoutAwareness + int64(len(payload))
			smetrics.NotegicMeter.Value(context.Background(), "realtime.outbound.queue.depth", int64(len(queue.messages)),
				attribute.String("queueType", "channel"),
				attribute.String("frameType", string(frame.Type)),
			)
			smetrics.NotegicMeter.Value(context.Background(), "realtime.outbound.queue.bytes", queue.queuedBytes,
				attribute.String("queueType", "channel"),
				attribute.String("frameType", string(frame.Type)),
			)
			q.trySignalWriterLocked()

			return nil
		}
	}

	if len(queue.messages) >= sconstants.RealtimeMaxOutboundFramesPerChannel ||
		queue.queuedBytes+int64(len(payload)) > sconstants.RealtimeMaxOutboundBytesPerChannel {
		return errors.New("realtime channel outbound queue is full")
	}

	queue.messages = append(queue.messages, struct {
		payload   []byte
		frameType realtimetypes.BinaryFrameType
	}{
		payload:   payload,
		frameType: frame.Type,
	})
	queue.queuedBytes += int64(len(payload))
	smetrics.NotegicMeter.Value(context.Background(), "realtime.outbound.queue.depth", int64(len(queue.messages)),
		attribute.String("queueType", "channel"),
		attribute.String("frameType", string(frame.Type)),
	)
	smetrics.NotegicMeter.Value(context.Background(), "realtime.outbound.queue.bytes", queue.queuedBytes,
		attribute.String("queueType", "channel"),
		attribute.String("frameType", string(frame.Type)),
	)
	q.trySignalWriterLocked()

	return nil
}

func (q *outboundQueue) clearChannel(connectorChannelId uint32) {
	q.mutex.Lock()
	defer q.mutex.Unlock()

	delete(q.channelQueues, connectorChannelId)
	smetrics.NotegicMeter.Value(context.Background(), "realtime.outbound.queue.depth", 0,
		attribute.String("queueType", "channel"),
	)
	smetrics.NotegicMeter.Value(context.Background(), "realtime.outbound.queue.bytes", 0,
		attribute.String("queueType", "channel"),
	)
	for index, channelId := range q.channelOrder {
		if channelId != connectorChannelId {
			continue
		}

		q.channelOrder = append(q.channelOrder[:index], q.channelOrder[index+1:]...)
		if q.channelOrderCursor >= len(q.channelOrder) {
			q.channelOrderCursor = 0
		}

		return
	}
}

func (q *outboundQueue) runWriter() {
	defer close(q.stopped)

	for {
		q.mutex.Lock()
		messageType := 0
		var payload []byte
		if len(q.controlQueue) > 0 {
			messageType = websocket.TextMessage
			payload = q.controlQueue[0]
			q.controlQueue = q.controlQueue[1:]
			q.controlQueuedBytes -= int64(len(payload))
			smetrics.NotegicMeter.Value(context.Background(), "realtime.outbound.queue.depth", int64(len(q.controlQueue)),
				attribute.String("queueType", "control"),
			)
			smetrics.NotegicMeter.Value(context.Background(), "realtime.outbound.queue.bytes", q.controlQueuedBytes,
				attribute.String("queueType", "control"),
			)
		} else {
			// Round robin through non-empty per-channel queues.
			for offset := 0; offset < len(q.channelOrder); offset++ {
				index := (q.channelOrderCursor + offset) % len(q.channelOrder)
				channelId := q.channelOrder[index]
				queue := q.channelQueues[channelId]
				if queue == nil || len(queue.messages) == 0 {
					continue
				}

				messageType = websocket.BinaryMessage
				message := queue.messages[0]
				payload = message.payload
				queue.messages = queue.messages[1:]
				queue.queuedBytes -= int64(len(message.payload))
				smetrics.NotegicMeter.Value(context.Background(), "realtime.outbound.queue.depth", int64(len(queue.messages)),
					attribute.String("queueType", "channel"),
					attribute.String("frameType", string(message.frameType)),
				)
				smetrics.NotegicMeter.Value(context.Background(), "realtime.outbound.queue.bytes", queue.queuedBytes,
					attribute.String("queueType", "channel"),
					attribute.String("frameType", string(message.frameType)),
				)
				q.channelOrderCursor = (index + 1) % len(q.channelOrder)

				break
			}
		}

		if messageType == 0 {
			wake := q.wake
			done := q.done
			q.mutex.Unlock()

			// wait until wake or done signal is triggered
			select {
			case <-wake:
				continue
			case <-done:
				return
			}
		}
		q.mutex.Unlock()

		q.writeMutex.Lock()
		err := q.connection.SetWriteDeadline(
			time.Now().Add(sconstants.RealtimeControlWriteTimeout),
		)
		if err == nil {
			err = q.connection.WriteMessage(messageType, payload)
		}
		q.writeMutex.Unlock()
		if err != nil {
			_ = q.connection.Close()

			return
		}
	}
}

func (q *outboundQueue) trySignalWriterLocked() {
	select {
	case q.wake <- struct{}{}:
	default:
	}
}
