package inputs

import enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"

type CreateUserSettingInput struct {
	Language             *enums.Language                `json:"language" gorm:"column:language;"`
	Density              *enums.UserSettingDensity      `json:"density" gorm:"column:density;"`
	StartSurface         *enums.UserSettingStartSurface `json:"startSurface" gorm:"column:start_surface;"`
	ReduceMotion         *bool                          `json:"reduceMotion" gorm:"column:reduce_motion;"`
	LineWrap             *bool                          `json:"lineWrap" gorm:"column:line_wrap;"`
	QuickInsert          *bool                          `json:"quickInsert" gorm:"column:quick_insert;"`
	PrivatePreviews      *bool                          `json:"privatePreviews" gorm:"column:private_previews;"`
	RoutineNudges        *bool                          `json:"routineNudges" gorm:"column:routine_nudges;"`
	SyncNotifications    *bool                          `json:"syncNotifications" gorm:"column:sync_notifications;"`
	QuietMode            *bool                          `json:"quietMode" gorm:"column:quiet_mode;"`
	QuietModeStartMinute *int64                         `json:"quietModeStartMinute" gorm:"column:quiet_mode_start_minute;"`
	QuietModeEndMinute   *int64                         `json:"quietModeEndMinute" gorm:"column:quiet_mode_end_minute;"`
}

type UpdateUserSettingInput struct {
	Language             *enums.Language                `json:"language" gorm:"column:language;"`
	Density              *enums.UserSettingDensity      `json:"density" gorm:"column:density;"`
	StartSurface         *enums.UserSettingStartSurface `json:"startSurface" gorm:"column:start_surface;"`
	ReduceMotion         *bool                          `json:"reduceMotion" gorm:"column:reduce_motion;"`
	LineWrap             *bool                          `json:"lineWrap" gorm:"column:line_wrap;"`
	QuickInsert          *bool                          `json:"quickInsert" gorm:"column:quick_insert;"`
	PrivatePreviews      *bool                          `json:"privatePreviews" gorm:"column:private_previews;"`
	RoutineNudges        *bool                          `json:"routineNudges" gorm:"column:routine_nudges;"`
	SyncNotifications    *bool                          `json:"syncNotifications" gorm:"column:sync_notifications;"`
	QuietMode            *bool                          `json:"quietMode" gorm:"column:quiet_mode;"`
	QuietModeStartMinute *int64                         `json:"quietModeStartMinute" gorm:"column:quiet_mode_start_minute;"`
	QuietModeEndMinute   *int64                         `json:"quietModeEndMinute" gorm:"column:quiet_mode_end_minute;"`
}

type PartialUpdateUserSettingInput = PartialUpdateInput[UpdateUserSettingInput]
