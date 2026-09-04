package material

import (
	"bytes"
	"context"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	pg "github.com/lib/pq"
	"gorm.io/gorm"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/materials"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	sconstants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	ssearchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"
	slogs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"
	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"
	sschemas "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/schemas"
	sscopes "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/scopes"
	stypes "github.com/HiIamJeff67/notegic-backend/shared/types"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres"
	materialsql "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/postgres/sqls/material"
	storage "github.com/HiIamJeff67/notegic-backend/runtimes/core/data/storage"
	apiexceptions "github.com/HiIamJeff67/notegic-backend/runtimes/core/exceptions"
)

type MaterialServiceInterface interface {
	GetMyMaterialById(ctx context.Context, requestDto *capi.GetMyMaterialByIdRequestDto) (*capi.GetMyMaterialByIdResponseDto, *cexceptions.Exception)
	GetMyMaterialAndItsParentById(ctx context.Context, requestDto *capi.GetMyMaterialAndItsParentByIdRequestDto) (*capi.GetMyMaterialAndItsParentByIdResponseDto, *cexceptions.Exception)
	GetMyMaterialsByParentSubShelfId(ctx context.Context, requestDto *capi.GetMyMaterialsByParentSubShelfIdRequestDto) (*capi.GetMyMaterialsByParentSubShelfIdResponseDto, *cexceptions.Exception)
	GetMyMaterialsByRootShelfId(ctx context.Context, requestDto *capi.GetMyMaterialsByRootShelfIdRequestDto) (*capi.GetMyMaterialsByRootShelfIdResponseDto, *cexceptions.Exception)
	CreateMyMaterial(ctx context.Context, requestDto *capi.CreateMyMaterialRequestDto) (*capi.CreateMyMaterialResponseDto, *cexceptions.Exception)
	UpdateMyMaterialById(ctx context.Context, requestDto *capi.UpdateMyMaterialByIdRequestDto) (*capi.UpdateMyMaterialByIdResponseDto, *cexceptions.Exception)
	SaveMyMaterialById(ctx context.Context, requestDto *capi.SaveMyMaterialByIdRequestDto) (*capi.SaveMyMaterialByIdResponseDto, *cexceptions.Exception)
	MoveMyMaterialById(ctx context.Context, requestDto *capi.MoveMyMaterialByIdRequestDto) (*capi.MoveMyMaterialByIdResponseDto, *cexceptions.Exception)
	MoveMyMaterialsByIds(ctx context.Context, requestDto *capi.MoveMyMaterialsByIdsRequestDto) (*capi.MoveMyMaterialsByIdsResponseDto, *cexceptions.Exception)
	RestoreMyMaterialById(ctx context.Context, requestDto *capi.RestoreMyMaterialByIdRequestDto) (*capi.RestoreMyMaterialByIdResponseDto, *cexceptions.Exception)
	RestoreMyMaterialsByIds(ctx context.Context, requestDto *capi.RestoreMyMaterialsByIdsRequestDto) (*capi.RestoreMyMaterialsByIdsResponseDto, *cexceptions.Exception)
	DeleteMyMaterialById(ctx context.Context, requestDto *capi.DeleteMyMaterialByIdRequestDto) (*capi.DeleteMyMaterialByIdResponseDto, *cexceptions.Exception)
	DeleteMyMaterialsByIds(ctx context.Context, requestDto *capi.DeleteMyMaterialsByIdsRequestDto) (*capi.DeleteMyMaterialsByIdsResponseDto, *cexceptions.Exception)

	/* ============================== GraphQL Methods ============================== */
	SearchPrivateMaterials(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchMaterialInput) (*cgqlmodels.SearchMaterialConnection, *cexceptions.Exception)
}

type MaterialService struct {
	validator          *validator.Validate
	db                 *gorm.DB
	storage            storage.StorageInterface
	materialScope      sscopes.MaterialScopeInterface
	subShelfRepository srepositories.SubShelfRepositoryInterface
	materialRepository srepositories.MaterialRepositoryInterface
	storageKeySalt     string
	materialException  apiexceptions.MaterialException
	storageException   apiexceptions.StorageException
	searchException    apiexceptions.SearchException
}

