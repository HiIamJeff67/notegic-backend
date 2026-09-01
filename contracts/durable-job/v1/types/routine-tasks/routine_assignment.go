package routinetasktypes

import (
	"time"

	"github.com/google/uuid"
)

type RoutineAssignment struct {
	RoutineId         uuid.UUID               `json:"routineId"`
	RoutineRecordId   uuid.UUID               `json:"routineRecordId"`
	DefinitionVersion int64                   `json:"definitionVersion"`
	ScheduledAt       time.Time               `json:"scheduledAt"`
	RoutineTasks      []RoutineTaskAssignment `json:"routineTasks"`
}
