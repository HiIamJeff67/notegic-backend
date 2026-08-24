package schemas

import (
	"time"

	"github.com/google/uuid"
)

type UserQuota struct {
	Id                      uuid.UUID `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid();"`
	UserId                  uuid.UUID `json:"userId" gorm:"column:user_id; type:uuid; not null; unique;"`
	RoutineTaskCostUnitUsed int64     `json:"routineTaskCostUnitUsed" gorm:"column:routine_task_cost_unit_used; type:bigint; not null; default:0; check:user_quota_check_routine_task_cost_unit_used_non_negative,routine_task_cost_unit_used >= 0;"`
	CycleStartedAt          time.Time `json:"cycleStartedAt" gorm:"column:cycle_started_at; type:timestamptz; not null;"`
	NextResetAt             time.Time `json:"nextResetAt" gorm:"column:next_reset_at; type:timestamptz; not null; index;"`
	UpdatedAt               time.Time `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt               time.Time `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	User User `gorm:"foreignKey:UserId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

func (UserQuota) TableName() string {
	return "UserQuotaTable"
}
