package blocks

import (
	"context"
	"database/sql"
	"errors"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"
	"net/http"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	pg "github.com/lib/pq"
	"gorm.io/gorm"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	constants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	searchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/block-packs"
	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	scopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	blockpacksql "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/sqls/block_pack"
)

type BlockPackServiceInterface interface {
	GetMyBlockPackById(ctx context.Context, requestDto *capi.GetMyBlockPackByIdRequestDto) (*capi.GetMyBlockPackByIdResponseDto, *cexceptions.Exception)
	GetMyBlockPackAndItsParentById(ctx context.Context, requestDto *capi.GetMyBlockPackAndItsParentByIdRequestDto) (*capi.GetMyBlockPackAndItsParentByIdResponseDto, *cexceptions.Exception)
	GetMyBlockPacksByParentSubShelfId(ctx context.Context, requestDto *capi.GetMyBlockPacksByParentSubShelfIdRequestDto) (*capi.GetMyBlockPacksByParentSubShelfIdResponseDto, *cexceptions.Exception)
	GetAllMyBlockPacksByRootShelfId(ctx context.Context, requestDto *capi.GetAllMyBlockPacksByRootShelfIdRequestDto) (*capi.GetAllMyBlockPacksByRootShelfIdResponseDto, *cexceptions.Exception)
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

	SearchPrivateBlockPacks(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchBlockPackInput) (*cgqlmodels.SearchBlockPackConnection, *cexceptions.Exception)
}

type BlockPackService struct {
	validator           *validator.Validate
	db                  *gorm.DB
	blockPackScope      scopes.BlockPackScopeInterface
	subShelfRepository  repositories.SubShelfRepositoryInterface
	blockPackRepository repositories.BlockPackRepositoryInterface
}

func NewBlockPackService(
	validator *validator.Validate,
	db *gorm.DB,
	blockPackScope scopes.BlockPackScopeInterface,
	subShelfRepository repositories.SubShelfRepositoryInterface,
	blockPackRepository repositories.BlockPackRepositoryInterface,
) BlockPackServiceInterface {
	return &BlockPackService{
		validator:           validator,
		db:                  db,
		blockPackScope:      blockPackScope,
		subShelfRepository:  subShelfRepository,
		blockPackRepository: blockPackRepository,
	}
}

/* ============================== Main Methods ============================== */

func (s *BlockPackService) GetMyBlockPackById(
	ctx context.Context, requestDto *capi.GetMyBlockPackByIdRequestDto,
) (*capi.GetMyBlockPackByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
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

	onlyDeleted := types.Ternary_Neutral
	if requestDto.Param.IsDeleted != nil {
		if *requestDto.Param.IsDeleted {
			onlyDeleted = types.Ternary_Positive
		} else {
			onlyDeleted = types.Ternary_Negative
		}
	}

	blockPack, exception := s.blockPackRepository.CheckPermissionAndGetOneById(
		requestDto.Param.BlockPackId,
		actorUserId,
		[]schemas.BlockPackRelation{schemas.BlockPackRelation_YjsDocument},
		allowedPermissions,
		options.WithDB(db),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(onlyDeleted),
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
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
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

	onlyDeleted := types.Ternary_Neutral
	if requestDto.Param.IsDeleted != nil {
		if *requestDto.Param.IsDeleted {
			onlyDeleted = types.Ternary_Positive
		} else {
			onlyDeleted = types.Ternary_Negative
		}
	}

	resDto := capi.GetMyBlockPackAndItsParentByIdResponseDto{}
	var parentSubShelfPath types.UUIDArray
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
			return nil, apiexceptions.NewBlockPackException().NotFound().WithOrigin(err)
		}

		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(ctx, err, "Failed to scan BlockPack and its parent response")
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
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
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

	onlyDeleted := types.Ternary_Neutral
	if requestDto.Param.AreDeleted != nil {
		if *requestDto.Param.AreDeleted {
			onlyDeleted = types.Ternary_Positive
		} else {
			onlyDeleted = types.Ternary_Negative
		}
	}

	resDto := capi.GetMyBlockPacksByParentSubShelfIdResponseDto{}
	result := db.Model(&schemas.BlockPack{}).
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
		).Scopes(scopes.NewBlockPackScope().FilterOnlyDeleted(onlyDeleted)).
		Order("name ASC").
		Limit(int(data.MaxBlockPackOfSubShelf)).
		Scan(&resDto)
	if err := result.Error; err != nil {
		return nil, apiexceptions.NewBlockPackException().NotFound().WithOrigin(err)
	}

	return &resDto, nil
}

