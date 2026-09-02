package coretypes

import "github.com/google/uuid"

type CreatableRoutineTaskDependency struct {
	RoutineTaskId         uuid.UUID `json:"routineTaskId" validate:"required"`
	PreviousRoutineTaskId uuid.UUID `json:"previousRoutineTaskId" validate:"required"`
	Description           string    `json:"description" validate:"max=128"`
	Progress              int32     `json:"progress" validate:"min=0,max=100"`
}