func NewMaterialService(
	validator *validator.Validate,
	db *gorm.DB,
	storage storage.StorageInterface,
	materialScope sscopes.MaterialScopeInterface,
	subShelfRepository srepositories.SubShelfRepositoryInterface,
	materialRepository srepositories.MaterialRepositoryInterface,
	storageKeySalt string,
	materialException apiexceptions.MaterialException,
	storageException apiexceptions.StorageException,
	searchException apiexceptions.SearchException,
) MaterialServiceInterface {
	return &MaterialService{
		validator:          validator,
		db:                 db,
		storage:            storage,
		materialScope:      materialScope,
		subShelfRepository: subShelfRepository,
		materialRepository: materialRepository,
		storageKeySalt:     storageKeySalt,
		materialException:  materialException,
		storageException:   storageException,
		searchException:    searchException,
	}
}

func (s *MaterialService) GetMyMaterialById(
	ctx context.Context, requestDto *capi.GetMyMaterialByIdRequestDto,
) (*capi.GetMyMaterialByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.materialException.InvalidDto().WithOrigin(err)
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

	material, exception := s.materialRepository.GetOneById(
		requestDto.Param.MaterialId,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(onlyDeleted),
	)
	if exception != nil {
		return nil, exception
	}

	downloadURL, err := s.storage.PresignGetObjectByKey(ctx, material.ContentKey, nil)
	if err != nil {
		slogs.NotegicLogger.Error(ctx, err, "Failed to presign Material object")
	}

	return &capi.GetMyMaterialByIdResponseDto{
		Id:               material.Id,
		ParentSubShelfId: material.ParentSubShelfId,
		Name:             material.Name,
		Size:             material.Size,
		ContentType:      material.ContentType,
		ParseMediaType:   material.ParseMediaType,
		DownloadURL:      downloadURL,
		DeletedAt:        material.DeletedAt,
		UpdatedAt:        material.UpdatedAt,
		CreatedAt:        material.CreatedAt,
	}, nil
}

func (s *MaterialService) GetMyMaterialAndItsParentById(
	ctx context.Context, requestDto *capi.GetMyMaterialAndItsParentByIdRequestDto,
) (*capi.GetMyMaterialAndItsParentByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.materialException.InvalidDto().WithOrigin(err)
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

	resDto := capi.GetMyMaterialAndItsParentByIdResponseDto{}
	var contentKey string
	err := db.Raw(materialsql.GetMyMaterialAndItsParentByIdSQL,
		requestDto.Param.MaterialId, actorUserId, pg.Array(allowedPermissions), onlyDeleted,
	).Row().
		Scan(&resDto.Id,
			&resDto.Name,
			&resDto.Size,
			&resDto.ContentType,
			&resDto.ParseMediaType,
			&contentKey,
			&resDto.DeletedAt,
			&resDto.UpdatedAt,
			&resDto.CreatedAt,
			&resDto.RootShelfId,
			&resDto.ParentSubShelfId,
			&resDto.ParentSubShelfName,
			&resDto.ParentSubShelfPrevSubShelfId,
			&resDto.ParentSubShelfPath,
			&resDto.ParentSubShelfDeletedAt,
			&resDto.ParentSubShelfUpdatedAt,
			&resDto.ParentSubShelfCreatedAt,
		)
	if err != nil {
		return nil, s.materialException.NotFound().WithOrigin(err)
	}
	if len(strings.TrimSpace(contentKey)) == 0 {
		return nil, s.materialException.NotFound()
	}

	downloadURL, err := s.storage.PresignGetObjectByKey(ctx, contentKey, nil)
	if err != nil {
		slogs.NotegicLogger.Error(ctx, err, "Failed to presign Material object")
	}
	resDto.DownloadURL = downloadURL // could be empty string

	return &resDto, nil
}

