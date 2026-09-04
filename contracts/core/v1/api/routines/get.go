package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type RoutineResponseDto struct {
	Id               uuid.UUID             `json:"id"`
	StationId        uuid.UUID             `json:"stationId"`
	Title            string                `json:"title"`
	Description      string                `json:"description"`
	Status           cenums.RoutineStatus  `json:"status"`
	Phase            *cenums.RoutinePhase  `json:"phase"`
	IsPinned         bool                  `json:"isPinned"`
	ScheduledStartAt time.Time             `json:"scheduledStartAt"`
	ScheduledEndAt   time.Time             `json:"scheduledEndAt"`
	Period           *cenums.RoutinePeriod `json:"period"`
	Timezone         string                `json:"timezone"`
	DeletedAt        *time.Time            `json:"deletedAt"`
	UpdatedAt        time.Time             `json:"updatedAt"`
	CreatedAt        time.Time             `json:"createdAt"`
	TagIds           []uuid.UUID           `json:"tagIds"`
	TaskIds          []uuid.UUID           `json:"taskIds"`
	ItemIds          []uuid.UUID           `json:"itemIds"`
}

type GetMyRoutineByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RoutineId uuid.UUID `json:"routineId" validate:"required"`
			IsDeleted *bool     `json:"isDeleted" validate:"omitnil"`
		},
		struct{},
	]
}
type GetMyRoutineByIdResponseDto = RoutineResponseDto

type GetMyRoutinesByStationIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			StationId  uuid.UUID `json:"stationId" validate:"required"`
			AreDeleted *bool     `json:"areDeleted" validate:"omitnil"`
		},
		struct{},
	]
}
type GetMyRoutinesByStationIdResponseDto []RoutineResponseDto

type GetMyRoutinesByTimeRangeRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			From       time.Time   `json:"from" validate:"required"`
			To         time.Time   `json:"to" validate:"required"`
			StationIds []uuid.UUID `json:"stationIds" validate:"required,min=1,max=1024"`
			AreDeleted *bool       `json:"areDeleted" validate:"omitnil"`
		},
		struct{},
	]
}
type GetMyRoutinesByTimeRangeResponseDto []RoutineResponseDto
