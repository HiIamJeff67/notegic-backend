package coretypes

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type UpdatableRoutineTask struct {
	RoutineTaskId uuid.UUID `json:"routineTaskId" validate:"required"`
	Values        struct {
		RoutineId              *uuid.UUID                 `json:"routineId" validate:"omitnil"`
		Title                  *string                    `json:"title" validate:"omitnil,min=1,max=128"`
		Purpose                *cenums.RoutineTaskPurpose `json:"purpose" validate:"omitnil,isroutinetaskpurpose"`
		Payload                *datatypes.JSON            `json:"payload" validate:"omitnil,max=16777216"`
		Priority               *int32                     `json:"priority" validate:"omitnil,min=0,max=100"`
		MaxAttempts            *int32                     `json:"maxAttempts" validate:"omitnil,min=1,max=20"`
		PreviousRoutineTaskIds *[]uuid.UUID               `json:"previousRoutineTaskIds" validate:"omitnil,dive,required"`
	} `json:"values"`
	SetNull *map[string]bool `json:"setNull,omitempty"`
}
