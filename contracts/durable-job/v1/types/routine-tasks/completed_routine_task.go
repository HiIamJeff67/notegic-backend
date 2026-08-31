package routinetasktypes

import (
	"time"

	"github.com/google/uuid"
)

type CompletedRoutineTask struct {
	RoutineTaskId       uuid.UUID            `json:"routineTaskId" validate:"required"`
	RoutineTaskRecordId uuid.UUID            `json:"routineTaskRecordId" validate:"required"`
	RoutineRecordId     uuid.UUID            `json:"routineRecordId" validate:"required"`
	CompletedAt         time.Time            `json:"completedAt" validate:"required"`
	PreparedTask        *PreparedRoutineTask `json:"preparedTask" validate:"required"`
	ExecutionResult     *ExecutionResult     `json:"executionResult,omitempty"`
}
