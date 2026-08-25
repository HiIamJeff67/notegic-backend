package apicontract

import (
	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type RestoreMyMaterialByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			MaterialId uuid.UUID `json:"materialId" validate:"required"`
		},
		struct{},
	]
}

type RestoreMyMaterialByIdResponseDto = MaterialResponseDto

type RestoreMyMaterialsByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			MaterialIds []uuid.UUID `json:"materialIds" validate:"required,min=1,max=1024"`
		},
		struct{},
		struct{},
	]
}

type RestoreMyMaterialsByIdsResponseDto []MaterialResponseDto
