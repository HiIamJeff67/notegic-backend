package shelves

import (
	"context"
	"errors"
	"net/http"
	"slices"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/root-shelves"
	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
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
	apiexceptions "github.com/HiIamJeff67/notegic-backend/internal/core/exceptions"
)

type RootShelfServiceInterface interface {
	GetMyRootShelfById(ctx context.Context, requestDto *capi.GetMyRootShelfByIdRequestDto) (*capi.GetMyRootShelfByIdResponseDto, *cexceptions.Exception)
	CreateRootShelf(ctx context.Context, requestDto *capi.CreateRootShelfRequestDto) (*capi.CreateRootShelfResponseDto, *cexceptions.Exception)
	CreateRootShelves(ctx context.Context, requestDto *capi.CreateRootShelvesRequestDto) (*capi.CreateRootShelvesResponseDto, *cexceptions.Exception)
	UpdateMyRootShelfById(ctx context.Context, requestDto *capi.UpdateMyRootShelfByIdRequestDto) (*capi.UpdateMyRootShelfByIdResponseDto, *cexceptions.Exception)
	UpdateMyRootShelvesByIds(ctx context.Context, requestDto *capi.UpdateMyRootShelvesByIdsRequestDto) (*capi.UpdateMyRootShelvesByIdsResponseDto, *cexceptions.Exception)
	RestoreMyRootShelfById(ctx context.Context, requestDto *capi.RestoreMyRootShelfByIdRequestDto) (*capi.RestoreMyRootShelfByIdResponseDto, *cexceptions.Exception)
	RestoreMyRootShelvesByIds(ctx context.Context, requestDto *capi.RestoreMyRootShelvesByIdsRequestDto) (*capi.RestoreMyRootShelvesByIdsResponseDto, *cexceptions.Exception)
	DeleteMyRootShelfById(ctx context.Context, requestDto *capi.DeleteMyRootShelfByIdRequestDto) (*capi.DeleteMyRootShelfByIdResponseDto, *cexceptions.Exception)
	DeleteMyRootShelvesByIds(ctx context.Context, requestDto *capi.DeleteMyRootShelvesByIdsRequestDto) (*capi.DeleteMyRootShelvesByIdsResponseDto, *cexceptions.Exception)

	GetMyRootShelfPermission(ctx context.Context, requestDto *capi.GetMyRootShelfPermissionRequestDto) (*capi.GetMyRootShelfPermissionResponseDto, *cexceptions.Exception)
	CreateMyRootShelfPermission(ctx context.Context, requestDto *capi.CreateMyRootShelfPermissionRequestDto) (*capi.CreateMyRootShelfPermissionResponseDto, *cexceptions.Exception)
	UpsertMyRootShelfPermission(ctx context.Context, requestDto *capi.UpsertMyRootShelfPermissionRequestDto) (*capi.UpsertMyRootShelfPermissionResponseDto, *cexceptions.Exception)
	UpsertMyRootShelfPermissions(ctx context.Context, requestDto *capi.UpsertMyRootShelfPermissionsRequestDto) (*capi.UpsertMyRootShelfPermissionsResponseDto, *cexceptions.Exception)
	UpdateMyRootShelfPermission(ctx context.Context, requestDto *capi.UpdateMyRootShelfPermissionRequestDto) (*capi.UpdateMyRootShelfPermissionResponseDto, *cexceptions.Exception)
	TransferMyRootShelfOwnership(ctx context.Context, requestDto *capi.TransferMyRootShelfOwnershipRequestDto) (*capi.TransferMyRootShelfOwnershipResponseDto, *cexceptions.Exception)
	DeleteMyRootShelfPermission(ctx context.Context, requestDto *capi.DeleteMyRootShelfPermissionRequestDto) (*capi.DeleteMyRootShelfPermissionResponseDto, *cexceptions.Exception)
	DeleteMyRootShelfPermissions(ctx context.Context, requestDto *capi.DeleteMyRootShelfPermissionsRequestDto) (*capi.DeleteMyRootShelfPermissionsResponseDto, *cexceptions.Exception)
	LeaveMyRootShelf(ctx context.Context, requestDto *capi.LeaveMyRootShelfRequestDto) *cexceptions.Exception
	LeaveMyRootShelves(ctx context.Context, requestDto *capi.LeaveMyRootShelvesRequestDto) *cexceptions.Exception

	SearchPrivateRootShelves(ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchRootShelfInput) (*cgqlmodels.SearchRootShelfConnection, *cexceptions.Exception)
}

type RootShelfService struct {
	validator                *validator.Validate
	db                       *gorm.DB
	rootShelfScope           sscopes.RootShelfScopeInterface
	rootShelfRepository      srepositories.RootShelfRepositoryInterface
	usersToShelvesRepository srepositories.UsersToShelvesRepositoryInterface
	blockPackRepository      srepositories.BlockPackRepositoryInterface
}

func NewRootShelfService(
	validator *validator.Validate,
	db *gorm.DB,
	rootShelfScope sscopes.RootShelfScopeInterface,
	rootShelfRepository srepositories.RootShelfRepositoryInterface,
	usersToShelvesRepository srepositories.UsersToShelvesRepositoryInterface,
	blockPackRepository srepositories.BlockPackRepositoryInterface,
) RootShelfServiceInterface {
	if db == nil {
		db = data.DB
	}
	return &RootShelfService{
		validator:                validator,
		db:                       db,
		rootShelfScope:           rootShelfScope,
		rootShelfRepository:      rootShelfRepository,
		usersToShelvesRepository: usersToShelvesRepository,
		blockPackRepository:      blockPackRepository,
	}
}

/* ============================== Auxiliary Functions ============================== */

