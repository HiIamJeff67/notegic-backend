package apicontract

import coreapicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
import enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"

type GetMySettingRequestDto struct {
	coreapicontract.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct{},
		struct{},
	]
}

type GetMySettingResponseDto struct {
	Language             enumcontract.Language                `json:"language"`
	Density              enumcontract.UserSettingDensity      `json:"density"`
	StartSurface         enumcontract.UserSettingStartSurface `json:"startSurface"`
	ReduceMotion         bool                                 `json:"reduceMotion"`
	LineWrap             bool                                 `json:"lineWrap"`
	QuickInsert          bool                                 `json:"quickInsert"`
	PrivatePreviews      bool                                 `json:"privatePreviews"`
	RoutineNudges        bool                                 `json:"routineNudges"`
	SyncNotifications    bool                                 `json:"syncNotifications"`
	QuietMode            bool                                 `json:"quietMode"`
	QuietModeStartMinute int64                                `json:"quietModeStartMinute"`
	QuietModeEndMinute   int64                                `json:"quietModeEndMinute"`
}
