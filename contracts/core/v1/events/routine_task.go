package eventscontract

import (
	"time"

	"github.com/google/uuid"

	croutinetasktypes "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/types/routine-tasks"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type RoutineTaskCompletedData struct {
	RoutineTaskId       uuid.UUID                          `json:"routineTaskId"`
	RoutineTaskRecordId uuid.UUID                          `json:"routineTaskRecordId"`
	RoutineRecordId     uuid.UUID                          `json:"routineRecordId"`
	RoutineId           uuid.UUID                          `json:"routineId"`
	ActorUserPublicId   uuid.UUID                          `json:"actorUserPublicId"`
	Purpose             cenums.RoutineTaskPurpose          `json:"purpose"`
	WorkerId            uuid.UUID                          `json:"workerId"`
	Attempt             int32                              `json:"attempt"`
	CompletedAt         time.Time                          `json:"completedAt"`
	ExecutionResult     *croutinetasktypes.ExecutionResult `json:"executionResult,omitempty"`
}
