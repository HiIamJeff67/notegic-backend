package apicontract

import (
	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type LeaveMyRootShelfRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RootShelfId uuid.UUID `json:"rootShelfId" validate:"required"`
		},
		struct{},
	]
}

type LeaveMyRootShelfResponseDto struct{}

type LeaveMyRootShelvesRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RootShelves []struct {
				RootShelfId uuid.UUID `json:"rootShelfId" validate:"required"`
			} `json:"rootShelves" validate:"required,min=1,max=1024,dive"`
		},
		struct{},
		struct{},
	]
}

type LeaveMyRootShelvesResponseDto struct{}
