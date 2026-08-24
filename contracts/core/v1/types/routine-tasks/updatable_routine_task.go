package coretypes

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type UpdatableRoutineTask struct {
	RoutineTaskId uuid.UUID `json:"routineTaskId" validate:"required"`
	Values        struct {
		RoutineId       *uuid.UUID                       `json:"routineId" validate:"omitnil"`
		Title           *string                          `json:"title" validate:"omitnil,min=1,max=128"`
		Purpose         *enumcontract.RoutineTaskPurpose `json:"purpose" validate:"omitnil,isroutinetaskpurpose"`
		Payload         *datatypes.JSON                  `json:"payload" validate:"omitnil,max=16777216"`
		Priority        *int32                           `json:"priority" validate:"omitnil,min=0,max=100"`
		MaxAttempts     *int32                           `json:"maxAttempts" validate:"omitnil,min=1,max=20"`
		Period          *enumcontract.RoutinePeriod      `json:"period" validate:"omitnil,isroutineperiod"`
		NextScheduledAt *time.Time                       `json:"nextScheduledAt" validate:"omitnil"`
	} `json:"values"`
	SetNull *map[string]bool `json:"setNull,omitempty"`
}
