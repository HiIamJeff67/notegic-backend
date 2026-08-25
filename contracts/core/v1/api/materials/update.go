package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/materials"
)

type UpdateMyMaterialByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Values  coretypes.UpdatableMaterial `json:"values"`
			SetNull *map[string]bool            `json:"setNull,omitempty"`
		},
		struct {
			MaterialId uuid.UUID `json:"materialId" validate:"required"`
		},
		struct{},
	]
}

type UpdateMyMaterialByIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
