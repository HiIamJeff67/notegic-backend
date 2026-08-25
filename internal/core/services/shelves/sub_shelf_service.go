package shelves

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	pg "github.com/lib/pq"
	"gorm.io/gorm"

	cblockpacks "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/block-packs"
	csubshelves "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/sub-shelves"
	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	ssearchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	storage "github.com/HiIamJeff67/notegic-backend/internal/core/data/storage"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
)

type SubShelfServiceInterface interface {
	GetMySubShelfById(ctx context.Context, requestDto *csubshelves.GetMySubShelfByIdRequestDto) (*csubshelves.GetMySubShelfByIdResponseDto, *cexceptions.Exception)
	GetMySubShelvesByPrevSubShelfId(ctx context.Context, requestDto *csubshelves.GetMySubShelvesByPrevSubShelfIdRequestDto) (*csubshelves.GetMySubShelvesByPrevSubShelfIdResponseDto, *cexceptions.Exception)
	GetAllMySubShelvesByRootShelfId(ctx context.Context, requestDto *csubshelves.GetAllMySubShelvesByRootShelfIdRequestDto) (*csubshelves.GetAllMySubShelvesByRootShelfIdResponseDto, *cexceptions.Exception)
	GetMySubShelvesAndItemsByPrevSubShelfId(ctx context.Context, requestDto *csubshelves.GetMySubShelvesAndItemsByPrevSubShelfIdRequestDto) (*csubshelves.GetMySubShelvesAndItemsByPrevSubShelfIdResponseDto, *cexceptions.Exception)
	CreateSubShelfByRootShelfId(ctx context.Context, requestDto *csubshelves.CreateSubShelfByRootShelfIdRequestDto) (*csubshelves.CreateSubShelfByRootShelfIdResponseDto, *cexceptions.Exception)
	CreateSubShelvesByRootShelfIds(ctx context.Context, requestDto *csubshelves.CreateSubShelvesByRootShelfIdsRequestDto) (*csubshelves.CreateSubShelvesByRootShelfIdsResponseDto, *cexceptions.Exception)
	UpdateMySubShelfById(ctx context.Context, requestDto *csubshelves.UpdateMySubShelfByIdRequestDto) (*csubshelves.UpdateMySubShelfByIdResponseDto, *cexceptions.Exception)
	UpdateMySubShelvesByIds(ctx context.Context, requestDto *csubshelves.UpdateMySubShelvesByIdsRequestDto) (*csubshelves.UpdateMySubShelvesByIdsResponseDto, *cexceptions.Exception)
	MoveMySubShelfByRootShelfId(ctx context.Context, requestDto *csubshelves.MoveMySubShelfByRootShelfIdRequestDto) (*csubshelves.MoveMySubShelfByRootShelfIdResponseDto, *cexceptions.Exception)
	MoveMySubShelvesByRootShelfId(ctx context.Context, requestDto *csubshelves.MoveMySubShelvesByRootShelfIdRequestDto) (*csubshelves.MoveMySubShelvesByRootShelfIdResponseDto, *cexceptions.Exception)
	MoveMySubShelvesByRootShelfIds(ctx context.Context, requestDto *csubshelves.MoveMySubShelvesByRootShelfIdsRequestDto) (*csubshelves.MoveMySubShelvesByRootShelfIdsResponseDto, *cexceptions.Exception)
	RestoreMySubShelfById(ctx context.Context, requestDto *csubshelves.RestoreMySubShelfByIdRequestDto) (*csubshelves.RestoreMySubShelfByIdResponseDto, *cexceptions.Exception)
	RestoreMySubShelvesByIds(ctx context.Context, requestDto *csubshelves.RestoreMySubShelvesByIdsRequestDto) (*csubshelves.RestoreMySubShelvesByIdsResponseDto, *cexceptions.Exception)
	DeleteMySubShelfById(ctx context.Context, requestDto *csubshelves.DeleteMySubShelfByIdRequestDto) (*csubshelves.DeleteMySubShelfByIdResponseDto, *cexceptions.Exception)
	DeleteMySubShelvesByIds(ctx context.Context, requestDto *csubshelves.DeleteMySubShelvesByIdsRequestDto) (*csubshelves.DeleteMySubShelvesByIdsResponseDto, *cexceptions.Exception)

	SearchPrivateSubShelves(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchSubShelfInput) (*cgqlmodels.SearchSubShelfConnection, *cexceptions.Exception)
}

type SubShelfService struct {
	validator           *validator.Validate
	db                  *gorm.DB
	storage             storage.StorageInterface
	subShelfScope       sscopes.SubShelfScopeInterface
	subShelfRepository  srepositories.SubShelfRepositoryInterface
	rootShelfRepository srepositories.RootShelfRepositoryInterface
	materialRepository  srepositories.MaterialRepositoryInterface
	blockPackRepository srepositories.BlockPackRepositoryInterface
}

func NewSubShelfService(
	validator *validator.Validate,
	db *gorm.DB,
	storage storage.StorageInterface,
	subShelfScope sscopes.SubShelfScopeInterface,
	subShelfRepository srepositories.SubShelfRepositoryInterface,
	rootShelfRepository srepositories.RootShelfRepositoryInterface,
	materialRepository srepositories.MaterialRepositoryInterface,
	blockPackRepository srepositories.BlockPackRepositoryInterface,
) SubShelfServiceInterface {
	if db == nil {
		db = data.DB
	}
	return &SubShelfService{
		validator:           validator,
		db:                  db,
		storage:             storage,
		subShelfScope:       subShelfScope,
		subShelfRepository:  subShelfRepository,
		rootShelfRepository: rootShelfRepository,
		materialRepository:  materialRepository,
		blockPackRepository: blockPackRepository,
	}
}

