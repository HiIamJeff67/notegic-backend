package workers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	cyjsworker "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1"
	cyjsworkerevents "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1/events"
	coreconfig "github.com/HiIamJeff67/notegic-backend/internal/core/configs"

	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	metrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"

	yjsworkerproducers "github.com/HiIamJeff67/notegic-backend/internal/core/transports/yjsworker/producers"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
)

type YjsMaintenanceWorker struct {
	db              *gorm.DB
	commandProducer *yjsworkerproducers.YjsMaintenanceCommandProducer
	strategy        *yjsMaintenanceStrategy
	kafkaConfig     platformkafka.ConsumerConfig
	slots           chan struct{}
}

type yjsMaintenanceRequest struct {
	hint    coreevents.YjsMaintenanceHintData
	attempt int
}

type yjsMaintenanceStrategy struct {
	config   coreconfig.YjsMaintenanceStrategyConfig
	mutex    sync.Mutex
	pending  map[uuid.UUID]coreevents.YjsMaintenanceHintData
	requests map[uuid.UUID]yjsMaintenanceRequest
	attempts map[uuid.UUID]int
	inFlight map[uuid.UUID]struct{}
	notify   chan struct{}
}

func NewYjsMaintenanceWorker(
	db *gorm.DB,
	producer *yjsworkerproducers.YjsMaintenanceCommandProducer,
	config coreconfig.YjsMaintenanceStrategyConfig,
	kafkaConfig platformkafka.ConsumerConfig,
) *YjsMaintenanceWorker {
	return &YjsMaintenanceWorker{
		db:              db,
		commandProducer: producer,
		strategy:        newYjsMaintenanceStrategy(config),
		kafkaConfig:     kafkaConfig,
		slots:           make(chan struct{}, config.MaximumDispatchWorkers),
	}
}

func newYjsMaintenanceStrategy(config coreconfig.YjsMaintenanceStrategyConfig) *yjsMaintenanceStrategy {
	return &yjsMaintenanceStrategy{
		config:   config,
		pending:  make(map[uuid.UUID]coreevents.YjsMaintenanceHintData),
		requests: make(map[uuid.UUID]yjsMaintenanceRequest),
		attempts: make(map[uuid.UUID]int),
		inFlight: make(map[uuid.UUID]struct{}),
		notify:   make(chan struct{}, 1),
	}
}

func (w *YjsMaintenanceWorker) Start(ctx context.Context) func() {
	hintConsumer, err := platformkafka.NewConsumer(w.kafkaConfig, coreevents.CoreYjsMaintenanceHintTopic.String())
	if err != nil {
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(ctx, err, "failed to create Core Yjs maintenance hint consumer")
		}
		return func() {}
	}
	resultConsumer, err := platformkafka.NewConsumer(w.kafkaConfig, cyjsworkerevents.CoreYjsWorkerMaintenanceResultTopic.String())
	if err != nil {
		hintConsumer.Close()
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(ctx, err, "failed to create Core Yjs maintenance result consumer")
		}
		return func() {}
	}

	workerCtx, cancel := context.WithCancel(ctx)
	var waitGroup sync.WaitGroup
	waitGroup.Add(3)
	go func() {
		defer waitGroup.Done()
		if err := hintConsumer.Run(workerCtx, w.consumeHint); err != nil && workerCtx.Err() == nil && logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(workerCtx, err, "Core Yjs maintenance hint consumer stopped")
		}
	}()
	go func() {
		defer waitGroup.Done()
		if err := resultConsumer.Run(workerCtx, w.consumeResult); err != nil && workerCtx.Err() == nil && logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(workerCtx, err, "Core Yjs maintenance result consumer stopped")
		}
	}()
	go func() {
		defer waitGroup.Done()
		w.dispatch(workerCtx)
	}()

	return func() {
		cancel()
		hintConsumer.Close()
		resultConsumer.Close()
		waitGroup.Wait()
	}
}

