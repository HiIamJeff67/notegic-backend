package apicontract

import (
	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type LeaveMyStationRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			StationId uuid.UUID `json:"stationId" validate:"required"`
		},
		struct{},
	]
}

type LeaveMyStationResponseDto struct{}

type LeaveMyStationsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Stations []struct {
				StationId uuid.UUID `json:"stationId" validate:"required"`
			} `json:"stations" validate:"required,min=1,max=1024,dive"`
		},
		struct{},
		struct{},
	]
}

type LeaveMyStationsResponseDto struct{}