func (s *MaterialService) GetMyMaterialsByParentSubShelfId(
	ctx context.Context, requestDto *capi.GetMyMaterialsByParentSubShelfIdRequestDto,
) (*capi.GetMyMaterialsByParentSubShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.materialException.InvalidDto().WithOrigin(err)
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

	materials := []sschemas.Material{}
	result := db.Model(&sschemas.Material{}).
		Joins(`LEFT JOIN "SubShelfTable" ss ON "MaterialTable".parent_sub_shelf_id = ss.id`).
		Joins(`LEFT JOIN "UsersToShelvesTable" uts ON ss.root_shelf_id = uts.root_shelf_id`).
		Where("ss.id = ? AND uts.user_id = ? AND uts.permission IN ?",
			requestDto.Param.ParentSubShelfId,
			actorUserId,
			allowedPermissions,
		).Scopes(sscopes.NewMaterialScope().FilterOnlyDeleted(onlyDeleted)).
		Order("name ASC").
		Limit(int(data.MaxMaterialsOfSubShelf)).
		Find(&materials)
	if err := result.Error; err != nil {
		return nil, s.materialException.NotFound().WithOrigin(err)
	}

	resDto := capi.GetMyMaterialsByParentSubShelfIdResponseDto{}
	for _, material := range materials {
		downloadURL, err := s.storage.PresignGetObjectByKey(ctx, material.ContentKey, nil)
		if err != nil {
			slogs.NotegicLogger.Error(ctx, err, "Failed to presign Material object")
		}
		resDto = append(resDto, capi.GetMyMaterialByIdResponseDto{
			Id:               material.Id,
			ParentSubShelfId: material.ParentSubShelfId,
			Name:             material.Name,
			Size:             material.Size,
			ContentType:      material.ContentType,
			ParseMediaType:   material.ParseMediaType,
			DownloadURL:      downloadURL,
			DeletedAt:        material.DeletedAt,
			UpdatedAt:        material.UpdatedAt,
			CreatedAt:        material.CreatedAt,
		})
	}

	return &resDto, nil
}

func (s *MaterialService) GetMyMaterialsByRootShelfId(
	ctx context.Context, requestDto *capi.GetMyMaterialsByRootShelfIdRequestDto,
) (*capi.GetMyMaterialsByRootShelfIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.materialException.InvalidDto().WithOrigin(err)
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

	materials := []sschemas.Material{}
	result := db.Model(&sschemas.Material{}).
		Joins(`LEFT JOIN "SubShelfTable" ss ON "MaterialTable".parent_sub_shelf_id = ss.id`).
		Joins(`LEFT JOIN "UsersToShelvesTable" uts ON ss.root_shelf_id = uts.root_shelf_id`).
		Where("ss.root_shelf_id = ? AND uts.user_id = ? AND uts.permission IN ?",
			requestDto.Param.RootShelfId, actorUserId, allowedPermissions,
		).Scopes(sscopes.NewMaterialScope().FilterOnlyDeleted(onlyDeleted)).
		Limit(int(data.MaxMaterialsOfRootShelf)).
		Order("name ASC").
		Find(&materials)
	if err := result.Error; err != nil {
		return nil, s.materialException.NotFound()
	}

	resDto := capi.GetMyMaterialsByRootShelfIdResponseDto{}
	for _, material := range materials {
		downloadURL, err := s.storage.PresignGetObjectByKey(ctx, material.ContentKey, nil)
		if err != nil {
			slogs.NotegicLogger.Error(ctx, err, "Failed to presign Material object")
		}
		resDto = append(resDto, capi.GetMyMaterialByIdResponseDto{
			Id:               material.Id,
			ParentSubShelfId: material.ParentSubShelfId,
			Name:             material.Name,
			Size:             material.Size,
			ContentType:      material.ContentType,
			ParseMediaType:   material.ParseMediaType,
			DownloadURL:      downloadURL,
			DeletedAt:        material.DeletedAt,
			UpdatedAt:        material.UpdatedAt,
			CreatedAt:        material.CreatedAt,
		})
	}

	return &resDto, nil
}

