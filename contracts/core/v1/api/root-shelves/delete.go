package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type DeleteMyRootShelfByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RootShelfId uuid.UUID `json:"rootShelfId" validate:"required"`
		},
		struct{},
		struct{},
	]
}

type DeleteMyRootShelfByIdResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}

type DeleteMyRootShelvesByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RootShelfIds []uuid.UUID `json:"rootShelfIds" validate:"required,min=1,max=1024"`
		},
		struct{},
		struct{},
	]
}

type DeleteMyRootShelvesByIdsResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}
