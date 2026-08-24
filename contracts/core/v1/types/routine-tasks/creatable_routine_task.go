package coretypes

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type CreatableRoutineTask struct {
	RoutineId       uuid.UUID                       `json:"routineId" validate:"required"`
	Title           string                          `json:"title" validate:"required,min=1,max=128"`
	Purpose         enumcontract.RoutineTaskPurpose `json:"purpose" validate:"required,isroutinetaskpurpose"`
	Payload         datatypes.JSON                  `json:"payload" validate:"omitempty,max=16777216"`
	Priority        int32                           `json:"priority" validate:"omitempty,min=0,max=100"`
	MaxAttempts     int32                           `json:"maxAttempts" validate:"omitempty,min=1,max=20"`
	Period          *enumcontract.RoutinePeriod     `json:"period" validate:"omitnil,isroutineperiod"`
	NextScheduledAt time.Time                       `json:"nextScheduledAt" validate:"required"`
}
