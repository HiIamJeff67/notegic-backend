package apicontract

import (
	"time"

	coreapicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type UpdateMySettingRequestDto struct {
	coreapicontract.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Values struct {
				Language             *enumcontract.Language                `json:"language" validate:"omitnil,islanguage"`
				Density              *enumcontract.UserSettingDensity      `json:"density" validate:"omitnil,oneof=Comfortable Balanced Compact"`
				StartSurface         *enumcontract.UserSettingStartSurface `json:"startSurface" validate:"omitnil,oneof=Dashboard Routines"`
				ReduceMotion         *bool                                 `json:"reduceMotion"`
				LineWrap             *bool                                 `json:"lineWrap"`
				QuickInsert          *bool                                 `json:"quickInsert"`
				PrivatePreviews      *bool                                 `json:"privatePreviews"`
				RoutineNudges        *bool                                 `json:"routineNudges"`
				SyncNotifications    *bool                                 `json:"syncNotifications"`
				QuietMode            *bool                                 `json:"quietMode"`
				QuietModeStartMinute *int64                                `json:"quietModeStartMinute" validate:"omitnil,min=0,max=1439"`
				QuietModeEndMinute   *int64                                `json:"quietModeEndMinute" validate:"omitnil,min=0,max=1439"`
			} `json:"values"`
			SetNull *map[string]bool `json:"setNull,omitempty"`
		},
		struct{},
		struct{},
	]
}

type UpdateMySettingResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
