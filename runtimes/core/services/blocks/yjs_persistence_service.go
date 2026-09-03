package blocks

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"gorm.io/gorm"

	cyjsworker "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1"

	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	smetrics "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/metrics"
	straces "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/traces"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
)

type YjsPersistenceServiceInterface interface {
	LoadDocument(ctx context.Context, blockPackId uuid.UUID) (*cyjsworker.YjsDocumentState, error)
	AppendUpdate(ctx context.Context, blockPackId uuid.UUID, persistenceBatchId uuid.UUID, originConnectionId *uuid.UUID, payload []byte) (int64, error)
	GetCompactableYjsDocumentWithUpdates(ctx context.Context, blockPackId uuid.UUID) (*cyjsworker.YjsCompactionInput, error)
	ApplyCompactedYjsDocument(ctx context.Context, blockPackId uuid.UUID, result cyjsworker.YjsCompactionResult) (bool, error)
}

type YjsPersistenceService struct {
	db                     *gorm.DB
	blockPackYjsRepository srepositories.BlockPackYjsRepositoryInterface
}

func NewYjsPersistenceService(
	db *gorm.DB,
	blockPackYjsRepository srepositories.BlockPackYjsRepositoryInterface,
) YjsPersistenceServiceInterface {
	return &YjsPersistenceService{
		db:                     db,
		blockPackYjsRepository: blockPackYjsRepository,
	}
}

func (s *YjsPersistenceService) LoadDocument(
	ctx context.Context, blockPackId uuid.UUID,
) (state *cyjsworker.YjsDocumentState, err error) {
	start := time.Now()
	ctx, span := straces.NotegicTracer.Start(ctx, "yjs.document.load")
	defer func() { straces.NotegicTracer.End(span, err) }()

	db := s.db.WithContext(ctx)

	document, updates, err := s.blockPackYjsRepository.LoadDocumentAndUpdates(
		blockPackId,
		srepositories.WithDB(db),
	)
	if err != nil {
		smetrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
			attribute.String("operation", "document.load"),
			attribute.String("outcome", "error"),
		)
		smetrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
			attribute.String("operation", "document.load"),
			attribute.String("outcome", "error"),
		)
		slogs.NotegicLogger.Error(ctx, err, "failed to load Yjs document", attribute.String("operation", "document.load"))

		return nil, err
	}

	state = &cyjsworker.YjsDocumentState{
		Snapshot:               document.Snapshot,
		StateVector:            document.StateVector,
		LastUpdateSequence:     document.LastUpdateSequence,
		CompactedUntilSequence: document.CompactedUntilSequence,
		ProjectedUntilSequence: document.ProjectedUntilSequence,
		Updates:                make([]cyjsworker.YjsDocumentUpdate, len(updates)),
	}
	for index, update := range updates {
		state.Updates[index] = cyjsworker.YjsDocumentUpdate{
			UpdateSequence: update.UpdateSequence,
			Payload:        update.Payload,
		}
	}

	span.SetAttributes(attribute.Int("yjs.update_count", len(updates)))
	smetrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
		attribute.String("operation", "document.load"),
		attribute.String("outcome", "success"),
	)
	smetrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
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
	ctx, span := straces.NotegicTracer.Start(ctx, "yjs.update.append")
	defer func() { straces.NotegicTracer.End(span, err) }()

	db := s.db.WithContext(ctx)

	updateSequence, err = s.blockPackYjsRepository.AppendUpdate(
		blockPackId,
		sinputs.AppendBlockPackYjsUpdateInput{
			PersistenceBatchId: persistenceBatchId,
			OriginConnectionId: originConnectionId,
			Payload:            payload,
		},
		srepositories.WithDB(db),
	)
	if err != nil {
		smetrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
			attribute.String("operation", "update.append"),
			attribute.String("outcome", "error"),
		)
		smetrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
			attribute.String("operation", "update.append"),
			attribute.String("outcome", "error"),
		)
		smetrics.NotegicMeter.Bytes(ctx, "yjs.payload.bytes", int64(len(payload)), attribute.String("operation", "update.append"))
		slogs.NotegicLogger.Error(ctx, err, "failed to append Yjs update", attribute.String("operation", "update.append"))

		return 0, err
	}

	span.SetAttributes(attribute.Int("yjs.payload_bytes", len(payload)))
	smetrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
		attribute.String("operation", "update.append"),
		attribute.String("outcome", "success"),
	)
	smetrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
		attribute.String("operation", "update.append"),
		attribute.String("outcome", "success"),
	)
	smetrics.NotegicMeter.Bytes(ctx, "yjs.payload.bytes", int64(len(payload)), attribute.String("operation", "update.append"))

	return updateSequence, nil
}

