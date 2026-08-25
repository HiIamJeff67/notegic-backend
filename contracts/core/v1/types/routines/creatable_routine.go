package coretypes

import (
	"time"

	"github.com/google/uuid"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type CreatableRoutine struct {
	Id               *uuid.UUID           `json:"id" validate:"omitnil"`
	StationId        uuid.UUID            `json:"stationId" validate:"required"`
	Title            string               `json:"title" validate:"required,min=1,max=128"`
	Description      string               `json:"description" validate:"max=1024"`
	Status           *enums.RoutineStatus `json:"status" validate:"omitnil,isroutinestatus"`
	IsPinned         *bool                `json:"isPinned" validate:"omitnil"`
	ScheduledStartAt *time.Time           `json:"scheduledStartAt" validate:"omitnil"`
	ScheduledEndAt   *time.Time           `json:"scheduledEndAt" validate:"omitnil"`
	Period           *enums.RoutinePeriod `json:"period" validate:"omitnil,isroutineperiod"`
	Timezone         *string              `json:"timezone" validate:"omitnil,max=64,istimezone"`
}
