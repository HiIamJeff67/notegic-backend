package routines

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	validator "github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
	constants "github.com/HiIamJeff67/notegic-backend/shared/constants"
	types "github.com/HiIamJeff67/notegic-backend/shared/types"

	searchcursor "github.com/HiIamJeff67/notegic-backend/shared/lib/searchcursor"

	apicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/stations"
	gqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	schemas "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
	scopes "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/scopes"
)

type StationServiceInterface interface {
	GetMyStationById(ctx context.Context, requestDto *apicontract.GetMyStationByIdRequestDto) (*apicontract.GetMyStationByIdResponseDto, *exceptions.Exception)
	GetAllMyStations(ctx context.Context, requestDto *apicontract.GetAllMyStationsRequestDto) (*apicontract.GetAllMyStationsResponseDto, *exceptions.Exception)
	CreateStation(ctx context.Context, requestDto *apicontract.CreateStationRequestDto) (*apicontract.CreateStationResponseDto, *exceptions.Exception)
	CreateStations(ctx context.Context, requestDto *apicontract.CreateStationsRequestDto) (*apicontract.CreateStationsResponseDto, *exceptions.Exception)
	UpdateMyStationById(ctx context.Context, requestDto *apicontract.UpdateMyStationByIdRequestDto) (*apicontract.UpdateMyStationByIdResponseDto, *exceptions.Exception)
	UpdateMyStationsByIds(ctx context.Context, requestDto *apicontract.UpdateMyStationsByIdsRequestDto) (*apicontract.UpdateMyStationsByIdsResponseDto, *exceptions.Exception)
	RestoreMyStationById(ctx context.Context, requestDto *apicontract.RestoreMyStationByIdRequestDto) (*apicontract.RestoreMyStationByIdResponseDto, *exceptions.Exception)
	RestoreMyStationsByIds(ctx context.Context, requestDto *apicontract.RestoreMyStationsByIdsRequestDto) (*apicontract.RestoreMyStationsByIdsResponseDto, *exceptions.Exception)
	DeleteMyStationById(ctx context.Context, requestDto *apicontract.DeleteMyStationByIdRequestDto) (*apicontract.DeleteMyStationByIdResponseDto, *exceptions.Exception)
	DeleteMyStationsByIds(ctx context.Context, requestDto *apicontract.DeleteMyStationsByIdsRequestDto) (*apicontract.DeleteMyStationsByIdsResponseDto, *exceptions.Exception)
	HardDeleteMyStationById(ctx context.Context, requestDto *apicontract.HardDeleteMyStationByIdRequestDto) (*apicontract.HardDeleteMyStationByIdResponseDto, *exceptions.Exception)
	HardDeleteMyStationsByIds(ctx context.Context, requestDto *apicontract.HardDeleteMyStationsByIdsRequestDto) (*apicontract.HardDeleteMyStationsByIdsResponseDto, *exceptions.Exception)

	VisualizeMyTotalCount(ctx context.Context, requestDto *apicontract.VisualizeMyTotalCountRequestDto) (*apicontract.VisualizeMyTotalCountResponseDto, *exceptions.Exception)

	GetMyStationPermission(ctx context.Context, requestDto *apicontract.GetMyStationPermissionRequestDto) (*apicontract.GetMyStationPermissionResponseDto, *exceptions.Exception)
	CreateMyStationPermission(ctx context.Context, requestDto *apicontract.CreateMyStationPermissionRequestDto) (*apicontract.CreateMyStationPermissionResponseDto, *exceptions.Exception)
	UpsertMyStationPermission(ctx context.Context, requestDto *apicontract.UpsertMyStationPermissionRequestDto) (*apicontract.UpsertMyStationPermissionResponseDto, *exceptions.Exception)
	UpsertMyStationPermissions(ctx context.Context, requestDto *apicontract.UpsertMyStationPermissionsRequestDto) (*apicontract.UpsertMyStationPermissionsResponseDto, *exceptions.Exception)
	UpdateMyStationPermission(ctx context.Context, requestDto *apicontract.UpdateMyStationPermissionRequestDto) (*apicontract.UpdateMyStationPermissionResponseDto, *exceptions.Exception)
	TransferMyStationOwnership(ctx context.Context, requestDto *apicontract.TransferMyStationOwnershipRequestDto) (*apicontract.TransferMyStationOwnershipResponseDto, *exceptions.Exception)
	DeleteMyStationPermission(ctx context.Context, requestDto *apicontract.DeleteMyStationPermissionRequestDto) (*apicontract.DeleteMyStationPermissionResponseDto, *exceptions.Exception)
	DeleteMyStationPermissions(ctx context.Context, requestDto *apicontract.DeleteMyStationPermissionsRequestDto) (*apicontract.DeleteMyStationPermissionsResponseDto, *exceptions.Exception)
	LeaveMyStation(ctx context.Context, requestDto *apicontract.LeaveMyStationRequestDto) *exceptions.Exception
	LeaveMyStations(ctx context.Context, requestDto *apicontract.LeaveMyStationsRequestDto) *exceptions.Exception

	SearchPrivateStations(ctx context.Context, userId uuid.UUID, gqlInput gqlmodels.SearchStationInput) (*gqlmodels.SearchStationConnection, *exceptions.Exception)
}

type StationService struct {
	validator                 *validator.Validate
	db                        *gorm.DB
	stationScope              scopes.StationScopeInterface
	stationRepository         repositories.StationRepositoryInterface
	usersToStationsRepository repositories.UsersToStationsRepositoryInterface
}

func NewStationService(
	validator *validator.Validate,
	db *gorm.DB,
	stationScope scopes.StationScopeInterface,
	stationRepository repositories.StationRepositoryInterface,
	usersToStationsRepository repositories.UsersToStationsRepositoryInterface,
) StationServiceInterface {
	if db == nil {
		db = data.DB
	}
	return &StationService{
		validator:                 validator,
		db:                        db,
		stationScope:              stationScope,
		stationRepository:         stationRepository,
		usersToStationsRepository: usersToStationsRepository,
	}
}

/* ============================== Auxiliary Functions ============================== */

type stationPermissionValues struct {
	UserPublicId uuid.UUID
	Permission   enums.AccessControlPermission
	UpdatedAt    time.Time
	CreatedAt    time.Time
}