/* ============================== Auxiliary Functions ============================== */

func newSubShelfResponseDto(subShelf sschemas.SubShelf) csubshelves.SubShelfResponseDto {
	return csubshelves.SubShelfResponseDto{
		Id:             subShelf.Id,
		Name:           subShelf.Name,
		RootShelfId:    subShelf.RootShelfId,
		PrevSubShelfId: subShelf.PrevSubShelfId,
		Path:           []uuid.UUID(subShelf.Path),
		DeletedAt:      subShelf.DeletedAt,
		UpdatedAt:      subShelf.UpdatedAt,
		CreatedAt:      subShelf.CreatedAt,
	}
}

/* ============================== Service Methods for SubShelf ============================== */

func (s *SubShelfService) GetMySubShelfById(
	ctx context.Context, requestDto *csubshelves.GetMySubShelfByIdRequestDto,
) (*csubshelves.GetMySubShelfByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
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

	subShelf, exception := s.subShelfRepository.GetOneById(
		requestDto.Param.SubShelfId,
		actorUserId,
		nil,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(onlyDeleted),
	)
	if exception != nil {
		return nil, exception
	}

	responseDto := newSubShelfResponseDto(*subShelf)
	return &responseDto, nil
}

func (s *SubShelfService) GetMySubShelvesByPrevSubShelfId(
	ctx context.Context, requestDto *csubshelves.GetMySubShelvesByPrevSubShelfIdRequestDto,
) (*csubshelves.GetMySubShelvesByPrevSubShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
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

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	responseDto := make(csubshelves.GetMySubShelvesByPrevSubShelfIdResponseDto, 0)
	subQuery := db.Model(&sschemas.UsersToShelves{}).
		Select("1").
		Where(`root_shelf_id = "SubShelfTable".root_shelf_id AND user_id = ? AND permission IN ?`,
			actorUserId, allowedPermissions,
		)
	var subShelves []sschemas.SubShelf
	result := db.Model(&sschemas.SubShelf{}).
		Where("prev_sub_shelf_id = ? AND EXISTS (?)", requestDto.Param.PrevSubShelfId, subQuery).
		Scopes(sscopes.NewSubShelfScope().FilterOnlyDeleted(onlyDeleted)).
		Order(`"SubShelfTable".name ASC`).
		Limit(int(data.MaxSubShelvesOfSubShelf)).
		Find(&subShelves)
	if err := result.Error; err != nil {
		return nil, apiexceptions.NewShelfException().NotFound().WithOrigin(err)
	}
	for _, subShelf := range subShelves {
		responseDto = append(responseDto, newSubShelfResponseDto(subShelf))
	}
	return &responseDto, nil
}

func (s *SubShelfService) GetAllMySubShelvesByRootShelfId(
	ctx context.Context, requestDto *csubshelves.GetAllMySubShelvesByRootShelfIdRequestDto,
) (*csubshelves.GetAllMySubShelvesByRootShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
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

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	var subShelves []sschemas.SubShelf
	subQuery := db.Model(&sschemas.UsersToShelves{}).
		Select("1").
		Where(`root_shelf_id = "SubShelfTable".root_shelf_id AND user_id = ? AND permission IN ?`,
			actorUserId, allowedPermissions,
		)
	result := db.Model(&sschemas.SubShelf{}).
		Where("root_shelf_id = ? AND EXISTS (?)",
			requestDto.Param.RootShelfId, subQuery,
		).Scopes(sscopes.NewSubShelfScope().FilterOnlyDeleted(onlyDeleted)).
		Order(`"SubShelfTable".name ASC`).
		Limit(int(data.MaxSubShelvesOfSubShelf)).
		Find(&subShelves)
	if err := result.Error; err != nil {
		return nil, apiexceptions.NewShelfException().NotFound().WithOrigin(err)
	}

	responseDto := make(csubshelves.GetAllMySubShelvesByRootShelfIdResponseDto, 0, len(subShelves))
	for _, subShelf := range subShelves {
		responseDto = append(responseDto, newSubShelfResponseDto(subShelf))
	}
	return &responseDto, nil
}

