package inputs

import (
	"time"

	"github.com/google/uuid"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type CreateRoutineInput struct {
	Id               *uuid.UUID           `json:"id" gorm:"column:id;"`
	Title            string               `json:"title" gorm:"column:title;"`
	Description      string               `json:"description" gorm:"column:description;"`
	Status           *enums.RoutineStatus `json:"status" gorm:"column:status;"`
	IsPinned         *bool                `json:"isPinned" gorm:"column:is_pinned;"`
	ScheduledStartAt *time.Time           `json:"scheduledStartAt" gorm:"column:scheduled_start_at;"`
	ScheduledEndAt   *time.Time           `json:"scheduledEndAt" gorm:"column:scheduled_end_at;"`
	Period           *enums.RoutinePeriod `json:"period" gorm:"column:period;"`
	Timezone         *string              `json:"timezone" gorm:"column:timezone;"`
}

type CreateRoutineByStationIdInput struct {
	Id               *uuid.UUID           `json:"id" gorm:"column:id;"`
	StationId        uuid.UUID            `json:"stationId" gorm:"column:station_id;"`
	Title            string               `json:"title" gorm:"column:title;"`
	Description      string               `json:"description" gorm:"column:description;"`
	Status           *enums.RoutineStatus `json:"status" gorm:"column:status;"`
	IsPinned         *bool                `json:"isPinned" gorm:"column:is_pinned;"`
	ScheduledStartAt *time.Time           `json:"scheduledStartAt" gorm:"column:scheduled_start_at;"`
	ScheduledEndAt   *time.Time           `json:"scheduledEndAt" gorm:"column:scheduled_end_at;"`
	Period           *enums.RoutinePeriod `json:"period" gorm:"column:period;"`
	Timezone         *string              `json:"timezone" gorm:"column:timezone;"`
}

type UpdateRoutineInput struct {
	StationId        *uuid.UUID           `json:"stationId" gorm:"column:station_id;"`
	Title            *string              `json:"title" gorm:"column:title;"`
	Description      *string              `json:"description" gorm:"column:description;"`
	Status           *enums.RoutineStatus `json:"status" gorm:"column:status;"`
	IsPinned         *bool                `json:"isPinned" gorm:"column:is_pinned;"`
	ScheduledStartAt *time.Time           `json:"scheduledStartAt" gorm:"column:scheduled_start_at;"`
	ScheduledEndAt   *time.Time           `json:"scheduledEndAt" gorm:"column:scheduled_end_at;"`
	Period           *enums.RoutinePeriod `json:"period" gorm:"column:period;"`
	Timezone         *string              `json:"timezone" gorm:"column:timezone;"`
}

type PartialUpdateRoutineInput = PartialUpdateInput[UpdateRoutineInput]

type UpdateRoutineByIdInput struct {
	Id                 uuid.UUID                              `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateRoutineInput] `json:"partialUpdateInput"`
}

/* ============================== System Only Input ============================== */

type BulkCheckRoutinePermissionInput struct {
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
}

type BulkCreateRoutineInput struct {
	UserId           uuid.UUID            `json:"userId" gorm:"column:user_id;"`
	Id               *uuid.UUID           `json:"id" gorm:"column:id;"`
	StationId        uuid.UUID            `json:"stationId" gorm:"column:station_id;"`
	Title            string               `json:"title" gorm:"column:title;"`
	Description      string               `json:"description" gorm:"column:description;"`
	Status           *enums.RoutineStatus `json:"status" gorm:"column:status;"`
	IsPinned         *bool                `json:"isPinned" gorm:"column:is_pinned;"`
	ScheduledStartAt *time.Time           `json:"scheduledStartAt" gorm:"column:scheduled_start_at;"`
	ScheduledEndAt   *time.Time           `json:"scheduledEndAt" gorm:"column:scheduled_end_at;"`
	Period           *enums.RoutinePeriod `json:"period" gorm:"column:period;"`
	Timezone         *string              `json:"timezone" gorm:"column:timezone;"`
}

type BulkUpdateRoutineInput struct {
	UserId             uuid.UUID                              `json:"userId" gorm:"column:user_id;"`
	Id                 uuid.UUID                              `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateRoutineInput] `json:"partialUpdateInput"`
}
