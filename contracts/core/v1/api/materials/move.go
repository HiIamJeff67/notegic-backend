package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type MoveMyMaterialByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			MaterialId                  uuid.UUID `json:"materialId" validate:"required"`
			DestinationParentSubShelfId uuid.UUID `json:"destinationParentSubShelfId" validate:"required"`
		},
		struct{},
		struct{},
	]
}

type MoveMyMaterialByIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}

type MoveMyMaterialsByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			MaterialIds                 []uuid.UUID `json:"materialIds" validate:"required,min=1,max=100"`
			DestinationParentSubShelfId uuid.UUID   `json:"destinationParentSubShelfId" validate:"required"`
		},
		struct{},
		struct{},
	]
}

type MoveMyMaterialsByIdsResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