func (w *YjsMaintenanceWorker) consumeHint(
	ctx context.Context,
	_ platformkafka.ConsumerRecord,
	event cevent.EventEnvelope[json.RawMessage],
) error {
	if event.EventType != coreevents.EventType_YjsMaintenanceHint ||
		event.AggregateType != coreevents.AggregateType_BlockPack ||
		event.AggregateId == uuid.Nil || event.KafkaKey != event.AggregateId.String() {
		return &platformkafka.ConsumerError{Classification: platformkafka.ErrorClassification_SchemaIncompatible, Origin: errors.New("invalid Yjs maintenance hint envelope")}
	}
	var hint coreevents.YjsMaintenanceHintData
	if err := json.Unmarshal(event.Data, &hint); err != nil {
		return &platformkafka.ConsumerError{Classification: platformkafka.ErrorClassification_SchemaIncompatible, Origin: fmt.Errorf("decode Yjs maintenance hint: %w", err)}
	}
	if hint.BlockPackId != event.AggregateId || hint.DocumentId == uuid.Nil || hint.LatestUpdateSequence < 0 || hint.CompactedUntilSequence < 0 || hint.ProjectedUntilSequence < -1 {
		return &platformkafka.ConsumerError{Classification: platformkafka.ErrorClassification_SchemaIncompatible, Origin: errors.New("invalid Yjs maintenance hint data")}
	}
	if err := w.enqueue(hint); err != nil {
		return &platformkafka.ConsumerError{Classification: platformkafka.ErrorClassification_Transient, Origin: err}
	}
	if metrics.NotegicMeter != nil {
		metrics.NotegicMeter.Value(ctx, "yjs.maintenance.queue.size", int64(w.pendingCount()))
	}
	return nil
}

func (w *YjsMaintenanceWorker) consumeResult(
	ctx context.Context,
	_ platformkafka.ConsumerRecord,
	event cevent.EventEnvelope[json.RawMessage],
) error {
	if event.EventType != cyjsworkerevents.EventType_YjsMaintenanceCompleted || event.AggregateType != cyjsworkerevents.AggregateType_BlockPack || event.AggregateId == uuid.Nil || event.KafkaKey != event.AggregateId.String() {
		return &platformkafka.ConsumerError{Classification: platformkafka.ErrorClassification_SchemaIncompatible, Origin: errors.New("invalid Yjs maintenance result envelope")}
	}
	var result cyjsworkerevents.YjsMaintenanceResultData
	if err := json.Unmarshal(event.Data, &result); err != nil {
		return &platformkafka.ConsumerError{Classification: platformkafka.ErrorClassification_SchemaIncompatible, Origin: fmt.Errorf("decode Yjs maintenance result: %w", err)}
	}
	if result.RequestId == uuid.Nil || result.BlockPackId != event.AggregateId || result.DocumentId == uuid.Nil || result.TargetSequence < 0 || (result.Operation != cyjsworkerevents.YjsMaintenanceOperation_Compact && result.Operation != cyjsworkerevents.YjsMaintenanceOperation_Project) {
		return &platformkafka.ConsumerError{Classification: platformkafka.ErrorClassification_SchemaIncompatible, Origin: errors.New("invalid Yjs maintenance result data")}
	}
	if result.Success {
		request, exists := w.complete(result.RequestId)
		if exists && result.Operation == cyjsworkerevents.YjsMaintenanceOperation_Compact && result.CompactedUntilSequence > request.hint.CompactedUntilSequence {
			request.hint.CompactedUntilSequence = result.CompactedUntilSequence
			request.hint.UncompactedUpdateCount = request.hint.LatestUpdateSequence - result.CompactedUntilSequence
			compactedAt := time.Now().UTC()
			request.hint.LastCompactedAt = &compactedAt
			if result.ProjectedUntilSequence > request.hint.ProjectedUntilSequence {
				request.hint.ProjectedUntilSequence = result.ProjectedUntilSequence
			}
			if request.hint.ProjectedUntilSequence < request.hint.LatestUpdateSequence {
				w.retry(ctx, request.hint)
			}
		}
		if metrics.NotegicMeter != nil {
			metrics.NotegicMeter.Count(ctx, "yjs.maintenance.result.success", 1)
		}
		return nil
	}
	if request, exists := w.fail(result.RequestId); exists && request.attempt < w.strategy.config.MaximumRequestAttempts {
		w.retry(ctx, request.hint)
	}
	if logs.NotegicLogger != nil {
		logs.NotegicLogger.Error(ctx, errors.New(result.Error), "Core Yjs maintenance request failed")
	}
	return nil
}