func (s *SubShelfService) GetMySubShelvesAndItemsByPrevSubShelfId(
	ctx context.Context, requestDto *csubshelves.GetMySubShelvesAndItemsByPrevSubShelfIdRequestDto,
) (*csubshelves.GetMySubShelvesAndItemsByPrevSubShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}

	db := s.db.WithContext(ctx)

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
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

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	resDto := csubshelves.GetMySubShelvesAndItemsByPrevSubShelfIdResponseDto{}
	subQuery := db.Model(&sschemas.UsersToShelves{}).
		Select("1").
		Where(`root_shelf_id = "SubShelfTable".root_shelf_id AND user_id = ? AND permission IN ?`,
			actorUserId, allowedPermissions,
		)
	var subShelves []sschemas.SubShelf
	resultOfGettingSubShelves := db.Model(&sschemas.SubShelf{}).
		Where("prev_sub_shelf_id = ? AND EXISTS (?)",
			requestDto.Param.PrevSubShelfId, subQuery,
		).Scopes(sscopes.NewSubShelfScope().FilterOnlyDeleted(onlyDeleted)).
		Order(`"SubShelfTable".name ASC`).
		Limit(int(data.MaxSubShelvesOfSubShelf)).
		Find(&subShelves)
	if err := resultOfGettingSubShelves.Error; err != nil {
		return nil, apiexceptions.NewShelfException().NotFound().WithOrigin(err)
	}
	for _, subShelf := range subShelves {
		resDto.SubShelves = append(resDto.SubShelves, newSubShelfResponseDto(subShelf))
	}

	materials := []sschemas.Material{}
	resultOfGettingMaterials := db.Model(&sschemas.Material{}).
		Joins(`LEFT JOIN "SubShelfTable" ss ON "MaterialTable".parent_sub_shelf_id = ss.id`).
		Joins(`LEFT JOIN "UsersToShelvesTable" uts ON ss.root_shelf_id = uts.root_shelf_id`).
		Where("ss.id = ? AND uts.user_id = ? AND uts.permission IN ?",
			requestDto.Param.PrevSubShelfId,
			actorUserId,
			allowedPermissions,
		).Scopes(sscopes.NewMaterialScope().FilterOnlyDeleted(onlyDeleted)).
		Order(`"MaterialTable".name ASC`).
		Limit(int(data.MaxMaterialsOfSubShelf)).
		Find(&materials)
	if err := resultOfGettingMaterials.Error; err != nil {
		return nil, apiexceptions.NewMaterialException().NotFound().WithOrigin(err)
	}

	for _, material := range materials {
		downloadURL, err := s.storage.PresignGetObjectByKey(ctx, material.ContentKey, nil)
		if err != nil {
			return nil, apiexceptions.NewStorageException().FailedToPresignedGetObject(material.ContentKey).WithOrigin(err)
		}
		resDto.Materials = append(resDto.Materials, csubshelves.SubShelfMaterialResponseDto{
			Id:               material.Id,
			ParentSubShelfId: material.ParentSubShelfId,
			Name:             material.Name,
			Size:             material.Size,
			ContentType:      material.ContentType.String(),
			ParseMediaType:   material.ParseMediaType,
			DownloadUrl:      downloadURL,
			DeletedAt:        material.DeletedAt,
			UpdatedAt:        material.UpdatedAt,
			CreatedAt:        material.CreatedAt,
		})
	}

	var blockPacks []cblockpacks.GetMyBlockPackByIdResponseDto
	resultOfGettingBlockPacks := db.Model(&sschemas.BlockPack{}).
		Joins(`LEFT JOIN "SubShelfTable" ss ON "BlockPackTable".parent_sub_shelf_id = ss.id`).
		Joins(`LEFT JOIN "UsersToShelvesTable" uts ON ss.root_shelf_id = uts.root_shelf_id`).
		Where("ss.id = ? AND uts.user_id = ? AND uts.permission IN ?",
			requestDto.Param.PrevSubShelfId,
			actorUserId,
			allowedPermissions,
		).Scopes(sscopes.NewBlockPackScope().FilterOnlyDeleted(onlyDeleted)).
		Order(`"BlockPackTable".name ASC`).
		Limit(int(data.MaxBlockPackOfSubShelf)).
		Scan(&blockPacks)
	if err := resultOfGettingBlockPacks.Error; err != nil {
		return nil, apiexceptions.NewBlockPackException().NotFound().WithOrigin(err)
	}

	for _, blockPack := range blockPacks {
		var icon *string
		if blockPack.Icon != nil {
			value := string(*blockPack.Icon)
			icon = &value
		}
		resDto.BlockPacks = append(resDto.BlockPacks, csubshelves.SubShelfBlockPackResponseDto{
			Id:                     blockPack.Id,
			ParentSubShelfId:       blockPack.ParentSubShelfId,
			Name:                   blockPack.Name,
			Icon:                   icon,
			HeaderBackgroundUrl:    blockPack.HeaderBackgroundURL,
			BlockCount:             blockPack.BlockCount,
			LastUpdateSequence:     blockPack.LastUpdateSequence,
			CompactedUntilSequence: blockPack.CompactedUntilSequence,
			ProjectedUntilSequence: blockPack.ProjectedUntilSequence,
			IsProjectionCurrent:    blockPack.IsProjectionCurrent,
			DeletedAt:              blockPack.DeletedAt,
			UpdatedAt:              blockPack.UpdatedAt,
			CreatedAt:              blockPack.CreatedAt,
		})
	}

	return &resDto, nil
}

