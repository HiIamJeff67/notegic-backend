package apicontract

import (
	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/routine-task-dependencies"
)

type UpdateRoutineTaskDependencyByRoutineIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		coretypes.UpdatableRoutineTaskDependency,
		struct {
			RoutineId uuid.UUID `json:"routineId" validate:"required"`
		},
		struct{},
	]
}

type UpdateRoutineTaskDependencyByRoutineIdResponseDto = coretypes.RoutineTaskDependency

type UpdateRoutineTaskDependenciesByRoutineIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Dependencies []coretypes.UpdatableRoutineTaskDependency `json:"dependencies" validate:"required,min=1,max=1024,dive"`
		},
		struct {
			RoutineId uuid.UUID `json:"routineId" validate:"required"`
		},
		struct{},
	]
}

type UpdateRoutineTaskDependenciesByRoutineIdResponseDto []coretypes.RoutineTaskDependency
