package blocks

import (
	"context"
	"database/sql"
	"errors"
	"net/http"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	pg "github.com/lib/pq"
	"gorm.io/gorm"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/block-packs"
	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	ssearchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	general "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/general"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres"
	blockpacksql "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/sqls/block_pack"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/core/exceptions"
)

type BlockPackServiceInterface interface {
	GetMyBlockPackById(ctx context.Context, requestDto *capi.GetMyBlockPackByIdRequestDto) (*capi.GetMyBlockPackByIdResponseDto, *cexceptions.Exception)
	GetMyBlockPackAndItsParentById(ctx context.Context, requestDto *capi.GetMyBlockPackAndItsParentByIdRequestDto) (*capi.GetMyBlockPackAndItsParentByIdResponseDto, *cexceptions.Exception)
	GetMyBlockPacksByParentSubShelfId(ctx context.Context, requestDto *capi.GetMyBlockPacksByParentSubShelfIdRequestDto) (*capi.GetMyBlockPacksByParentSubShelfIdResponseDto, *cexceptions.Exception)
	GetMyBlockPacksByRootShelfId(ctx context.Context, requestDto *capi.GetMyBlockPacksByRootShelfIdRequestDto) (*capi.GetMyBlockPacksByRootShelfIdResponseDto, *cexceptions.Exception)
	CreateBlockPack(ctx context.Context, requestDto *capi.CreateBlockPackRequestDto) (*capi.CreateBlockPackResponseDto, *cexceptions.Exception)
	CreateBlockPacks(ctx context.Context, requestDto *capi.CreateBlockPacksRequestDto) (*capi.CreateBlockPacksResponseDto, *cexceptions.Exception)
	UpdateMyBlockPackById(ctx context.Context, requestDto *capi.UpdateMyBlockPackByIdRequestDto) (*capi.UpdateMyBlockPackByIdResponseDto, *cexceptions.Exception)
	UpdateMyBlockPacksByIds(ctx context.Context, requestDto *capi.UpdateMyBlockPacksByIdsRequestDto) (*capi.UpdateMyBlockPacksByIdsResponseDto, *cexceptions.Exception)
	MoveMyBlockPackByParentSubShelfId(ctx context.Context, requestDto *capi.MoveMyBlockPackByParentSubShelfIdRequestDto) (*capi.MoveMyBlockPackByParentSubShelfIdResponseDto, *cexceptions.Exception)
	MoveMyBlockPacksByParentSubShelfId(ctx context.Context, requestDto *capi.MoveMyBlockPacksByParentSubShelfIdRequestDto) (*capi.MoveMyBlockPacksByParentSubShelfIdResponseDto, *cexceptions.Exception)
	MoveMyBlockPacksByParentSubShelfIds(ctx context.Context, requestDto *capi.MoveMyBlockPacksByParentSubShelfIdsRequestDto) (*capi.MoveMyBlockPacksByParentSubShelfIdsResponseDto, *cexceptions.Exception)
	RestoreMyBlockPackById(ctx context.Context, requestDto *capi.RestoreMyBlockPackByIdRequestDto) (*capi.RestoreMyBlockPackByIdResponseDto, *cexceptions.Exception)
	RestoreMyBlockPacksByIds(ctx context.Context, requestDto *capi.RestoreMyBlockPacksByIdsRequestDto) (*capi.RestoreMyBlockPacksByIdsResponseDto, *cexceptions.Exception)
	DeleteMyBlockPackById(ctx context.Context, requestDto *capi.DeleteMyBlockPackByIdRequestDto) (*capi.DeleteMyBlockPackByIdResponseDto, *cexceptions.Exception)
	DeleteMyBlockPacksByIds(ctx context.Context, requestDto *capi.DeleteMyBlockPacksByIdsRequestDto) (*capi.DeleteMyBlockPacksByIdsResponseDto, *cexceptions.Exception)

	/* ============================== GraphQL Methods ============================== */
	SearchPrivateBlockPacks(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchBlockPackInput) (*cgqlmodels.SearchBlockPackConnection, *cexceptions.Exception)
}

type BlockPackService struct {
	validator                        *validator.Validate
	db                               *gorm.DB
	blockPackScope                   sscopes.BlockPackScopeInterface
	subShelfRepository               srepositories.SubShelfRepositoryInterface
	blockPackRepository              srepositories.BlockPackRepositoryInterface
	accessRevocationOutboxRepository general.OutboxEventRepositoryInterface[coreevents.BlockPackAccessRevokedData]
	resourceOutboxRepository         general.OutboxEventRepositoryInterface[coreevents.ResourceChangedData]
	yjsOutboxRepository              general.OutboxEventRepositoryInterface[coreevents.YjsMaintenanceHintData]
	blockPackException               apiexceptions.BlockPackException
	searchException                  apiexceptions.SearchException
}

func NewBlockPackService(
	validator *validator.Validate,
	db *gorm.DB,
	blockPackScope sscopes.BlockPackScopeInterface,
	subShelfRepository srepositories.SubShelfRepositoryInterface,
	blockPackRepository srepositories.BlockPackRepositoryInterface,
	accessRevocationOutboxRepository general.OutboxEventRepositoryInterface[coreevents.BlockPackAccessRevokedData],
	resourceOutboxRepository general.OutboxEventRepositoryInterface[coreevents.ResourceChangedData],
	yjsOutboxRepository general.OutboxEventRepositoryInterface[coreevents.YjsMaintenanceHintData],
	blockPackException apiexceptions.BlockPackException,
	searchException apiexceptions.SearchException,
) BlockPackServiceInterface {
	return &BlockPackService{
		validator:                        validator,
		db:                               db,
		blockPackScope:                   blockPackScope,
		subShelfRepository:               subShelfRepository,
		blockPackRepository:              blockPackRepository,
		accessRevocationOutboxRepository: accessRevocationOutboxRepository,
		resourceOutboxRepository:         resourceOutboxRepository,
		yjsOutboxRepository:              yjsOutboxRepository,
		blockPackException:               blockPackException,
		searchException:                  searchException,
	}
}

