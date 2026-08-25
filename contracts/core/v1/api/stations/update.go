package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/stations"
)

type UpdateMyStationByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Values struct {
				Name                *string `json:"name,omitempty" validate:"omitnil,min=1,max=128"`
				Description         *string `json:"description,omitempty" validate:"omitnil,max=1024"`
				Icon                *string `json:"icon,omitempty" validate:"omitnil,issupportedicon"`
				HeaderBackgroundURL *string `json:"headerBackgroundURL,omitempty" validate:"omitnil,isimageurl"`
			} `json:"values"`
			SetNull *map[string]bool `json:"setNull,omitempty"`
		},
		struct {
			StationId uuid.UUID `json:"stationId" validate:"required"`
		},
		struct{},
	]
}

type UpdateMyStationByIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}

type UpdateMyStationsByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			UpdatedStations []coretypes.UpdatableStation `json:"updatedStations" validate:"required,min=1,max=1024,dive"`
		},
		struct{},
		struct{},
	]
}

type UpdateMyStationsByIdsResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
