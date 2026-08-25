package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/routine-tags"
)

type UpdateMyRoutineTagByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Values struct {
				Name  *string `json:"name" validate:"omitnil,min=1,max=128"`
				Color *string `json:"color" validate:"omitnil,ishexcodecolor"`
				Icon  *string `json:"icon" validate:"omitnil,issupportedicon"`
			} `json:"values"`
			SetNull *map[string]bool `json:"setNull,omitempty"`
		},
		struct {
			RoutineTagId uuid.UUID `json:"routineTagId" validate:"required"`
		},
		struct{},
	]
}

type UpdateMyRoutineTagByIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}

type UpdateMyRoutineTagsByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			UpdatedRoutineTags []coretypes.UpdatableRoutineTag `json:"updatedRoutineTags" validate:"required,min=1,max=1024,dive"`
		},
		struct{},
		struct{},
	]
}

type UpdateMyRoutineTagsByIdsResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
