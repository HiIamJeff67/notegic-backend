package durablejobeventscontract

import (
	"time"

	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type RoutineTaskRunningData struct {
	RoutineTaskId       uuid.UUID                 `json:"routineTaskId"`
	RoutineTaskRecordId uuid.UUID                 `json:"routineTaskRecordId"`
	RoutineId           uuid.UUID                 `json:"routineId"`
	ActorUserPublicId   uuid.UUID                 `json:"actorUserPublicId"`
	Purpose             cenums.RoutineTaskPurpose `json:"purpose"`
	Attempt             int32                     `json:"attempt"`
	StartedAt           time.Time                 `json:"startedAt"`
}
