package apicontract

import (
	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type RestoreMySubShelfByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			SubShelfId uuid.UUID `json:"subShelfId" validate:"required"`
		},
		struct{},
	]
}

type RestoreMySubShelfByIdResponseDto = SubShelfResponseDto

type RestoreMySubShelvesByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			SubShelfIds []uuid.UUID `json:"subShelfIds" validate:"required,min=1,max=1024"`
		},
		struct{},
		struct{},
	]
}

type RestoreMySubShelvesByIdsResponseDto []SubShelfResponseDto