func (w *YjsMaintenanceWorker) dispatch(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.strategy.notify:
			w.dispatchPending(ctx)
		}
	}
}

func (w *YjsMaintenanceWorker) dispatchPending(ctx context.Context) {
	for {
		hints := w.dequeueBatch(w.strategy.config.MaximumDispatchBatch)
		if len(hints) == 0 {
			return
		}
		var waitGroup sync.WaitGroup
		for _, hint := range hints {
			w.slots <- struct{}{}
			waitGroup.Add(1)
			go func(hint coreevents.YjsMaintenanceHintData) {
				defer waitGroup.Done()
				defer func() { <-w.slots }()
				if err := w.dispatchHint(ctx, hint); err != nil {
					if logs.NotegicLogger != nil {
						logs.NotegicLogger.Error(ctx, err, "failed to dispatch Core Yjs maintenance request")
					}
					w.retry(ctx, hint)
				}
			}(hint)
		}
		waitGroup.Wait()
	}
}

func (w *YjsMaintenanceWorker) dispatchHint(ctx context.Context, hint coreevents.YjsMaintenanceHintData) error {
	operation := cyjsworkerevents.YjsMaintenanceOperation_Project
	if hint.UncompactedUpdateCount >= cyjsworker.YjsCompactionUpdateThreshold || (hint.CompactedUntilSequence < hint.LatestUpdateSequence && hint.LastCompactedAt == nil) {
		operation = cyjsworkerevents.YjsMaintenanceOperation_Compact
	}
	if operation == cyjsworkerevents.YjsMaintenanceOperation_Project && hint.ProjectedUntilSequence >= hint.LatestUpdateSequence {
		return nil
	}

	var document schemas.BlockPackYjsDocument
	result := w.db.WithContext(ctx).Select("id, block_pack_id, last_update_sequence, compacted_until_sequence, projected_until_sequence").Where("block_pack_id = ? AND deleted_at IS NULL", hint.BlockPackId).First(&document)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil
	}
	if result.Error != nil {
		return fmt.Errorf("load Yjs maintenance state: %w", result.Error)
	}
	targetSequence := hint.LatestUpdateSequence
	if targetSequence > document.LastUpdateSequence {
		targetSequence = document.LastUpdateSequence
	}
	if (operation == cyjsworkerevents.YjsMaintenanceOperation_Compact && targetSequence <= document.CompactedUntilSequence) || (operation == cyjsworkerevents.YjsMaintenanceOperation_Project && targetSequence <= document.ProjectedUntilSequence) {
		return nil
	}

	requestId := uuid.New()
	w.track(requestId, hint)
	request := cyjsworkerevents.YjsMaintenanceCommandData{
		RequestId:      requestId,
		BlockPackId:    hint.BlockPackId,
		DocumentId:     document.Id,
		Operation:      operation,
		TargetSequence: targetSequence,
		CorrelationId:  uuid.New().String(),
	}
	source := cevent.EventEnvelope[json.RawMessage]{EventId: uuid.New(), Trace: cevent.TraceMetadata{}}
	if err := w.commandProducer.Produce(ctx, source, request); err != nil {
		w.complete(requestId)
		return err
	}
	return nil
}

func (w *YjsMaintenanceWorker) retry(ctx context.Context, hint coreevents.YjsMaintenanceHintData) {
	go func() {
		timer := time.NewTimer(time.Second)
		defer timer.Stop()
		select {
		case <-ctx.Done():
		case <-timer.C:
			if err := w.enqueue(hint); err != nil && logs.NotegicLogger != nil {
				logs.NotegicLogger.Error(ctx, err, "failed to requeue Core Yjs maintenance hint")
			}
		}
	}()
}

