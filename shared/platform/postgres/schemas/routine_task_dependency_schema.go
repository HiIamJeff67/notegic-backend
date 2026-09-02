package schemas

import (
	"time"

	"github.com/google/uuid"

	postgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type RoutineTaskDependency struct {
	RoutineTaskId         uuid.UUID `json:"routineTaskId" gorm:"column:routine_task_id; type:uuid; primaryKey;"`
	PreviousRoutineTaskId uuid.UUID `json:"previousRoutineTaskId" gorm:"column:previous_routine_task_id; type:uuid; primaryKey; index:routine_dependency_idx_previous_routine_task_id;"`
	Description           string    `json:"description" gorm:"column:description; size:128; not null; default:''; check:routine_task_dependency_check_description_length,char_length(description) <= 128;"`
	Progress              int32     `json:"progress" gorm:"column:progress; type:integer; not null; default:0; check:routine_task_dependency_check_progress,progress >= 0 AND progress <= 100;"`
	UpdatedAt             time.Time `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt             time.Time `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	RoutineTask         RoutineTask `json:"routineTask" gorm:"foreignKey:RoutineTaskId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	PreviousRoutineTask RoutineTask `json:"previousRoutineTask" gorm:"foreignKey:PreviousRoutineTaskId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

func (RoutineTaskDependency) TableName() string {
	return postgres.TableName_RoutineDependencyTable.String()
}