func (s *YjsPersistenceService) GetCompactableYjsDocumentWithUpdates(
	ctx context.Context, blockPackId uuid.UUID,
) (input *cyjsworker.YjsCompactionInput, err error) {
	start := time.Now()
	ctx, span := straces.NotegicTracer.Start(ctx, "yjs.compaction.load")
	defer func() { straces.NotegicTracer.End(span, err) }()

	db := s.db.WithContext(ctx)

	document, updates, err := s.blockPackYjsRepository.GetCompactableYjsDocumentWithUpdates(
		blockPackId,
		srepositories.WithDB(db),
	)
	if err != nil || document == nil {
		if err != nil {
			slogs.NotegicLogger.Error(ctx, err, "failed to load compactable Yjs document", attribute.String("operation", "compaction.load"))
		}
		smetrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
			attribute.String("operation", "compaction.load"),
			attribute.String("outcome", "error"),
		)
		smetrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
			attribute.String("operation", "compaction.load"),
			attribute.String("outcome", "error"),
		)

		return nil, err
	}

	input = &cyjsworker.YjsCompactionInput{
		Snapshot:                   document.Snapshot,
		StateVector:                document.StateVector,
		BaseCompactedUntilSequence: document.CompactedUntilSequence,
		CutoffSequence:             document.LastUpdateSequence,
		Updates:                    make([]cyjsworker.YjsDocumentUpdate, len(updates)),
	}
	for index, update := range updates {
		input.Updates[index] = cyjsworker.YjsDocumentUpdate{
			UpdateSequence: update.UpdateSequence,
			Payload:        update.Payload,
		}
	}

	span.SetAttributes(attribute.Int("yjs.update_count", len(updates)))
	smetrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
		attribute.String("operation", "compaction.load"),
		attribute.String("outcome", "success"),
	)
	smetrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
		attribute.String("operation", "compaction.load"),
		attribute.String("outcome", "success"),
	)

	return input, nil
}

func (s *YjsPersistenceService) ApplyCompactedYjsDocument(
	ctx context.Context, blockPackId uuid.UUID, result cyjsworker.YjsCompactionResult,
) (applied bool, err error) {
	start := time.Now()
	ctx, span := straces.NotegicTracer.Start(ctx, "yjs.compaction.apply")
	defer func() { straces.NotegicTracer.End(span, err) }()

	db := s.db.WithContext(ctx)

	applied, err = s.blockPackYjsRepository.ApplyCompactedYjsDocument(
		blockPackId,
		sinputs.ApplyCompactedBlockPackYjsDocumentInput{
			BaseCompactedUntilSequence: result.BaseCompactedUntilSequence,
			CutoffSequence:             result.CutoffSequence,
			Snapshot:                   result.Snapshot,
			StateVector:                result.StateVector,
		},
		srepositories.WithDB(db),
	)
	if err != nil {
		smetrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
			attribute.String("operation", "compaction.apply"),
			attribute.String("outcome", "error"),
		)
		smetrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
			attribute.String("operation", "compaction.apply"),
			attribute.String("outcome", "error"),
		)
		smetrics.NotegicMeter.Bytes(ctx, "yjs.payload.bytes", int64(len(result.Snapshot)), attribute.String("operation", "compaction.apply"))
		slogs.NotegicLogger.Error(ctx, err, "failed to apply compacted Yjs document", attribute.String("operation", "compaction.apply"))

		return false, err
	}

	span.SetAttributes(attribute.Bool("yjs.applied", applied))
	smetrics.NotegicMeter.Count(ctx, "yjs.operation.count", 1,
		attribute.String("operation", "compaction.apply"),
		attribute.String("outcome", "success"),
	)
	smetrics.NotegicMeter.Duration(ctx, "yjs.operation.duration", time.Since(start),
		attribute.String("operation", "compaction.apply"),
		attribute.String("outcome", "success"),
	)
	smetrics.NotegicMeter.Bytes(ctx, "yjs.payload.bytes", int64(len(result.Snapshot)), attribute.String("operation", "compaction.apply"))

	return applied, nil
}
