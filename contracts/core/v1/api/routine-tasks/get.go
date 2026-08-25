package apicontract

import (
	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type GetMyRoutineTaskByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RoutineTaskId uuid.UUID `json:"routineTaskId" validate:"required"`
			IsDeleted     *bool     `json:"isDeleted" validate:"omitnil"`
		},
		struct{},
	]
}
type GetMyRoutineTaskByIdResponseDto = RoutineTaskResponseDto
type GetAllMyRoutineTasksByRoutineIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RoutineIds []uuid.UUID `json:"routineIds" validate:"required,min=1,max=1024"`
			AreDeleted *bool       `json:"areDeleted" validate:"omitnil"`
		},
		struct{},
	]
}
type GetAllMyRoutineTasksByRoutineIdsResponseDto []RoutineTaskResponseDto
type GetAllMyRoutineTasksRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			AreDeleted *bool `json:"areDeleted" validate:"omitnil"`
		},
		struct{},
	]
}
type GetAllMyRoutineTasksResponseDto []RoutineTaskResponseDto
