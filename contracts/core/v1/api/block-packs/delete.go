package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type DeleteMyBlockPackByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			BlockPackId uuid.UUID `json:"blockPackId" validate:"required"`
		},
		struct{},
	]
}

type DeleteMyBlockPackByIdResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}

type DeleteMyBlockPacksByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			BlockPackIds []uuid.UUID `json:"blockPackIds" validate:"required,min=1,max=1024"`
		},
		struct{},
		struct{},
	]
}

type DeleteMyBlockPacksByIdsResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}
