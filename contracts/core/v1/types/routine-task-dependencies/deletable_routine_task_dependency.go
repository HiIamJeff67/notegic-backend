package coretypes

import "github.com/google/uuid"

type DeletableRoutineTaskDependency struct {
	RoutineTaskId         uuid.UUID `json:"routineTaskId" validate:"required"`
	PreviousRoutineTaskId uuid.UUID `json:"previousRoutineTaskId" validate:"required"`
}
