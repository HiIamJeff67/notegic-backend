package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/routines"
)

type UpdateMyRoutineByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		coretypes.UpdatableRoutine,
		struct {
			RoutineId uuid.UUID `json:"routineId" validate:"required"`
		},
		struct{},
	]
}
type UpdateMyRoutineByIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
type UpdateMyRoutinesByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			UpdatedRoutines []coretypes.UpdatableRoutine `json:"updatedRoutines" validate:"required,min=1,max=1024,dive"`
		},
		struct{},
		struct{},
	]
}
type UpdateMyRoutinesByIdsResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
