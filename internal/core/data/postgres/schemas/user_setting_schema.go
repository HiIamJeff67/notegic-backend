package schemas

import (
	"time"

	"github.com/google/uuid"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type UserSetting struct {
	Id                   uuid.UUID                     `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid();"`
	UserId               uuid.UUID                     `json:"userId" gorm:"column:user_id; type:uuid; not null; unique;"`
	Language             enums.Language                `json:"language" gorm:"column:language; type:\"Language\"; not null; default:'English';"`
	Density              enums.UserSettingDensity      `json:"density" gorm:"column:density; type:\"UserSettingDensity\"; not null; default:'Balanced';"`
	StartSurface         enums.UserSettingStartSurface `json:"startSurface" gorm:"column:start_surface; type:\"UserSettingStartSurface\"; not null; default:'Dashboard';"`
	ReduceMotion         bool                          `json:"reduceMotion" gorm:"column:reduce_motion; type:boolean; not null; default:false;"`
	LineWrap             bool                          `json:"lineWrap" gorm:"column:line_wrap; type:boolean; not null; default:true;"`
	QuickInsert          bool                          `json:"quickInsert" gorm:"column:quick_insert; type:boolean; not null; default:true;"`
	PrivatePreviews      bool                          `json:"privatePreviews" gorm:"column:private_previews; type:boolean; not null; default:false;"`
	RoutineNudges        bool                          `json:"routineNudges" gorm:"column:routine_nudges; type:boolean; not null; default:true;"`
	SyncNotifications    bool                          `json:"syncNotifications" gorm:"column:sync_notifications; type:boolean; not null; default:true;"`
	QuietMode            bool                          `json:"quietMode" gorm:"column:quiet_mode; type:boolean; not null; default:true;"`
	QuietModeStartMinute int64                         `json:"quietModeStartMinute" gorm:"column:quiet_mode_start_minute; type:bigint; not null; default:1320;"`
	QuietModeEndMinute   int64                         `json:"quietModeEndMinute" gorm:"column:quiet_mode_end_minute; type:bigint; not null; default:480;"`
	UpdatedAt            time.Time                     `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
}

// User Setting Table Name
func (UserSetting) TableName() string {
	return "UserSettingTable"
}
