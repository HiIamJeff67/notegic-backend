package routinetasktypes

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

// PreparedRoutineTask is the runtime-neutral result of DurableJob preparation.
// Core remains responsible for validating permissions and applying the payload
// to its own database transaction.
type PreparedRoutineTask struct {
	RoutineTaskId       uuid.UUID                 `json:"routineTaskId" validate:"required"`
	RoutineTaskRecordId uuid.UUID                 `json:"routineTaskRecordId" validate:"required"`
	RoutineId           uuid.UUID                 `json:"routineId" validate:"required"`
	ActorUserId         uuid.UUID                 `json:"actorUserId" validate:"required"`
	ActorUserPublicId   uuid.UUID                 `json:"actorUserPublicId" validate:"required"`
	Attempt             int32                     `json:"attempt" validate:"gte=1"`
	Purpose             cenums.RoutineTaskPurpose `json:"purpose" validate:"required"`
	Payload             json.RawMessage           `json:"payload" validate:"required"`
	PreparedAt          time.Time                 `json:"preparedAt" validate:"required"`
}