func (w *YjsMaintenanceWorker) enqueue(hint coreevents.YjsMaintenanceHintData) error {
	w.strategy.mutex.Lock()
	defer w.strategy.mutex.Unlock()
	if _, exists := w.strategy.pending[hint.BlockPackId]; !exists && len(w.strategy.pending) >= w.strategy.config.MaximumPendingHints {
		return errors.New("Core Yjs maintenance queue is full")
	}
	if current, exists := w.strategy.pending[hint.BlockPackId]; !exists || hint.LatestUpdateSequence >= current.LatestUpdateSequence {
		w.strategy.pending[hint.BlockPackId] = hint
	}
	select {
	case w.strategy.notify <- struct{}{}:
	default:
	}
	return nil
}

func (w *YjsMaintenanceWorker) dequeueBatch(limit int) []coreevents.YjsMaintenanceHintData {
	w.strategy.mutex.Lock()
	defer w.strategy.mutex.Unlock()
	result := make([]coreevents.YjsMaintenanceHintData, 0, limit)
	for len(result) < limit {
		var selectedId uuid.UUID
		var selected coreevents.YjsMaintenanceHintData
		selectedScore := int64(-1)
		for id, hint := range w.strategy.pending {
			if _, exists := w.strategy.inFlight[id]; exists {
				continue
			}
			score := hint.UncompactedUpdateCount*4 + (hint.LatestUpdateSequence-hint.ProjectedUntilSequence)*3
			if hint.LastCompactedAt == nil {
				score += 100_000
			} else if age := time.Since(*hint.LastCompactedAt); age > 0 {
				score += int64(age / time.Minute)
			}
			score += int64((hint.SnapshotBytes + hint.StateVectorBytes) / 1024)
			if score > selectedScore {
				selectedId, selected, selectedScore = id, hint, score
			}
		}
		if selectedId == uuid.Nil {
			break
		}
		delete(w.strategy.pending, selectedId)
		result = append(result, selected)
	}
	return result
}

func (w *YjsMaintenanceWorker) track(requestId uuid.UUID, hint coreevents.YjsMaintenanceHintData) {
	w.strategy.mutex.Lock()
	defer w.strategy.mutex.Unlock()
	attempt := w.strategy.attempts[hint.BlockPackId] + 1
	w.strategy.attempts[hint.BlockPackId] = attempt
	w.strategy.requests[requestId] = yjsMaintenanceRequest{hint: hint, attempt: attempt}
	w.strategy.inFlight[hint.BlockPackId] = struct{}{}
}

func (w *YjsMaintenanceWorker) complete(requestId uuid.UUID) (yjsMaintenanceRequest, bool) {
	w.strategy.mutex.Lock()
	defer w.strategy.mutex.Unlock()
	request, exists := w.strategy.requests[requestId]
	if exists {
		delete(w.strategy.attempts, request.hint.BlockPackId)
		delete(w.strategy.inFlight, request.hint.BlockPackId)
	}
	delete(w.strategy.requests, requestId)
	return request, exists
}

func (w *YjsMaintenanceWorker) fail(requestId uuid.UUID) (yjsMaintenanceRequest, bool) {
	w.strategy.mutex.Lock()
	defer w.strategy.mutex.Unlock()
	request, exists := w.strategy.requests[requestId]
	if exists {
		delete(w.strategy.requests, requestId)
		delete(w.strategy.inFlight, request.hint.BlockPackId)
		if request.attempt >= w.strategy.config.MaximumRequestAttempts {
			delete(w.strategy.attempts, request.hint.BlockPackId)
		}
	}
	return request, exists
}

func (w *YjsMaintenanceWorker) pendingCount() int {
	w.strategy.mutex.Lock()
	defer w.strategy.mutex.Unlock()
	return len(w.strategy.pending)
}
