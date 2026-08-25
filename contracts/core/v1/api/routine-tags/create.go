package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/routine-tags"
)

type CreateRoutineTagRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		coretypes.CreatableRoutineTag,
		struct{},
		struct{},
	]
}

type CreateRoutineTagResponseDto struct {
	Id        uuid.UUID `json:"id"`
	CreatedAt time.Time `json:"createdAt"`
}

type CreateRoutineTagsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			CreatedRoutineTags []coretypes.CreatableRoutineTag `json:"createdRoutineTags" validate:"required,min=1,max=1024,dive"`
		},
		struct{},
		struct{},
	]
}

type CreateRoutineTagsResponseDto struct {
	Ids       []uuid.UUID `json:"ids"`
	CreatedAt time.Time   `json:"createdAt"`
}
