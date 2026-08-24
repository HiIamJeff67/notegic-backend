package durablejobeventscontract

import (
	"time"

	"github.com/google/uuid"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type RoutineTaskRunningData struct {
	RoutineTaskId       uuid.UUID                `json:"routineTaskId"`
	RoutineTaskRecordId uuid.UUID                `json:"routineTaskRecordId"`
	RoutineId           uuid.UUID                `json:"routineId"`
	ActorUserPublicId   uuid.UUID                `json:"actorUserPublicId"`
	Purpose             enums.RoutineTaskPurpose `json:"purpose"`
	Attempt             int32                    `json:"attempt"`
	StartedAt           time.Time                `json:"startedAt"`
}