func (s *SubShelfService) CreateSubShelfByRootShelfId(
	ctx context.Context, requestDto *csubshelves.CreateSubShelfByRootShelfIdRequestDto,
) (*csubshelves.CreateSubShelfByRootShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	newSubShelfId, exception := s.subShelfRepository.CreateOneByRootShelfId(
		requestDto.Body.RootShelfId,
		actorUserId,
		sinputs.CreateSubShelfInput{
			Id:             requestDto.Body.Id,
			Name:           requestDto.Body.Name,
			PrevSubShelfId: requestDto.Body.PrevSubShelfId,
		},
		srepositories.WithDB(s.db.WithContext(ctx)),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &csubshelves.CreateSubShelfByRootShelfIdResponseDto{
		Id:        *newSubShelfId,
		CreatedAt: time.Now(),
	}, nil
}

func (s *SubShelfService) CreateSubShelvesByRootShelfIds(
	ctx context.Context, requestDto *csubshelves.CreateSubShelvesByRootShelfIdsRequestDto,
) (*csubshelves.CreateSubShelvesByRootShelfIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
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

	input := make([]sinputs.CreateSubShelfByRootShelfIdInput, len(requestDto.Body.CreatedSubShelves))
	for index, createdSubShelf := range requestDto.Body.CreatedSubShelves {
		input[index] = sinputs.CreateSubShelfByRootShelfIdInput{
			Id:             createdSubShelf.Id,
			RootShelfId:    createdSubShelf.RootShelfId,
			PrevSubShelfId: createdSubShelf.PrevSubShelfId,
			Name:           createdSubShelf.Name,
		}
	}
	newSubShelfIds, exception := s.subShelfRepository.CreateManyByRootShelfIds(
		actorUserId,
		input,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &csubshelves.CreateSubShelvesByRootShelfIdsResponseDto{
		Ids:       newSubShelfIds,
		CreatedAt: time.Now(),
	}, nil
}

func (s *SubShelfService) UpdateMySubShelfById(
	ctx context.Context, requestDto *csubshelves.UpdateMySubShelfByIdRequestDto,
) (*csubshelves.UpdateMySubShelfByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
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

	subShelf, exception := s.subShelfRepository.UpdateOneById(
		requestDto.Param.SubShelfId,
		actorUserId,
		sinputs.PartialUpdateSubShelfInput{
			Values: sinputs.UpdateSubShelfInput{
				Name: requestDto.Body.Values.Name,
			},
			SetNull: requestDto.Body.SetNull,
		},
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &csubshelves.UpdateMySubShelfByIdResponseDto{
		UpdatedAt: subShelf.UpdatedAt,
	}, nil
}

func (s *SubShelfService) UpdateMySubShelvesByIds(
	ctx context.Context, requestDto *csubshelves.UpdateMySubShelvesByIdsRequestDto,
) (*csubshelves.UpdateMySubShelvesByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
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

	input := make([]sinputs.UpdateSubShelfByIdInput, len(requestDto.Body.UpdatedSubShelves))
	for index, updatedSubShelf := range requestDto.Body.UpdatedSubShelves {
		input[index] = sinputs.UpdateSubShelfByIdInput{
			Id: updatedSubShelf.SubShelfId,
			PartialUpdateInput: sinputs.PartialUpdateInput[sinputs.UpdateSubShelfInput]{
				Values: sinputs.UpdateSubShelfInput{
					Name: updatedSubShelf.Values.Name,
				},
				SetNull: updatedSubShelf.SetNull,
			},
		}
	}
	exception = s.subShelfRepository.UpdateManyByIds(
		actorUserId,
		input,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &csubshelves.UpdateMySubShelvesByIdsResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *SubShelfService) MoveMySubShelfByRootShelfId(
	ctx context.Context, requestDto *csubshelves.MoveMySubShelfByRootShelfIdRequestDto,
) (*csubshelves.MoveMySubShelfByRootShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	if requestDto.Body.DestinationSubShelfId != nil &&
		requestDto.Body.SourceSubShelfId == *requestDto.Body.DestinationSubShelfId {
		return nil, apiexceptions.NewShelfException().NoChanges()
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

	from, exception := s.subShelfRepository.CheckPermissionAndGetOneById(
		requestDto.Body.SourceSubShelfId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception = cexceptions.Cover(exception, []cexceptions.Pair{
		{First: from.RootShelfId != requestDto.Body.SourceRootShelfId, Second: apiexceptions.NewShelfException().NotFound()},
	}); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPackIds, exception := s.blockPackRepository.GetIdsBySubShelfIdsAndDescendants(
		[]uuid.UUID{from.Id},
		srepositories.WithTransactionDB(tx),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if requestDto.Body.DestinationSubShelfId != nil {
		to, exception := s.subShelfRepository.CheckPermissionAndGetOneById(
			*requestDto.Body.DestinationSubShelfId,
			actorUserId,
			nil,
			allowedPermissions,
			srepositories.WithTransactionDB(tx),
			srepositories.WithAllowedPermissions(allowedPermissions),
			srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
			srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
		)
		if exception = cexceptions.Cover(exception, []cexceptions.Pair{
			{First: to.RootShelfId != requestDto.Body.DestinationRootShelfId, Second: apiexceptions.NewShelfException().NotFound()},
			{
				First: len(from.Path)+len(to.Path) > int(data.MaxSubShelvesOfRootShelf),
				Second: apiexceptions.NewShelfException().MaximumDepthExceeded(
					int32(len(from.Path)+len(to.Path)),
					data.MaxSubShelvesOfRootShelf,
				),
			},
		}); exception != nil {
			tx.Rollback()
			return nil, exception
		}

		// check if to.Path contain any from.Id, if it's true, then it means the user is trying to move the sub shelf to its child
		for _, parent := range to.Path {
			if parent == requestDto.Body.SourceSubShelfId {
				tx.Rollback()
				return nil, apiexceptions.NewShelfException().InsertParentIntoItsChildren(
					requestDto.Body.DestinationSubShelfId,
					requestDto.Body.SourceSubShelfId,
				)
			}
		}

		to.Path = append(to.Path, to.Id)
		result := tx.Exec(`
			UPDATE "SubShelfTable"
			SET "root_shelf_id" = ?, "prev_sub_shelf_id" = ?, "path" = ?, "updated_at" = NOW()
			WHERE id = ? AND deleted_at IS NULL`,
			requestDto.Body.DestinationRootShelfId, requestDto.Body.DestinationSubShelfId, pg.Array(to.Path),
			requestDto.Body.SourceSubShelfId,
		)
		if err := result.Error; err != nil {
			tx.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(err)
		}
	} else {
		result := tx.Exec(`
			UPDATE "SubShelfTable"
			SET "root_shelf_id" = ?, "prev_sub_shelf_id" = ?, "path" = ?, "updated_at" = NOW()
			WHERE id = ? AND deleted_at IS NULL`,
			requestDto.Body.DestinationRootShelfId, nil, pg.Array([]uuid.UUID{}), requestDto.Body.SourceSubShelfId,
		)
		if err := result.Error; err != nil {
			tx.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(err)
		}
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		requestDto.Body.SourceSubShelfId.String(),
		blockPackIds,
		nil,
		coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"MoveMySubShelfByRootShelfId",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &csubshelves.MoveMySubShelfByRootShelfIdResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *SubShelfService) MoveMySubShelvesByRootShelfId(
	ctx context.Context, requestDto *csubshelves.MoveMySubShelvesByRootShelfIdRequestDto,
) (*csubshelves.MoveMySubShelvesByRootShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
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

	froms, exception := s.subShelfRepository.CheckPermissionsAndGetManyByIds(
		requestDto.Body.SourceSubShelfIds,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	for _, from := range froms {
		if from.RootShelfId != requestDto.Body.SourceRootShelfId {
			tx.Rollback()
			return nil, apiexceptions.NewShelfException().NotFound()
		}
	}
	sourceSubShelfIds := make([]uuid.UUID, len(froms))
	for index, from := range froms {
		sourceSubShelfIds[index] = from.Id
	}
	blockPackIds, exception := s.blockPackRepository.GetIdsBySubShelfIdsAndDescendants(
		sourceSubShelfIds,
		srepositories.WithTransactionDB(tx),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if requestDto.Body.DestinationSubShelfId != nil {
		to, exception := s.subShelfRepository.CheckPermissionAndGetOneById(
			*requestDto.Body.DestinationSubShelfId,
			actorUserId,
			nil,
			allowedPermissions,
			srepositories.WithTransactionDB(tx),
			srepositories.WithAllowedPermissions(allowedPermissions),
			srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
			srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
		)
		if exception = cexceptions.Cover(exception, []cexceptions.Pair{
			{First: to.RootShelfId != requestDto.Body.DestinationRootShelfId, Second: apiexceptions.NewShelfException().NotFound()},
		}); exception != nil {
			tx.Rollback()
			return nil, exception
		}

		if to.Path == nil {
			to.Path = []uuid.UUID{}
		}

		sourceSubShelfIdMap := make(map[uuid.UUID]bool)
		for _, from := range froms {
			if len(from.Path)+len(to.Path) > int(data.MaxSubShelvesOfRootShelf) {
				apiexceptions.NewShelfException().MaximumDepthExceeded(
					int32(len(from.Path)+len(to.Path)),
					data.MaxSubShelvesOfRootShelf,
				)
				// sourceSubShelfIdMap[from.Id] = false
			} else if from.Id == to.Id { // handling inserting node to itself here
				apiexceptions.NewShelfException().InsertParentIntoItsChildren(to.Id, from.Id)
				// sourceSubShelfIdMap[from.Id] = false
			} else {
				sourceSubShelfIdMap[from.Id] = true
			}
		}

		for _, parentId := range to.Path { // handling inserting node to its children here
			if sourceSubShelfIdMap[parentId] {
				apiexceptions.NewShelfException().InsertParentIntoItsChildren(
					requestDto.Body.DestinationSubShelfId,
					parentId,
				)
				sourceSubShelfIdMap[parentId] = false // has to mark the sub shelf as invalid
			}
		}

		validSourceSubShelfIds := []uuid.UUID{}
		for sourceSubShelfId, exist := range sourceSubShelfIdMap {
			if exist {
				validSourceSubShelfIds = append(validSourceSubShelfIds, sourceSubShelfId)
			}
		}

		to.Path = append(to.Path, to.Id)
		result := tx.Exec(`
			UPDATE "SubShelfTable"
			SET "root_shelf_id" = ?, "prev_sub_shelf_id" = ?, "path" = ?, "updated_at" = NOW()
			WHERE id IN ? AND deleted_at IS NULL`,
			requestDto.Body.DestinationRootShelfId, requestDto.Body.DestinationSubShelfId, pg.Array(to.Path), validSourceSubShelfIds,
		)
		if err := result.Error; err != nil {
			tx.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(err)
		}
	} else {
		validSourceSubShelfIds := []uuid.UUID{}
		for _, from := range froms {
			validSourceSubShelfIds = append(validSourceSubShelfIds, from.Id)
		}

		result := tx.Exec(`
			UPDATE "SubShelfTable"
			SET "root_shelf_id" = ?, "prev_sub_shelf_id" = ?, "path" = ?, "updated_at" = NOW()
			WHERE id IN ? AND deleted_at IS NULL`,
			requestDto.Body.DestinationRootShelfId, nil, pg.Array([]uuid.UUID{}), validSourceSubShelfIds,
		)
		if err := result.Error; err != nil {
			tx.Rollback()
			return nil, apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(err)
		}
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		"sub-shelf-bulk-move",
		blockPackIds,
		nil,
		coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"MoveMySubShelvesByRootShelfId",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &csubshelves.MoveMySubShelvesByRootShelfIdResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *SubShelfService) MoveMySubShelvesByRootShelfIds(
	ctx context.Context, requestDto *csubshelves.MoveMySubShelvesByRootShelfIdsRequestDto,
) (*csubshelves.MoveMySubShelvesByRootShelfIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
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

	var destinationSubShelfIds []uuid.UUID
	var sourceSubShelfIds []uuid.UUID
	var rootShelfIds []uuid.UUID
	hasSubShelfIdSeen := make(map[uuid.UUID]bool)                               // use to do the first cleaning duplicated sub shelves in requestDto
	destinationSubShelfIdToSourceSubShelfIds := make(map[uuid.UUID][]uuid.UUID) // destination sub shelf -> { all source sub shelves... }
	for _, movedSubShelf := range requestDto.Body.MovedSubShelves {
		if movedSubShelf.DestinationSubShelfId != nil {
			destinationSubShelfIds = append(destinationSubShelfIds, *movedSubShelf.DestinationSubShelfId)
			for _, sourceSubShelfId := range movedSubShelf.SourceSubShelfIds {
				if !hasSubShelfIdSeen[sourceSubShelfId] {
					hasSubShelfIdSeen[sourceSubShelfId] = true
					sourceSubShelfIds = append(sourceSubShelfIds, sourceSubShelfId)
					destinationSubShelfIdToSourceSubShelfIds[*movedSubShelf.DestinationSubShelfId] = append(destinationSubShelfIdToSourceSubShelfIds[*movedSubShelf.DestinationSubShelfId], sourceSubShelfId)
				}
			}
		} else {
			for _, sourceSubShelfId := range movedSubShelf.SourceSubShelfIds {
				if !hasSubShelfIdSeen[sourceSubShelfId] {
					hasSubShelfIdSeen[sourceSubShelfId] = true
					sourceSubShelfIds = append(sourceSubShelfIds, sourceSubShelfId)
					destinationSubShelfIdToSourceSubShelfIds[uuid.Nil] = append(destinationSubShelfIdToSourceSubShelfIds[uuid.Nil], sourceSubShelfId)
				}
			}
		}
		rootShelfIds = append(rootShelfIds, movedSubShelf.SourceRootShelfId)
		rootShelfIds = append(rootShelfIds, movedSubShelf.DestinationRootShelfId)
	}

	isRootShelfValid := make(map[uuid.UUID]bool)
	validRootShelves, _, exception := s.rootShelfRepository.CheckPermissionsAndGetManyByIds(
		rootShelfIds,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	for _, validRootShelf := range validRootShelves {
		isRootShelfValid[validRootShelf.Id] = true
	}

	validSourceSubShelfMap := make(map[uuid.UUID]sschemas.SubShelf)
	validSourceSubShelves, exception := s.subShelfRepository.CheckPermissionsAndGetManyByIds(
		sourceSubShelfIds,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	for _, validSourceSubShelf := range validSourceSubShelves {
		if isRootShelfValid[validSourceSubShelf.RootShelfId] {
			validSourceSubShelfMap[validSourceSubShelf.Id] = validSourceSubShelf
		}
	}

	var finalValidDestinationSubShelves []sschemas.SubShelf
	validDestinationSubShelves, exception := s.subShelfRepository.CheckPermissionsAndGetManyByIds(
		destinationSubShelfIds,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithLockingStrength(srepositories.LockingStrengthNoKeyUpdate),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	for _, validDestinationSubShelf := range validDestinationSubShelves {
		if isRootShelfValid[validDestinationSubShelf.RootShelfId] {
			finalValidDestinationSubShelves = append(finalValidDestinationSubShelves, validDestinationSubShelf)
		}
	}

	sourceSubShelfIdMap := make(map[uuid.UUID]bool)
	for _, to := range finalValidDestinationSubShelves {
		sourceSubShelfIds, exist := destinationSubShelfIdToSourceSubShelfIds[to.Id] // get the destination of the current sub shelf
		if !exist {                                                                 // if it does not exist a direction from the current sub shelf to the source
			continue // it means the current sub shelf is either an invalid sub shelf or have no source sub shelf pointing to it, then we just continue on other sub shelves
		}

		for _, sourceSubShelfId := range sourceSubShelfIds {
			from, exist := validSourceSubShelfMap[sourceSubShelfId]
			if !exist {
				continue
			}

			if len(from.Path)+len(to.Path) > int(data.MaxSubShelvesOfRootShelf) {
				apiexceptions.NewShelfException().MaximumDepthExceeded(
					int32(len(from.Path)+len(to.Path)),
					data.MaxSubShelvesOfRootShelf,
				)
				// sourceSubShelfIdMap[sourceSubShelfId] = false
			} else if from.Id == to.Id { // handling inserting node to itself here
				apiexceptions.NewShelfException().InsertParentIntoItsChildren(to.Id, from.Id)
				// sourceSubShelfIdMap[sourceSubShelfId] = false
			} else {
				sourceSubShelfIdMap[from.Id] = true
			}
		}

		for _, parentId := range to.Path { // handling inserting node to its children here
			// once we iterated through the source sub shelves of the current destination sub shelf
			// we have the complete source sub shelf recorded in the sourceSubShelfIdMap now
			if sourceSubShelfIdMap[parentId] {
				apiexceptions.NewShelfException().InsertParentIntoItsChildren(
					to.Id,
					parentId,
				)
				sourceSubShelfIdMap[parentId] = false
			}
		}
	}

	validSourceSubShelfIds := make([]uuid.UUID, 0, len(validSourceSubShelfMap))
	for sourceSubShelfId := range validSourceSubShelfMap {
		validSourceSubShelfIds = append(validSourceSubShelfIds, sourceSubShelfId)
	}
	blockPackIds, exception := s.blockPackRepository.GetIdsBySubShelfIdsAndDescendants(
		validSourceSubShelfIds,
		srepositories.WithTransactionDB(tx),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	var valuePlaceholders []string
	var valueArgs []interface{}
	for _, to := range finalValidDestinationSubShelves {
		sourceSubShelfIds, exist := destinationSubShelfIdToSourceSubShelfIds[to.Id]
		if !exist {
			continue
		}

		for _, sourceSubShelfId := range sourceSubShelfIds {
			from, exist := validSourceSubShelfMap[sourceSubShelfId]
			if !exist {
				continue
			}

			path := to.Path
			path = append(path, to.Id)
			valuePlaceholders = append(valuePlaceholders, "(?::uuid, ?::uuid, ?::uuid, ?::uuid[])")
			valueArgs = append(valueArgs,
				from.Id,
				to.Id,
				to.RootShelfId,
				path,
			)
		}
	}

	sql := fmt.Sprintf(`
		UPDATE "SubShelfTable" AS s
		SET
			root_shelf_id = COALESCE(s.root_shelf_id, v.dest_root_shelf_id::uuid),
			prev_sub_shelf_id = v.dest_sub_shelf_id::uuid,
			path = COALESCE(s.path, v.path::uuid[]),
			updated_at = NOW()
		FROM (VALUES %s) AS v(source_id, dest_sub_shelf_id, dest_root_shelf_id, path)
		WHERE s.id = v.source_id::uuid AND s.deleted_at IS NULL
	`, strings.Join(valuePlaceholders, ","))
	result := tx.Exec(sql, valueArgs...)
	if exception := cexceptions.Cover(nil, []cexceptions.Pair{
		{First: result.Error != nil, Second: apiexceptions.NewShelfException().FailedToUpdate().WithOrigin(result.Error)},
		{First: result.RowsAffected == 0, Second: apiexceptions.NewShelfException().NoChanges()},
	}); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		"sub-shelf-multi-root-move",
		blockPackIds,
		nil,
		coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"MoveMySubShelvesByRootShelfIds",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, apiexceptions.NewShelfException().FailedToCommitTransaction().WithOrigin(err)
	}

	return &csubshelves.MoveMySubShelvesByRootShelfIdsResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *SubShelfService) RestoreMySubShelfById(
	ctx context.Context, requestDto *csubshelves.RestoreMySubShelfByIdRequestDto,
) (*csubshelves.RestoreMySubShelfByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
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

	restoredSubShelf, exception := s.subShelfRepository.RestoreSoftDeletedOneById(
		requestDto.Param.SubShelfId,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	responseDto := newSubShelfResponseDto(*restoredSubShelf)
	return &responseDto, nil
}

func (s *SubShelfService) RestoreMySubShelvesByIds(
	ctx context.Context, requestDto *csubshelves.RestoreMySubShelvesByIdsRequestDto,
) (*csubshelves.RestoreMySubShelvesByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
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

	restoredSubShelves, exception := s.subShelfRepository.RestoreSoftDeletedManyByIds(
		requestDto.Body.SubShelfIds,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	resDto := csubshelves.RestoreMySubShelvesByIdsResponseDto{}
	for _, restoredSubShelf := range restoredSubShelves {
		resDto = append(resDto, newSubShelfResponseDto(restoredSubShelf))
	}
	return &resDto, nil
}

func (s *SubShelfService) DeleteMySubShelfById(
	ctx context.Context, requestDto *csubshelves.DeleteMySubShelfByIdRequestDto,
) (*csubshelves.DeleteMySubShelfByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, cexceptions.New(
			"TransactionBeginFailed",
			"SubShelf",
			"DeleteMySubShelfById",
			"Failed to begin the sub shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	blockPackIds, exception := s.blockPackRepository.GetIdsByParentSubShelfIds(
		[]uuid.UUID{requestDto.Param.SubShelfId},
		srepositories.WithTransactionDB(tx),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	exception = s.subShelfRepository.SoftDeleteOneById(
		requestDto.Param.SubShelfId,
		actorUserId,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		requestDto.Param.SubShelfId.String(),
		blockPackIds,
		nil,
		coreevents.BlockPackAccessRevocationReason_ResourceUnavailable,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMySubShelfById",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"SubShelf",
			"DeleteMySubShelfById",
			"Failed to commit the sub shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &csubshelves.DeleteMySubShelfByIdResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *SubShelfService) DeleteMySubShelvesByIds(
	ctx context.Context, requestDto *csubshelves.DeleteMySubShelvesByIdsRequestDto,
) (*csubshelves.DeleteMySubShelvesByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, cexceptions.New(
			"TransactionBeginFailed",
			"SubShelf",
			"DeleteMySubShelvesByIds",
			"Failed to begin the sub shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	blockPackIds, exception := s.blockPackRepository.GetIdsByParentSubShelfIds(
		requestDto.Body.SubShelfIds,
		srepositories.WithTransactionDB(tx),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	exception = s.subShelfRepository.SoftDeleteManyByIds(
		requestDto.Body.SubShelfIds,
		actorUserId,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		"sub-shelf-bulk-delete",
		blockPackIds,
		nil,
		coreevents.BlockPackAccessRevocationReason_ResourceUnavailable,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMySubShelvesByIds",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"SubShelf",
			"DeleteMySubShelvesByIds",
			"Failed to commit the sub shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &csubshelves.DeleteMySubShelvesByIdsResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

/* ============================== Service Methods for GraphQL SubShelf ============================== */

func (s *SubShelfService) SearchPrivateSubShelves(
	ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchSubShelfInput,
) (*cgqlmodels.SearchSubShelfConnection, *cexceptions.Exception) {
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

	query := db.Model(&sschemas.SubShelf{}).
		Select(`"SubShelfTable".*`).
		Joins(`INNER JOIN "UsersToShelvesTable" uts ON "SubShelfTable".root_shelf_id = uts.root_shelf_id`).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, allowedPermissions).
		Scopes(s.subShelfScope.FilterOnlyDeleted(onlyDeleted))

	if gqlInput.RootShelfID != nil {
		query = query.Where(
			`"SubShelfTable".root_shelf_id = ?`,
			*gqlInput.RootShelfID,
		)
	}

	if gqlInput.PrevSubShelfID != nil {
		query = query.Where(
			`"SubShelfTable".prev_sub_shelf_id = ?`,
			*gqlInput.PrevSubShelfID,
		)
	}

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			`"SubShelfTable".name ILIKE ?`,
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := ssearchcursor.Decode[cgqlmodels.SearchSubShelfCursorFields](*gqlInput.After)
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToDecode().WithOrigin(err)
		}

		query = query.Where(`"SubShelfTable".id > ?`, searchCursor.Fields.ID)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		cending := cgqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == cgqlmodels.SearchSortOrderDesc {
			cending = cgqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case cgqlmodels.SearchSubShelfSortByName:
			query = query.Order(`"SubShelfTable".name ` + cending).
				Order(`cardinality("SubShelfTable".path) ` + cending).
				Order(`"SubShelfTable".updated_at ` + cending).
				Order(`"SubShelfTable".created_at ` + cending)
		case cgqlmodels.SearchSubShelfSortByPathLength:
			query = query.Order(`cardinality("SubShelfTable".path) ` + cending).
				Order(`"SubShelfTable".name ` + cending).
				Order(`"SubShelfTable".updated_at ` + cending).
				Order(`"SubShelfTable".created_at ` + cending)
		case cgqlmodels.SearchSubShelfSortByLastUpdate:
			query = query.Order(`"SubShelfTable".updated_at ` + cending).
				Order(`"SubShelfTable".name ` + cending).
				Order(`cardinality("SubShelfTable".path) ` + cending).
				Order(`"SubShelfTable".created_at ` + cending)
		case cgqlmodels.SearchSubShelfSortByCreatedAt:
			query = query.Order(`"SubShelfTable".created_at ` + cending).
				Order(`"SubShelfTable".name ` + cending).
				Order(`cardinality("SubShelfTable".path) ` + cending).
				Order(`"SubShelfTable".updated_at ` + cending)
		default:
			query = query.Order(`"SubShelfTable".name ` + cending).
				Order(`cardinality("SubShelfTable".path) ` + cending).
				Order(`"SubShelfTable".updated_at ` + cending).
				Order(`"SubShelfTable".created_at ` + cending)
		}
	}

	limit := sconstants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, sconstants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var subShelves []sschemas.SubShelf
	if err := query.Scopes(s.subShelfScope.IncludePreloads(
		[]sschemas.SubShelfRelation{
			sschemas.SubShelfRelation_NextSubShelves,
			sschemas.SubShelfRelation_Items,
		},
	)).Find(&subShelves).Error; err != nil {
		return nil, apiexceptions.NewShelfException().NotFound().WithOrigin(err)
	}

	hasNextPage := len(subShelves) > limit
	searchEdges := make([]*cgqlmodels.SearchSubShelfEdge, len(subShelves))

	for index, subShelf := range subShelves {
		searchCursor := ssearchcursor.SearchCursor[cgqlmodels.SearchSubShelfCursorFields]{
			Fields: cgqlmodels.SearchSubShelfCursorFields{
				ID: subShelf.Id,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, apiexceptions.NewSearchException().FailedToEncode().WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, apiexceptions.NewSearchException().FailedToUnmarshalSearchCursor()
		}

		searchEdges[index] = &cgqlmodels.SearchSubShelfEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                subShelf.ToPrivateSubShelf(),
		}
	}

	if hasNextPage {
		searchEdges = searchEdges[:limit]
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

	return &cgqlmodels.SearchSubShelfConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
