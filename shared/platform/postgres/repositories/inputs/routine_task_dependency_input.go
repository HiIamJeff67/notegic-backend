package inputs

import "github.com/google/uuid"

type CreateRoutineTaskDependencyInput struct {
	RoutineTaskId         uuid.UUID `json:"routineTaskId" gorm:"column:routine_task_id;"`
	PreviousRoutineTaskId uuid.UUID `json:"previousRoutineTaskId" gorm:"column:previous_routine_task_id;"`
	Description           string    `json:"description" gorm:"column:description;"`
	Progress              int32     `json:"progress" gorm:"column:progress;"`
}

type UpdateRoutineTaskDependencyInput struct {
	RoutineTaskId         uuid.UUID `json:"routineTaskId" gorm:"column:routine_task_id;"`
	PreviousRoutineTaskId uuid.UUID `json:"previousRoutineTaskId" gorm:"column:previous_routine_task_id;"`
	Description           string    `json:"description" gorm:"column:description;"`
	Progress              int32     `json:"progress" gorm:"column:progress;"`
}

type RoutineTaskDependencyKey struct {
	RoutineTaskId         uuid.UUID `json:"routineTaskId"`
	PreviousRoutineTaskId uuid.UUID `json:"previousRoutineTaskId"`
}
