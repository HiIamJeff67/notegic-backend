package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type GetMyStationByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			StationId uuid.UUID `json:"stationId" validate:"required"`
			IsDeleted *bool     `json:"isDeleted,omitempty" validate:"omitnil"`
		},
		struct{},
	]
}

type GetMyStationByIdResponseDto struct {
	Id                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	Description         string     `json:"description"`
	Icon                *string    `json:"icon"`
	HeaderBackgroundURL *string    `json:"headerBackgroundURL"`
	Permission          string     `json:"permission"`
	RoutineCount        int64      `json:"routineCount"`
	DeletedAt           *time.Time `json:"deletedAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	CreatedAt           time.Time  `json:"createdAt"`
}

type GetAllMyStationsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct{},
		struct {
			AreDeleted *bool `json:"areDeleted,omitempty" validate:"omitnil"`
		},
	]
}

type StationSummaryResponseDto struct {
	Id                  uuid.UUID  `json:"id"`
	Name                string     `json:"name"`
	Icon                *string    `json:"icon"`
	HeaderBackgroundURL *string    `json:"headerBackgroundURL"`
	Permission          string     `json:"permission"`
	RoutineCount        int64      `json:"routineCount"`
	DeletedAt           *time.Time `json:"deletedAt"`
	UpdatedAt           time.Time  `json:"updatedAt"`
	CreatedAt           time.Time  `json:"createdAt"`
}

type GetAllMyStationsResponseDto []StationSummaryResponseDto