func (s *BlockPackService) GetMyBlockPackById(
	ctx context.Context, requestDto *capi.GetMyBlockPackByIdRequestDto,
) (*capi.GetMyBlockPackByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	onlyDeleted := stypes.Ternary_Neutral
	if requestDto.Param.IsDeleted != nil {
		if *requestDto.Param.IsDeleted {
			onlyDeleted = stypes.Ternary_Positive
		} else {
			onlyDeleted = stypes.Ternary_Negative
		}
	}

	blockPack, exception := s.blockPackRepository.CheckPermissionAndGetOneById(
		requestDto.Param.BlockPackId,
		actorUserId,
		[]sschemas.BlockPackRelation{sschemas.BlockPackRelation_YjsDocument},
		allowedPermissions,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(onlyDeleted),
	)
	if exception != nil {
		return nil, exception
	}

	resDto := &capi.GetMyBlockPackByIdResponseDto{
		Id:                  blockPack.Id,
		ParentSubShelfId:    blockPack.ParentSubShelfId,
		Name:                blockPack.Name,
		Icon:                blockPack.Icon,
		HeaderBackgroundURL: blockPack.HeaderBackgroundURL,
		BlockCount:          blockPack.BlockCount,
		DeletedAt:           blockPack.DeletedAt,
		UpdatedAt:           blockPack.UpdatedAt,
		CreatedAt:           blockPack.CreatedAt,
	}
	if blockPack.YjsDocument != nil {
		resDto.LastUpdateSequence = blockPack.YjsDocument.LastUpdateSequence
		resDto.CompactedUntilSequence = blockPack.YjsDocument.CompactedUntilSequence
		resDto.ProjectedUntilSequence = blockPack.YjsDocument.ProjectedUntilSequence
		resDto.IsProjectionCurrent = blockPack.YjsDocument.ProjectedUntilSequence >= blockPack.YjsDocument.LastUpdateSequence
	}

	return resDto, nil
}

