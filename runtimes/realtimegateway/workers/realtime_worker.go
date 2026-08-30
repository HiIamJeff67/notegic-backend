package workers

import (
	"context"
	"hash/fnv"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gorilla/websocket"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"

	realtimetypes "github.com/HiIamJeff67/notegic-backend/runtimes/realtimegateway/types"
)

type WorkerManagerInterface interface {
	Attach(frame realtimetypes.InternalFrame) bool
	Detach(frame realtimetypes.InternalFrame)
	Forward(frame realtimetypes.InternalFrame) bool
	SetFrameHandler(handler func(realtimetypes.InternalFrame))
}

type WorkerManager struct {
	workers []*realtimeWorker
	cancel  context.CancelFunc
	waiter  sync.WaitGroup
}

func NewWorkerManager(endpoints []string) *WorkerManager {
	ctx, cancel := context.WithCancel(context.Background())
	workers := make([]*realtimeWorker, 0, len(endpoints))

	for _, endpoint := range endpoints {
		worker := &realtimeWorker{
			endpoint:       endpoint,
			activeChannels: make(map[string]realtimetypes.InternalFrame),
			outbound:       make(chan realtimetypes.InternalFrame, sconstants.RealtimeWorkerQueueSize),
		}
		workers = append(workers, worker)
	}

	manager := &WorkerManager{
		workers: workers,
		cancel:  cancel,
	}
	for _, worker := range manager.workers {
		manager.waiter.Add(1)
		go func() {
			defer manager.waiter.Done()
			worker.run(ctx)
		}()
	}

	return manager
}

func (m *WorkerManager) Shutdown() {
	m.cancel()
	for _, worker := range m.workers {
		worker.close()
	}
	m.waiter.Wait()
}

func (m *WorkerManager) Attach(frame realtimetypes.InternalFrame) bool {
	worker := m.getWorker(frame.ChannelId)
	if worker == nil || !worker.ready.Load() {
		return false
	}

	worker.activeChannelsMutex.Lock()
	worker.activeChannels[worker.channelKey(frame)] = frame
	worker.activeChannelsMutex.Unlock()

	return worker.enqueue(frame)
}

func (m *WorkerManager) Detach(frame realtimetypes.InternalFrame) {
	worker := m.getWorker(frame.ChannelId)
	if worker == nil {
		return
	}

	worker.activeChannelsMutex.Lock()
	delete(worker.activeChannels, worker.channelKey(frame))
	worker.activeChannelsMutex.Unlock()

	if worker.ready.Load() {
		worker.enqueue(frame)
	}
}

func (m *WorkerManager) Forward(frame realtimetypes.InternalFrame) bool {
	worker := m.getWorker(frame.ChannelId)
	if worker == nil || !worker.ready.Load() {
		return false
	}

	return worker.enqueue(frame)
}

func (m *WorkerManager) SetFrameHandler(handler func(realtimetypes.InternalFrame)) {
	for _, worker := range m.workers {
		worker.frameHandlerMutex.Lock()
		worker.frameHandler = handler
		worker.frameHandlerMutex.Unlock()
	}
}

func (m *WorkerManager) getWorker(channelId [16]byte) *realtimeWorker {
	if len(m.workers) == 0 {
		return nil
	}

	hasher := fnv.New32a()
	_, _ = hasher.Write(channelId[:])

	return m.workers[int(hasher.Sum32())%len(m.workers)]
}

/* ============================== Realtime Worker Methods ============================== */

type realtimeWorker struct {
	endpoint string

	activeChannels      map[string]realtimetypes.InternalFrame
	activeChannelsMutex sync.RWMutex

	frameHandler      func(realtimetypes.InternalFrame)
	frameHandlerMutex sync.RWMutex

	outbound        chan realtimetypes.InternalFrame
	ready           atomic.Bool
	connectionMutex sync.Mutex
	connection      *websocket.Conn
}

func (w *realtimeWorker) channelKey(frame realtimetypes.InternalFrame) string {
	return frame.ConnectionId.String() + ":" + strconv.FormatUint(uint64(frame.ConnectorChannelId), 10)
}

func (w *realtimeWorker) enqueue(frame realtimetypes.InternalFrame) bool {
	select {
	case w.outbound <- frame:
		return true
	default:
		return false
	}
}

func (w *realtimeWorker) run(ctx context.Context) {
	for {
		connection, _, err := websocket.DefaultDialer.DialContext(ctx, w.endpoint, nil)
		if err != nil {
			w.ready.Store(false)
			if !wait(ctx, sconstants.RealtimeWorkerReconnectDelay) {
				return
			}

			continue
		}
		w.connectionMutex.Lock()
		w.connection = connection
		w.connectionMutex.Unlock()

		w.ready.Store(true)
		if !w.replayActiveChannels(connection) {
			w.ready.Store(false)
			connection.Close()
			w.connectionMutex.Lock()
			w.connection = nil
			w.connectionMutex.Unlock()
			if !wait(ctx, sconstants.RealtimeWorkerReconnectDelay) {
				return
			}

			continue
		}

		readError := make(chan struct{})
		go w.read(connection, readError)

		connected := true
		for connected {
			select {
			case <-ctx.Done():
				connected = false
			case <-readError:
				connected = false
			case frame := <-w.outbound:
				payload, err := frame.MarshalBytes()
				if err != nil || connection.SetWriteDeadline(time.Now().Add(sconstants.RealtimeControlWriteTimeout)) != nil ||
					connection.WriteMessage(websocket.BinaryMessage, payload) != nil {
					connected = false
				}
			}
		}

		w.ready.Store(false)
		connection.Close()
		w.connectionMutex.Lock()
		w.connection = nil
		w.connectionMutex.Unlock()
		if !wait(ctx, sconstants.RealtimeWorkerReconnectDelay) {
			return
		}
	}
}

func (w *realtimeWorker) close() {
	w.connectionMutex.Lock()
	defer w.connectionMutex.Unlock()

	if w.connection != nil {
		_ = w.connection.Close()
	}
}

func wait(ctx context.Context, duration time.Duration) bool {
	timer := time.NewTimer(duration)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

func (w *realtimeWorker) replayActiveChannels(connection *websocket.Conn) bool {
	w.activeChannelsMutex.RLock()
	frames := make([]realtimetypes.InternalFrame, 0, len(w.activeChannels))
	for _, frame := range w.activeChannels {
		frames = append(frames, frame)
	}
	w.activeChannelsMutex.RUnlock()

	for _, frame := range frames {
		payload, err := frame.MarshalBytes()
		if err != nil || connection.SetWriteDeadline(time.Now().Add(sconstants.RealtimeControlWriteTimeout)) != nil ||
			connection.WriteMessage(websocket.BinaryMessage, payload) != nil {
			return false
		}
	}

	return true
}

func (w *realtimeWorker) read(connection *websocket.Conn, readError chan<- struct{}) {
	defer close(readError)

	for {
		messageType, payload, err := connection.ReadMessage()
		if err != nil || messageType != websocket.BinaryMessage {
			return
		}

		var frame realtimetypes.InternalFrame
		if err := frame.UnmarshalBytes(payload); err != nil || int(frame.Version) != sconstants.RealtimeWorkerProtocolVersion {
			return
		}

		w.frameHandlerMutex.RLock()
		frameHandler := w.frameHandler
		w.frameHandlerMutex.RUnlock()

		if frameHandler != nil {
			frameHandler(frame)
		}
	}
}
