package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type HardDeleteMyRoutineTagByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RoutineTagId uuid.UUID `json:"routineTagId" validate:"required"`
		},
		struct{},
	]
}

type HardDeleteMyRoutineTagByIdResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}

type HardDeleteMyRoutineTagsByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RoutineTagIds []uuid.UUID `json:"routineTagIds" validate:"required,min=1,max=1024"`
		},
		struct{},
		struct{},
	]
}

type HardDeleteMyRoutineTagsByIdsResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}
