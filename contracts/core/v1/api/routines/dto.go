package apicontract

import (
	"time"

	"github.com/google/uuid"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type RoutineResponseDto struct {
	Id               uuid.UUID            `json:"id"`
	StationId        uuid.UUID            `json:"stationId"`
	Title            string               `json:"title"`
	Description      string               `json:"description"`
	Status           enums.RoutineStatus  `json:"status"`
	IsPinned         bool                 `json:"isPinned"`
	ScheduledStartAt time.Time            `json:"scheduledStartAt"`
	ScheduledEndAt   time.Time            `json:"scheduledEndAt"`
	Period           *enums.RoutinePeriod `json:"period"`
	Timezone         string               `json:"timezone"`
	DeletedAt        *time.Time           `json:"deletedAt"`
	UpdatedAt        time.Time            `json:"updatedAt"`
	CreatedAt        time.Time            `json:"createdAt"`
	TagIds           []uuid.UUID          `json:"tagIds"`
	TaskIds          []uuid.UUID          `json:"taskIds"`
	ItemIds          []uuid.UUID          `json:"itemIds"`
}