func (s *BlockPackService) GetAllMyBlockPacksByRootShelfId(
	ctx context.Context, requestDto *capi.GetAllMyBlockPacksByRootShelfIdRequestDto,
) (*capi.GetAllMyBlockPacksByRootShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
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

	onlyDeleted := types.Ternary_Neutral
	if requestDto.Param.AreDeleted != nil {
		if *requestDto.Param.AreDeleted {
			onlyDeleted = types.Ternary_Positive
		} else {
			onlyDeleted = types.Ternary_Negative
		}
	}

	resDto := capi.GetAllMyBlockPacksByRootShelfIdResponseDto{}
	result := db.Model(&schemas.BlockPack{}).
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
		).Scopes(scopes.NewBlockPackScope().FilterOnlyDeleted(onlyDeleted)).
		Limit(int(data.MaxBlockPackOfRootShelf)).
		Order("name ASC").
		Scan(&resDto)
	if err := result.Error; err != nil {
		return nil, apiexceptions.NewBlockPackException().NotFound().WithOrigin(err)
	}

	return &resDto, nil
}

func (s *BlockPackService) CreateBlockPack(
	ctx context.Context, requestDto *capi.CreateBlockPackRequestDto,
) (*capi.CreateBlockPackResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
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
		inputs.CreateBlockPackInput{
			Id:                  requestDto.Body.Id,
			Name:                requestDto.Body.Name,
			Icon:                requestDto.Body.Icon,
			HeaderBackgroundURL: requestDto.Body.HeaderBackgroundURL,
		},
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	document := schemas.BlockPackYjsDocument{BlockPackId: *newBlockPackId}
	if err := tx.Create(&document).Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewBlockPackException().FailedToCreate().WithOrigin(err)
	}
	if err := repositories.NewOutboxEventRepository().EnqueueYjsMaintenanceHint(
		tx,
		uuid.NewString(),
		*newBlockPackId,
		"block_pack_created",
	); err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewBlockPackException().FailedToCreate().WithOrigin(err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewBlockPackException().FailedToCommitTransaction().WithOrigin(err)
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
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
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

	input := make([]inputs.CreateBlockPackBySubShelfIdInput, len(requestDto.Body.CreatedBlockPacks))
	for index, createdBlockPack := range requestDto.Body.CreatedBlockPacks {
		input[index] = inputs.CreateBlockPackBySubShelfIdInput{
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
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithLockingStrength(options.LockingStrengthNoKeyUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	documents := make([]schemas.BlockPackYjsDocument, len(newBlockPackIds))
	for index, newBlockPackId := range newBlockPackIds {
		documents[index] = schemas.BlockPackYjsDocument{BlockPackId: newBlockPackId}
	}
	if err := tx.CreateInBatches(&documents, constants.MaxBatchCreateBlockSize).Error; err != nil {
		tx.Rollback()

		return nil, apiexceptions.NewBlockPackException().FailedToCreate().WithOrigin(err)
	}
	if err := repositories.NewOutboxEventRepository().EnqueueManyYjsMaintenanceHints(
		tx,
		uuid.NewString(),
		newBlockPackIds,
		"block_pack_created",
	); err != nil {
		tx.Rollback()

		return nil, apiexceptions.NewBlockPackException().FailedToCreate().WithOrigin(err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()

		return nil, apiexceptions.NewBlockPackException().FailedToCommitTransaction().WithOrigin(err)
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
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, apiexceptions.NewBlockPackException().FailedToUpdate().WithOrigin(tx.Error)
	}
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
		inputs.PartialUpdateBlockPackInput{
			Values: inputs.UpdateBlockPackInput{
				Name:                requestDto.Body.Values.Name,
				Icon:                requestDto.Body.Values.Icon,
				HeaderBackgroundURL: requestDto.Body.Values.HeaderBackgroundURL,
			},
			SetNull: requestDto.Body.SetNull,
		},
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := repositories.NewOutboxEventRepository().EnqueueBlockPackChanged(
		tx,
		requestDto.Param.BlockPackId.String(),
		[]uuid.UUID{requestDto.Param.BlockPackId},
	); err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewBlockPackException().FailedToCreate().WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewBlockPackException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.UpdateMyBlockPackByIdResponseDto{
		UpdatedAt: blockPack.UpdatedAt,
	}, nil
}

func (s *BlockPackService) UpdateMyBlockPacksByIds(
	ctx context.Context, requestDto *capi.UpdateMyBlockPacksByIdsRequestDto,
) (*capi.UpdateMyBlockPacksByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, apiexceptions.NewBlockPackException().FailedToUpdate().WithOrigin(tx.Error)
	}
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

	input := make([]inputs.UpdateBlockPackByIdInput, len(requestDto.Body.UpdatedBlockPacks))
	for index, updatedBlockPack := range requestDto.Body.UpdatedBlockPacks {
		input[index] = inputs.UpdateBlockPackByIdInput{
			Id: updatedBlockPack.BlockPackId,
			PartialUpdateInput: inputs.PartialUpdateInput[inputs.UpdateBlockPackInput]{
				Values: inputs.UpdateBlockPackInput{
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
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPackIds := make([]uuid.UUID, len(requestDto.Body.UpdatedBlockPacks))
	for index, updatedBlockPack := range requestDto.Body.UpdatedBlockPacks {
		blockPackIds[index] = updatedBlockPack.BlockPackId
	}
	if err := repositories.NewOutboxEventRepository().EnqueueBlockPackChanged(
		tx,
		"block-pack-bulk-update",
		blockPackIds,
	); err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewBlockPackException().FailedToCreate().WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewBlockPackException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &capi.UpdateMyBlockPacksByIdsResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *BlockPackService) MoveMyBlockPackByParentSubShelfId(
	ctx context.Context, requestDto *capi.MoveMyBlockPackByParentSubShelfIdRequestDto,
) (*capi.MoveMyBlockPackByParentSubShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
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
	if tx.Error != nil {
		return nil, cexceptions.New(
			"TransactionBeginFailed",
			"BlockPack",
			"MoveMyBlockPackByParentSubShelfId",
			"Failed to begin the block pack transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	_, exception = s.blockPackRepository.UpdateOneById(
		requestDto.Body.BlockPackId,
		actorUserId,
		inputs.PartialUpdateBlockPackInput{
			Values: inputs.UpdateBlockPackInput{
				ParentSubShelfId: &requestDto.Body.DestinationParentSubShelfId,
			},
			SetNull: nil,
		},
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := repositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		requestDto.Body.BlockPackId.String(),
		[]uuid.UUID{requestDto.Body.BlockPackId},
		nil,
		coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
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
	if err := repositories.NewOutboxEventRepository().EnqueueBlockPackChanged(
		tx,
		requestDto.Body.BlockPackId.String(),
		[]uuid.UUID{requestDto.Body.BlockPackId},
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
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	input := make([]inputs.UpdateBlockPackByIdInput, len(requestDto.Body.BlockPackIds))
	for index, blockPackId := range requestDto.Body.BlockPackIds {
		input[index] = inputs.UpdateBlockPackByIdInput{
			Id: blockPackId,
			PartialUpdateInput: inputs.PartialUpdateInput[inputs.UpdateBlockPackInput]{
				Values: inputs.UpdateBlockPackInput{
					ParentSubShelfId: &requestDto.Body.DestinationParentSubShelfId,
				},
			},
		}
	}
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, cexceptions.New(
			"TransactionBeginFailed",
			"BlockPack",
			"MoveMyBlockPacksByParentSubShelfId",
			"Failed to begin the block pack transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	exception = s.blockPackRepository.UpdateManyByIds(
		actorUserId,
		input,
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := repositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		"block-pack-bulk-move",
		requestDto.Body.BlockPackIds,
		nil,
		coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
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
	if err := repositories.NewOutboxEventRepository().EnqueueBlockPackChanged(
		tx,
		"block-pack-bulk-move",
		requestDto.Body.BlockPackIds,
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
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	input := make([]inputs.UpdateBlockPackByIdInput, 0)
	for _, movedBlockPack := range requestDto.Body.MovedBlockPacks {
		for _, blockPackId := range movedBlockPack.BlockPackIds {
			input = append(input, inputs.UpdateBlockPackByIdInput{
				Id: blockPackId,
				PartialUpdateInput: inputs.PartialUpdateInput[inputs.UpdateBlockPackInput]{
					Values: inputs.UpdateBlockPackInput{
						ParentSubShelfId: &movedBlockPack.DestinationParentSubShelfId,
					},
				},
			})
		}
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, cexceptions.New(
			"TransactionBeginFailed",
			"BlockPack",
			"MoveMyBlockPacksByParentSubShelfIds",
			"Failed to begin the block pack transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	if exception = s.blockPackRepository.UpdateManyByIds(
		actorUserId,
		input,
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPackIds := make([]uuid.UUID, len(input))
	for index, movedBlockPack := range input {
		blockPackIds[index] = movedBlockPack.Id
	}
	if err := repositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		"block-pack-multi-parent-move",
		blockPackIds,
		nil,
		coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
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
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
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
		options.WithDB(db),
		options.WithAllowedPermissions(allowedPermissions),
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
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
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
		options.WithDB(db),
		options.WithAllowedPermissions(allowedPermissions),
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
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
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
	if tx.Error != nil {
		return nil, cexceptions.New(
			"TransactionBeginFailed",
			"BlockPack",
			"DeleteMyBlockPackById",
			"Failed to begin the block pack transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	if exception = s.blockPackRepository.SoftDeleteOneById(
		requestDto.Param.BlockPackId,
		actorUserId,
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := repositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		requestDto.Param.BlockPackId.String(),
		[]uuid.UUID{requestDto.Param.BlockPackId},
		nil,
		coreevents.BlockPackAccessRevocationReason_ResourceUnavailable,
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
	if err := repositories.NewOutboxEventRepository().EnqueueBlockPackDeleted(
		tx,
		requestDto.Param.BlockPackId.String(),
		[]uuid.UUID{requestDto.Param.BlockPackId},
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
		return nil, apiexceptions.NewBlockPackException().InvalidDto().WithOrigin(err)
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
	if tx.Error != nil {
		return nil, cexceptions.New(
			"TransactionBeginFailed",
			"BlockPack",
			"DeleteMyBlockPacksByIds",
			"Failed to begin the block pack transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	if exception = s.blockPackRepository.SoftDeleteManyByIds(
		requestDto.Body.BlockPackIds,
		actorUserId,
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := repositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		"block-pack-bulk-delete",
		requestDto.Body.BlockPackIds,
		nil,
		coreevents.BlockPackAccessRevocationReason_ResourceUnavailable,
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
	if err := repositories.NewOutboxEventRepository().EnqueueBlockPackDeleted(
		tx,
		"block-pack-bulk-delete",
		requestDto.Body.BlockPackIds,
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

/* ============================== Service Methods for GraphQL BlockPack ============================== */

func (s *BlockPackService) SearchPrivateBlockPacks(
	ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchBlockPackInput,
) (*cgqlmodels.SearchBlockPackConnection, *cexceptions.Exception) {
	startTime := time.Now()
	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	onlyDeleted := types.Ternary_Negative
	if gqlInput.IsDeletedAt != nil && *gqlInput.IsDeletedAt {
		onlyDeleted = types.Ternary_Positive
	}

	query := db.Model(&schemas.BlockPack{}).
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
		searchCursor, err := searchcursor.Decode[cgqlmodels.SearchBlockPackCursorFields](*gqlInput.After)
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToDecode().WithOrigin(err)
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

	limit := constants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, constants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var blockPacks []schemas.BlockPack
	if err := query.Scopes(s.blockPackScope.IncludePreloads(
		[]schemas.BlockPackRelation{
			schemas.BlockPackRelation_Blocks,
		},
	)).Find(&blockPacks).Error; err != nil {
		return nil, apiexceptions.NewBlockPackException().NotFound().WithOrigin(err)
	}

	hasNextPage := len(blockPacks) > limit
	searchEdges := make([]*cgqlmodels.SearchBlockPackEdge, len(blockPacks))

	for index, blockPack := range blockPacks {
		searchCursor := searchcursor.SearchCursor[cgqlmodels.SearchBlockPackCursorFields]{
			Fields: cgqlmodels.SearchBlockPackCursorFields{
				ID: blockPack.Id,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToEncode().WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, apiexceptions.NewSearchException().FailedToUnmarshalSearchCursor()
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
