package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/block-packs"
	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type UpdateMyBlockPackByIdRequestDto struct {
	coreapicontract.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Values struct {
				Name                *string                     `json:"name" validate:"omitnil,min=1,max=128"`
				Icon                *enumcontract.SupportedIcon `json:"icon" validate:"omitnil,issupportedicon"`
				HeaderBackgroundURL *string                     `json:"headerBackgroundURL" validate:"omitnil"`
			} `json:"values"`
			SetNull *map[string]bool `json:"setNull,omitempty"`
		},
		struct {
			BlockPackId uuid.UUID `json:"blockPackId" validate:"required"`
		},
		struct{},
	]
}

type UpdateMyBlockPackByIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}

type UpdateMyBlockPacksByIdsRequestDto struct {
	coreapicontract.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			UpdatedBlockPacks []coretypes.UpdatableBlockPack `json:"updatedBlockPacks" validate:"required,min=1,max=1024,dive"`
		},
		struct{},
		struct{},
	]
}

type UpdateMyBlockPacksByIdsResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
