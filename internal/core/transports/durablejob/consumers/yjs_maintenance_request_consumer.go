package durablejobconsumers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	durablejobeventscontract "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/events"
	eventcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	yjsworkereventscontract "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1/events"

	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	durablejobproducers "github.com/HiIamJeff67/notegic-backend/internal/core/transports/durablejob/producers"
	yjsworkerproducers "github.com/HiIamJeff67/notegic-backend/internal/core/transports/yjsworker/producers"
)

type YjsMaintenanceRequestConsumer struct {
	db              *gorm.DB
	commandProducer *yjsworkerproducers.YjsMaintenanceCommandProducer
	resultProducer  *durablejobproducers.YjsMaintenanceResultProducer
	kafkaConfig     platformkafka.ConsumerConfig
}

func NewYjsMaintenanceRequestConsumer(
	db *gorm.DB,
	commandProducer *yjsworkerproducers.YjsMaintenanceCommandProducer,
	resultProducer *durablejobproducers.YjsMaintenanceResultProducer,
	kafkaConfig platformkafka.ConsumerConfig,
) *YjsMaintenanceRequestConsumer {
	return &YjsMaintenanceRequestConsumer{
		db:              db,
		commandProducer: commandProducer,
		resultProducer:  resultProducer,
		kafkaConfig:     kafkaConfig,
	}
}

func (c *YjsMaintenanceRequestConsumer) Start(ctx context.Context) func() {
	consumer, err := platformkafka.NewConsumer(
		c.kafkaConfig,
		durablejobeventscontract.DurableJobCoreYjsMaintenanceRequestTopic.String(),
	)
	if err != nil {
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(ctx, err, "failed to create Yjs maintenance request consumer")
		}

		return func() {}
	}

	workerCtx, cancel := context.WithCancel(ctx)
	go func() {
		if err := consumer.Run(workerCtx, c.consume); err != nil && workerCtx.Err() == nil && logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(workerCtx, err, "Yjs maintenance request consumer stopped")
		}
	}()

	return func() {
		cancel()
		consumer.Close()
	}
}

func (c *YjsMaintenanceRequestConsumer) consume(
	ctx context.Context,
	_ platformkafka.ConsumerRecord,
	event eventcontract.EventEnvelope[json.RawMessage],
) error {
	if event.EventType != durablejobeventscontract.EventType_YjsMaintenanceRequested ||
		event.AggregateType != durablejobeventscontract.AggregateType_BlockPack ||
		event.AggregateId == uuid.Nil ||
		event.KafkaKey != event.AggregateId.String() {
		return &platformkafka.ConsumerError{
			Classification: platformkafka.ErrorClassification_SchemaIncompatible,
			Origin:         errors.New("invalid Yjs maintenance request envelope"),
		}
	}

	var request durablejobeventscontract.YjsMaintenanceRequestData
	if err := json.Unmarshal(event.Data, &request); err != nil {
		return &platformkafka.ConsumerError{
			Classification: platformkafka.ErrorClassification_SchemaIncompatible,
			Origin:         fmt.Errorf("decode Yjs maintenance request: %w", err),
		}
	}
	if request.RequestId == uuid.Nil || request.BlockPackId != event.AggregateId || request.DocumentId == uuid.Nil ||
		(request.Operation != yjsworkereventscontract.YjsMaintenanceOperation_Compact && request.Operation != yjsworkereventscontract.YjsMaintenanceOperation_Project) ||
		request.TargetSequence < 0 {
		return &platformkafka.ConsumerError{
			Classification: platformkafka.ErrorClassification_SchemaIncompatible,
			Origin:         errors.New("invalid Yjs maintenance request data"),
		}
	}

	var document schemas.BlockPackYjsDocument
	result := c.db.WithContext(ctx).
		Select("id, block_pack_id, last_update_sequence, compacted_until_sequence, projected_until_sequence").
		Where("block_pack_id = ? AND deleted_at IS NULL", request.BlockPackId).
		First(&document)
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		if err := c.resultProducer.Produce(ctx, event, yjsworkereventscontract.YjsMaintenanceResultData{
			RequestId:              request.RequestId,
			BlockPackId:            request.BlockPackId,
			DocumentId:             request.DocumentId,
			Operation:              request.Operation,
			TargetSequence:         request.TargetSequence,
			Success:                true,
			CompactedUntilSequence: 0,
			ProjectedUntilSequence: 0,
		}); err != nil {
			return fmt.Errorf("produce Yjs maintenance no-op result: %w", err)
		}

		return nil
	}
	if result.Error != nil {
		return &platformkafka.ConsumerError{
			Classification: platformkafka.ErrorClassification_Transient,
			Origin:         fmt.Errorf("load Yjs maintenance state: %w", result.Error),
		}
	}
	if document.Id != request.DocumentId {
		return &platformkafka.ConsumerError{
			Classification: platformkafka.ErrorClassification_SchemaIncompatible,
			Origin:         errors.New("Yjs maintenance request document does not match Core state"),
		}
	}

	targetSequence := request.TargetSequence
	if targetSequence > document.LastUpdateSequence {
		targetSequence = document.LastUpdateSequence
	}
	if (request.Operation == yjsworkereventscontract.YjsMaintenanceOperation_Compact &&
		targetSequence <= document.CompactedUntilSequence) ||
		(request.Operation == yjsworkereventscontract.YjsMaintenanceOperation_Project &&
			targetSequence <= document.ProjectedUntilSequence) {
		if err := c.resultProducer.Produce(ctx, event, yjsworkereventscontract.YjsMaintenanceResultData{
			RequestId:              request.RequestId,
			BlockPackId:            request.BlockPackId,
			DocumentId:             document.Id,
			Operation:              request.Operation,
			TargetSequence:         request.TargetSequence,
			Success:                true,
			CompactedUntilSequence: document.CompactedUntilSequence,
			ProjectedUntilSequence: document.ProjectedUntilSequence,
		}); err != nil {
			return fmt.Errorf("produce Yjs maintenance no-op result: %w", err)
		}

		return nil
	}

	if err := c.commandProducer.Produce(ctx, event, request, document.Id, targetSequence); err != nil {
		return fmt.Errorf("produce Yjs maintenance command: %w", err)
	}

	return nil
}
