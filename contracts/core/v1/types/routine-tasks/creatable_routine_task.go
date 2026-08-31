package coretypes

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type CreatableRoutineTask struct {
	RoutineId              uuid.UUID                 `json:"routineId" validate:"required"`
	Title                  string                    `json:"title" validate:"required,min=1,max=128"`
	Purpose                cenums.RoutineTaskPurpose `json:"purpose" validate:"required,isroutinetaskpurpose"`
	Payload                datatypes.JSON            `json:"payload" validate:"omitempty,max=16777216"`
	Priority               int32                     `json:"priority" validate:"omitempty,min=0,max=100"`
	MaxAttempts            int32                     `json:"maxAttempts" validate:"omitempty,min=1,max=20"`
	PreviousRoutineTaskIds []uuid.UUID               `json:"previousRoutineTaskIds" validate:"omitempty,dive,required"`
}
