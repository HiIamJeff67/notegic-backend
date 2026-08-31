package schemas

import (
	"github.com/google/uuid"

	postgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type RoutineTaskDependency struct {
	RoutineTaskId         uuid.UUID `json:"routineTaskId" gorm:"column:routine_task_id; type:uuid; primaryKey;"`
	PreviousRoutineTaskId uuid.UUID `json:"previousRoutineTaskId" gorm:"column:previous_routine_task_id; type:uuid; primaryKey; index:routine_dependency_idx_previous_routine_task_id;"`

	RoutineTask         RoutineTask `json:"routineTask" gorm:"foreignKey:RoutineTaskId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	PreviousRoutineTask RoutineTask `json:"previousRoutineTask" gorm:"foreignKey:PreviousRoutineTaskId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

func (RoutineTaskDependency) TableName() string {
	return postgres.TableName_RoutineDependencyTable.String()
}