func (s *MaterialService) CreateMyMaterial(
	ctx context.Context, requestDto *capi.CreateMyMaterialRequestDto,
) (*capi.CreateMyMaterialResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.materialException.InvalidDto().WithOrigin(err)
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
	actorUserPublicId, exception := contexts.GetActorUserPublicId(ctx)
	if exception != nil {
		return nil, exception
	}

	newMaterialId := uuid.New()
	newContentKey := s.storage.GetKey(
		actorUserPublicId.String(),
		newMaterialId.String(),
		s.storageKeySalt,
	)
	zeroSize := int64(0)
	_, exception = s.materialRepository.CreateOneBySubShelfId(
		requestDto.Body.ParentSubShelfId,
		actorUserId,
		sinputs.CreateMaterialInput{
			Id:             newMaterialId,
			Name:           requestDto.Body.Name,
			Size:           zeroSize,
			ContentKey:     newContentKey,
			ParseMediaType: "",
		},
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	newContentFile := bytes.NewReader([]byte{})

	object, err := s.storage.NewObject(newContentKey, newContentFile, zeroSize)
	if err != nil {
		return nil, s.storageException.FailedToReadObjectBytes().WithOrigin(err)
	}

	if err := s.storage.PutObjectByKey(ctx, newContentKey, object); err != nil {
		return nil, s.storageException.FailedToPutObject(object).WithOrigin(err)
	}

	return &capi.CreateMyMaterialResponseDto{
		Id:        newMaterialId,
		CreatedAt: time.Now(),
	}, nil
}

func (s *MaterialService) UpdateMyMaterialById(
	ctx context.Context, requestDto *capi.UpdateMyMaterialByIdRequestDto,
) (*capi.UpdateMyMaterialByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.materialException.InvalidDto().WithOrigin(err)
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

	material, exception := s.materialRepository.UpdateOneById(
		requestDto.Param.MaterialId,
		actorUserId,
		sinputs.PartialUpdateMaterialInput{
			Values: sinputs.UpdateMaterialInput{
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

	return &capi.UpdateMyMaterialByIdResponseDto{
		UpdatedAt: material.UpdatedAt,
	}, nil
}

func (s *MaterialService) SaveMyMaterialById(
	ctx context.Context, requestDto *capi.SaveMyMaterialByIdRequestDto,
) (*capi.SaveMyMaterialByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.materialException.InvalidDto().WithOrigin(err)
	}
	// check if there does exist a file in the requestDto
	if requestDto.Body.ContentFile == nil {
		return nil, s.materialException.InvalidDto()
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
	actorUserPublicId, exception := contexts.GetActorUserPublicId(ctx)
	if exception != nil {
		return nil, exception
	}

	partialUpdate := sinputs.PartialUpdateMaterialInput{
		Values: sinputs.UpdateMaterialInput{
			// content key remain the same here
		},
		SetNull: nil,
	}
	var contentKey = s.storage.GetKey(
		actorUserPublicId.String(),
		requestDto.Param.MaterialId.String(),
		s.storageKeySalt,
	)

	fileHeaderSize := int64(len(requestDto.Body.ContentFile))

	// extract the data in it and get its content type, parse media type, and actual size, etc.
	object, err := s.storage.NewObject(contentKey, bytes.NewReader(requestDto.Body.ContentFile), fileHeaderSize)
	if err != nil {
		return nil, s.storageException.FailedToReadObjectBytes().WithOrigin(err)
	}
	if object == nil {
		return nil, s.materialException.CannotGetFileObjects()
	}

	size := object.Size
	contentType, err := cenums.ConvertStringToMaterialContentType(object.ContentType)
	if err != nil {
		return nil, s.materialException.InvalidType(object.ContentType).WithOrigin(err)
	}
	partialUpdate.Values.ParseMediaType = &object.ParseMediaType
	partialUpdate.Values.Size = &size
	partialUpdate.Values.ContentType = contentType

	material, exception := s.materialRepository.UpdateOneById(
		requestDto.Param.MaterialId,
		actorUserId,
		partialUpdate,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	// if there does exist a file, then put the file at the end to ensure the entire operation is consistent
	if err := s.storage.PutObjectByKey(ctx, material.ContentKey, object); err != nil {
		return nil, s.storageException.FailedToPutObject(object).WithOrigin(err)
	}

	return &capi.SaveMyMaterialByIdResponseDto{
		UpdatedAt: material.UpdatedAt,
	}, nil
}

func (s *MaterialService) MoveMyMaterialById(
	ctx context.Context, requestDto *capi.MoveMyMaterialByIdRequestDto,
) (*capi.MoveMyMaterialByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.materialException.InvalidDto().WithOrigin(err)
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

	result := db.Exec(materialsql.MoveMyMaterialByIdSQL,
		requestDto.Body.DestinationParentSubShelfId,
		requestDto.Body.MaterialId,
		actorUserId,
		pg.Array(allowedPermissions),
		requestDto.Body.DestinationParentSubShelfId,
		actorUserId,
		pg.Array(allowedPermissions),
	)
	if err := result.Error; err != nil {
		return nil, s.materialException.FailedToUpdate().WithOrigin(err)
	}
	if result.RowsAffected == 0 {
		return nil, s.materialException.NoChanges()
	}

	return &capi.MoveMyMaterialByIdResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *MaterialService) MoveMyMaterialsByIds(
	ctx context.Context, requestDto *capi.MoveMyMaterialsByIdsRequestDto,
) (*capi.MoveMyMaterialsByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.materialException.InvalidDto().WithOrigin(err)
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

	result := db.Exec(materialsql.MoveMyMaterialsByIdsSQL,
		requestDto.Body.DestinationParentSubShelfId,
		requestDto.Body.MaterialIds,
		actorUserId,
		pg.Array(allowedPermissions),
		requestDto.Body.DestinationParentSubShelfId,
		actorUserId,
		pg.Array(allowedPermissions),
	)
	if err := result.Error; err != nil {
		return nil, s.materialException.FailedToUpdate().WithOrigin(err)
	}
	if result.RowsAffected == 0 {
		return nil, s.materialException.NoChanges()
	}

	return &capi.MoveMyMaterialsByIdsResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *MaterialService) RestoreMyMaterialById(
	ctx context.Context, requestDto *capi.RestoreMyMaterialByIdRequestDto,
) (*capi.RestoreMyMaterialByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.materialException.InvalidDto().WithOrigin(err)
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

	restoredMaterial, exception := s.materialRepository.RestoreSoftDeletedOneById(
		requestDto.Param.MaterialId,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	downloadURL, err := s.storage.PresignGetObjectByKey(ctx, restoredMaterial.ContentKey, nil)
	if err != nil {
		slogs.NotegicLogger.Error(ctx, err, "Failed to presign Material object")
	}

	return &capi.RestoreMyMaterialByIdResponseDto{
		Id:               restoredMaterial.Id,
		ParentSubShelfId: restoredMaterial.ParentSubShelfId,
		Name:             restoredMaterial.Name,
		Size:             restoredMaterial.Size,
		ContentType:      restoredMaterial.ContentType,
		ParseMediaType:   restoredMaterial.ParseMediaType,
		DownloadURL:      downloadURL,
		DeletedAt:        restoredMaterial.DeletedAt,
		UpdatedAt:        restoredMaterial.UpdatedAt,
		CreatedAt:        restoredMaterial.CreatedAt,
	}, nil
}

func (s *MaterialService) RestoreMyMaterialsByIds(
	ctx context.Context, requestDto *capi.RestoreMyMaterialsByIdsRequestDto,
) (*capi.RestoreMyMaterialsByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.materialException.InvalidDto().WithOrigin(err)
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

	restoredMaterials, exception := s.materialRepository.RestoreSoftDeletedManyByIds(
		requestDto.Body.MaterialIds,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	resDto := capi.RestoreMyMaterialsByIdsResponseDto{}
	for _, restoredMaterial := range restoredMaterials {
		downloadURL, err := s.storage.PresignGetObjectByKey(ctx, restoredMaterial.ContentKey, nil)
		if err != nil {
			slogs.NotegicLogger.Error(ctx, err, "Failed to presign Material object")
		}
		resDto = append(resDto, capi.RestoreMyMaterialByIdResponseDto{
			Id:               restoredMaterial.Id,
			ParentSubShelfId: restoredMaterial.ParentSubShelfId,
			Name:             restoredMaterial.Name,
			Size:             restoredMaterial.Size,
			ContentType:      restoredMaterial.ContentType,
			ParseMediaType:   restoredMaterial.ParseMediaType,
			DownloadURL:      downloadURL,
			DeletedAt:        restoredMaterial.DeletedAt,
			UpdatedAt:        restoredMaterial.UpdatedAt,
			CreatedAt:        restoredMaterial.CreatedAt,
		})
	}
	return &resDto, nil
}

func (s *MaterialService) DeleteMyMaterialById(
	ctx context.Context, requestDto *capi.DeleteMyMaterialByIdRequestDto,
) (*capi.DeleteMyMaterialByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.materialException.InvalidDto().WithOrigin(err)
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

	exception = s.materialRepository.SoftDeleteOneById(
		requestDto.Param.MaterialId,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.DeleteMyMaterialByIdResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *MaterialService) DeleteMyMaterialsByIds(
	ctx context.Context, requestDto *capi.DeleteMyMaterialsByIdsRequestDto,
) (*capi.DeleteMyMaterialsByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, s.materialException.InvalidDto().WithOrigin(err)
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

	exception = s.materialRepository.SoftDeleteManyByIds(
		requestDto.Body.MaterialIds,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.DeleteMyMaterialsByIdsResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

/* ============================== GraphQL Methods ============================== */

func (s *MaterialService) SearchPrivateMaterials(
	ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchMaterialInput,
) (*cgqlmodels.SearchMaterialConnection, *cexceptions.Exception) {
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

	query := db.Model(&sschemas.Material{}).
		Select(`"MaterialTable".*`).
		Joins(`INNER JOIN "SubShelfTable" ss ON ss.id = "MaterialTable".parent_sub_shelf_id`).
		Joins(`INNER JOIN "UsersToShelvesTable" uts ON uts.root_shelf_id = ss.root_shelf_id`).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, allowedPermissions).
		Scopes(s.materialScope.FilterOnlyDeleted(onlyDeleted))

	if gqlInput.ParentSubShelfID != nil {
		query = query.Where(`"MaterialTable".parent_sub_shelf_id = ?`, *gqlInput.ParentSubShelfID)
	}

	if gqlInput.RootShelfID != nil {
		query = query.Where("ss.root_shelf_id = ?", *gqlInput.RootShelfID)
	}

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			`"MaterialTable".name ILIKE ? OR "MaterialTable".content_type::text ILIKE ? OR "MaterialTable".parse_media_type ILIKE ?`,
			"%"+gqlInput.Query+"%",
			"%"+gqlInput.Query+"%",
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := ssearchcursor.Decode[cgqlmodels.SearchMaterialCursorFields](*gqlInput.After)
		if err != nil {
			return nil, s.searchException.FailedToDecode().WithOrigin(err)
		}

		query = query.Where(`"MaterialTable".id > ?`, searchCursor.Fields.ID)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		cending := cgqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == cgqlmodels.SearchSortOrderDesc {
			cending = cgqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case cgqlmodels.SearchMaterialSortByName:
			query = query.Order(`"MaterialTable".name ` + cending).
				Order(`"MaterialTable".updated_at ` + cending).
				Order(`"MaterialTable".created_at ` + cending)
		case cgqlmodels.SearchMaterialSortBySize:
			query = query.Order(`"MaterialTable".size ` + cending).
				Order(`"MaterialTable".name ` + cending).
				Order(`"MaterialTable".updated_at ` + cending).
				Order(`"MaterialTable".created_at ` + cending)
		case cgqlmodels.SearchMaterialSortByContentType:
			query = query.Order(`"MaterialTable".content_type ` + cending).
				Order(`"MaterialTable".name ` + cending).
				Order(`"MaterialTable".updated_at ` + cending).
				Order(`"MaterialTable".created_at ` + cending)
		case cgqlmodels.SearchMaterialSortByLastUpdate:
			query = query.Order(`"MaterialTable".updated_at ` + cending).
				Order(`"MaterialTable".name ` + cending).
				Order(`"MaterialTable".created_at ` + cending)
		case cgqlmodels.SearchMaterialSortByCreatedAt:
			query = query.Order(`"MaterialTable".created_at ` + cending).
				Order(`"MaterialTable".name ` + cending).
				Order(`"MaterialTable".updated_at ` + cending)
		default:
			query = query.Order(`"MaterialTable".name ` + cending).
				Order(`"MaterialTable".updated_at ` + cending).
				Order(`"MaterialTable".created_at ` + cending)
		}
	}

	limit := sconstants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, sconstants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var materials []sschemas.Material
	if err := query.Find(&materials).Error; err != nil {
		return nil, s.materialException.NotFound().WithOrigin(err)
	}

	hasNextPage := len(materials) > limit
	searchEdges := make([]*cgqlmodels.SearchMaterialEdge, len(materials))

	for index, material := range materials {
		searchCursor := ssearchcursor.SearchCursor[cgqlmodels.SearchMaterialCursorFields]{
			Fields: cgqlmodels.SearchMaterialCursorFields{
				ID: material.Id,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, s.searchException.FailedToEncode().WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, s.searchException.FailedToUnmarshalSearchCursor()
		}

		searchEdges[index] = &cgqlmodels.SearchMaterialEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                material.ToPrivateMaterial(),
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

	return &cgqlmodels.SearchMaterialConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
