package user

import (
	"context"
	"net/http"

	validator "github.com/go-playground/validator/v10"
	"gorm.io/gorm"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/user-settings"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"

	srepositories "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories"
	sinputs "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres/repositories/inputs"

	contexts "github.com/HiIamJeff67/notegic-backend/runtimes/core/contexts"
)

type UserSettingServiceInterface interface {
	GetMySetting(ctx context.Context, requestDto *capi.GetMySettingRequestDto) (*capi.GetMySettingResponseDto, *cexceptions.Exception)
	UpdateMySetting(ctx context.Context, requestDto *capi.UpdateMySettingRequestDto) (*capi.UpdateMySettingResponseDto, *cexceptions.Exception)
}

type UserSettingService struct {
	validator             *validator.Validate
	db                    *gorm.DB
	userSettingRepository srepositories.UserSettingRepositoryInterface
}

func NewUserSettingService(
	validator *validator.Validate,
	db *gorm.DB,
	userSettingRepository srepositories.UserSettingRepositoryInterface,
) UserSettingServiceInterface {
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
		srepositories.WithDB(db),
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
	var language *cenums.Language
	if requestDto.Body.Values.Language != nil {
		language = requestDto.Body.Values.Language
	}
	var density *cenums.UserSettingDensity
	if requestDto.Body.Values.Density != nil {
		density = requestDto.Body.Values.Density
	}
	var startSurface *cenums.UserSettingStartSurface
	if requestDto.Body.Values.StartSurface != nil {
		startSurface = requestDto.Body.Values.StartSurface
	}
	updatedUserSetting, exception := s.userSettingRepository.UpdateOneByUserId(
		actorUserId,
		sinputs.PartialUpdateUserSettingInput{
			Values: sinputs.UpdateUserSettingInput{
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
		srepositories.WithDB(db),
	)
	if exception != nil {
		return nil, exception
	}

	return &capi.UpdateMySettingResponseDto{
		UpdatedAt: updatedUserSetting.UpdatedAt,
	}, nil
}
