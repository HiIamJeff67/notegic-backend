package apicontract

import (
	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/routine-task-dependencies"
)

type DeleteRoutineTaskDependencyByRoutineIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		coretypes.DeletableRoutineTaskDependency,
		struct {
			RoutineId uuid.UUID `json:"routineId" validate:"required"`
		},
		struct{},
	]
}

type DeleteRoutineTaskDependencyByRoutineIdResponseDto = DeleteRoutineTaskDependenciesByRoutineIdResponseDto

type DeleteRoutineTaskDependenciesByRoutineIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			Dependencies []coretypes.DeletableRoutineTaskDependency `json:"dependencies" validate:"required,min=1,max=1024,dive"`
		},
		struct {
			RoutineId uuid.UUID `json:"routineId" validate:"required"`
		},
		struct{},
	]
}

type DeleteRoutineTaskDependenciesByRoutineIdResponseDto struct {
	DeletedCount int64 `json:"deletedCount"`
}
