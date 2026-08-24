package blocks

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	yjsworkercontract "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1"

	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	metrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"
	traces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"

	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
)

type YjsPersistenceServiceInterface interface {
	LoadDocument(ctx context.Context, blockPackId uuid.UUID) (*yjsworkercontract.YjsDocumentState, error)
	AppendUpdate(ctx context.Context, blockPackId uuid.UUID, persistenceBatchId uuid.UUID, originConnectionId *uuid.UUID, payload []byte) (int64, error)
	GetCompactableYjsDocumentWithUpdates(ctx context.Context, blockPackId uuid.UUID) (*yjsworkercontract.YjsCompactionInput, error)
	ApplyCompactedYjsDocument(ctx context.Context, blockPackId uuid.UUID, result yjsworkercontract.YjsCompactionResult) (bool, error)
}

type YjsPersistenceService struct {
	db                     *gorm.DB
	blockPackYjsRepository repositories.BlockPackYjsRepositoryInterface
}

func NewYjsPersistenceService(db *gorm.DB) YjsPersistenceServiceInterface {
	return &YjsPersistenceService{
		db:                     db,
		blockPackYjsRepository: repositories.NewBlockPackYjsRepository(),
	}
}

func (s *YjsPersistenceService) LoadDocument(
	ctx context.Context, blockPackId uuid.UUID,
) (state *yjsworkercontract.YjsDocumentState, err error) {
	start := time.Now()
	ctx, span := traces.NotegicTracer.Start(ctx, "yjs.document.load")
	defer func() { traces.NotegicTracer.End(span, err) }()

	db := s.db.WithContext(ctx)

	document, updates, err := s.blockPackYjsRepository.LoadDocumentAndUpdates(
		blockPackId,
		options.WithDB(db),
	)
	if err != nil {
		metrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
			attribute.String("operation", "document.load"),
			attribute.String("outcome", "error"),
		)
		metrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
			attribute.String("operation", "document.load"),
			attribute.String("outcome", "error"),
		)
		logs.NotegicLogger.Error(ctx, err, "failed to load Yjs document", attribute.String("operation", "document.load"))

		return nil, err
	}

	state = &yjsworkercontract.YjsDocumentState{
		Snapshot:               document.Snapshot,
		StateVector:            document.StateVector,
		LastUpdateSequence:     document.LastUpdateSequence,
		CompactedUntilSequence: document.CompactedUntilSequence,
		ProjectedUntilSequence: document.ProjectedUntilSequence,
		Updates:                make([]yjsworkercontract.YjsDocumentUpdate, len(updates)),
	}
	for index, update := range updates {
		state.Updates[index] = yjsworkercontract.YjsDocumentUpdate{
			UpdateSequence: update.UpdateSequence,
			Payload:        update.Payload,
		}
	}

	span.SetAttributes(attribute.Int("yjs.update_count", len(updates)))
	metrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
		attribute.String("operation", "document.load"),
		attribute.String("outcome", "success"),
	)
	metrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
		attribute.String("operation", "document.load"),
		attribute.String("outcome", "success"),
	)

	return state, nil
}

func (s *YjsPersistenceService) AppendUpdate(
	ctx context.Context,
	blockPackId uuid.UUID,
	persistenceBatchId uuid.UUID,
	originConnectionId *uuid.UUID,
	payload []byte,
) (updateSequence int64, err error) {
	start := time.Now()
	ctx, span := traces.NotegicTracer.Start(ctx, "yjs.update.append")
	defer func() { traces.NotegicTracer.End(span, err) }()

	db := s.db.WithContext(ctx)

	updateSequence, err = s.blockPackYjsRepository.AppendUpdate(
		blockPackId,
		inputs.AppendBlockPackYjsUpdateInput{
			PersistenceBatchId: persistenceBatchId,
			OriginConnectionId: originConnectionId,
			Payload:            payload,
		},
		options.WithDB(db),
	)
	if err != nil {
		metrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
			attribute.String("operation", "update.append"),
			attribute.String("outcome", "error"),
		)
		metrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
			attribute.String("operation", "update.append"),
			attribute.String("outcome", "error"),
		)
		metrics.NotegicMeter.Bytes(ctx, "yjs.payload.bytes", int64(len(payload)), attribute.String("operation", "update.append"))
		logs.NotegicLogger.Error(ctx, err, "failed to append Yjs update", attribute.String("operation", "update.append"))

		return 0, err
	}

	span.SetAttributes(attribute.Int("yjs.payload_bytes", len(payload)))
	metrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
		attribute.String("operation", "update.append"),
		attribute.String("outcome", "success"),
	)
	metrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
		attribute.String("operation", "update.append"),
		attribute.String("outcome", "success"),
	)
	metrics.NotegicMeter.Bytes(ctx, "yjs.payload.bytes", int64(len(payload)), attribute.String("operation", "update.append"))

	return updateSequence, nil
}

