package routinetasktypes

import (
	"time"

	"github.com/google/uuid"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type FailedRoutineTask struct {
	RoutineTaskId       uuid.UUID                        `json:"routineTaskId" validate:"required"`
	RoutineTaskRecordId uuid.UUID                        `json:"routineTaskRecordId" validate:"required"`
	FailedAt            time.Time                        `json:"failedAt" validate:"required"`
	ErrorCode           enums.RoutineTaskRecordErrorCode `json:"errorCode" validate:"required"`
	ErrorReason         string                           `json:"errorReason" validate:"required,max=256"`
}
