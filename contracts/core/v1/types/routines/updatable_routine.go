package coretypes

import (
	"time"

	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type UpdatableRoutine struct {
	RoutineId uuid.UUID `json:"routineId" validate:"required"`
	Values    struct {
		StationId        *uuid.UUID            `json:"stationId" validate:"omitnil"`
		Title            *string               `json:"title" validate:"omitnil,min=1,max=128"`
		Description      *string               `json:"description" validate:"omitnil,max=1024"`
		Status           *cenums.RoutineStatus `json:"status" validate:"omitnil,isroutinestatus"`
		IsPinned         *bool                 `json:"isPinned" validate:"omitnil"`
		ScheduledStartAt *time.Time            `json:"scheduledStartAt" validate:"omitnil"`
		ScheduledEndAt   *time.Time            `json:"scheduledEndAt" validate:"omitnil"`
		Period           *cenums.RoutinePeriod `json:"period" validate:"omitnil,isroutineperiod"`
		Timezone         *string               `json:"timezone" validate:"omitnil,max=64,istimezone"`
	} `json:"values"`
	SetNull *map[string]bool `json:"setNull,omitempty"`
}
