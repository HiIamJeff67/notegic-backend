package adaptersconsumers

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	apicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/blocks"
	eventcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	crepositories "github.com/HiIamJeff67/notegic-backend/contracts/types/models/repositories"
	yjsworkercontract "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1"
	yjsworkereventscontract "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1/events"

	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	blockservices "github.com/HiIamJeff67/notegic-backend/internal/core/services/blocks"
)

type YjsCommandConsumer struct {
	db                     *gorm.DB
	yjsPersistenceService  blockservices.YjsPersistenceServiceInterface
	blockService           blockservices.BlockServiceInterface
	blockPackYjsRepository repositories.BlockPackYjsRepositoryInterface
	kafkaConfig            platformkafka.ConsumerConfig
}

func NewYjsCommandConsumer(
	db *gorm.DB,
	yjsPersistenceService blockservices.YjsPersistenceServiceInterface,
	blockService blockservices.BlockServiceInterface,
	kafkaConfig platformkafka.ConsumerConfig,
) *YjsCommandConsumer {
	return &YjsCommandConsumer{
		db:                     db,
		yjsPersistenceService:  yjsPersistenceService,
		blockService:           blockService,
		blockPackYjsRepository: repositories.NewBlockPackYjsRepository(),
		kafkaConfig:            kafkaConfig,
	}
}

func (c *YjsCommandConsumer) Start(ctx context.Context) func() {
	consumer, err := platformkafka.NewConsumer(c.kafkaConfig, yjsworkereventscontract.YjsWorkerCoreCommandTopic.String())
	if err != nil {
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(ctx, err, "Failed to create YjsWorker command consumer")
		}

		return func() {}
	}

	workerCtx, cancel := context.WithCancel(ctx)
	go func() {
		if err := consumer.Run(workerCtx, c.consume); err != nil && workerCtx.Err() == nil && logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(workerCtx, err, "YjsWorker command consumer stopped")
		}
	}()

	return func() {
		cancel()
		consumer.Close()
	}
}

func (c *YjsCommandConsumer) consume(
	ctx context.Context,
	_ platformkafka.ConsumerRecord,
	event eventcontract.EventEnvelope[json.RawMessage],
) error {
	var command yjsworkercontract.CommandEnvelope[json.RawMessage]
	if err := json.Unmarshal(event.Data, &command); err != nil {
		return &platformkafka.ConsumerError{
			Classification: platformkafka.ErrorClassification_SchemaIncompatible,
			Origin:         fmt.Errorf("decode YjsWorker command: %w", err),
		}
	}
	if command.SchemaVersion != yjsworkercontract.Version || command.CommandId == uuid.Nil ||
		command.BlockPackId == uuid.Nil || command.CommandType == "" || command.Producer != "yjs-worker" {
		return c.writeReply(ctx, command, nil, &yjsworkercontract.Error{
			Code:      "InvalidCommand",
			Message:   "the YjsWorker command envelope is invalid",
			Retryable: false,
		})
	}
	if command.BlockPackId != event.AggregateId || command.BlockPackId.String() != event.KafkaKey {
		return c.writeReply(ctx, command, nil, &yjsworkercontract.Error{
			Code:      "InvalidCommand",
			Message:   "the YjsWorker command partition key is invalid",
			Retryable: false,
		})
	}

	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin YjsWorker command transaction: %w", tx.Error)
	}

	data, exception, err := c.execute(ctx, tx, command)
	if err != nil {
		tx.Rollback()

		return err
	}
	if err := c.enqueueReply(tx, command, data, exception); err != nil {
		tx.Rollback()

		return err
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit YjsWorker command transaction: %w", err)
	}

	return nil
}

