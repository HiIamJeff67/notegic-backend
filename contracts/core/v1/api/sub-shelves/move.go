package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/sub-shelves"
)

type MoveMySubShelfByRootShelfIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			SourceRootShelfId      uuid.UUID  `json:"sourceRootShelfId" validate:"required"`
			SourceSubShelfId       uuid.UUID  `json:"sourceSubShelfId" validate:"required"`
			DestinationRootShelfId uuid.UUID  `json:"destinationRootShelfId" validate:"required"`
			DestinationSubShelfId  *uuid.UUID `json:"destinationSubShelfId" validate:"omitnil"`
		},
		struct{},
		struct{},
	]
}

type MoveMySubShelfByRootShelfIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}

type MoveMySubShelvesByRootShelfIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		coretypes.MovableSubShelf,
		struct{},
		struct{},
	]
}

type MoveMySubShelvesByRootShelfIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}

type MoveMySubShelvesByRootShelfIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			MovedSubShelves []coretypes.MovableSubShelf `json:"movedSubShelves" validate:"required,min=1,max=1024,dive"`
		},
		struct{},
		struct{},
	]
}

type MoveMySubShelvesByRootShelfIdsResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