func (s *BlockPackService) GetMyBlockPackAndItsParentById(
	ctx context.Context, requestDto *capi.GetMyBlockPackAndItsParentByIdRequestDto,
) (*capi.GetMyBlockPackAndItsParentByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	onlyDeleted := stypes.Ternary_Neutral
	if requestDto.Param.IsDeleted != nil {
		if *requestDto.Param.IsDeleted {
			onlyDeleted = stypes.Ternary_Positive
		} else {
			onlyDeleted = stypes.Ternary_Negative
		}
	}

	resDto := capi.GetMyBlockPackAndItsParentByIdResponseDto{}
	var parentSubShelfPath stypes.UUIDArray
	err := db.Raw(blockpacksql.GetMyBlockPackAndItsParentByIdSQL,
		requestDto.Param.BlockPackId, actorUserId, pg.Array(allowedPermissions), onlyDeleted,
	).Row().
		Scan(&resDto.Id,
			&resDto.Name,
			&resDto.Icon,
			&resDto.HeaderBackgroundURL,
			&resDto.BlockCount,
			&resDto.LastUpdateSequence,
			&resDto.CompactedUntilSequence,
			&resDto.ProjectedUntilSequence,
			&resDto.IsProjectionCurrent,
			&resDto.DeletedAt,
			&resDto.UpdatedAt,
			&resDto.CreatedAt,
			&resDto.RootShelfId,
			&resDto.Permission,
			&resDto.ParentSubShelfId,
			&resDto.ParentSubShelfName,
			&resDto.ParentSubShelfPrevSubShelfId,
			&parentSubShelfPath,
			&resDto.ParentSubShelfDeletedAt,
			&resDto.ParentSubShelfUpdatedAt,
			&resDto.ParentSubShelfCreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, s.blockPackException.NotFound().WithOrigin(err)
		}

		if slogs.NotegicLogger != nil {
			slogs.NotegicLogger.Error(ctx, err, "Failed to scan BlockPack and its parent response")
		}

		return nil, cexceptions.New(
			"FailedToRead",
			"BlockPack",
			"Repository",
			"Failed to read BlockPack and its parent",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	resDto.ParentSubShelfPath = []uuid.UUID(parentSubShelfPath)

	return &resDto, nil
}

func (s *BlockPackService) GetMyBlockPacksByParentSubShelfId(
	ctx context.Context, requestDto *capi.GetMyBlockPacksByParentSubShelfIdRequestDto,
) (*capi.GetMyBlockPacksByParentSubShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	onlyDeleted := stypes.Ternary_Neutral
	if requestDto.Param.AreDeleted != nil {
		if *requestDto.Param.AreDeleted {
			onlyDeleted = stypes.Ternary_Positive
		} else {
			onlyDeleted = stypes.Ternary_Negative
		}
	}

	resDto := capi.GetMyBlockPacksByParentSubShelfIdResponseDto{}
	result := db.Model(&sschemas.BlockPack{}).
		Select(`
			"BlockPackTable".*,
			COALESCE(ydoc.last_update_sequence, 0) AS last_update_sequence,
			COALESCE(ydoc.compacted_until_sequence, 0) AS compacted_until_sequence,
			COALESCE(ydoc.projected_until_sequence, -1) AS projected_until_sequence,
			COALESCE(ydoc.projected_until_sequence, -1) >= COALESCE(ydoc.last_update_sequence, 0) AS is_projection_current
		`).
		Joins(`LEFT JOIN "BlockPackYjsDocumentTable" ydoc ON ydoc.block_pack_id = "BlockPackTable".id AND ydoc.deleted_at IS NULL`).
		Joins(`LEFT JOIN "SubShelfTable" ss ON "BlockPackTable".parent_sub_shelf_id = ss.id`).
		Joins(`LEFT JOIN "UsersToShelvesTable" uts ON ss.root_shelf_id = uts.root_shelf_id`).
		Where("ss.id = ? AND uts.user_id = ? AND uts.permission IN ?",
			requestDto.Param.ParentSubShelfId,
			actorUserId,
			allowedPermissions,
		).Scopes(sscopes.NewBlockPackScope().FilterOnlyDeleted(onlyDeleted)).
		Order("name ASC").
		Limit(int(data.MaxBlockPackOfSubShelf)).
		Scan(&resDto)
	if err := result.Error; err != nil {
		return nil, s.blockPackException.NotFound().WithOrigin(err)
	}

	return &resDto, nil
}

func (s *BlockPackService) GetMyBlockPacksByRootShelfId(
	ctx context.Context, requestDto *capi.GetMyBlockPacksByRootShelfIdRequestDto,
) (*capi.GetMyBlockPacksByRootShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	onlyDeleted := stypes.Ternary_Neutral
	if requestDto.Param.AreDeleted != nil {
		if *requestDto.Param.AreDeleted {
			onlyDeleted = stypes.Ternary_Positive
		} else {
			onlyDeleted = stypes.Ternary_Negative
		}
	}

	resDto := capi.GetMyBlockPacksByRootShelfIdResponseDto{}
	result := db.Model(&sschemas.BlockPack{}).
		Select(`
			"BlockPackTable".*,
			COALESCE(ydoc.last_update_sequence, 0) AS last_update_sequence,
			COALESCE(ydoc.compacted_until_sequence, 0) AS compacted_until_sequence,
			COALESCE(ydoc.projected_until_sequence, -1) AS projected_until_sequence,
			COALESCE(ydoc.projected_until_sequence, -1) >= COALESCE(ydoc.last_update_sequence, 0) AS is_projection_current
		`).
		Joins(`LEFT JOIN "BlockPackYjsDocumentTable" ydoc ON ydoc.block_pack_id = "BlockPackTable".id AND ydoc.deleted_at IS NULL`).
		Joins(`LEFT JOIN "SubShelfTable" ss ON "BlockPackTable".parent_sub_shelf_id = ss.id`).
		Joins(`LEFT JOIN "UsersToShelvesTable" uts ON ss.root_shelf_id = uts.root_shelf_id`).
		Where("ss.root_shelf_id = ? AND uts.user_id = ? AND uts.permission IN ?",
			requestDto.Param.RootShelfId, actorUserId, allowedPermissions,
		).Scopes(sscopes.NewBlockPackScope().FilterOnlyDeleted(onlyDeleted)).
		Limit(int(data.MaxBlockPackOfRootShelf)).
		Order("name ASC").
		Scan(&resDto)
	if err := result.Error; err != nil {
		return nil, s.blockPackException.NotFound().WithOrigin(err)
	}

	return &resDto, nil
}

func (s *BlockPackService) CreateBlockPack(
	ctx context.Context, requestDto *capi.CreateBlockPackRequestDto,
) (*capi.CreateBlockPackResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()

	newBlockPackId, exception := s.blockPackRepository.CreateOneBySubShelfId(
		requestDto.Body.ParentSubShelfId,
		actorUserId,
		sinputs.CreateBlockPackInput{
			Id:                  requestDto.Body.Id,
			Name:                requestDto.Body.Name,
			Icon:                requestDto.Body.Icon,
			HeaderBackgroundURL: requestDto.Body.HeaderBackgroundURL,
		},
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	document := sschemas.BlockPackYjsDocument{BlockPackId: *newBlockPackId}
	if err := tx.Create(&document).Error; err != nil {
		tx.Rollback()
		return nil, s.blockPackException.FailedToCreate().WithOrigin(err)
	}
	correlationId := uuid.NewString()
	var yjsDocument sschemas.BlockPackYjsDocument
	if err := tx.Model(&sschemas.BlockPackYjsDocument{}).
		Where("block_pack_id = ? AND deleted_at IS NULL", *newBlockPackId).
		First(&yjsDocument).Error; err != nil {
		tx.Rollback()
		return nil, s.blockPackException.FailedToCreate().WithOrigin(err)
	}
	if err := s.yjsOutboxRepository.EnqueueOutboxEvents(
		tx,
		coreevents.CoreYjsMaintenanceHintTopic,
		[]cevent.EventEnvelope[coreevents.YjsMaintenanceHintData]{
			{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     coreevents.EventType_YjsMaintenanceHint,
				AggregateType: coreevents.AggregateType_BlockPack,
				AggregateId:   yjsDocument.BlockPackId,
				KafkaKey:      yjsDocument.BlockPackId.String(),
				OccurredAt:    time.Now().UTC(),
				CorrelationId: correlationId,
				Data: coreevents.YjsMaintenanceHintData{
					BlockPackId:            yjsDocument.BlockPackId,
					DocumentId:             yjsDocument.Id,
					LatestUpdateSequence:   yjsDocument.LastUpdateSequence,
					CompactedUntilSequence: yjsDocument.CompactedUntilSequence,
					ProjectedUntilSequence: yjsDocument.ProjectedUntilSequence,
					LastCompactedAt:        yjsDocument.LastCompactedAt,
					UncompactedUpdateCount: yjsDocument.LastUpdateSequence - yjsDocument.CompactedUntilSequence,
					SnapshotBytes:          len(yjsDocument.Snapshot),
					StateVectorBytes:       len(yjsDocument.StateVector),
					Reason:                 "block_pack_created",
				},
			},
		},
	); err != nil {
		tx.Rollback()
		return nil, s.blockPackException.FailedToCreate().WithOrigin(err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, s.blockPackException.FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.CreateBlockPackResponseDto{
		Id:        *newBlockPackId,
		CreatedAt: time.Now(),
	}, nil
}

func (s *BlockPackService) CreateBlockPacks(
	ctx context.Context, requestDto *capi.CreateBlockPacksRequestDto,
) (*capi.CreateBlockPacksResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()

	input := make([]sinputs.CreateBlockPackBySubShelfIdInput, len(requestDto.Body.CreatedBlockPacks))
	for index, createdBlockPack := range requestDto.Body.CreatedBlockPacks {
		input[index] = sinputs.CreateBlockPackBySubShelfIdInput{
			Id:                  createdBlockPack.Id,
			ParentSubShelfId:    createdBlockPack.ParentSubShelfId,
			Name:                createdBlockPack.Name,
			Icon:                createdBlockPack.Icon,
			HeaderBackgroundURL: createdBlockPack.HeaderBackgroundURL,
		}
	}
	newBlockPackIds, exception := s.blockPackRepository.CreateManyBySubShelfIds(
		actorUserId,
		input,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	documents := make([]sschemas.BlockPackYjsDocument, len(newBlockPackIds))
	for index, newBlockPackId := range newBlockPackIds {
		documents[index] = sschemas.BlockPackYjsDocument{BlockPackId: newBlockPackId}
	}
	if err := tx.CreateInBatches(&documents, sconstants.MaxBatchCreateBlockSize).Error; err != nil {
		tx.Rollback()

		return nil, s.blockPackException.FailedToCreate().WithOrigin(err)
	}
	if len(newBlockPackIds) > 0 {
		correlationId := uuid.NewString()
		var yjsDocuments []sschemas.BlockPackYjsDocument
		if err := tx.Model(&sschemas.BlockPackYjsDocument{}).
			Where("block_pack_id IN ? AND deleted_at IS NULL", newBlockPackIds).
			Find(&yjsDocuments).Error; err != nil {
			tx.Rollback()
			return nil, s.blockPackException.FailedToCreate().WithOrigin(err)
		}
		if len(yjsDocuments) != len(newBlockPackIds) {
			tx.Rollback()
			return nil, s.blockPackException.FailedToCreate().WithOrigin(errors.New("Yjs maintenance hints require documents for every BlockPack ID"))
		}
		yjsEvents := make([]cevent.EventEnvelope[coreevents.YjsMaintenanceHintData], 0, len(yjsDocuments))
		for _, document := range yjsDocuments {
			yjsEvents = append(yjsEvents, cevent.EventEnvelope[coreevents.YjsMaintenanceHintData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_YjsMaintenanceHint,
			AggregateType: coreevents.AggregateType_BlockPack,
			AggregateId:   document.BlockPackId,
			KafkaKey:      document.BlockPackId.String(),
			OccurredAt:    time.Now().UTC(),
			CorrelationId: correlationId,
			Data: coreevents.YjsMaintenanceHintData{
				BlockPackId:            document.BlockPackId,
				DocumentId:             document.Id,
				LatestUpdateSequence:   document.LastUpdateSequence,
				CompactedUntilSequence: document.CompactedUntilSequence,
				ProjectedUntilSequence: document.ProjectedUntilSequence,
				LastCompactedAt:        document.LastCompactedAt,
				UncompactedUpdateCount: document.LastUpdateSequence - document.CompactedUntilSequence,
				SnapshotBytes:          len(document.Snapshot),
				StateVectorBytes:       len(document.StateVector),
				Reason:                 "block_pack_created",
			},
			})
		}
		if err := s.yjsOutboxRepository.EnqueueOutboxEvents(
			tx,
			coreevents.CoreYjsMaintenanceHintTopic,
			yjsEvents,
		); err != nil {
			tx.Rollback()

			return nil, s.blockPackException.FailedToCreate().WithOrigin(err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()

		return nil, s.blockPackException.FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.CreateBlockPacksResponseDto{
		Ids:       newBlockPackIds,
		CreatedAt: time.Now(),
	}, nil
}

func (s *BlockPackService) UpdateMyBlockPackById(
	ctx context.Context, requestDto *capi.UpdateMyBlockPackByIdRequestDto,
) (*capi.UpdateMyBlockPackByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	blockPack, exception := s.blockPackRepository.UpdateOneById(
		requestDto.Param.BlockPackId,
		actorUserId,
		sinputs.PartialUpdateBlockPackInput{
			Values: sinputs.UpdateBlockPackInput{
				Name:                requestDto.Body.Values.Name,
				Icon:                requestDto.Body.Values.Icon,
				HeaderBackgroundURL: requestDto.Body.Values.HeaderBackgroundURL,
			},
			SetNull: requestDto.Body.SetNull,
		},
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPackChangedEvents := []cevent.EventEnvelope[coreevents.ResourceChangedData]{
		{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_BlockPackChanged,
			AggregateType: coreevents.AggregateType_BlockPack,
			AggregateId:   requestDto.Param.BlockPackId,
			KafkaKey:      requestDto.Param.BlockPackId.String(),
			OccurredAt:    time.Now().UTC(),
			CorrelationId: requestDto.Param.BlockPackId.String(),
			Data: coreevents.ResourceChangedData{
				ResourceId: requestDto.Param.BlockPackId,
				Change:     coreevents.ResourceEventChange_Updated,
			},
		},
	}
	if err := s.resourceOutboxRepository.EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		blockPackChangedEvents,
	); err != nil {
		tx.Rollback()
		return nil, s.blockPackException.FailedToCreate().WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, s.blockPackException.FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.UpdateMyBlockPackByIdResponseDto{
		UpdatedAt: blockPack.UpdatedAt,
	}, nil
}

func (s *BlockPackService) UpdateMyBlockPacksByIds(
	ctx context.Context, requestDto *capi.UpdateMyBlockPacksByIdsRequestDto,
) (*capi.UpdateMyBlockPacksByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	input := make([]sinputs.UpdateBlockPackByIdInput, len(requestDto.Body.UpdatedBlockPacks))
	for index, updatedBlockPack := range requestDto.Body.UpdatedBlockPacks {
		input[index] = sinputs.UpdateBlockPackByIdInput{
			Id: updatedBlockPack.BlockPackId,
			PartialUpdateInput: sinputs.PartialUpdateInput[sinputs.UpdateBlockPackInput]{
				Values: sinputs.UpdateBlockPackInput{
					Name:                updatedBlockPack.Values.Name,
					Icon:                updatedBlockPack.Values.Icon,
					HeaderBackgroundURL: updatedBlockPack.Values.HeaderBackgroundURL,
				},
			},
		}
	}
	exception = s.blockPackRepository.UpdateManyByIds(
		actorUserId,
		input,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPackIds := make([]uuid.UUID, len(requestDto.Body.UpdatedBlockPacks))
	for index, updatedBlockPack := range requestDto.Body.UpdatedBlockPacks {
		blockPackIds[index] = updatedBlockPack.BlockPackId
	}
	blockPackChangedEvents := make([]cevent.EventEnvelope[coreevents.ResourceChangedData], 0, len(blockPackIds))
	for _, blockPackId := range blockPackIds {
		blockPackChangedEvents = append(blockPackChangedEvents, cevent.EventEnvelope[coreevents.ResourceChangedData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_BlockPackChanged,
			AggregateType: coreevents.AggregateType_BlockPack,
			AggregateId:   blockPackId,
			KafkaKey:      blockPackId.String(),
			OccurredAt:    time.Now().UTC(),
			CorrelationId: "block-pack-bulk-update",
			Data: coreevents.ResourceChangedData{
				ResourceId: blockPackId,
				Change:     coreevents.ResourceEventChange_Updated,
			},
		})
	}
	if err := s.resourceOutboxRepository.EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		blockPackChangedEvents,
	); err != nil {
		tx.Rollback()
		return nil, s.blockPackException.FailedToCreate().WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, s.blockPackException.FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.UpdateMyBlockPacksByIdsResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *BlockPackService) MoveMyBlockPackByParentSubShelfId(
	ctx context.Context, requestDto *capi.MoveMyBlockPackByParentSubShelfIdRequestDto,
) (*capi.MoveMyBlockPackByParentSubShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()
	_, exception = s.blockPackRepository.UpdateOneById(
		requestDto.Body.BlockPackId,
		actorUserId,
		sinputs.PartialUpdateBlockPackInput{
			Values: sinputs.UpdateBlockPackInput{
				ParentSubShelfId: &requestDto.Body.DestinationParentSubShelfId,
			},
			SetNull: nil,
		},
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := s.accessRevocationOutboxRepository.EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		[]cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData]{
			{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     coreevents.EventType_BlockPackAccessRevoked,
				AggregateType: coreevents.AggregateType_BlockPack,
				AggregateId:   requestDto.Body.BlockPackId,
				KafkaKey:      requestDto.Body.BlockPackId.String(),
				OccurredAt:    time.Now().UTC(),
				CorrelationId: requestDto.Body.BlockPackId.String(),
				Data: coreevents.BlockPackAccessRevokedData{
					Reason: coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
				},
			},
		},
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"MoveMyBlockPackByParentSubShelfId",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := s.resourceOutboxRepository.EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		[]cevent.EventEnvelope[coreevents.ResourceChangedData]{
			{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     coreevents.EventType_BlockPackChanged,
				AggregateType: coreevents.AggregateType_BlockPack,
				AggregateId:   requestDto.Body.BlockPackId,
				KafkaKey:      requestDto.Body.BlockPackId.String(),
				OccurredAt:    time.Now().UTC(),
				CorrelationId: requestDto.Body.BlockPackId.String(),
				Data: coreevents.ResourceChangedData{
					ResourceId: requestDto.Body.BlockPackId,
					Change:     coreevents.ResourceEventChange_Updated,
				},
			},
		},
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"MoveMyBlockPackByParentSubShelfId",
			"Failed to create resource event",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"BlockPack",
			"MoveMyBlockPackByParentSubShelfId",
			"Failed to commit the block pack transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &capi.MoveMyBlockPackByParentSubShelfIdResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *BlockPackService) MoveMyBlockPacksByParentSubShelfId(
	ctx context.Context, requestDto *capi.MoveMyBlockPacksByParentSubShelfIdRequestDto,
) (*capi.MoveMyBlockPacksByParentSubShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	input := make([]sinputs.UpdateBlockPackByIdInput, len(requestDto.Body.BlockPackIds))
	for index, blockPackId := range requestDto.Body.BlockPackIds {
		input[index] = sinputs.UpdateBlockPackByIdInput{
			Id: blockPackId,
			PartialUpdateInput: sinputs.PartialUpdateInput[sinputs.UpdateBlockPackInput]{
				Values: sinputs.UpdateBlockPackInput{
					ParentSubShelfId: &requestDto.Body.DestinationParentSubShelfId,
				},
			},
		}
	}
	tx := s.db.WithContext(ctx).Begin()
	exception = s.blockPackRepository.UpdateManyByIds(
		actorUserId,
		input,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPackAccessRevokedEvents := make([]cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData], 0, len(requestDto.Body.BlockPackIds))
	for _, blockPackId := range requestDto.Body.BlockPackIds {
		blockPackAccessRevokedEvents = append(blockPackAccessRevokedEvents, cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_BlockPackAccessRevoked,
			AggregateType: coreevents.AggregateType_BlockPack,
			AggregateId:   blockPackId,
			KafkaKey:      blockPackId.String(),
			OccurredAt:    time.Now().UTC(),
			CorrelationId: "block-pack-bulk-move",
			Data: coreevents.BlockPackAccessRevokedData{
				Reason: coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
			},
		})
	}
	if err := s.accessRevocationOutboxRepository.EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		blockPackAccessRevokedEvents,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"MoveMyBlockPacksByParentSubShelfId",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	blockPackChangedEvents := make([]cevent.EventEnvelope[coreevents.ResourceChangedData], 0, len(requestDto.Body.BlockPackIds))
	for _, blockPackId := range requestDto.Body.BlockPackIds {
		blockPackChangedEvents = append(blockPackChangedEvents, cevent.EventEnvelope[coreevents.ResourceChangedData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_BlockPackChanged,
			AggregateType: coreevents.AggregateType_BlockPack,
			AggregateId:   blockPackId,
			KafkaKey:      blockPackId.String(),
			OccurredAt:    time.Now().UTC(),
			CorrelationId: "block-pack-bulk-move",
			Data: coreevents.ResourceChangedData{
				ResourceId: blockPackId,
				Change:     coreevents.ResourceEventChange_Updated,
			},
		})
	}
	if err := s.resourceOutboxRepository.EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		blockPackChangedEvents,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"MoveMyBlockPacksByParentSubShelfId",
			"Failed to create resource event",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"BlockPack",
			"MoveMyBlockPacksByParentSubShelfId",
			"Failed to commit the block pack transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &capi.MoveMyBlockPacksByParentSubShelfIdResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *BlockPackService) MoveMyBlockPacksByParentSubShelfIds(
	ctx context.Context, requestDto *capi.MoveMyBlockPacksByParentSubShelfIdsRequestDto,
) (*capi.MoveMyBlockPacksByParentSubShelfIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	input := make([]sinputs.UpdateBlockPackByIdInput, 0)
	for _, movedBlockPack := range requestDto.Body.MovedBlockPacks {
		for _, blockPackId := range movedBlockPack.BlockPackIds {
			input = append(input, sinputs.UpdateBlockPackByIdInput{
				Id: blockPackId,
				PartialUpdateInput: sinputs.PartialUpdateInput[sinputs.UpdateBlockPackInput]{
					Values: sinputs.UpdateBlockPackInput{
						ParentSubShelfId: &movedBlockPack.DestinationParentSubShelfId,
					},
				},
			})
		}
	}

	tx := s.db.WithContext(ctx).Begin()
	if exception = s.blockPackRepository.UpdateManyByIds(
		actorUserId,
		input,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPackIds := make([]uuid.UUID, len(input))
	for index, movedBlockPack := range input {
		blockPackIds[index] = movedBlockPack.Id
	}
	blockPackAccessRevokedEvents := make([]cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData], 0, len(blockPackIds))
	for _, blockPackId := range blockPackIds {
		blockPackAccessRevokedEvents = append(blockPackAccessRevokedEvents, cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_BlockPackAccessRevoked,
			AggregateType: coreevents.AggregateType_BlockPack,
			AggregateId:   blockPackId,
			KafkaKey:      blockPackId.String(),
			OccurredAt:    time.Now().UTC(),
			CorrelationId: "block-pack-multi-parent-move",
			Data: coreevents.BlockPackAccessRevokedData{
				Reason: coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
			},
		})
	}
	if err := s.accessRevocationOutboxRepository.EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		blockPackAccessRevokedEvents,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"MoveMyBlockPacksByParentSubShelfIds",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"BlockPack",
			"MoveMyBlockPacksByParentSubShelfIds",
			"Failed to commit the block pack transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &capi.MoveMyBlockPacksByParentSubShelfIdsResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *BlockPackService) RestoreMyBlockPackById(
	ctx context.Context, requestDto *capi.RestoreMyBlockPackByIdRequestDto,
) (*capi.RestoreMyBlockPackByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	restoredBlockPack, exception := s.blockPackRepository.RestoreSoftDeletedOneById(
		requestDto.Param.BlockPackId,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.RestoreMyBlockPackByIdResponseDto{
		Id:                  restoredBlockPack.Id,
		ParentSubShelfId:    restoredBlockPack.ParentSubShelfId,
		Name:                restoredBlockPack.Name,
		Icon:                restoredBlockPack.Icon,
		HeaderBackgroundURL: restoredBlockPack.HeaderBackgroundURL,
		BlockCount:          restoredBlockPack.BlockCount,
		DeletedAt:           restoredBlockPack.DeletedAt,
		UpdatedAt:           restoredBlockPack.UpdatedAt,
		CreatedAt:           restoredBlockPack.CreatedAt,
	}, nil
}

func (s *BlockPackService) RestoreMyBlockPacksByIds(
	ctx context.Context, requestDto *capi.RestoreMyBlockPacksByIdsRequestDto,
) (*capi.RestoreMyBlockPacksByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	restoredBlockPacks, exception := s.blockPackRepository.RestoreSoftDeletedManyByIds(
		requestDto.Body.BlockPackIds,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	resDto := capi.RestoreMyBlockPacksByIdsResponseDto{}
	for _, restoredBlockPack := range restoredBlockPacks {
		resDto = append(resDto, capi.RestoreMyBlockPackByIdResponseDto{
			Id:                  restoredBlockPack.Id,
			ParentSubShelfId:    restoredBlockPack.ParentSubShelfId,
			Name:                restoredBlockPack.Name,
			Icon:                restoredBlockPack.Icon,
			HeaderBackgroundURL: restoredBlockPack.HeaderBackgroundURL,
			BlockCount:          restoredBlockPack.BlockCount,
			DeletedAt:           restoredBlockPack.DeletedAt,
			UpdatedAt:           restoredBlockPack.UpdatedAt,
			CreatedAt:           restoredBlockPack.CreatedAt,
		})
	}

	return &resDto, nil
}

func (s *BlockPackService) DeleteMyBlockPackById(
	ctx context.Context, requestDto *capi.DeleteMyBlockPackByIdRequestDto,
) (*capi.DeleteMyBlockPackByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()
	if exception = s.blockPackRepository.SoftDeleteOneById(
		requestDto.Param.BlockPackId,
		actorUserId,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := s.accessRevocationOutboxRepository.EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		[]cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData]{
			{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     coreevents.EventType_BlockPackAccessRevoked,
				AggregateType: coreevents.AggregateType_BlockPack,
				AggregateId:   requestDto.Param.BlockPackId,
				KafkaKey:      requestDto.Param.BlockPackId.String(),
				OccurredAt:    time.Now().UTC(),
				CorrelationId: requestDto.Param.BlockPackId.String(),
				Data: coreevents.BlockPackAccessRevokedData{
					Reason: coreevents.BlockPackAccessRevocationReason_ResourceUnavailable,
				},
			},
		},
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMyBlockPackById",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := s.resourceOutboxRepository.EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		[]cevent.EventEnvelope[coreevents.ResourceChangedData]{
			{
				SchemaVersion: cevent.Version,
				EventId:       uuid.New(),
				EventType:     coreevents.EventType_BlockPackDeleted,
				AggregateType: coreevents.AggregateType_BlockPack,
				AggregateId:   requestDto.Param.BlockPackId,
				KafkaKey:      requestDto.Param.BlockPackId.String(),
				OccurredAt:    time.Now().UTC(),
				CorrelationId: requestDto.Param.BlockPackId.String(),
				Data: coreevents.ResourceChangedData{
					ResourceId: requestDto.Param.BlockPackId,
					Change:     coreevents.ResourceEventChange_Deleted,
				},
			},
		},
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMyBlockPackById",
			"Failed to create resource event",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"BlockPack",
			"DeleteMyBlockPackById",
			"Failed to commit the block pack transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &capi.DeleteMyBlockPackByIdResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *BlockPackService) DeleteMyBlockPacksByIds(
	ctx context.Context, requestDto *capi.DeleteMyBlockPacksByIdsRequestDto,
) (*capi.DeleteMyBlockPacksByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.blockPackException.InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()
	if exception = s.blockPackRepository.SoftDeleteManyByIds(
		requestDto.Body.BlockPackIds,
		actorUserId,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPackAccessRevokedEvents := make([]cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData], 0, len(requestDto.Body.BlockPackIds))
	for _, blockPackId := range requestDto.Body.BlockPackIds {
		blockPackAccessRevokedEvents = append(blockPackAccessRevokedEvents, cevent.EventEnvelope[coreevents.BlockPackAccessRevokedData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_BlockPackAccessRevoked,
			AggregateType: coreevents.AggregateType_BlockPack,
			AggregateId:   blockPackId,
			KafkaKey:      blockPackId.String(),
			OccurredAt:    time.Now().UTC(),
			CorrelationId: "block-pack-bulk-delete",
			Data: coreevents.BlockPackAccessRevokedData{
				Reason: coreevents.BlockPackAccessRevocationReason_ResourceUnavailable,
			},
		})
	}
	if err := s.accessRevocationOutboxRepository.EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		blockPackAccessRevokedEvents,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMyBlockPacksByIds",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	blockPackDeletedEvents := make([]cevent.EventEnvelope[coreevents.ResourceChangedData], 0, len(requestDto.Body.BlockPackIds))
	for _, blockPackId := range requestDto.Body.BlockPackIds {
		blockPackDeletedEvents = append(blockPackDeletedEvents, cevent.EventEnvelope[coreevents.ResourceChangedData]{
			SchemaVersion: cevent.Version,
			EventId:       uuid.New(),
			EventType:     coreevents.EventType_BlockPackDeleted,
			AggregateType: coreevents.AggregateType_BlockPack,
			AggregateId:   blockPackId,
			KafkaKey:      blockPackId.String(),
			OccurredAt:    time.Now().UTC(),
			CorrelationId: "block-pack-bulk-delete",
			Data: coreevents.ResourceChangedData{
				ResourceId: blockPackId,
				Change:     coreevents.ResourceEventChange_Deleted,
			},
		})
	}
	if err := s.resourceOutboxRepository.EnqueueOutboxEvents(
		tx,
		coreevents.CoreLifecycleTopic,
		blockPackDeletedEvents,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMyBlockPacksByIds",
			"Failed to create resource events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"BlockPack",
			"DeleteMyBlockPacksByIds",
			"Failed to commit the block pack transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &capi.DeleteMyBlockPacksByIdsResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

/* ============================== GraphQL Methods ============================== */

func (s *BlockPackService) SearchPrivateBlockPacks(
	ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchBlockPackInput,
) (*cgqlmodels.SearchBlockPackConnection, *cexceptions.Exception) {
	startTime := time.Now()
	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	onlyDeleted := stypes.Ternary_Negative
	if gqlInput.IsDeletedAt != nil && *gqlInput.IsDeletedAt {
		onlyDeleted = stypes.Ternary_Positive
	}

	query := db.Model(&sschemas.BlockPack{}).
		Select(`"BlockPackTable".*`).
		Joins(`INNER JOIN "SubShelfTable" ss ON ss.id = "BlockPackTable".parent_sub_shelf_id`).
		Joins(`INNER JOIN "UsersToShelvesTable" uts ON uts.root_shelf_id = ss.root_shelf_id`).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, allowedPermissions).
		Scopes(s.blockPackScope.FilterOnlyDeleted(onlyDeleted))

	if gqlInput.ParentSubShelfID != nil {
		query = query.Where(`"BlockPackTable".parent_sub_shelf_id = ?`, *gqlInput.ParentSubShelfID)
	}

	if gqlInput.RootShelfID != nil {
		query = query.Where("ss.root_shelf_id = ?", *gqlInput.RootShelfID)
	}

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			`"BlockPackTable".name ILIKE ?`,
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := ssearchcursor.Decode[cgqlmodels.SearchBlockPackCursorFields](*gqlInput.After)
		if err != nil {
			return nil, s.searchException.FailedToDecode().WithOrigin(err)
		}

		query = query.Where(`"BlockPackTable".id > ?`, searchCursor.Fields.ID)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		cending := cgqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == cgqlmodels.SearchSortOrderDesc {
			cending = cgqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case cgqlmodels.SearchBlockPackSortByName:
			query = query.Order(`"BlockPackTable".name ` + cending).
				Order(`"BlockPackTable".updated_at ` + cending).
				Order(`"BlockPackTable".created_at ` + cending)
		case cgqlmodels.SearchBlockPackSortByBlockCount:
			query = query.Order(`"BlockPackTable".block_count ` + cending).
				Order(`"BlockPackTable".name ` + cending).
				Order(`"BlockPackTable".updated_at ` + cending).
				Order(`"BlockPackTable".created_at ` + cending)
		case cgqlmodels.SearchBlockPackSortByLastUpdate:
			query = query.Order(`"BlockPackTable".updated_at ` + cending).
				Order(`"BlockPackTable".name ` + cending).
				Order(`"BlockPackTable".created_at ` + cending)
		case cgqlmodels.SearchBlockPackSortByCreatedAt:
			query = query.Order(`"BlockPackTable".created_at ` + cending).
				Order(`"BlockPackTable".name ` + cending).
				Order(`"BlockPackTable".updated_at ` + cending)
		default:
			query = query.Order(`"BlockPackTable".name ` + cending).
				Order(`"BlockPackTable".updated_at ` + cending).
				Order(`"BlockPackTable".created_at ` + cending)
		}
	}

	limit := sconstants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, sconstants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var blockPacks []sschemas.BlockPack
	if err := query.Scopes(s.blockPackScope.IncludePreloads(
		[]sschemas.BlockPackRelation{
			sschemas.BlockPackRelation_Blocks,
		},
	)).Find(&blockPacks).Error; err != nil {
		return nil, s.blockPackException.NotFound().WithOrigin(err)
	}

	hasNextPage := len(blockPacks) > limit
	searchEdges := make([]*cgqlmodels.SearchBlockPackEdge, len(blockPacks))

	for index, blockPack := range blockPacks {
		searchCursor := ssearchcursor.SearchCursor[cgqlmodels.SearchBlockPackCursorFields]{
			Fields: cgqlmodels.SearchBlockPackCursorFields{
				ID: blockPack.Id,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, s.searchException.FailedToEncode().WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, s.searchException.FailedToUnmarshalSearchCursor()
		}

		searchEdges[index] = &cgqlmodels.SearchBlockPackEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                blockPack.ToPrivateBlockPack(),
		}
	}

	searchPageInfo := &cgqlmodels.SearchPageInfo{
		HasNextPage:     hasNextPage,
		HasPreviousPage: gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0,
	}

	if len(searchEdges) > 0 {
		searchPageInfo.StartEncodedSearchCursor = &searchEdges[0].EncodedSearchCursor
		searchPageInfo.EndEncodedSearchCursor = &searchEdges[len(searchEdges)-1].EncodedSearchCursor
	}

	searchTime := float64(time.Since(startTime).Nanoseconds()) / 1e6
	if hasNextPage {
		searchEdges = searchEdges[:limit]
	}

	return &cgqlmodels.SearchBlockPackConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