func (s *YjsPersistenceService) GetCompactableYjsDocumentWithUpdates(
	ctx context.Context, blockPackId uuid.UUID,
) (input *yjsworkercontract.YjsCompactionInput, err error) {
	start := time.Now()
	ctx, span := traces.NotegicTracer.Start(ctx, "yjs.compaction.load")
	defer func() { traces.NotegicTracer.End(span, err) }()

	db := s.db.WithContext(ctx)

	document, updates, err := s.blockPackYjsRepository.GetCompactableYjsDocumentWithUpdates(
		blockPackId,
		options.WithDB(db),
	)
	if err != nil || document == nil {
		if err != nil {
			logs.NotegicLogger.Error(ctx, err, "failed to load compactable Yjs document", attribute.String("operation", "compaction.load"))
		}
		metrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
			attribute.String("operation", "compaction.load"),
			attribute.String("outcome", "error"),
		)
		metrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
			attribute.String("operation", "compaction.load"),
			attribute.String("outcome", "error"),
		)

		return nil, err
	}

	input = &yjsworkercontract.YjsCompactionInput{
		Snapshot:                   document.Snapshot,
		StateVector:                document.StateVector,
		BaseCompactedUntilSequence: document.CompactedUntilSequence,
		CutoffSequence:             document.LastUpdateSequence,
		Updates:                    make([]yjsworkercontract.YjsDocumentUpdate, len(updates)),
	}
	for index, update := range updates {
		input.Updates[index] = yjsworkercontract.YjsDocumentUpdate{
			UpdateSequence: update.UpdateSequence,
			Payload:        update.Payload,
		}
	}

	span.SetAttributes(attribute.Int("yjs.update_count", len(updates)))
	metrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
		attribute.String("operation", "compaction.load"),
		attribute.String("outcome", "success"),
	)
	metrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
		attribute.String("operation", "compaction.load"),
		attribute.String("outcome", "success"),
	)

	return input, nil
}

func (s *YjsPersistenceService) ApplyCompactedYjsDocument(
	ctx context.Context, blockPackId uuid.UUID, result yjsworkercontract.YjsCompactionResult,
) (applied bool, err error) {
	start := time.Now()
	ctx, span := traces.NotegicTracer.Start(ctx, "yjs.compaction.apply")
	defer func() { traces.NotegicTracer.End(span, err) }()

	db := s.db.WithContext(ctx)

	applied, err = s.blockPackYjsRepository.ApplyCompactedYjsDocument(
		blockPackId,
		inputs.ApplyCompactedBlockPackYjsDocumentInput{
			BaseCompactedUntilSequence: result.BaseCompactedUntilSequence,
			CutoffSequence:             result.CutoffSequence,
			Snapshot:                   result.Snapshot,
			StateVector:                result.StateVector,
		},
		options.WithDB(db),
	)
	if err != nil {
		metrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
			attribute.String("operation", "compaction.apply"),
			attribute.String("outcome", "error"),
		)
		metrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
			attribute.String("operation", "compaction.apply"),
			attribute.String("outcome", "error"),
		)
		metrics.NotegicMeter.Bytes(ctx, "yjs.payload.bytes", int64(len(result.Snapshot)), attribute.String("operation", "compaction.apply"))
		logs.NotegicLogger.Error(ctx, err, "failed to apply compacted Yjs document", attribute.String("operation", "compaction.apply"))

		return false, err
	}

	span.SetAttributes(attribute.Bool("yjs.applied", applied))
	metrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
		attribute.String("operation", "compaction.apply"),
		attribute.String("outcome", "success"),
	)
	metrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
		attribute.String("operation", "compaction.apply"),
		attribute.String("outcome", "success"),
	)
	metrics.NotegicMeter.Bytes(ctx, "yjs.payload.bytes", int64(len(result.Snapshot)), attribute.String("operation", "compaction.apply"))

	return applied, nil
}
