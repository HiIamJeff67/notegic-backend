package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type DeleteMyStationByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			StationId uuid.UUID `json:"stationId" validate:"required"`
		},
		struct{},
		struct{},
	]
}

type DeleteMyStationByIdResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}

type DeleteMyStationsByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			StationIds []uuid.UUID `json:"stationIds" validate:"required,min=1,max=1024"`
		},
		struct{},
		struct{},
	]
}

type DeleteMyStationsByIdsResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}

type HardDeleteMyStationByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			StationId uuid.UUID `json:"stationId" validate:"required"`
		},
		struct{},
		struct{},
	]
}

type HardDeleteMyStationByIdResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}

type HardDeleteMyStationsByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			StationIds []uuid.UUID `json:"stationIds" validate:"required,min=1,max=1024"`
		},
		struct{},
		struct{},
	]
}

type HardDeleteMyStationsByIdsResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}
