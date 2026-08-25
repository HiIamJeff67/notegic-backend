package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type RestoreMyRoutineByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RoutineId uuid.UUID `json:"routineId" validate:"required"`
		},
		struct{},
		struct{},
	]
}
type RestoreMyRoutineByIdResponseDto = RoutineResponseDto
type RestoreMyRoutinesByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RoutineIds []uuid.UUID `json:"routineIds" validate:"required,min=1,max=1024"`
		},
		struct{},
		struct{},
	]
}
type RestoreMyRoutinesByIdsResponseDto []RoutineResponseDto
type DeleteMyRoutineByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RoutineId uuid.UUID `json:"routineId" validate:"required"`
		},
		struct{},
		struct{},
	]
}
type DeleteMyRoutineByIdResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}
type DeleteMyRoutinesByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RoutineIds []uuid.UUID `json:"routineIds" validate:"required,min=1,max=1024"`
		},
		struct{},
		struct{},
	]
}
type DeleteMyRoutinesByIdsResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}
type HardDeleteMyRoutineByIdRequestDto = DeleteMyRoutineByIdRequestDto
type HardDeleteMyRoutineByIdResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}
type HardDeleteMyRoutinesByIdsRequestDto = DeleteMyRoutinesByIdsRequestDto
type HardDeleteMyRoutinesByIdsResponseDto struct {
	DeletedAt time.Time `json:"deletedAt"`
}
