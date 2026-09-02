package apicontract

import (
	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/routine-task-dependencies"
)

type CreateRoutineTaskDependencyByRoutineIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		coretypes.CreatableRoutineTaskDependency,
		struct {
			RoutineId uuid.UUID `json:"routineId" validate:"required"`
		},
		struct{},
	]
}

type CreateRoutineTaskDependencyByRoutineIdResponseDto = coretypes.RoutineTaskDependency

type CreateRoutineTaskDependenciesByRoutineIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Dependencies []coretypes.CreatableRoutineTaskDependency `json:"dependencies" validate:"required,min=1,max=1024,dive"`
		},
		struct {
			RoutineId uuid.UUID `json:"routineId" validate:"required"`
		},
		struct{},
	]
}

type CreateRoutineTaskDependenciesByRoutineIdResponseDto []coretypes.RoutineTaskDependency
