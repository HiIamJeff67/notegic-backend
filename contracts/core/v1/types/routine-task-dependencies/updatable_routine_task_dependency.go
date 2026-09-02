package coretypes

import "github.com/google/uuid"

type UpdatableRoutineTaskDependency struct {
	RoutineTaskId         uuid.UUID `json:"routineTaskId" validate:"required"`
	PreviousRoutineTaskId uuid.UUID `json:"previousRoutineTaskId" validate:"required"`
	Description           *string   `json:"description" validate:"omitnil,max=128"`
	Progress              *int32    `json:"progress" validate:"omitnil,min=0,max=100"`
}
