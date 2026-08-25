package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type GetMyRoutineTagByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct {
			RoutineTagId uuid.UUID `json:"routineTagId" validate:"required"`
			IsDeleted    *bool     `json:"isDeleted" validate:"omitnil"`
		},
		struct{},
	]
}

type RoutineTagResponseDto struct {
	Id        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Color     string    `json:"color"`
	Icon      *string   `json:"icon"`
	UpdatedAt time.Time `json:"updatedAt"`
	CreatedAt time.Time `json:"createdAt"`
}

type GetMyRoutineTagByIdResponseDto = RoutineTagResponseDto

type GetAllMyRoutineTagsRequestDto struct {
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

type GetAllMyRoutineTagsResponseDto []RoutineTagResponseDto
