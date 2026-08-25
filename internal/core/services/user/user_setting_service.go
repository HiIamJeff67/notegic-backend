package user

import (
	"context"
	inputs "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories/inputs"
	"net/http"

	validator "github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-settings"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	contexts "github.com/HiIamJeff67/notegic-backend/internal/core/contexts"
	data "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres"
	repositories "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/repositories"
	options "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
)

type UserSettingServiceInterface interface {
	GetMySetting(ctx context.Context, requestDto *capi.GetMySettingRequestDto) (*capi.GetMySettingResponseDto, *cexceptions.Exception)
	UpdateMySetting(ctx context.Context, requestDto *capi.UpdateMySettingRequestDto) (*capi.UpdateMySettingResponseDto, *cexceptions.Exception)
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
	requestDto *capi.GetMySettingRequestDto,
) (*capi.GetMySettingResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, cexceptions.New(
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

	return &capi.GetMySettingResponseDto{
		Language:             userSetting.Language,
		Density:              userSetting.Density,
		StartSurface:         userSetting.StartSurface,
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
	requestDto *capi.UpdateMySettingRequestDto,
) (*capi.UpdateMySettingResponseDto, *cexceptions.Exception) {
	if err := s.validator.Struct(requestDto); err != nil {
		return nil, cexceptions.New(
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
		language = requestDto.Body.Values.Language
	}
	var density *enums.UserSettingDensity
	if requestDto.Body.Values.Density != nil {
		density = requestDto.Body.Values.Density
	}
	var startSurface *enums.UserSettingStartSurface
	if requestDto.Body.Values.StartSurface != nil {
		startSurface = requestDto.Body.Values.StartSurface
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

	return &capi.UpdateMySettingResponseDto{
		UpdatedAt: updatedUserSetting.UpdatedAt,
	}, nil
}