func (s *RootShelfService) saveMyRootShelfPermission(
	ctx context.Context,
	actorUserId uuid.UUID,
	rootShelfId uuid.UUID,
	targetUserPublicId uuid.UUID,
	permission cenums.AccessControlPermission,
	requireExisting *bool,
) (*capi.RootShelfPermissionResponseDto, *cexceptions.Exception) {
	if permission == cenums.AccessControlPermission_Owner {
		return nil, cexceptions.New(
			"PermissionDenied",
			"RootShelf",
			"ManagePermission",
			"You do not have permission to manage this root shelf",
			http.StatusBadRequest,
		)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, cexceptions.New(
			"TransactionBeginFailed",
			"RootShelf",
			"SaveMyRootShelfPermission",
			"Failed to begin the root shelf permission transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}

	rootShelf, actorPermission, exception := s.rootShelfRepository.CheckPermissionAndGetOneById(
		rootShelfId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	var targetUser sschemas.User
	if result := tx.Where("public_id = ?", targetUserPublicId).First(&targetUser); result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}

	targetPermission, targetException := s.usersToShelvesRepository.GetOne(
		rootShelf.Id,
		targetUser.Id,
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if targetException != nil && !errors.Is(targetException.Origin(), gorm.ErrRecordNotFound) {
		tx.Rollback()
		return nil, targetException
	}
	if requireExisting != nil && *requireExisting != (targetPermission != nil) {
		tx.Rollback()
		if *requireExisting {
			return nil, targetException
		}
		return nil, cexceptions.New(
			"NoChanges",
			"RootShelf",
			"Manage",
			"No root shelf changes were applied",
			http.StatusNotModified,
		)
	}
	if targetPermission != nil && targetPermission.Permission == cenums.AccessControlPermission_Owner {
		tx.Rollback()
		return nil, cexceptions.New(
			"PermissionDenied",
			"RootShelf",
			"ManagePermission",
			"You do not have permission to manage this root shelf",
			http.StatusBadRequest,
		)
	}
	if actorPermission != cenums.AccessControlPermission_Owner && (permission == cenums.AccessControlPermission_Admin || targetPermission != nil && targetPermission.Permission == cenums.AccessControlPermission_Admin) {
		tx.Rollback()
		return nil, cexceptions.New(
			"PermissionDenied",
			"RootShelf",
			"ManagePermission",
			"You do not have permission to manage this root shelf",
			http.StatusBadRequest,
		)
	}

	var relation *sschemas.UsersToShelves
	if targetPermission == nil {
		relation, exception = s.usersToShelvesRepository.CreateOne(
			rootShelf.Id,
			targetUser.Id,
			permission,
			srepositories.WithTransactionDB(tx),
		)
	} else {
		relation, exception = s.usersToShelvesRepository.UpdateOne(
			rootShelf.Id,
			targetUser.Id,
			permission,
			srepositories.WithTransactionDB(tx),
		)
	}
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	blockPacks, exception := s.blockPackRepository.GetManyByRootShelfIds(
		[]uuid.UUID{rootShelf.Id},
		srepositories.WithTransactionDB(tx),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPackIds := make([]uuid.UUID, len(blockPacks))
	for index, blockPack := range blockPacks {
		blockPackIds[index] = blockPack.Id
	}
	if targetPermission != nil &&
		slices.Index(cenums.AllAccessControlPermissions, permission) <
			slices.Index(cenums.AllAccessControlPermissions, targetPermission.Permission) {
		if err := srepositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
			tx,
			rootShelf.Id.String(),
			blockPackIds,
			[]uuid.UUID{targetUser.PublicId},
			coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
		); err != nil {
			tx.Rollback()
			return nil, cexceptions.New(
				"FailedToCreate",
				"Outbox",
				"SaveMyRootShelfPermission",
				"Failed to create lifecycle outbox events",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueRootShelfPermissionChanged(
		tx,
		rootShelf.Id.String(),
		rootShelf.Id,
		targetUser.PublicId,
		permission.String(),
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"SaveMyRootShelfPermission",
			"Failed to create resource event",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"RootShelf",
			"Manage",
			"Failed to commit the root shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	return &capi.RootShelfPermissionResponseDto{
		UserPublicId: targetUser.PublicId,
		Permission:   relation.Permission.String(),
		UpdatedAt:    relation.UpdatedAt,
		CreatedAt:    relation.CreatedAt,
	}, nil
}

/* ============================== Service Methods for RootShelf ============================== */

func (s *RootShelfService) GetMyRootShelfById(
	ctx context.Context, requestDto *capi.GetMyRootShelfByIdRequestDto,
) (*capi.GetMyRootShelfByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	db := s.db.WithContext(ctx)

	onlyDeleted := stypes.Ternary_Neutral
	if requestDto.Param.IsDeleted != nil {
		if *requestDto.Param.IsDeleted {
			onlyDeleted = stypes.Ternary_Positive
		} else {
			onlyDeleted = stypes.Ternary_Negative
		}
	}

	shelf, permission, exception := s.rootShelfRepository.GetOneById(
		requestDto.Param.RootShelfId,
		actorUserId,
		nil,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(onlyDeleted),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.GetMyRootShelfByIdResponseDto{
		Id:             shelf.Id,
		Name:           shelf.Name,
		Permission:     permission.String(),
		SubShelfCount:  shelf.SubShelfCount,
		ItemCount:      shelf.ItemCount,
		LastAnalyzedAt: shelf.LastAnalyzedAt,
		DeletedAt:      shelf.DeletedAt,
		UpdatedAt:      shelf.UpdatedAt,
		CreatedAt:      shelf.CreatedAt,
	}, nil
}

func (s *RootShelfService) CreateRootShelf(
	ctx context.Context, requestDto *capi.CreateRootShelfRequestDto,
) (*capi.CreateRootShelfResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	db := s.db.WithContext(ctx)

	now := time.Now()
	newRootShelfId, exception := s.rootShelfRepository.CreateOne(
		actorUserId,
		sinputs.CreateRootShelfInput{
			Id:             requestDto.Body.Id,
			Name:           requestDto.Body.Name,
			LastAnalyzedAt: &now,
		},
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.CreateRootShelfResponseDto{
		Id:             *newRootShelfId,
		LastAnalyzedAt: now,
		CreatedAt:      time.Now(),
	}, nil
}

func (s *RootShelfService) CreateRootShelves(
	ctx context.Context, requestDto *capi.CreateRootShelvesRequestDto,
) (*capi.CreateRootShelvesResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	db := s.db.WithContext(ctx)

	now := time.Now()
	input := make([]sinputs.CreateRootShelfInput, len(requestDto.Body.RootShelves))
	for index, createdRootShelf := range requestDto.Body.RootShelves {
		input[index] = sinputs.CreateRootShelfInput{
			Id:             createdRootShelf.Id,
			Name:           createdRootShelf.Name,
			LastAnalyzedAt: &now,
		}
	}
	newRootShelfIds, exception := s.rootShelfRepository.CreateMany(
		actorUserId,
		input,
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.CreateRootShelvesResponseDto{
		Ids:            newRootShelfIds,
		LastAnalyzedAt: now,
		CreatedAt:      time.Now(),
	}, nil
}

func (s *RootShelfService) UpdateMyRootShelfById(
	ctx context.Context, requestDto *capi.UpdateMyRootShelfByIdRequestDto,
) (*capi.UpdateMyRootShelfByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	db := s.db.WithContext(ctx)

	rootShelf, exception := s.rootShelfRepository.UpdateOneById(
		requestDto.Param.RootShelfId,
		actorUserId,
		sinputs.PartialUpdateRootShelfInput{
			Values: sinputs.UpdateRootShelfInput{
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

	return &capi.UpdateMyRootShelfByIdResponseDto{
		UpdatedAt: rootShelf.UpdatedAt,
	}, nil
}

func (s *RootShelfService) UpdateMyRootShelvesByIds(
	ctx context.Context,
	requestDto *capi.UpdateMyRootShelvesByIdsRequestDto,
) (*capi.UpdateMyRootShelvesByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	db := s.db.WithContext(ctx)
	input := make([]sinputs.UpdateRootShelfByIdInput, len(requestDto.Body.UpdatedRootShelves))
	for index, updatedRootShelf := range requestDto.Body.UpdatedRootShelves {
		input[index] = sinputs.UpdateRootShelfByIdInput{
			Id: updatedRootShelf.RootShelfId,
			PartialUpdateInput: sinputs.PartialUpdateInput[sinputs.UpdateRootShelfInput]{
				Values: sinputs.UpdateRootShelfInput{
					Name: updatedRootShelf.Values.Name,
				},
				SetNull: updatedRootShelf.SetNull,
			},
		}
	}
	exception = s.rootShelfRepository.UpdateManyByIds(
		actorUserId,
		input,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.UpdateMyRootShelvesByIdsResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *RootShelfService) RestoreMyRootShelfById(
	ctx context.Context,
	requestDto *capi.RestoreMyRootShelfByIdRequestDto,
) (*capi.RestoreMyRootShelfByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	db := s.db.WithContext(ctx)

	restoredRootShelf, exception := s.rootShelfRepository.RestoreSoftDeletedOneById(
		requestDto.Body.RootShelfId,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.RestoreMyRootShelfByIdResponseDto{
		Id:             restoredRootShelf.Id,
		Name:           restoredRootShelf.Name,
		SubShelfCount:  restoredRootShelf.SubShelfCount,
		ItemCount:      restoredRootShelf.ItemCount,
		LastAnalyzedAt: restoredRootShelf.LastAnalyzedAt,
		DeletedAt:      restoredRootShelf.DeletedAt,
		UpdatedAt:      restoredRootShelf.UpdatedAt,
		CreatedAt:      restoredRootShelf.CreatedAt,
	}, nil
}

func (s *RootShelfService) RestoreMyRootShelvesByIds(
	ctx context.Context,
	requestDto *capi.RestoreMyRootShelvesByIdsRequestDto,
) (*capi.RestoreMyRootShelvesByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	db := s.db.WithContext(ctx)

	restoredRootShelves, exception := s.rootShelfRepository.RestoreSoftDeletedManyByIds(
		requestDto.Body.RootShelfIds,
		actorUserId,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	responseDto := capi.RestoreMyRootShelvesByIdsResponseDto{}
	for _, restoredRootShelf := range restoredRootShelves {
		responseDto = append(responseDto, capi.RestoreMyRootShelfByIdResponseDto{
			Id:             restoredRootShelf.Id,
			Name:           restoredRootShelf.Name,
			SubShelfCount:  restoredRootShelf.SubShelfCount,
			ItemCount:      restoredRootShelf.ItemCount,
			LastAnalyzedAt: restoredRootShelf.LastAnalyzedAt,
			DeletedAt:      restoredRootShelf.DeletedAt,
			UpdatedAt:      restoredRootShelf.UpdatedAt,
			CreatedAt:      restoredRootShelf.CreatedAt,
		})
	}

	return &responseDto, nil
}

func (s *RootShelfService) DeleteMyRootShelfById(
	ctx context.Context,
	requestDto *capi.DeleteMyRootShelfByIdRequestDto,
) (*capi.DeleteMyRootShelfByIdResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()

	rootShelf, permission, exception := s.rootShelfRepository.CheckPermissionAndGetOneById(
		requestDto.Body.RootShelfId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPacks, exception := s.blockPackRepository.GetManyByRootShelfIds(
		[]uuid.UUID{rootShelf.Id},
		srepositories.WithTransactionDB(tx),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPackIds := make([]uuid.UUID, len(blockPacks))
	for index, blockPack := range blockPacks {
		blockPackIds[index] = blockPack.Id
	}
	var rootShelfMemberPublicIds []uuid.UUID
	if permission == cenums.AccessControlPermission_Owner {
		var relations []sschemas.UsersToShelves
		result := tx.
			Preload(string(sschemas.UsersToShelvesRelation_User)).
			Where("root_shelf_id = ?", rootShelf.Id).
			Find(&relations)
		if result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New(
				"FailedToRead",
				"RootShelf",
				"DeleteMyRootShelfById",
				"Failed to resolve root shelf members",
				http.StatusInternalServerError,
				true,
			).WithOrigin(result.Error)
		}
		rootShelfMemberPublicIds = make([]uuid.UUID, 0, len(relations))
		for _, relation := range relations {
			if relation.User.PublicId != uuid.Nil {
				rootShelfMemberPublicIds = append(rootShelfMemberPublicIds, relation.User.PublicId)
			}
		}
	}

	if permission == cenums.AccessControlPermission_Owner {
		result := tx.
			Model(&sschemas.RootShelf{}).
			Where("id = ?", rootShelf.Id).
			Update("deleted_at", time.Now())
		if result.Error != nil {
			tx.Rollback()
			return nil, cexceptions.New(
				"FailedToUpdate",
				"RootShelf",
				"Manage",
				"Failed to update the root shelf",
				http.StatusInternalServerError,
				true,
			).WithOrigin(result.Error)
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return nil, cexceptions.New(
				"NoChanges",
				"RootShelf",
				"Manage",
				"No root shelf changes were applied",
				http.StatusNotModified,
			)
		}
	} else {
		exception = s.usersToShelvesRepository.DeleteOne(
			rootShelf.Id,
			actorUserId,
			srepositories.WithTransactionDB(tx),
		)
		if exception != nil {
			tx.Rollback()
			return nil, exception
		}
	}

	var targetUserPublicIds []uuid.UUID
	if permission != cenums.AccessControlPermission_Owner {
		actorUserPublicId, exception := contexts.GetActorUserPublicId(ctx)
		if exception != nil {
			tx.Rollback()
			return nil, exception
		}
		targetUserPublicIds = []uuid.UUID{actorUserPublicId}
	}
	reason := coreevents.BlockPackAccessRevocationReason_PermissionRevoked
	if permission == cenums.AccessControlPermission_Owner {
		reason = coreevents.BlockPackAccessRevocationReason_ResourceUnavailable
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		rootShelf.Id.String(),
		blockPackIds,
		targetUserPublicIds,
		reason,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMyRootShelfById",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if permission == cenums.AccessControlPermission_Owner {
		if err := srepositories.NewOutboxEventRepository().EnqueueRootShelfDeleted(
			tx,
			rootShelf.Id.String(),
			rootShelf.Id,
			rootShelfMemberPublicIds,
		); err != nil {
			tx.Rollback()
			return nil, cexceptions.New(
				"FailedToCreate",
				"Outbox",
				"DeleteMyRootShelfById",
				"Failed to create resource events",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}
	} else if len(targetUserPublicIds) > 0 {
		if err := srepositories.NewOutboxEventRepository().EnqueueRootShelfPermissionRevoked(
			tx,
			rootShelf.Id.String(),
			rootShelf.Id,
			targetUserPublicIds[0],
		); err != nil {
			tx.Rollback()
			return nil, cexceptions.New(
				"FailedToCreate",
				"Outbox",
				"DeleteMyRootShelfById",
				"Failed to create resource event",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"RootShelf",
			"Manage",
			"Failed to commit the root shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	return &capi.DeleteMyRootShelfByIdResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *RootShelfService) DeleteMyRootShelvesByIds(
	ctx context.Context,
	requestDto *capi.DeleteMyRootShelvesByIdsRequestDto,
) (*capi.DeleteMyRootShelvesByIdsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, cexceptions.New(
			"TransactionBeginFailed",
			"RootShelf",
			"DeleteMyRootShelvesByIds",
			"Failed to begin the root shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	blockPacks, exception := s.blockPackRepository.GetManyByRootShelfIds(
		requestDto.Body.RootShelfIds,
		srepositories.WithTransactionDB(tx),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPackIds := make([]uuid.UUID, len(blockPacks))
	for index, blockPack := range blockPacks {
		blockPackIds[index] = blockPack.Id
	}
	var rootShelfRelations []sschemas.UsersToShelves
	if result := tx.
		Preload(string(sschemas.UsersToShelvesRelation_User)).
		Where("root_shelf_id IN ?", requestDto.Body.RootShelfIds).
		Find(&rootShelfRelations); result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToRead",
			"RootShelf",
			"DeleteMyRootShelvesByIds",
			"Failed to resolve root shelf members",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	rootShelfMemberPublicIdsById := make(map[uuid.UUID][]uuid.UUID, len(requestDto.Body.RootShelfIds))
	for _, relation := range rootShelfRelations {
		if relation.User.PublicId != uuid.Nil {
			rootShelfMemberPublicIdsById[relation.RootShelfId] = append(
				rootShelfMemberPublicIdsById[relation.RootShelfId],
				relation.User.PublicId,
			)
		}
	}

	exception = s.rootShelfRepository.SoftDeleteManyByIds(
		requestDto.Body.RootShelfIds,
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
		"root-shelf-bulk-delete",
		blockPackIds,
		nil,
		coreevents.BlockPackAccessRevocationReason_ResourceUnavailable,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMyRootShelvesByIds",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueManyRootShelfDeleted(
		tx,
		"root-shelf-bulk-delete",
		requestDto.Body.RootShelfIds,
		rootShelfMemberPublicIdsById,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMyRootShelvesByIds",
			"Failed to create resource events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"RootShelf",
			"DeleteMyRootShelvesByIds",
			"Failed to commit the root shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &capi.DeleteMyRootShelvesByIdsResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *RootShelfService) GetMyRootShelfPermission(
	ctx context.Context, requestDto *capi.GetMyRootShelfPermissionRequestDto,
) (*capi.GetMyRootShelfPermissionResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	db := s.db.WithContext(ctx)
	if _, _, exception = s.rootShelfRepository.CheckPermissionAndGetOneById(
		requestDto.Param.RootShelfId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithDB(db),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	); exception != nil {
		return nil, exception
	}

	var targetUser sschemas.User
	if result := db.Where("public_id = ?", requestDto.Param.UserPublicId).First(&targetUser); result.Error != nil {
		return nil, cexceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}
	relation, exception := s.usersToShelvesRepository.GetOne(
		requestDto.Param.RootShelfId,
		targetUser.Id,
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.GetMyRootShelfPermissionResponseDto{
		UserPublicId: targetUser.PublicId,
		Permission:   relation.Permission.String(),
		UpdatedAt:    relation.UpdatedAt,
		CreatedAt:    relation.CreatedAt,
	}, nil
}

func (s *RootShelfService) CreateMyRootShelfPermission(
	ctx context.Context, requestDto *capi.CreateMyRootShelfPermissionRequestDto,
) (*capi.CreateMyRootShelfPermissionResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	permission, err := cenums.ConvertStringToAccessControlPermission(requestDto.Body.Permission)
	if err != nil {
		return nil, cexceptions.InvalidInput("RootShelf").WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	requireExisting := false
	return s.saveMyRootShelfPermission(ctx, actorUserId, requestDto.Param.RootShelfId, requestDto.Param.UserPublicId, *permission, &requireExisting)
}

func (s *RootShelfService) UpsertMyRootShelfPermission(
	ctx context.Context, requestDto *capi.UpsertMyRootShelfPermissionRequestDto,
) (*capi.UpsertMyRootShelfPermissionResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	permission, err := cenums.ConvertStringToAccessControlPermission(requestDto.Body.Permission)
	if err != nil {
		return nil, cexceptions.InvalidInput("RootShelf").WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	return s.saveMyRootShelfPermission(ctx, actorUserId, requestDto.Param.RootShelfId, requestDto.Param.UserPublicId, *permission, nil)
}

func (s *RootShelfService) UpsertMyRootShelfPermissions(
	ctx context.Context, requestDto *capi.UpsertMyRootShelfPermissionsRequestDto,
) (*capi.UpsertMyRootShelfPermissionsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	userPublicIds := make([]uuid.UUID, len(requestDto.Body.Permissions))
	permissionByPublicId := make(map[uuid.UUID]cenums.AccessControlPermission, len(requestDto.Body.Permissions))
	for index, input := range requestDto.Body.Permissions {
		permission, err := cenums.ConvertStringToAccessControlPermission(input.Permission)
		if err != nil {
			return nil, cexceptions.InvalidInput("RootShelf").WithOrigin(err)
		}
		if *permission == cenums.AccessControlPermission_Owner {
			return nil, cexceptions.New(
				"PermissionDenied",
				"RootShelf",
				"ManagePermission",
				"You do not have permission to manage this root shelf",
				http.StatusBadRequest,
			)
		}
		if _, exists := permissionByPublicId[input.UserPublicId]; exists {
			return nil, cexceptions.New(
				"InvalidRequest",
				"RootShelf",
				"ValidateRequest",
				"Root shelf request is invalid",
				http.StatusBadRequest,
			)
		}

		userPublicIds[index] = input.UserPublicId
		permissionByPublicId[input.UserPublicId] = *permission
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, cexceptions.New(
			"TransactionBeginFailed",
			"RootShelf",
			"Manage",
			"Failed to begin the root shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}

	rootShelf, actorPermission, exception := s.rootShelfRepository.CheckPermissionAndGetOneById(
		requestDto.Param.RootShelfId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	var targetUsers []sschemas.User
	result := tx.
		Model(&sschemas.User{}).
		Select("id, public_id").
		Where("public_id IN ?", userPublicIds).
		Find(&targetUsers)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}
	if len(targetUsers) != len(userPublicIds) {
		tx.Rollback()
		return nil, cexceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		)
	}

	userByPublicId := make(map[uuid.UUID]sschemas.User, len(targetUsers))
	userById := make(map[uuid.UUID]sschemas.User, len(targetUsers))
	for _, user := range targetUsers {
		userByPublicId[user.PublicId] = user
		userById[user.Id] = user
	}

	userIds := make([]uuid.UUID, len(userPublicIds))
	for index, userPublicId := range userPublicIds {
		userIds[index] = userByPublicId[userPublicId].Id
	}

	existingPermissions, exception := s.usersToShelvesRepository.GetMany(
		rootShelf.Id,
		userIds,
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	existingPermissionByUserId := make(map[uuid.UUID]cenums.AccessControlPermission, len(existingPermissions))
	for _, existingPermission := range existingPermissions {
		existingPermissionByUserId[existingPermission.UserId] = existingPermission.Permission
	}

	permissions := make([]cenums.AccessControlPermission, len(userIds))
	for index, userId := range userIds {
		user := userById[userId]
		permission := permissionByPublicId[user.PublicId]
		if existingPermissionByUserId[userId] == cenums.AccessControlPermission_Owner {
			tx.Rollback()
			return nil, cexceptions.New(
				"PermissionDenied",
				"RootShelf",
				"ManagePermission",
				"You do not have permission to manage this root shelf",
				http.StatusBadRequest,
			)
		}
		if actorPermission != cenums.AccessControlPermission_Owner &&
			(permission == cenums.AccessControlPermission_Admin ||
				existingPermissionByUserId[userId] == cenums.AccessControlPermission_Admin) {
			tx.Rollback()
			return nil, cexceptions.New(
				"PermissionDenied",
				"RootShelf",
				"ManagePermission",
				"You do not have permission to manage this root shelf",
				http.StatusBadRequest,
			)
		}

		permissions[index] = permission
	}

	updatedPermissions, exception := s.usersToShelvesRepository.UpsertMany(
		rootShelf.Id,
		userIds,
		permissions,
		srepositories.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	userPublicIdByUserId := make(map[uuid.UUID]uuid.UUID, len(userById))
	for userId, user := range userById {
		userPublicIdByUserId[userId] = user.PublicId
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueManyRootShelfPermissionChanges(
		tx,
		rootShelf.Id.String(),
		rootShelf.Id,
		updatedPermissions,
		userPublicIdByUserId,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"UpsertMyRootShelfPermissions",
			"Failed to create resource events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	if err := tx.Commit().Error; err != nil {
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"RootShelf",
			"Manage",
			"Failed to commit the root shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	updatedPermissionByUserId := make(map[uuid.UUID]sschemas.UsersToShelves, len(updatedPermissions))
	for _, updatedPermission := range updatedPermissions {
		updatedPermissionByUserId[updatedPermission.UserId] = updatedPermission
	}

	responseDtos := make([]capi.RootShelfPermissionResponseDto, len(userIds))
	for index, userId := range userIds {
		user := userById[userId]
		updatedPermission := updatedPermissionByUserId[userId]
		responseDtos[index] = capi.RootShelfPermissionResponseDto{
			UserPublicId: user.PublicId,
			Permission:   updatedPermission.Permission.String(),
			UpdatedAt:    updatedPermission.UpdatedAt,
			CreatedAt:    updatedPermission.CreatedAt,
		}
	}

	return &capi.UpsertMyRootShelfPermissionsResponseDto{
		Permissions: responseDtos,
	}, nil
}

func (s *RootShelfService) UpdateMyRootShelfPermission(
	ctx context.Context, requestDto *capi.UpdateMyRootShelfPermissionRequestDto,
) (*capi.UpdateMyRootShelfPermissionResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	permission, err := cenums.ConvertStringToAccessControlPermission(requestDto.Body.Permission)
	if err != nil {
		return nil, cexceptions.InvalidInput("RootShelf").WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	requireExisting := true
	return s.saveMyRootShelfPermission(ctx, actorUserId, requestDto.Param.RootShelfId, requestDto.Param.UserPublicId, *permission, &requireExisting)
}

func (s *RootShelfService) TransferMyRootShelfOwnership(
	ctx context.Context,
	requestDto *capi.TransferMyRootShelfOwnershipRequestDto,
) (*capi.TransferMyRootShelfOwnershipResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, cexceptions.New(
			"TransactionBeginFailed",
			"RootShelf",
			"TransferMyRootShelfOwnership",
			"Failed to begin the root shelf ownership transfer transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	rootShelf, permission, exception := s.rootShelfRepository.CheckPermissionAndGetOneById(
		requestDto.Param.RootShelfId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if permission != cenums.AccessControlPermission_Owner {
		tx.Rollback()
		return nil, cexceptions.New(
			"PermissionDenied",
			"RootShelf",
			"ManagePermission",
			"You do not have permission to manage this root shelf",
			http.StatusBadRequest,
		)
	}

	var actorUser sschemas.User
	if result := tx.Select("id, public_id").Where("id = ?", actorUserId).First(&actorUser); result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}
	var targetUser sschemas.User
	if result := tx.Select("id, public_id").Where("public_id = ?", requestDto.Body.TargetUserPublicId).First(&targetUser); result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}
	if targetUser.Id == actorUserId {
		tx.Rollback()
		return nil, cexceptions.New(
			"NoChanges",
			"RootShelf",
			"Manage",
			"No root shelf changes were applied",
			http.StatusNotModified,
		)
	}

	targetMembership, exception := s.usersToShelvesRepository.GetOne(
		rootShelf.Id,
		targetUser.Id,
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if targetMembership.Permission == cenums.AccessControlPermission_Owner {
		tx.Rollback()
		return nil, cexceptions.New(
			"NoChanges",
			"RootShelf",
			"Manage",
			"No root shelf changes were applied",
			http.StatusNotModified,
		)
	}

	var accounts []sschemas.UserAccount
	result := tx.
		Clauses(clause.Locking{Strength: srepositories.LockingStrengthUpdate}).
		Where("user_id IN ?", []uuid.UUID{actorUserId, targetUser.Id}).
		Order("user_id").
		Find(&accounts)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToUpdate",
			"RootShelf",
			"Manage",
			"Failed to update the root shelf",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	if len(accounts) != 2 {
		tx.Rollback()
		return nil, cexceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		)
	}

	var maximumSubscribers int32
	result = tx.
		Model(&sschemas.User{}).
		Select(`"PlanLimitationTable".max_realtime_room_subscriber_count`).
		Joins(`INNER JOIN "PlanLimitationTable" ON "PlanLimitationTable".key = "UserTable".plan`).
		Where(`"UserTable".id = ?`, targetUser.Id).
		Scan(&maximumSubscribers)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToUpdate",
			"RootShelf",
			"Manage",
			"Failed to update the root shelf",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 || maximumSubscribers <= 0 {
		tx.Rollback()
		return nil, cexceptions.New(
			"PermissionDenied",
			"RootShelf",
			"ManagePermission",
			"You do not have permission to manage this root shelf",
			http.StatusBadRequest,
		)
	}

	var blockPackIds []uuid.UUID
	result = tx.
		Model(&sschemas.BlockPack{}).
		Select(`"BlockPackTable".id`).
		Joins(`INNER JOIN "SubShelfTable" ON "SubShelfTable".id = "BlockPackTable".parent_sub_shelf_id`).
		Where(`"SubShelfTable".root_shelf_id = ?`, rootShelf.Id).
		Where(`"BlockPackTable".deleted_at IS NULL`).
		Where(`"SubShelfTable".deleted_at IS NULL`).
		Find(&blockPackIds)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"QueryFailed",
			"BlockPack",
			"ManageRootShelf",
			"Failed to retrieve root shelf block packs",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}

	if _, exception = s.usersToShelvesRepository.UpdateOne(
		rootShelf.Id,
		actorUserId,
		cenums.AccessControlPermission_Admin,
		srepositories.WithTransactionDB(tx),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	newOwnerMembership, exception := s.usersToShelvesRepository.UpdateOne(
		rootShelf.Id,
		targetUser.Id,
		cenums.AccessControlPermission_Owner,
		srepositories.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	result = tx.Model(&sschemas.RootShelf{}).
		Where("id = ?", rootShelf.Id).
		Update("owner_id", targetUser.Id)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToUpdate",
			"RootShelf",
			"Manage",
			"Failed to update the root shelf",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, cexceptions.New(
			"NotFound",
			"RootShelf",
			"Manage",
			"Root shelf was not found",
			http.StatusNotFound,
		)
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueRootShelfPermissionChanged(
		tx,
		rootShelf.Id.String(),
		rootShelf.Id,
		actorUser.PublicId,
		cenums.AccessControlPermission_Admin.String(),
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"TransferMyRootShelfOwnership",
			"Failed to create resource events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueRootShelfPermissionChanged(
		tx,
		rootShelf.Id.String(),
		rootShelf.Id,
		targetUser.PublicId,
		cenums.AccessControlPermission_Owner.String(),
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"TransferMyRootShelfOwnership",
			"Failed to create resource events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"RootShelf",
			"Manage",
			"Failed to commit the root shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &capi.TransferMyRootShelfOwnershipResponseDto{
		RootShelfId:               rootShelf.Id,
		PreviousOwnerUserPublicId: actorUser.PublicId,
		NewOwnerUserPublicId:      targetUser.PublicId,
		UpdatedAt:                 newOwnerMembership.UpdatedAt,
	}, nil
}

func (s *RootShelfService) DeleteMyRootShelfPermission(
	ctx context.Context, requestDto *capi.DeleteMyRootShelfPermissionRequestDto,
) (*capi.DeleteMyRootShelfPermissionResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	tx := s.db.WithContext(ctx).Begin()

	rootShelf, actorPermission, exception := s.rootShelfRepository.CheckPermissionAndGetOneById(
		requestDto.Param.RootShelfId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	var targetUser sschemas.User
	result := tx.
		Model(&sschemas.User{}).
		Where("public_id = ?", requestDto.Param.UserPublicId).
		First(&targetUser)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}

	targetPermission, exception := s.usersToShelvesRepository.GetOne(
		rootShelf.Id,
		targetUser.Id,
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if targetPermission.Permission == cenums.AccessControlPermission_Owner {
		tx.Rollback()
		return nil, cexceptions.New(
			"PermissionDenied",
			"RootShelf",
			"ManagePermission",
			"You do not have permission to manage this root shelf",
			http.StatusBadRequest,
		)
	}
	if actorPermission != cenums.AccessControlPermission_Owner &&
		targetPermission.Permission == cenums.AccessControlPermission_Admin {
		tx.Rollback()
		return nil, cexceptions.New(
			"PermissionDenied",
			"RootShelf",
			"ManagePermission",
			"You do not have permission to manage this root shelf",
			http.StatusBadRequest,
		)
	}

	exception = s.usersToShelvesRepository.DeleteOne(
		rootShelf.Id,
		targetUser.Id,
		srepositories.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPacks, exception := s.blockPackRepository.GetManyByRootShelfIds(
		[]uuid.UUID{rootShelf.Id},
		srepositories.WithTransactionDB(tx),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPackIds := make([]uuid.UUID, len(blockPacks))
	for index, blockPack := range blockPacks {
		blockPackIds[index] = blockPack.Id
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		rootShelf.Id.String(),
		blockPackIds,
		[]uuid.UUID{targetUser.PublicId},
		coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMyRootShelfPermission",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueRootShelfPermissionRevoked(
		tx,
		rootShelf.Id.String(),
		rootShelf.Id,
		targetUser.PublicId,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMyRootShelfPermission",
			"Failed to create resource event",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"RootShelf",
			"Manage",
			"Failed to commit the root shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	return &capi.DeleteMyRootShelfPermissionResponseDto{}, nil
}

func (s *RootShelfService) DeleteMyRootShelfPermissions(
	ctx context.Context, requestDto *capi.DeleteMyRootShelfPermissionsRequestDto,
) (*capi.DeleteMyRootShelfPermissionsResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}

	userPublicIdSet := make(map[uuid.UUID]struct{}, len(requestDto.Body.UserPublicIds))
	for _, userPublicId := range requestDto.Body.UserPublicIds {
		if _, exists := userPublicIdSet[userPublicId]; exists {
			return nil, cexceptions.New(
				"InvalidRequest",
				"RootShelf",
				"ValidateRequest",
				"Root shelf request is invalid",
				http.StatusBadRequest,
			)
		}

		userPublicIdSet[userPublicId] = struct{}{}
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
			"RootShelf",
			"Manage",
			"Failed to begin the root shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}

	rootShelf, actorPermission, exception := s.rootShelfRepository.CheckPermissionAndGetOneById(
		requestDto.Param.RootShelfId,
		actorUserId,
		nil,
		allowedPermissions,
		srepositories.WithTransactionDB(tx),
		srepositories.WithAllowedPermissions(allowedPermissions),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	var targetUsers []sschemas.User
	result := tx.
		Model(&sschemas.User{}).
		Select("id, public_id").
		Where("public_id IN ?", requestDto.Body.UserPublicIds).
		Find(&targetUsers)
	if result.Error != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}
	if len(targetUsers) != len(requestDto.Body.UserPublicIds) {
		tx.Rollback()
		return nil, cexceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		)
	}

	userIdByPublicId := make(map[uuid.UUID]uuid.UUID, len(targetUsers))
	for _, targetUser := range targetUsers {
		userIdByPublicId[targetUser.PublicId] = targetUser.Id
	}

	userIds := make([]uuid.UUID, len(requestDto.Body.UserPublicIds))
	for index, userPublicId := range requestDto.Body.UserPublicIds {
		userIds[index] = userIdByPublicId[userPublicId]
	}

	targetPermissions, exception := s.usersToShelvesRepository.GetMany(
		rootShelf.Id,
		userIds,
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if len(targetPermissions) != len(userIds) {
		tx.Rollback()
		return nil, cexceptions.New(
			"NotFound",
			"RootShelf",
			"Manage",
			"Root shelf was not found",
			http.StatusNotFound,
		)
	}

	for _, targetPermission := range targetPermissions {
		if targetPermission.Permission == cenums.AccessControlPermission_Owner {
			tx.Rollback()
			return nil, cexceptions.New(
				"PermissionDenied",
				"RootShelf",
				"ManagePermission",
				"You do not have permission to manage this root shelf",
				http.StatusBadRequest,
			)
		}
		if actorPermission != cenums.AccessControlPermission_Owner &&
			targetPermission.Permission == cenums.AccessControlPermission_Admin {
			tx.Rollback()
			return nil, cexceptions.New(
				"PermissionDenied",
				"RootShelf",
				"ManagePermission",
				"You do not have permission to manage this root shelf",
				http.StatusBadRequest,
			)
		}
	}

	exception = s.usersToShelvesRepository.DeleteMany(
		rootShelf.Id,
		userIds,
		srepositories.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPacks, exception := s.blockPackRepository.GetManyByRootShelfIds(
		[]uuid.UUID{rootShelf.Id},
		srepositories.WithTransactionDB(tx),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	blockPackIds := make([]uuid.UUID, len(blockPacks))
	for index, blockPack := range blockPacks {
		blockPackIds[index] = blockPack.Id
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		rootShelf.Id.String(),
		blockPackIds,
		requestDto.Body.UserPublicIds,
		coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMyRootShelfPermissions",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueManyRootShelfPermissionRevocations(
		tx,
		rootShelf.Id.String(),
		[]uuid.UUID{rootShelf.Id},
		requestDto.Body.UserPublicIds,
	); err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"DeleteMyRootShelfPermissions",
			"Failed to create resource events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, cexceptions.New(
			"TransactionCommitFailed",
			"RootShelf",
			"Manage",
			"Failed to commit the root shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	return &capi.DeleteMyRootShelfPermissionsResponseDto{}, nil
}

func (s *RootShelfService) LeaveMyRootShelf(
	ctx context.Context, requestDto *capi.LeaveMyRootShelfRequestDto,
) *cexceptions.Exception {
	if err := s.validator.Struct(requestDto); err != nil {
		return apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return exception
	}

	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return cexceptions.New(
			"TransactionBeginFailed",
			"RootShelf",
			"LeaveMyRootShelf",
			"Failed to begin the root shelf leave transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	blockPacks, exception := s.blockPackRepository.GetManyByRootShelfIds(
		[]uuid.UUID{requestDto.Param.RootShelfId},
		srepositories.WithTransactionDB(tx),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return exception
	}
	blockPackIds := make([]uuid.UUID, len(blockPacks))
	for index, blockPack := range blockPacks {
		blockPackIds[index] = blockPack.Id
	}
	rootShelf, permission, exception := s.rootShelfRepository.CheckPermissionAndGetOneById(
		requestDto.Param.RootShelfId,
		actorUserId,
		nil,
		nil,
		srepositories.WithTransactionDB(tx),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return exception
	}
	if permission == cenums.AccessControlPermission_Owner {
		tx.Rollback()
		return cexceptions.New(
			"PermissionDenied",
			"RootShelf",
			"ManagePermission",
			"You do not have permission to manage this root shelf",
			http.StatusBadRequest,
		)
	}
	if exception = s.usersToShelvesRepository.DeleteOne(
		rootShelf.Id,
		actorUserId,
		srepositories.WithTransactionDB(tx),
	); exception != nil {
		tx.Rollback()
		return exception
	}
	actorUserPublicId, exception := contexts.GetActorUserPublicId(ctx)
	if exception != nil {
		tx.Rollback()
		return exception
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		rootShelf.Id.String(),
		blockPackIds,
		[]uuid.UUID{actorUserPublicId},
		coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
	); err != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"LeaveMyRootShelf",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueRootShelfPermissionRevoked(
		tx,
		rootShelf.Id.String(),
		rootShelf.Id,
		actorUserPublicId,
	); err != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"LeaveMyRootShelf",
			"Failed to create resource event",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return cexceptions.New(
			"TransactionCommitFailed",
			"RootShelf",
			"Manage",
			"Failed to commit the root shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	return nil
}

func (s *RootShelfService) LeaveMyRootShelves(
	ctx context.Context, requestDto *capi.LeaveMyRootShelvesRequestDto,
) *cexceptions.Exception {
	if err := s.validator.Struct(requestDto); err != nil {
		return apiexceptions.NewShelfException().InvalidDto().WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return exception
	}
	rootShelfIdSet := make(map[uuid.UUID]struct{}, len(requestDto.Body.RootShelves))
	rootShelfIds := make([]uuid.UUID, len(requestDto.Body.RootShelves))
	for index, rootShelfRequestDto := range requestDto.Body.RootShelves {
		if _, exists := rootShelfIdSet[rootShelfRequestDto.RootShelfId]; exists {
			return cexceptions.New(
				"InvalidRequest",
				"RootShelf",
				"ValidateRequest",
				"Root shelf request is invalid",
				http.StatusBadRequest,
			)
		}
		rootShelfIdSet[rootShelfRequestDto.RootShelfId] = struct{}{}
		rootShelfIds[index] = rootShelfRequestDto.RootShelfId
	}
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return cexceptions.New(
			"TransactionBeginFailed",
			"RootShelf",
			"LeaveMyRootShelves",
			"Failed to begin the root shelf leave transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	blockPacks, exception := s.blockPackRepository.GetManyByRootShelfIds(
		rootShelfIds,
		srepositories.WithTransactionDB(tx),
		srepositories.WithOnlyDeleted(stypes.Ternary_Negative),
	)
	if exception != nil {
		tx.Rollback()
		return exception
	}
	blockPackIds := make([]uuid.UUID, len(blockPacks))
	for index, blockPack := range blockPacks {
		blockPackIds[index] = blockPack.Id
	}
	relations, exception := s.usersToShelvesRepository.GetManyByRootShelfIdsAndUserId(
		rootShelfIds,
		actorUserId,
		srepositories.WithTransactionDB(tx),
		srepositories.WithLockingStrength(srepositories.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return exception
	}
	if len(relations) != len(rootShelfIds) {
		tx.Rollback()
		return cexceptions.New(
			"NotFound",
			"RootShelf",
			"Manage",
			"Root shelf was not found",
			http.StatusNotFound,
		)
	}
	for _, relation := range relations {
		if relation.Permission == cenums.AccessControlPermission_Owner {
			tx.Rollback()
			return cexceptions.New(
				"PermissionDenied",
				"RootShelf",
				"ManagePermission",
				"You do not have permission to manage this root shelf",
				http.StatusBadRequest,
			)
		}
	}

	if exception = s.usersToShelvesRepository.DeleteManyByRootShelfIdsAndUserId(
		rootShelfIds,
		actorUserId,
		srepositories.WithTransactionDB(tx),
	); exception != nil {
		tx.Rollback()
		return exception
	}
	actorUserPublicId, exception := contexts.GetActorUserPublicId(ctx)
	if exception != nil {
		tx.Rollback()
		return exception
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueBlockPackAccessRevocations(
		tx,
		"root-shelf-bulk-leave",
		blockPackIds,
		[]uuid.UUID{actorUserPublicId},
		coreevents.BlockPackAccessRevocationReason_PermissionRevoked,
	); err != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"LeaveMyRootShelves",
			"Failed to create lifecycle outbox events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := srepositories.NewOutboxEventRepository().EnqueueManyRootShelfPermissionRevocations(
		tx,
		"root-shelf-bulk-leave",
		rootShelfIds,
		[]uuid.UUID{actorUserPublicId},
	); err != nil {
		tx.Rollback()
		return cexceptions.New(
			"FailedToCreate",
			"Outbox",
			"LeaveMyRootShelves",
			"Failed to create resource events",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return cexceptions.New(
			"TransactionCommitFailed",
			"RootShelf",
			"Manage",
			"Failed to commit the root shelf transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	return nil
}

/* ============================== Service Methods for GraphQL RootShelf ============================== */

func (s *RootShelfService) SearchPrivateRootShelves(
	ctx context.Context, userId uuid.UUID, gqlInput cgqlmodels.SearchRootShelfInput,
) (*cgqlmodels.SearchRootShelfConnection, *cexceptions.Exception) {
	type PrivateRootShelf struct {
		sschemas.RootShelf
		Permission cenums.AccessControlPermission `gorm:"column:permission"`
	}

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

	query := db.Model(&sschemas.RootShelf{}).
		Select(`"RootShelfTable".*, uts.permission AS permission`).
		Joins(`LEFT JOIN "UsersToShelvesTable" uts ON "RootShelfTable".id = uts.root_shelf_id`).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, allowedPermissions).
		Scopes(s.rootShelfScope.FilterOnlyDeleted(onlyDeleted))

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			"name ILIKE ?",
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := ssearchcursor.Decode[cgqlmodels.SearchRootShelfCursorFields](*gqlInput.After)
		if err != nil {
			return nil, cexceptions.New(
				"CursorDecodeFailed",
				"Search",
				"SearchPrivateRootShelves",
				"Failed to decode the search cursor",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}

		query = query.Where("id > ?", searchCursor.Fields.ID)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		var cending string = cgqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == cgqlmodels.SearchSortOrderDesc {
			cending = cgqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case cgqlmodels.SearchRootShelfSortByName:
			query = query.Order("name " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRootShelfSortByLastUpdate:
			query = query.Order("updated_at " + cending).
				Order("name " + cending).
				Order("created_at " + cending)
		case cgqlmodels.SearchRootShelfSortByCreatedAt:
			query = query.Order("created_at " + cending).
				Order("name " + cending).
				Order("updated_at " + cending)
		default:
			query = query.Order("name " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		}
	}

	limit := sconstants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, sconstants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var shelves []PrivateRootShelf
	if err := query.Scopes(s.rootShelfScope.IncludePreloads(
		[]sschemas.RootShelfRelation{
			sschemas.RootShelfRelation_UsersToShelves,
			sschemas.RootShelfRelation_Items,
		},
	)).Find(&shelves).Error; err != nil {
		return nil, cexceptions.New(
			"NotFound",
			"RootShelf",
			"Manage",
			"Root shelf was not found",
			http.StatusNotFound,
		).WithOrigin(err)
	}

	userIds := make([]uuid.UUID, 0)
	userIdsSeen := make(map[uuid.UUID]struct{})
	for _, shelf := range shelves {
		if _, exists := userIdsSeen[shelf.OwnerId]; !exists {
			userIds = append(userIds, shelf.OwnerId)
			userIdsSeen[shelf.OwnerId] = struct{}{}
		}

		for _, usersToShelf := range shelf.UsersToShelves {
			if _, exists := userIdsSeen[usersToShelf.UserId]; exists {
				continue
			}

			userIds = append(userIds, usersToShelf.UserId)
			userIdsSeen[usersToShelf.UserId] = struct{}{}
		}
	}

	users := make([]sschemas.User, 0, len(userIds))
	if len(userIds) > 0 {
		if err := db.
			Where("id IN ?", userIds).
			Find(&users).Error; err != nil {
			return nil, cexceptions.New(
				"NotFound",
				"User",
				"ResolveUser",
				"User was not found",
				http.StatusNotFound,
			).WithOrigin(err)
		}
	}

	publicUsersById := make(map[uuid.UUID]*cgqlmodels.PublicUser, len(users))
	for _, user := range users {
		publicUsersById[user.Id] = user.ToPublicUser()
	}

	hasNextPage := len(shelves) > limit
	searchEdges := make([]*cgqlmodels.SearchRootShelfEdge, len(shelves))

	for index, shelf := range shelves {
		searchCursor := ssearchcursor.SearchCursor[cgqlmodels.SearchRootShelfCursorFields]{
			Fields: cgqlmodels.SearchRootShelfCursorFields{
				ID: shelf.Id,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, cexceptions.New(
				"CursorEncodeFailed",
				"Search",
				"SearchPrivateRootShelves",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, cexceptions.New(
				"CursorEncodingFailed",
				"Search",
				"SearchPrivateRootShelves",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			)
		}

		privateRootShelf := shelf.RootShelf.ToPrivateRootShelf(shelf.Permission)
		owner, exists := publicUsersById[shelf.OwnerId]
		if !exists {
			return nil, cexceptions.New(
				"NotFound",
				"User",
				"ResolveUser",
				"User was not found",
				http.StatusNotFound,
			)
		}

		privateRootShelf.Owner = owner
		for _, usersToShelf := range shelf.UsersToShelves {
			if usersToShelf.UserId == shelf.OwnerId {
				continue
			}

			sharer, exists := publicUsersById[usersToShelf.UserId]
			if !exists {
				return nil, cexceptions.New(
					"NotFound",
					"User",
					"ResolveUser",
					"User was not found",
					http.StatusNotFound,
				)
			}

			privateRootShelf.Sharers = append(privateRootShelf.Sharers, sharer)
		}

		searchEdges[index] = &cgqlmodels.SearchRootShelfEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                privateRootShelf,
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

	return &cgqlmodels.SearchRootShelfConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
