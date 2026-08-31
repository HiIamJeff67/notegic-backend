package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type HardDeleteMyRoutineTaskByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RoutineTaskId uuid.UUID `json:"routineTaskId" validate:"required"`
		},
		struct{},
		struct{},
	]
}
type HardDeleteMyRoutineTaskByIdResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}
type HardDeleteMyRoutineTasksByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RoutineTaskIds []uuid.UUID `json:"routineTaskIds" validate:"required,min=1,max=1024"`
		},
		struct{},
		struct{},
	]
}
type HardDeleteMyRoutineTasksByIdsResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}
