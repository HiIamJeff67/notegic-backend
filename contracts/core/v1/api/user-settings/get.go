package apicontract

import coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
import cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

type GetMySettingRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct{},
		struct{},
	]
}

type GetMySettingResponseDto struct {
	Language             cenums.Language                `json:"language"`
	Density              cenums.UserSettingDensity      `json:"density"`
	StartSurface         cenums.UserSettingStartSurface `json:"startSurface"`
	ReduceMotion         bool                           `json:"reduceMotion"`
	LineWrap             bool                           `json:"lineWrap"`
	QuickInsert          bool                           `json:"quickInsert"`
	PrivatePreviews      bool                           `json:"privatePreviews"`
	RoutineNudges        bool                           `json:"routineNudges"`
	SyncNotifications    bool                           `json:"syncNotifications"`
	QuietMode            bool                           `json:"quietMode"`
	QuietModeStartMinute int64                          `json:"quietModeStartMinute"`
	QuietModeEndMinute   int64                          `json:"quietModeEndMinute"`
}
