package apicontract

import (
	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/routine-task-dependencies"
)

type GetRoutineTaskDependenciesByRoutineIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RoutineId uuid.UUID `json:"routineId" validate:"required"`
		},
		struct{},
	]
}

type GetRoutineTaskDependenciesByRoutineIdResponseDto []coretypes.RoutineTaskDependency
