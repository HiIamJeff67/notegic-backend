package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/block-packs"
)

type MoveMyBlockPackByParentSubShelfIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			BlockPackId                 uuid.UUID `json:"blockPackId" validate:"required"`
			DestinationParentSubShelfId uuid.UUID `json:"destinationParentSubShelfId" validate:"required"`
		},
		struct{},
		struct{},
	]
}

type MoveMyBlockPackByParentSubShelfIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}

type MoveMyBlockPacksByParentSubShelfIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		coretypes.MovableBlockPack,
		struct{},
		struct{},
	]
}

type MoveMyBlockPacksByParentSubShelfIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}

type MoveMyBlockPacksByParentSubShelfIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			MovedBlockPacks []coretypes.MovableBlockPack `json:"movedBlockPacks" validate:"required,min=1,max=1024,dive"`
		},
		struct{},
		struct{},
	]
}

type MoveMyBlockPacksByParentSubShelfIdsResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
