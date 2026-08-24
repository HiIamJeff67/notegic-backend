package user

import (
	"context"
	"net/http"

	validator "github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	exceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	apicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-settings"

	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/inputs"
	options "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/options"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type UserSettingServiceInterface interface {
	GetMySetting(ctx context.Context, requestDto *apicontract.GetMySettingRequestDto) (*apicontract.GetMySettingResponseDto, *exceptions.Exception)
	UpdateMySetting(ctx context.Context, requestDto *apicontract.UpdateMySettingRequestDto) (*apicontract.UpdateMySettingResponseDto, *exceptions.Exception)
}

type UserSettingService struct {
	validator             *validator.Validate
	db                    *gorm.DB
	userSettingRepository repositories.UserSettingRepositoryInterface
}

func NewUserSettingService(
	validator *validator.Validate,
	db *gorm.DB,
	userSettingRepository repositories.UserSettingRepositoryInterface,
) UserSettingServiceInterface {
	if db == nil {
		db = data.DB
	}
	return &UserSettingService{
		validator:             validator,
		db:                    db,
		userSettingRepository: userSettingRepository,
	}
}

/* ============================== Service Methods for UserSetting ============================== */

func (s *UserSettingService) GetMySetting(
	ctx context.Context,
	requestDto *apicontract.GetMySettingRequestDto,
) (*apicontract.GetMySettingResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"UserSetting",
			"GetMySetting",
			"User setting request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	db := s.db.WithContext(ctx)

	userSetting, exception := s.userSettingRepository.GetOneByUserId(
		actorUserId,
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.GetMySettingResponseDto{
		Language:             *userSetting.Language.ToContractable(),
		Density:              *userSetting.Density.ToContractable(),
		StartSurface:         *userSetting.StartSurface.ToContractable(),
		ReduceMotion:         userSetting.ReduceMotion,
		LineWrap:             userSetting.LineWrap,
		QuickInsert:          userSetting.QuickInsert,
		PrivatePreviews:      userSetting.PrivatePreviews,
		RoutineNudges:        userSetting.RoutineNudges,
		SyncNotifications:    userSetting.SyncNotifications,
		QuietMode:            userSetting.QuietMode,
		QuietModeStartMinute: userSetting.QuietModeStartMinute,
		QuietModeEndMinute:   userSetting.QuietModeEndMinute,
	}, nil
}

func (s *UserSettingService) UpdateMySetting(
	ctx context.Context,
	requestDto *apicontract.UpdateMySettingRequestDto,
) (*apicontract.UpdateMySettingResponseDto, *exceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, exceptions.New(
			"InvalidRequest",
			"UserSetting",
			"UpdateMySetting",
			"User setting request is invalid",
			http.StatusBadRequest,
		).WithOrigin(err)
	}
	actorUserId, exception := contexts.GetActorUserId(ctx)
	if exception != nil {
		return nil, exception
	}

	db := s.db.WithContext(ctx)
	var language *enums.Language
	if requestDto.Body.Values.Language != nil {
		language = enums.LanguageToStorable(requestDto.Body.Values.Language)
	}
	var density *enums.UserSettingDensity
	if requestDto.Body.Values.Density != nil {
		density = enums.UserSettingDensityToStorable(requestDto.Body.Values.Density)
	}
	var startSurface *enums.UserSettingStartSurface
	if requestDto.Body.Values.StartSurface != nil {
		startSurface = enums.UserSettingStartSurfaceToStorable(requestDto.Body.Values.StartSurface)
	}
	updatedUserSetting, exception := s.userSettingRepository.UpdateOneByUserId(
		actorUserId,
		inputs.PartialUpdateUserSettingInput{
			Values: inputs.UpdateUserSettingInput{
				Language:             language,
				Density:              density,
				StartSurface:         startSurface,
				ReduceMotion:         requestDto.Body.Values.ReduceMotion,
				LineWrap:             requestDto.Body.Values.LineWrap,
				QuickInsert:          requestDto.Body.Values.QuickInsert,
				PrivatePreviews:      requestDto.Body.Values.PrivatePreviews,
				RoutineNudges:        requestDto.Body.Values.RoutineNudges,
				SyncNotifications:    requestDto.Body.Values.SyncNotifications,
				QuietMode:            requestDto.Body.Values.QuietMode,
				QuietModeStartMinute: requestDto.Body.Values.QuietModeStartMinute,
				QuietModeEndMinute:   requestDto.Body.Values.QuietModeEndMinute,
			},
			SetNull: requestDto.Body.SetNull,
		},
		options.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &apicontract.UpdateMySettingResponseDto{
		UpdatedAt: updatedUserSetting.UpdatedAt,
	}, nil
}
