package apicontract

import (
	"time"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/routine-tasks"
)

type UpdateMyRoutineTaskByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		coretypes.UpdatableRoutineTask,
		struct{},
		struct{},
	]
}
type UpdateMyRoutineTaskByIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