func (c *YjsCommandConsumer) execute(
	ctx context.Context,
	tx *gorm.DB,
	command yjsworkercontract.CommandEnvelope[json.RawMessage],
) (json.RawMessage, *yjsworkercontract.Error, error) {
	switch command.CommandType {
	case yjsworkercontract.CommandType_LoadYjsDocument:
		state, err := c.yjsPersistenceService.LoadDocument(
			ctx,
			command.BlockPackId,
		)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return marshalYjsReplyData(yjsworkercontract.LoadYjsDocumentReplyDto{
				Found: false,
			})
		}
		if err != nil {
			return nil, nil, fmt.Errorf("load Yjs document: %w", err)
		}
		payload, err := state.MarshalBytes()
		if err != nil {
			return nil, &yjsworkercontract.Error{
				Code:      "InvalidDocument",
				Message:   "the persisted Yjs document is invalid",
				Retryable: false,
			}, nil
		}

		return marshalYjsReplyData(yjsworkercontract.LoadYjsDocumentReplyDto{
			Found:   true,
			Payload: payload,
		})
	case yjsworkercontract.CommandType_AppendYjsUpdate:
		var data yjsworkercontract.AppendYjsUpdateCommandDto
		if err := json.Unmarshal(command.Data, &data); err != nil || data.PersistenceBatchId == uuid.Nil || len(data.Payload) == 0 {
			return nil, &yjsworkercontract.Error{
				Code:      "InvalidCommand",
				Message:   "the Yjs update command is invalid",
				Retryable: false,
			}, nil
		}
		updateSequence, err := c.blockPackYjsRepository.AppendUpdate(
			command.BlockPackId,
			inputs.AppendBlockPackYjsUpdateInput{
				PersistenceBatchId: data.PersistenceBatchId,
				OriginConnectionId: data.OriginConnectionId,
				Payload:            data.Payload,
			},
			options.WithTransactionDB(tx),
		)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &yjsworkercontract.Error{
				Code:      "NotFound",
				Message:   "the Yjs document was not found",
				Retryable: false,
			}, nil
		}
		if err != nil {
			return nil, nil, fmt.Errorf("append Yjs update: %w", err)
		}
		if err := repositories.NewOutboxEventRepository().EnqueueYjsMaintenanceHint(
			tx,
			command.CorrelationId,
			command.BlockPackId,
			"yjs_update_persisted",
		); err != nil {
			return nil, nil, fmt.Errorf("enqueue Yjs maintenance hint: %w", err)
		}

		return marshalYjsReplyData(yjsworkercontract.AppendYjsUpdateReplyDto{
			UpdateSequence: updateSequence,
		})
	case yjsworkercontract.CommandType_LoadCompactableYjsDocument:
		input, err := c.yjsPersistenceService.GetCompactableYjsDocumentWithUpdates(
			ctx,
			command.BlockPackId,
		)
		if errors.Is(err, gorm.ErrRecordNotFound) || input == nil {
			return marshalYjsReplyData(yjsworkercontract.LoadCompactableYjsDocumentReplyDto{
				Found: false,
			})
		}
		if err != nil {
			return nil, nil, fmt.Errorf("load compactable Yjs document: %w", err)
		}
		payload, err := input.MarshalBytes()
		if err != nil {
			return nil, &yjsworkercontract.Error{
				Code:      "InvalidDocument",
				Message:   "the compactable Yjs document is invalid",
				Retryable: false,
			}, nil
		}

		return marshalYjsReplyData(yjsworkercontract.LoadCompactableYjsDocumentReplyDto{
			Found:   true,
			Payload: payload,
		})
	case yjsworkercontract.CommandType_ApplyCompactedYjsDocument:
		var data yjsworkercontract.ApplyCompactedYjsDocumentCommandDto
		var result yjsworkercontract.YjsCompactionResult
		if err := json.Unmarshal(command.Data, &data); err != nil || result.UnmarshalBytes(data.Payload) != nil {
			return nil, &yjsworkercontract.Error{
				Code:      "InvalidCommand",
				Message:   "the compacted Yjs document command is invalid",
				Retryable: false,
			}, nil
		}
		applied, err := c.blockPackYjsRepository.ApplyCompactedYjsDocument(
			command.BlockPackId,
			inputs.ApplyCompactedBlockPackYjsDocumentInput{
				BaseCompactedUntilSequence: result.BaseCompactedUntilSequence,
				CutoffSequence:             result.CutoffSequence,
				Snapshot:                   result.Snapshot,
				StateVector:                result.StateVector,
			},
			options.WithTransactionDB(tx),
		)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &yjsworkercontract.Error{
				Code:      "NotFound",
				Message:   "the Yjs document was not found",
				Retryable: false,
			}, nil
		}
		if err != nil {
			return nil, nil, fmt.Errorf("apply compacted Yjs document: %w", err)
		}

		return marshalYjsReplyData(yjsworkercontract.ApplyCompactedYjsDocumentReplyDto{
			Applied: applied,
		})
	case yjsworkercontract.CommandType_ApplyBlockProjection:
		var data yjsworkercontract.ApplyBlockProjectionCommandDto
		var requestDto apicontract.ApplyBlockProjectionRequestDto
		if err := json.Unmarshal(command.Data, &data); err != nil || json.Unmarshal(data.Projection, &requestDto) != nil {
			return nil, &yjsworkercontract.Error{
				Code:      "InvalidCommand",
				Message:   "the block projection command is invalid",
				Retryable: false,
			}, nil
		}
		responseDto, err := c.blockService.ApplyWithTransaction(
			ctx,
			tx,
			command.BlockPackId,
			requestDto,
		)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, &yjsworkercontract.Error{
				Code:      "NotFound",
				Message:   "the block pack was not found",
				Retryable: false,
			}, nil
		}
		if err != nil {
			return nil, nil, fmt.Errorf("apply block projection: %w", err)
		}

		return marshalYjsReplyData(yjsworkercontract.ApplyBlockProjectionReplyDto{
			Applied:                responseDto.Applied,
			ProjectedUntilSequence: responseDto.ProjectedUntilSequence,
		})
	default:
		return nil, &yjsworkercontract.Error{
			Code:      "UnsupportedCommand",
			Message:   "the YjsWorker command type is unsupported",
			Retryable: false,
		}, nil
	}
}