func (s *StationService) saveMyStationPermission(
	ctx context.Context,
	actorUserId uuid.UUID,
	stationId uuid.UUID,
	targetUserPublicId uuid.UUID,
	permission enums.AccessControlPermission,
	requireExisting *bool,
) (*stationPermissionValues, *exceptions.Exception) {
	if permission == enums.AccessControlPermission_Owner {
		return nil, exceptions.New(
			"PermissionDenied",
			"Station",
			"ManagePermission",
			"You do not have permission to manage this station",
			http.StatusBadRequest,
		)
	}
	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return nil, exceptions.New(
			"TransactionBeginFailed",
			"Station",
			"Manage",
			"Failed to begin the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	station, actorPermission, exception := s.stationRepository.CheckPermissionAndGetOneById(
		stationId,
		actorUserId,
		nil,
		allowedPermissions,
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(types.Ternary_Negative),
		options.WithLockingStrength(options.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	var targetUser schemas.User
	if result := tx.Where("public_id = ?", targetUserPublicId).First(&targetUser); result.Error != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}
	targetPermission, targetException := s.usersToStationsRepository.GetOne(
		station.Id,
		targetUser.Id,
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthUpdate),
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
		return nil, exceptions.New(
			"NoChanges",
			"Station",
			"Manage",
			"No station changes were applied",
			http.StatusNotModified,
		)
	}
	if targetPermission != nil && targetPermission.Permission == enums.AccessControlPermission_Owner {
		tx.Rollback()
		return nil, exceptions.New(
			"PermissionDenied",
			"Station",
			"ManagePermission",
			"You do not have permission to manage this station",
			http.StatusBadRequest,
		)
	}
	if actorPermission != enums.AccessControlPermission_Owner && (permission == enums.AccessControlPermission_Admin || targetPermission != nil && targetPermission.Permission == enums.AccessControlPermission_Admin) {
		tx.Rollback()
		return nil, exceptions.New(
			"PermissionDenied",
			"Station",
			"ManagePermission",
			"You do not have permission to manage this station",
			http.StatusBadRequest,
		)
	}
	var relation *schemas.UsersToStations
	if targetPermission == nil {
		relation, exception = s.usersToStationsRepository.CreateOne(
			station.Id,
			targetUser.Id,
			permission,
			options.WithTransactionDB(tx),
		)
	} else {
		relation, exception = s.usersToStationsRepository.UpdateOne(
			station.Id,
			targetUser.Id,
			permission,
			options.WithTransactionDB(tx),
		)
	}
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"TransactionCommitFailed",
			"Station",
			"Manage",
			"Failed to commit the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	return &stationPermissionValues{
		UserPublicId: targetUser.PublicId,
		Permission:   relation.Permission,
		UpdatedAt:    relation.UpdatedAt,
		CreatedAt:    relation.CreatedAt,
	}, nil
}

/* ============================== Service Methods for Station ============================== */

func (s *StationService) GetMyStationById(
	ctx context.Context,
	requestDto *apicontract.GetMyStationByIdRequestDto,
) (*apicontract.GetMyStationByIdResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
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

	onlyDeleted := types.Ternary_Neutral
	if requestDto.Param.IsDeleted != nil {
		if *requestDto.Param.IsDeleted {
			onlyDeleted = types.Ternary_Positive
		} else {
			onlyDeleted = types.Ternary_Negative
		}
	}

	station, permission, exception := s.stationRepository.GetOneById(
		requestDto.Param.StationId,
		actorUserId,
		nil,
		options.WithDB(db),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(onlyDeleted),
	)
	if exception != nil {
		return nil, exception
	}

	var icon *string
	if station.Icon != nil {
		iconString := station.Icon.String()
		icon = &iconString
	}

	return &apicontract.GetMyStationByIdResponseDto{
		Id:                  station.Id,
		Name:                station.Name,
		Description:         station.Description,
		Icon:                icon,
		HeaderBackgroundURL: station.HeaderBackgroundURL,
		Permission:          permission.String(),
		RoutineCount:        station.RoutineCount,
		DeletedAt:           station.DeletedAt,
		UpdatedAt:           station.UpdatedAt,
		CreatedAt:           station.CreatedAt,
	}, nil
}

func (s *StationService) GetAllMyStations(
	ctx context.Context,
	requestDto *apicontract.GetAllMyStationsRequestDto,
) (*apicontract.GetAllMyStationsResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
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

	onlyDeleted := types.Ternary_Neutral
	if requestDto.Query.AreDeleted != nil {
		if *requestDto.Query.AreDeleted {
			onlyDeleted = types.Ternary_Positive
		} else {
			onlyDeleted = types.Ternary_Negative
		}
	}

	stations, permissions, exception := s.stationRepository.GetAllByUserId(
		actorUserId,
		nil,
		options.WithDB(db),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(onlyDeleted),
	)
	if exception != nil {
		return nil, exception
	}

	responseDto := make(apicontract.GetAllMyStationsResponseDto, len(stations))
	for index, station := range stations {
		var icon *string
		if station.Icon != nil {
			iconString := station.Icon.String()
			icon = &iconString
		}
		responseDto[index] = apicontract.StationSummaryResponseDto{
			Id:                  station.Id,
			Name:                station.Name,
			Icon:                icon,
			HeaderBackgroundURL: station.HeaderBackgroundURL,
			Permission:          permissions[index].String(),
			RoutineCount:        station.RoutineCount,
			DeletedAt:           station.DeletedAt,
			UpdatedAt:           station.UpdatedAt,
			CreatedAt:           station.CreatedAt,
		}
	}

	return &responseDto, nil
}

