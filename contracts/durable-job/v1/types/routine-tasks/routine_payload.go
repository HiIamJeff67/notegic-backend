package routinetasktypes

import (
	"time"

	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type GetRoutineRoutineTaskPayload struct {
	RoutineId uuid.UUID `json:"routineId" validate:"required"`
}

type DeleteRoutineRoutineTaskPayload struct {
	RoutineId uuid.UUID `json:"routineId" validate:"required"`
}

type CreateRoutineRoutineTaskPayload struct {
	Id               *uuid.UUID            `json:"id" validate:"omitnil"`
	StationId        uuid.UUID             `json:"stationId" validate:"required"`
	Title            string                `json:"title" validate:"required,min=1,max=128"`
	Pattern          RoutineTaskPattern    `json:"pattern" validate:"omitempty,dive"`
	Description      string                `json:"description" validate:"max=1024"`
	IsPinned         *bool                 `json:"isPinned" validate:"omitnil"`
	ScheduledStartAt *time.Time            `json:"scheduledStartAt" validate:"omitnil"`
	ScheduledEndAt   *time.Time            `json:"scheduledEndAt" validate:"omitnil"`
	Period           *cenums.RoutinePeriod `json:"period" validate:"omitnil,isroutineperiod"`
	Timezone         *string               `json:"timezone" validate:"omitnil,max=64,istimezone"`
}

type UpdateRoutineRoutineTaskPayload struct {
	RoutineId        uuid.UUID             `json:"routineId" validate:"required"`
	Title            *string               `json:"title" validate:"omitnil,min=1,max=128"`
	Pattern          RoutineTaskPattern    `json:"pattern" validate:"omitempty,dive"`
	Description      *string               `json:"description" validate:"omitnil,max=1024"`
	IsPinned         *bool                 `json:"isPinned" validate:"omitnil"`
	ScheduledStartAt *time.Time            `json:"scheduledStartAt" validate:"omitnil"`
	ScheduledEndAt   *time.Time            `json:"scheduledEndAt" validate:"omitnil"`
	Period           *cenums.RoutinePeriod `json:"period" validate:"omitnil,isroutineperiod"`
	Timezone         *string               `json:"timezone" validate:"omitnil,max=64,istimezone"`
}