func marshalYjsReplyData(data any) (json.RawMessage, *yjsworkercontract.Error, error) {
	payload, err := json.Marshal(data)
	if err != nil {
		return nil, nil, fmt.Errorf("marshal YjsWorker reply data: %w", err)
	}

	return payload, nil, nil
}

func (c *YjsCommandConsumer) writeReply(
	ctx context.Context,
	command yjsworkercontract.CommandEnvelope[json.RawMessage],
	data json.RawMessage,
	exception *yjsworkercontract.Error,
) error {
	tx := c.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return fmt.Errorf("begin invalid YjsWorker command transaction: %w", tx.Error)
	}
	if err := c.enqueueReply(tx, command, data, exception); err != nil {
		tx.Rollback()

		return err
	}
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("commit invalid YjsWorker command transaction: %w", err)
	}

	return nil
}

func (c *YjsCommandConsumer) enqueueReply(
	tx *gorm.DB,
	command yjsworkercontract.CommandEnvelope[json.RawMessage],
	data json.RawMessage,
	exception *yjsworkercontract.Error,
) error {
	if data == nil {
		data = json.RawMessage("{}")
	}

	reply := yjsworkercontract.ReplyEnvelope[json.RawMessage]{
		SchemaVersion: yjsworkercontract.Version,
		CommandId:     command.CommandId,
		CommandType:   command.CommandType,
		BlockPackId:   command.BlockPackId,
		CorrelationId: command.CorrelationId,
		CausationId:   &command.CommandId,
		Trace:         command.Trace,
		Producer:      "core",
		RespondedAt:   time.Now().UTC(),
		Data:          data,
		Error:         exception,
	}

	return crepositories.EnqueueOutboxEvents(
		tx,
		yjsworkereventscontract.CoreYjsWorkerReplyTopic,
		[]eventcontract.EventEnvelope[yjsworkercontract.ReplyEnvelope[json.RawMessage]]{
			{
				SchemaVersion: eventcontract.Version,
				EventId:       uuid.New(),
				EventType:     yjsworkereventscontract.EventType_YjsWorkerCommandCompleted,
				AggregateType: yjsworkereventscontract.AggregateType_BlockPack,
				AggregateId:   command.BlockPackId,
				KafkaKey:      command.BlockPackId.String(),
				OccurredAt:    time.Now().UTC(),
				CorrelationId: command.CorrelationId,
				CausationId:   &command.CommandId,
				Trace:         command.Trace,
				Data:          reply,
			},
		},
	)
}
