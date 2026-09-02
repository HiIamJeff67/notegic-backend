package coretypes

import (
	"time"

	"github.com/google/uuid"
)

type RoutineTaskDependency struct {
	RoutineTaskId         uuid.UUID `json:"routineTaskId"`
	PreviousRoutineTaskId uuid.UUID `json:"previousRoutineTaskId"`
	Description           string    `json:"description"`
	Progress              int32     `json:"progress"`
	UpdatedAt             time.Time `json:"updatedAt"`
	CreatedAt             time.Time `json:"createdAt"`
}