func (s *StationService) CreateStation(
	ctx context.Context,
	requestDto *apicontract.CreateStationRequestDto,
) (*apicontract.CreateStationResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	var icon *enums.SupportedIcon
	if requestDto.Body.Icon != nil {
		parsedIcon := enums.SupportedIcon(*requestDto.Body.Icon)
		icon = &parsedIcon
	}

	newStationId, exception := s.stationRepository.CreateOne(
		actorUserId,
		inputs.CreateStationInput{
			Id:                  requestDto.Body.Id,
			Name:                requestDto.Body.Name,
			Description:         requestDto.Body.Description,
			Icon:                icon,
			HeaderBackgroundURL: requestDto.Body.HeaderBackgroundURL,
		},
		options.WithDB(s.db.WithContext(ctx)),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.CreateStationResponseDto{
		Id:        *newStationId,
		CreatedAt: time.Now(),
	}, nil
}

func (s *StationService) CreateStations(
	ctx context.Context,
	requestDto *apicontract.CreateStationsRequestDto,
) (*apicontract.CreateStationsResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	input := make([]inputs.CreateStationInput, len(requestDto.Body.CreatedStations))
	for index, createdStation := range requestDto.Body.CreatedStations {
		var icon *enums.SupportedIcon
		if createdStation.Icon != nil {
			parsedIcon := enums.SupportedIcon(*createdStation.Icon)
			icon = &parsedIcon
		}
		input[index] = inputs.CreateStationInput{
			Id:                  createdStation.Id,
			Name:                createdStation.Name,
			Description:         createdStation.Description,
			Icon:                icon,
			HeaderBackgroundURL: createdStation.HeaderBackgroundURL,
		}
	}
	newStationIds, exception := s.stationRepository.CreateMany(
		actorUserId,
		input,
		options.WithDB(s.db.WithContext(ctx)),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.CreateStationsResponseDto{
		Ids:       newStationIds,
		CreatedAt: time.Now(),
	}, nil
}

func (s *StationService) UpdateMyStationById(
	ctx context.Context,
	requestDto *apicontract.UpdateMyStationByIdRequestDto,
) (*apicontract.UpdateMyStationByIdResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	var icon *enums.SupportedIcon
	if requestDto.Body.Values.Icon != nil {
		parsedIcon := enums.SupportedIcon(*requestDto.Body.Values.Icon)
		icon = &parsedIcon
	}

	updatedStation, exception := s.stationRepository.UpdateOneById(
		requestDto.Param.StationId,
		actorUserId,
		inputs.PartialUpdateStationInput{
			Values: inputs.UpdateStationInput{
				Name:                requestDto.Body.Values.Name,
				Description:         requestDto.Body.Values.Description,
				Icon:                icon,
				HeaderBackgroundURL: requestDto.Body.Values.HeaderBackgroundURL,
			},
			SetNull: requestDto.Body.SetNull,
		},
		options.WithDB(s.db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.UpdateMyStationByIdResponseDto{
		UpdatedAt: updatedStation.UpdatedAt,
	}, nil
}

func (s *StationService) UpdateMyStationsByIds(
	ctx context.Context,
	requestDto *apicontract.UpdateMyStationsByIdsRequestDto,
) (*apicontract.UpdateMyStationsByIdsResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	input := make([]inputs.UpdateStationByIdInput, len(requestDto.Body.UpdatedStations))
	for index, updatedStation := range requestDto.Body.UpdatedStations {
		var icon *enums.SupportedIcon
		if updatedStation.Values.Icon != nil {
			parsedIcon := enums.SupportedIcon(*updatedStation.Values.Icon)
			icon = &parsedIcon
		}
		input[index] = inputs.UpdateStationByIdInput{
			Id: updatedStation.StationId,
			PartialUpdateInput: inputs.PartialUpdateInput[inputs.UpdateStationInput]{
				Values: inputs.UpdateStationInput{
					Name:                updatedStation.Values.Name,
					Description:         updatedStation.Values.Description,
					Icon:                icon,
					HeaderBackgroundURL: updatedStation.Values.HeaderBackgroundURL,
				},
				SetNull: updatedStation.SetNull,
			},
		}
	}
	exception = s.stationRepository.UpdateManyByIds(
		actorUserId,
		input,
		options.WithDB(s.db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.UpdateMyStationsByIdsResponseDto{
		UpdatedAt: time.Now(),
	}, nil
}

func (s *StationService) RestoreMyStationById(
	ctx context.Context,
	requestDto *apicontract.RestoreMyStationByIdRequestDto,
) (*apicontract.RestoreMyStationByIdResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	restoredStation, exception := s.stationRepository.RestoreSoftDeletedOneById(
		requestDto.Body.StationId,
		actorUserId,
		options.WithDB(s.db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	var icon *string
	if restoredStation.Icon != nil {
		iconString := restoredStation.Icon.String()
		icon = &iconString
	}

	return &apicontract.RestoreMyStationByIdResponseDto{
		Id:                  restoredStation.Id,
		Name:                restoredStation.Name,
		Description:         restoredStation.Description,
		Icon:                icon,
		HeaderBackgroundURL: restoredStation.HeaderBackgroundURL,
		RoutineCount:        restoredStation.RoutineCount,
		DeletedAt:           restoredStation.DeletedAt,
		UpdatedAt:           restoredStation.UpdatedAt,
		CreatedAt:           restoredStation.CreatedAt,
	}, nil
}

func (s *StationService) RestoreMyStationsByIds(
	ctx context.Context,
	requestDto *apicontract.RestoreMyStationsByIdsRequestDto,
) (*apicontract.RestoreMyStationsByIdsResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	restoredStations, exception := s.stationRepository.RestoreSoftDeletedManyByIds(
		requestDto.Body.StationIds,
		actorUserId,
		options.WithDB(s.db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	responseDto := make(apicontract.RestoreMyStationsByIdsResponseDto, 0, len(restoredStations))
	for _, restoredStation := range restoredStations {
		var icon *string
		if restoredStation.Icon != nil {
			iconString := restoredStation.Icon.String()
			icon = &iconString
		}
		responseDto = append(responseDto, apicontract.RestoreMyStationByIdResponseDto{
			Id:                  restoredStation.Id,
			Name:                restoredStation.Name,
			Description:         restoredStation.Description,
			Icon:                icon,
			HeaderBackgroundURL: restoredStation.HeaderBackgroundURL,
			RoutineCount:        restoredStation.RoutineCount,
			DeletedAt:           restoredStation.DeletedAt,
			UpdatedAt:           restoredStation.UpdatedAt,
			CreatedAt:           restoredStation.CreatedAt,
		})
	}

	return &responseDto, nil
}

func (s *StationService) DeleteMyStationById(
	ctx context.Context,
	requestDto *apicontract.DeleteMyStationByIdRequestDto,
) (*apicontract.DeleteMyStationByIdResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
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

	station, permission, exception := s.stationRepository.CheckPermissionAndGetOneById(
		requestDto.Body.StationId,
		actorUserId,
		nil,
		allowedPermissions,
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(types.Ternary_Negative),
		options.WithLockingStrength(options.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if permission == enums.AccessControlPermission_Owner {
		result := tx.
			Model(&schemas.Station{}).
			Where("id = ?", station.Id).
			Update("deleted_at", time.Now())
		if result.Error != nil {
			tx.Rollback()
			return nil, exceptions.New(
				"FailedToUpdate",
				"Station",
				"Manage",
				"Failed to update the station",
				http.StatusInternalServerError,
				true,
			).WithOrigin(result.Error)
		}
		if result.RowsAffected == 0 {
			tx.Rollback()
			return nil, exceptions.New(
				"NoChanges",
				"Station",
				"Manage",
				"No station changes were applied",
				http.StatusNotModified,
			)
		}
	} else {
		exception = s.usersToStationsRepository.DeleteOne(
			station.Id,
			actorUserId,
			options.WithTransactionDB(tx),
		)
		if exception != nil {
			tx.Rollback()
			return nil, exception
		}
	}

	if err := tx.Commit().Error; err != nil {
		return nil, exceptions.New(
			"TransactionCommitFailed",
			"Station",
			"Manage",
			"Failed to commit the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &apicontract.DeleteMyStationByIdResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *StationService) DeleteMyStationsByIds(
	ctx context.Context,
	requestDto *apicontract.DeleteMyStationsByIdsRequestDto,
) (*apicontract.DeleteMyStationsByIdsResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	exception = s.stationRepository.SoftDeleteManyByIds(
		requestDto.Body.StationIds,
		actorUserId,
		options.WithDB(s.db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.DeleteMyStationsByIdsResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *StationService) HardDeleteMyStationById(
	ctx context.Context,
	requestDto *apicontract.HardDeleteMyStationByIdRequestDto,
) (*apicontract.HardDeleteMyStationByIdResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	exception = s.stationRepository.HardDeleteOneById(
		requestDto.Body.StationId,
		actorUserId,
		options.WithDB(s.db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.HardDeleteMyStationByIdResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

func (s *StationService) HardDeleteMyStationsByIds(
	ctx context.Context,
	requestDto *apicontract.HardDeleteMyStationsByIdsRequestDto,
) (*apicontract.HardDeleteMyStationsByIdsResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	exception = s.stationRepository.HardDeleteManyByIds(
		requestDto.Body.StationIds,
		actorUserId,
		options.WithDB(s.db.WithContext(ctx)),
		options.WithAllowedPermissions(allowedPermissions),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.HardDeleteMyStationsByIdsResponseDto{
		DeletedAt: time.Now(),
	}, nil
}

/* ============================== Service Methods for Visualization ============================== */

func (s *StationService) VisualizeMyTotalCount(
	ctx context.Context, requestDto *apicontract.VisualizeMyTotalCountRequestDto,
) (*apicontract.VisualizeMyTotalCountResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	permission, err := enums.ConvertStringToAccessControlPermission(requestDto.Query.Permission)
	if err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	db := s.db.WithContext(ctx)

	var totals struct {
		StationCount     int64 `gorm:"column:station_count;"`
		RoutineCount     int64 `gorm:"column:routine_count;"`
		RoutineTaskCount int64 `gorm:"column:routine_task_count;"`
		RoutineTagCount  int64 `gorm:"column:routine_tag_count;"`
	}

	if *permission == enums.AccessControlPermission_Owner {
		result := db.Model(&schemas.UserAccount{}).
			Select("station_count, routine_count, routine_tag_count").
			Where(`user_id = ?`, actorUserId).
			Scan(&totals)
		if result.Error != nil {
			return nil, exceptions.New(
				"NotFound",
				"Station",
				"Manage",
				"Station was not found",
				http.StatusNotFound,
			).WithOrigin(result.Error)
		}

		result = db.Model(&schemas.RoutineTask{}).
			Joins(`INNER JOIN "RoutineTable" routine ON routine.id = "RoutineTaskTable".routine_id`).
			Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = routine.station_id`).
			Joins(`INNER JOIN "StationTable" station ON station.id = routine.station_id AND station.deleted_at IS NULL`).
			Where("uts.user_id = ? AND uts.permission = ?", actorUserId, *permission).
			Count(&totals.RoutineTaskCount)
		if result.Error != nil {
			return nil, exceptions.New(
				"NotFound",
				"RoutineTask",
				"ManageStation",
				"Routine task was not found",
				http.StatusNotFound,
			).WithOrigin(result.Error)
		}

		return &apicontract.VisualizeMyTotalCountResponseDto{
			Data: []apicontract.TotalCountDatumResponseDto{
				{
					Id:    "station-total-count",
					X:     "Station Total Count",
					Value: totals.StationCount,
				},
				{
					Id:    "routine-total-count",
					X:     "Routine Total Count",
					Value: totals.RoutineCount,
				},
				{
					Id:    "routine-task-total-count",
					X:     "Routine Task Total Count",
					Value: totals.RoutineTaskCount,
				},
				{
					Id:    "routine-tag-total-count",
					X:     "Routine Tag Total Count",
					Value: totals.RoutineTagCount,
				},
			},
		}, nil
	}

	result := db.Model(&schemas.Station{}).
		Select(`
			COUNT(DISTINCT "StationTable".id) AS station_count,
			COALESCE(SUM("StationTable".routine_count), 0) AS routine_count
		`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = "StationTable".id`).
		Where("uts.user_id = ? AND uts.permission = ?", actorUserId, *permission).
		Where(`"StationTable".deleted_at IS NULL`).
		Scan(&totals)
	if result.Error != nil {
		return nil, exceptions.New(
			"NotFound",
			"Station",
			"Manage",
			"Station was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}

	result = db.Model(&schemas.RoutineTask{}).
		Joins(`INNER JOIN "RoutineTable" routine ON routine.id = "RoutineTaskTable".routine_id`).
		Joins(`INNER JOIN "UsersToStationsTable" uts ON uts.station_id = routine.station_id`).
		Joins(`INNER JOIN "StationTable" station ON station.id = routine.station_id AND station.deleted_at IS NULL`).
		Where("uts.user_id = ? AND uts.permission = ?", actorUserId, *permission).
		Count(&totals.RoutineTaskCount)
	if result.Error != nil {
		return nil, exceptions.New(
			"NotFound",
			"RoutineTask",
			"ManageStation",
			"Routine task was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}

	return &apicontract.VisualizeMyTotalCountResponseDto{
		Data: []apicontract.TotalCountDatumResponseDto{
			{
				Id:    "station-total-count",
				X:     "Station Total Count",
				Value: totals.StationCount,
			},
			{
				Id:    "routine-total-count",
				X:     "Routine Total Count",
				Value: totals.RoutineCount,
			},
			{
				Id:    "routine-task-total-count",
				X:     "Routine Task Total Count",
				Value: totals.RoutineTaskCount,
			},
			{
				Id:    "routine-tag-total-count",
				X:     "Routine Tag Total Count",
				Value: totals.RoutineTagCount,
			},
		},
	}, nil
}

/* ============================== Service Methods for Station Permissions ============================== */

func (s *StationService) GetMyStationPermission(
	ctx context.Context, requestDto *apicontract.GetMyStationPermissionRequestDto,
) (*apicontract.GetMyStationPermissionResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	allowedPermissions, exception := contexts.GetAllowedPermissions(ctx)
	if exception != nil {
		return nil, exception
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	db := s.db.WithContext(ctx)
	if _, _, exception = s.stationRepository.CheckPermissionAndGetOneById(
		requestDto.Param.StationId,
		actorUserId,
		nil,
		allowedPermissions,
		options.WithDB(db),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(types.Ternary_Negative),
	); exception != nil {
		return nil, exception
	}

	var targetUser schemas.User
	if result := db.Where("public_id = ?", requestDto.Param.UserPublicId).First(&targetUser); result.Error != nil {
		return nil, exceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}
	relation, exception := s.usersToStationsRepository.GetOne(
		requestDto.Param.StationId,
		targetUser.Id,
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.GetMyStationPermissionResponseDto{
		UserPublicId: targetUser.PublicId,
		Permission:   relation.Permission.String(),
		UpdatedAt:    relation.UpdatedAt,
		CreatedAt:    relation.CreatedAt,
	}, nil
}

func (s *StationService) CreateMyStationPermission(
	ctx context.Context, requestDto *apicontract.CreateMyStationPermissionRequestDto,
) (*apicontract.CreateMyStationPermissionResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	permission, err := enums.ConvertStringToAccessControlPermission(requestDto.Body.Permission)
	if err != nil {
		return nil, exceptions.InvalidInput("Station").WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	requireExisting := false
	responseDto, exception := s.saveMyStationPermission(
		ctx,
		actorUserId,
		requestDto.Param.StationId,
		requestDto.Param.UserPublicId,
		*permission,
		&requireExisting,
	)
	if exception != nil {
		return nil, exception
	}
	return &apicontract.CreateMyStationPermissionResponseDto{
		UserPublicId: responseDto.UserPublicId,
		Permission:   responseDto.Permission.String(),
		UpdatedAt:    responseDto.UpdatedAt,
		CreatedAt:    responseDto.CreatedAt,
	}, nil
}

func (s *StationService) UpsertMyStationPermission(
	ctx context.Context, requestDto *apicontract.UpsertMyStationPermissionRequestDto,
) (*apicontract.UpsertMyStationPermissionResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	permission, err := enums.ConvertStringToAccessControlPermission(requestDto.Body.Permission)
	if err != nil {
		return nil, exceptions.InvalidInput("Station").WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	responseDto, exception := s.saveMyStationPermission(
		ctx,
		actorUserId,
		requestDto.Param.StationId,
		requestDto.Param.UserPublicId,
		*permission,
		nil,
	)
	if exception != nil {
		return nil, exception
	}
	return &apicontract.UpsertMyStationPermissionResponseDto{
		UserPublicId: responseDto.UserPublicId,
		Permission:   responseDto.Permission.String(),
		UpdatedAt:    responseDto.UpdatedAt,
		CreatedAt:    responseDto.CreatedAt,
	}, nil
}

func (s *StationService) UpsertMyStationPermissions(
	ctx context.Context, requestDto *apicontract.UpsertMyStationPermissionsRequestDto,
) (*apicontract.UpsertMyStationPermissionsResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	userPublicIds := make([]uuid.UUID, len(requestDto.Body.Permissions))
	permissionByPublicId := make(map[uuid.UUID]enums.AccessControlPermission, len(requestDto.Body.Permissions))
	for index, input := range requestDto.Body.Permissions {
		permission, err := enums.ConvertStringToAccessControlPermission(input.Permission)
		if err != nil {
			return nil, exceptions.InvalidInput("Station").WithOrigin(err)
		}
		if *permission == enums.AccessControlPermission_Owner {
			return nil, exceptions.New(
				"PermissionDenied",
				"Station",
				"ManagePermission",
				"You do not have permission to manage this station",
				http.StatusBadRequest,
			)
		}
		if _, exists := permissionByPublicId[input.UserPublicId]; exists {
			return nil, exceptions.New(
				"InvalidRequest",
				"Station",
				"ValidateRequest",
				"Station request is invalid",
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
		return nil, exceptions.New(
			"TransactionBeginFailed",
			"Station",
			"Manage",
			"Failed to begin the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}

	station, actorPermission, exception := s.stationRepository.CheckPermissionAndGetOneById(
		requestDto.Param.StationId,
		actorUserId,
		nil,
		allowedPermissions,
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(types.Ternary_Negative),
		options.WithLockingStrength(options.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	var targetUsers []schemas.User
	result := tx.
		Model(&schemas.User{}).
		Select("id, public_id").
		Where("public_id IN ?", userPublicIds).
		Find(&targetUsers)
	if result.Error != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}
	if len(targetUsers) != len(userPublicIds) {
		tx.Rollback()
		return nil, exceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		)
	}

	userByPublicId := make(map[uuid.UUID]schemas.User, len(targetUsers))
	userById := make(map[uuid.UUID]schemas.User, len(targetUsers))
	for _, user := range targetUsers {
		userByPublicId[user.PublicId] = user
		userById[user.Id] = user
	}

	userIds := make([]uuid.UUID, len(userPublicIds))
	for index, userPublicId := range userPublicIds {
		userIds[index] = userByPublicId[userPublicId].Id
	}

	existingPermissions, exception := s.usersToStationsRepository.GetMany(
		station.Id,
		userIds,
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	existingPermissionByUserId := make(map[uuid.UUID]enums.AccessControlPermission, len(existingPermissions))
	for _, existingPermission := range existingPermissions {
		existingPermissionByUserId[existingPermission.UserId] = existingPermission.Permission
	}

	permissions := make([]enums.AccessControlPermission, len(userIds))
	for index, userId := range userIds {
		user := userById[userId]
		permission := permissionByPublicId[user.PublicId]
		if existingPermissionByUserId[userId] == enums.AccessControlPermission_Owner {
			tx.Rollback()
			return nil, exceptions.New(
				"PermissionDenied",
				"Station",
				"ManagePermission",
				"You do not have permission to manage this station",
				http.StatusBadRequest,
			)
		}
		if actorPermission != enums.AccessControlPermission_Owner &&
			(permission == enums.AccessControlPermission_Admin ||
				existingPermissionByUserId[userId] == enums.AccessControlPermission_Admin) {
			tx.Rollback()
			return nil, exceptions.New(
				"PermissionDenied",
				"Station",
				"ManagePermission",
				"You do not have permission to manage this station",
				http.StatusBadRequest,
			)
		}

		permissions[index] = permission
	}

	updatedPermissions, exception := s.usersToStationsRepository.UpsertMany(
		station.Id,
		userIds,
		permissions,
		options.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"TransactionCommitFailed",
			"Station",
			"Manage",
			"Failed to commit the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	updatedPermissionByUserId := make(map[uuid.UUID]schemas.UsersToStations, len(updatedPermissions))
	for _, updatedPermission := range updatedPermissions {
		updatedPermissionByUserId[updatedPermission.UserId] = updatedPermission
	}

	responseDtos := make([]apicontract.StationPermissionResponseDto, len(userIds))
	for index, userId := range userIds {
		user := userById[userId]
		updatedPermission := updatedPermissionByUserId[userId]
		responseDtos[index] = apicontract.StationPermissionResponseDto{
			UserPublicId: user.PublicId,
			Permission:   updatedPermission.Permission.String(),
			UpdatedAt:    updatedPermission.UpdatedAt,
			CreatedAt:    updatedPermission.CreatedAt,
		}
	}

	return &apicontract.UpsertMyStationPermissionsResponseDto{Permissions: responseDtos}, nil
}

func (s *StationService) UpdateMyStationPermission(
	ctx context.Context, requestDto *apicontract.UpdateMyStationPermissionRequestDto,
) (*apicontract.UpdateMyStationPermissionResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	permission, err := enums.ConvertStringToAccessControlPermission(requestDto.Body.Permission)
	if err != nil {
		return nil, exceptions.InvalidInput("Station").WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}
	requireExisting := true
	responseDto, exception := s.saveMyStationPermission(
		ctx,
		actorUserId,
		requestDto.Param.StationId,
		requestDto.Param.UserPublicId,
		*permission,
		&requireExisting,
	)
	if exception != nil {
		return nil, exception
	}
	return &apicontract.UpdateMyStationPermissionResponseDto{
		UserPublicId: responseDto.UserPublicId,
		Permission:   responseDto.Permission.String(),
		UpdatedAt:    responseDto.UpdatedAt,
		CreatedAt:    responseDto.CreatedAt,
	}, nil
}

func (s *StationService) TransferMyStationOwnership(
	ctx context.Context,
	requestDto *apicontract.TransferMyStationOwnershipRequestDto,
) (*apicontract.TransferMyStationOwnershipResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
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
		return nil, exceptions.New(
			"TransactionBeginFailed",
			"Station",
			"Manage",
			"Failed to begin the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	station, permission, exception := s.stationRepository.CheckPermissionAndGetOneById(
		requestDto.Param.StationId,
		actorUserId,
		nil,
		allowedPermissions,
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(types.Ternary_Negative),
		options.WithLockingStrength(options.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if permission != enums.AccessControlPermission_Owner {
		tx.Rollback()
		return nil, exceptions.New(
			"PermissionDenied",
			"Station",
			"ManagePermission",
			"You do not have permission to manage this station",
			http.StatusBadRequest,
		)
	}

	var actorUser schemas.User
	if result := tx.Select("id, public_id").Where("id = ?", actorUserId).First(&actorUser); result.Error != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}
	var targetUser schemas.User
	if result := tx.Select("id, public_id").Where("public_id = ?", requestDto.Body.TargetUserPublicId).First(&targetUser); result.Error != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}
	if targetUser.Id == actorUserId {
		tx.Rollback()
		return nil, exceptions.New(
			"NoChanges",
			"Station",
			"Manage",
			"No station changes were applied",
			http.StatusNotModified,
		)
	}

	targetMembership, exception := s.usersToStationsRepository.GetOne(
		station.Id,
		targetUser.Id,
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if targetMembership.Permission == enums.AccessControlPermission_Owner {
		tx.Rollback()
		return nil, exceptions.New(
			"NoChanges",
			"Station",
			"Manage",
			"No station changes were applied",
			http.StatusNotModified,
		)
	}

	var accounts []schemas.UserAccount
	result := tx.
		Clauses(clause.Locking{Strength: options.LockingStrengthUpdate}).
		Where("user_id IN ?", []uuid.UUID{actorUserId, targetUser.Id}).
		Order("user_id").
		Find(&accounts)
	if result.Error != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"FailedToUpdate",
			"Station",
			"Manage",
			"Failed to update the station",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	if len(accounts) != 2 {
		tx.Rollback()
		return nil, exceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		)
	}

	if _, exception = s.usersToStationsRepository.UpdateOne(
		station.Id,
		actorUserId,
		enums.AccessControlPermission_Admin,
		options.WithTransactionDB(tx),
	); exception != nil {
		tx.Rollback()
		return nil, exception
	}
	newOwnerMembership, exception := s.usersToStationsRepository.UpdateOne(
		station.Id,
		targetUser.Id,
		enums.AccessControlPermission_Owner,
		options.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	result = tx.Model(&schemas.Station{}).
		Where("id = ?", station.Id).
		Update("owner_id", targetUser.Id)
	if result.Error != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"FailedToUpdate",
			"Station",
			"Manage",
			"Failed to update the station",
			http.StatusInternalServerError,
			true,
		).WithOrigin(result.Error)
	}
	if result.RowsAffected == 0 {
		tx.Rollback()
		return nil, exceptions.New(
			"NotFound",
			"Station",
			"Manage",
			"Station was not found",
			http.StatusNotFound,
		)
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"TransactionCommitFailed",
			"Station",
			"Manage",
			"Failed to commit the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &apicontract.TransferMyStationOwnershipResponseDto{
		StationId:                 station.Id,
		PreviousOwnerUserPublicId: actorUser.PublicId,
		NewOwnerUserPublicId:      targetUser.PublicId,
		UpdatedAt:                 newOwnerMembership.UpdatedAt,
	}, nil
}

func (s *StationService) DeleteMyStationPermission(
	ctx context.Context, requestDto *apicontract.DeleteMyStationPermissionRequestDto,
) (*apicontract.DeleteMyStationPermissionResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
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
		return nil, exceptions.New(
			"TransactionBeginFailed",
			"Station",
			"Manage",
			"Failed to begin the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}

	station, actorPermission, exception := s.stationRepository.CheckPermissionAndGetOneById(
		requestDto.Param.StationId,
		actorUserId,
		nil,
		allowedPermissions,
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(types.Ternary_Negative),
		options.WithLockingStrength(options.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	var targetUser schemas.User
	result := tx.
		Model(&schemas.User{}).
		Where("public_id = ?", requestDto.Param.UserPublicId).
		First(&targetUser)
	if result.Error != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}

	targetPermission, exception := s.usersToStationsRepository.GetOne(
		station.Id,
		targetUser.Id,
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if targetPermission.Permission == enums.AccessControlPermission_Owner {
		tx.Rollback()
		return nil, exceptions.New(
			"PermissionDenied",
			"Station",
			"ManagePermission",
			"You do not have permission to manage this station",
			http.StatusBadRequest,
		)
	}
	if actorPermission != enums.AccessControlPermission_Owner &&
		targetPermission.Permission == enums.AccessControlPermission_Admin {
		tx.Rollback()
		return nil, exceptions.New(
			"PermissionDenied",
			"Station",
			"ManagePermission",
			"You do not have permission to manage this station",
			http.StatusBadRequest,
		)
	}

	exception = s.usersToStationsRepository.DeleteOne(
		station.Id,
		targetUser.Id,
		options.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"TransactionCommitFailed",
			"Station",
			"Manage",
			"Failed to commit the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &apicontract.DeleteMyStationPermissionResponseDto{}, nil
}

func (s *StationService) DeleteMyStationPermissions(
	ctx context.Context, requestDto *apicontract.DeleteMyStationPermissionsRequestDto,
) (*apicontract.DeleteMyStationPermissionsResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}

	userPublicIdSet := make(map[uuid.UUID]struct{}, len(requestDto.Body.UserPublicIds))
	for _, userPublicId := range requestDto.Body.UserPublicIds {
		if _, exists := userPublicIdSet[userPublicId]; exists {
			return nil, exceptions.New(
				"InvalidRequest",
				"Station",
				"ValidateRequest",
				"Station request is invalid",
				http.StatusBadRequest,
			)
		}

		userPublicIdSet[userPublicId] = struct{}{}
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
		return nil, exceptions.New(
			"TransactionBeginFailed",
			"Station",
			"Manage",
			"Failed to begin the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}

	station, actorPermission, exception := s.stationRepository.CheckPermissionAndGetOneById(
		requestDto.Param.StationId,
		actorUserId,
		nil,
		allowedPermissions,
		options.WithTransactionDB(tx),
		options.WithAllowedPermissions(allowedPermissions),
		options.WithOnlyDeleted(types.Ternary_Negative),
		options.WithLockingStrength(options.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	var targetUsers []schemas.User
	result := tx.
		Model(&schemas.User{}).
		Select("id, public_id").
		Where("public_id IN ?", requestDto.Body.UserPublicIds).
		Find(&targetUsers)
	if result.Error != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"NotFound",
			"User",
			"ResolveUser",
			"User was not found",
			http.StatusNotFound,
		).WithOrigin(result.Error)
	}
	if len(targetUsers) != len(requestDto.Body.UserPublicIds) {
		tx.Rollback()
		return nil, exceptions.New(
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

	targetPermissions, exception := s.usersToStationsRepository.GetMany(
		station.Id,
		userIds,
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}
	if len(targetPermissions) != len(userIds) {
		tx.Rollback()
		return nil, exceptions.New(
			"NotFound",
			"Station",
			"Manage",
			"Station was not found",
			http.StatusNotFound,
		)
	}

	for _, targetPermission := range targetPermissions {
		if targetPermission.Permission == enums.AccessControlPermission_Owner {
			tx.Rollback()
			return nil, exceptions.New(
				"PermissionDenied",
				"Station",
				"ManagePermission",
				"You do not have permission to manage this station",
				http.StatusBadRequest,
			)
		}
		if actorPermission != enums.AccessControlPermission_Owner &&
			targetPermission.Permission == enums.AccessControlPermission_Admin {
			tx.Rollback()
			return nil, exceptions.New(
				"PermissionDenied",
				"Station",
				"ManagePermission",
				"You do not have permission to manage this station",
				http.StatusBadRequest,
			)
		}
	}

	exception = s.usersToStationsRepository.DeleteMany(
		station.Id,
		userIds,
		options.WithTransactionDB(tx),
	)
	if exception != nil {
		tx.Rollback()
		return nil, exception
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, exceptions.New(
			"TransactionCommitFailed",
			"Station",
			"Manage",
			"Failed to commit the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}

	return &apicontract.DeleteMyStationPermissionsResponseDto{}, nil
}

func (s *StationService) LeaveMyStation(
	ctx context.Context, requestDto *apicontract.LeaveMyStationRequestDto,
) *exceptions.Exception {
	if err := s.validator.Struct(requestDto); err != nil {
		return exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return exception
	}
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return exceptions.New(
			"TransactionBeginFailed",
			"Station",
			"Manage",
			"Failed to begin the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	station, permission, exception := s.stationRepository.CheckPermissionAndGetOneById(
		requestDto.Param.StationId,
		actorUserId,
		nil,
		nil,
		options.WithTransactionDB(tx),
		options.WithOnlyDeleted(types.Ternary_Negative),
		options.WithLockingStrength(options.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return exception
	}
	if permission == enums.AccessControlPermission_Owner {
		tx.Rollback()
		return exceptions.New(
			"PermissionDenied",
			"Station",
			"ManagePermission",
			"You do not have permission to manage this station",
			http.StatusBadRequest,
		)
	}
	if exception = s.usersToStationsRepository.DeleteOne(
		station.Id,
		actorUserId,
		options.WithTransactionDB(tx),
	); exception != nil {
		tx.Rollback()
		return exception
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return exceptions.New(
			"TransactionCommitFailed",
			"Station",
			"Manage",
			"Failed to commit the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	return nil
}

func (s *StationService) LeaveMyStations(
	ctx context.Context, requestDto *apicontract.LeaveMyStationsRequestDto,
) *exceptions.Exception {
	if err := s.validator.Struct(requestDto); err != nil {
		return exceptions.New(
			"InvalidRequest",
			"Station",
			"ValidateRequest",
			"Station request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return exception
	}
	stationIdSet := make(map[uuid.UUID]struct{}, len(requestDto.Body.Stations))
	stationIds := make([]uuid.UUID, len(requestDto.Body.Stations))
	for index, stationRequestDto := range requestDto.Body.Stations {
		if _, exists := stationIdSet[stationRequestDto.StationId]; exists {
			return exceptions.New(
				"InvalidRequest",
				"Station",
				"ValidateRequest",
				"Station request is invalid",
				http.StatusBadRequest,
			)
		}
		stationIdSet[stationRequestDto.StationId] = struct{}{}
		stationIds[index] = stationRequestDto.StationId
	}
	tx := s.db.WithContext(ctx).Begin()
	if tx.Error != nil {
		return exceptions.New(
			"TransactionBeginFailed",
			"Station",
			"Manage",
			"Failed to begin the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(tx.Error)
	}
	relations, exception := s.usersToStationsRepository.GetManyByStationIdsAndUserId(
		stationIds,
		actorUserId,
		options.WithTransactionDB(tx),
		options.WithLockingStrength(options.LockingStrengthUpdate),
	)
	if exception != nil {
		tx.Rollback()
		return exception
	}
	if len(relations) != len(stationIds) {
		tx.Rollback()
		return exceptions.New(
			"NotFound",
			"Station",
			"Manage",
			"Station was not found",
			http.StatusNotFound,
		)
	}
	for _, relation := range relations {
		if relation.Permission == enums.AccessControlPermission_Owner {
			tx.Rollback()
			return exceptions.New(
				"PermissionDenied",
				"Station",
				"ManagePermission",
				"You do not have permission to manage this station",
				http.StatusBadRequest,
			)
		}
	}

	if exception = s.usersToStationsRepository.DeleteManyByStationIdsAndUserId(
		stationIds,
		actorUserId,
		options.WithTransactionDB(tx),
	); exception != nil {
		tx.Rollback()
		return exception
	}
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return exceptions.New(
			"TransactionCommitFailed",
			"Station",
			"Manage",
			"Failed to commit the station transaction",
			http.StatusInternalServerError,
			true,
		).WithOrigin(err)
	}
	return nil
}

/* ============================== Service Methods for GraphQL Station ============================== */

func (s *StationService) SearchPrivateStations(
	ctx context.Context, userId uuid.UUID, gqlInput gqlmodels.SearchStationInput,
) (*gqlmodels.SearchStationConnection, *exceptions.Exception) {
	type PrivateStation struct {
		schemas.Station
		Permission enums.AccessControlPermission `gorm:"column:permission"`
	}

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

	query := db.Model(&schemas.Station{}).
		Select(`"StationTable".*, uts.permission AS permission`).
		Joins(`LEFT JOIN "UsersToStationsTable" uts ON "StationTable".id = uts.station_id`).
		Where("uts.user_id = ? AND uts.permission IN ?", userId, allowedPermissions).
		Scopes(s.stationScope.FilterOnlyDeleted(onlyDeleted))

	if len(strings.ReplaceAll(gqlInput.Query, " ", "")) > 0 {
		query = query.Where(
			"name ILIKE ?",
			"%"+gqlInput.Query+"%",
		)
	}
	if gqlInput.After != nil && len(strings.ReplaceAll(*gqlInput.After, " ", "")) > 0 {
		searchCursor, err := searchcursor.Decode[gqlmodels.SearchStationCursorFields](*gqlInput.After)
		if err != nil {
			return nil, exceptions.New(
				"CursorDecodeFailed",
				"Search",
				"SearchPrivateStations",
				"Failed to decode the search cursor",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}

		query = query.Where("id > ?", searchCursor.Fields.ID)
	}

	if gqlInput.SortBy != nil && gqlInput.SortOrder != nil {
		var cending string = gqlmodels.SearchSortOrderAsc.String()
		if *gqlInput.SortOrder == gqlmodels.SearchSortOrderDesc {
			cending = gqlmodels.SearchSortOrderDesc.String()
		}

		switch *gqlInput.SortBy {
		case gqlmodels.SearchStationSortByName:
			query = query.Order("name " + cending).
				Order("routine_count " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case gqlmodels.SearchStationSortByRoutineCount:
			query = query.Order("routine_count " + cending).
				Order("name " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		case gqlmodels.SearchStationSortByLastUpdate:
			query = query.Order("updated_at " + cending).
				Order("name " + cending).
				Order("routine_count " + cending).
				Order("created_at " + cending)
		case gqlmodels.SearchStationSortByCreatedAt:
			query = query.Order("created_at " + cending).
				Order("name " + cending).
				Order("routine_count " + cending).
				Order("updated_at " + cending)
		default:
			query = query.Order("name " + cending).
				Order("routine_count " + cending).
				Order("updated_at " + cending).
				Order("created_at " + cending)
		}
	}

	limit := constants.DefaultSearchLimit
	if gqlInput.First != nil && *gqlInput.First > 0 {
		limit = int(*gqlInput.First)
	}
	limit = min(limit, constants.MaxSearchLimit)
	query = query.Limit(limit + 1)

	var stations []PrivateStation
	if err := query.Find(&stations).Error; err != nil {
		return nil, exceptions.New(
			"NotFound",
			"Station",
			"Manage",
			"Station was not found",
			http.StatusNotFound,
		).WithOrigin(err)
	}

	hasNextPage := len(stations) > limit
	searchEdges := make([]*gqlmodels.SearchStationEdge, len(stations))

	for index, station := range stations {
		searchCursor := searchcursor.SearchCursor[gqlmodels.SearchStationCursorFields]{
			Fields: gqlmodels.SearchStationCursorFields{
				ID: station.Id,
			},
		}
		encodedSearchCursor, err := searchCursor.Encode()
		if err != nil {
			return nil, exceptions.New(
				"CursorEncodeFailed",
				"Search",
				"SearchPrivateStations",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			).WithOrigin(err)
		}
		if encodedSearchCursor == nil {
			return nil, exceptions.New(
				"CursorEncodingFailed",
				"Search",
				"SearchPrivateStations",
				"Failed to encode the search cursor",
				http.StatusInternalServerError,
				true,
			)
		}

		searchEdges[index] = &gqlmodels.SearchStationEdge{
			EncodedSearchCursor: *encodedSearchCursor,
			Node:                station.Station.ToPrivateSearchableStation(station.Permission),
		}
	}

	searchPageInfo := &gqlmodels.SearchPageInfo{
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

	return &gqlmodels.SearchStationConnection{
		SearchEdges:    searchEdges,
		SearchPageInfo: searchPageInfo,
		TotalCount:     int32(len(searchEdges)),
		SearchTime:     searchTime,
	}, nil
}
