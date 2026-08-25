package schemas

import (
	"time"

	"github.com/google/uuid"

	postgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

// UserQuota is the shared per-user quota projection persisted in
// UserQuotaTable. It is mutable state, not a PostgreSQL VIEW, because claim
// transactions must atomically update and lock its counter. CostUnitUsed is
// intentionally generic; the existing database column remains unchanged.
type UserQuota struct {
	Id             uuid.UUID `json:"id" gorm:"column:id;type:uuid;primaryKey;not null;default:gen_random_uuid();"`
	UserId         uuid.UUID `json:"userId" gorm:"column:user_id;type:uuid;not null;unique;"`
	CostUnitUsed   int64     `json:"costUnitUsed" gorm:"column:routine_task_cost_unit_used;type:bigint;not null;default:0;check:user_quota_check_routine_task_cost_unit_used_non_negative,routine_task_cost_unit_used >= 0;"`
	CycleStartedAt time.Time `json:"cycleStartedAt" gorm:"column:cycle_started_at;type:timestamptz;not null;"`
	NextResetAt    time.Time `json:"nextResetAt" gorm:"column:next_reset_at;type:timestamptz;not null;index;"`
	UpdatedAt      time.Time `json:"updatedAt" gorm:"column:updated_at;type:timestamptz;not null;autoUpdateTime:true;"`
	CreatedAt      time.Time `json:"createdAt" gorm:"column:created_at;type:timestamptz;not null;autoCreateTime:true;"`
}

func (UserQuota) TableName() string {
	return postgres.TableName_UserQuotaTable.String()
}
