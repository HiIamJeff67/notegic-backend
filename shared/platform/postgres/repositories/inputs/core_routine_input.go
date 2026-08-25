package inputs

import (
	"time"

	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type CreateRoutineInput struct {
	Id               *uuid.UUID            `json:"id" gorm:"column:id;"`
	Title            string                `json:"title" gorm:"column:title;"`
	Description      string                `json:"description" gorm:"column:description;"`
	Status           *cenums.RoutineStatus `json:"status" gorm:"column:status;"`
	IsPinned         *bool                 `json:"isPinned" gorm:"column:is_pinned;"`
	ScheduledStartAt *time.Time            `json:"scheduledStartAt" gorm:"column:scheduled_start_at;"`
	ScheduledEndAt   *time.Time            `json:"scheduledEndAt" gorm:"column:scheduled_end_at;"`
	Period           *cenums.RoutinePeriod `json:"period" gorm:"column:period;"`
	Timezone         *string               `json:"timezone" gorm:"column:timezone;"`
}

type CreateRoutineByStationIdInput struct {
	Id               *uuid.UUID            `json:"id" gorm:"column:id;"`
	StationId        uuid.UUID             `json:"stationId" gorm:"column:station_id;"`
	Title            string                `json:"title" gorm:"column:title;"`
	Description      string                `json:"description" gorm:"column:description;"`
	Status           *cenums.RoutineStatus `json:"status" gorm:"column:status;"`
	IsPinned         *bool                 `json:"isPinned" gorm:"column:is_pinned;"`
	ScheduledStartAt *time.Time            `json:"scheduledStartAt" gorm:"column:scheduled_start_at;"`
	ScheduledEndAt   *time.Time            `json:"scheduledEndAt" gorm:"column:scheduled_end_at;"`
	Period           *cenums.RoutinePeriod `json:"period" gorm:"column:period;"`
	Timezone         *string               `json:"timezone" gorm:"column:timezone;"`
}

type UpdateRoutineByIdInput struct {
	Id                 uuid.UUID                              `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateRoutineInput] `json:"partialUpdateInput"`
}
