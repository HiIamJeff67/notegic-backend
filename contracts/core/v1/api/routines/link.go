package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/routines"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type LinkRoutineTagByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RoutineId    uuid.UUID `json:"routineId" validate:"required"`
			RoutineTagId uuid.UUID `json:"routineTagId" validate:"required"`
			IsUnlink     bool      `json:"isUnlink"`
		},
		struct{},
		struct{},
	]
}
type LinkRoutineTagByIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
type LinkRoutineTagsByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			LinkedRoutinesAndTags []coretypes.LinkableRoutineAndTag `json:"linkedRoutinesAndTags" validate:"required,min=1,max=1024,dive"`
			IsUnlink              bool                              `json:"isUnlink"`
		},
		struct{},
		struct{},
	]
}
type LinkRoutineTagsByIdsResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
type LinkRoutineItemByIdRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RoutineId uuid.UUID       `json:"routineId" validate:"required"`
			ItemId    uuid.UUID       `json:"itemId" validate:"required"`
			ItemType  cenums.ItemType `json:"itemType" validate:"required,isitemtype"`
			IsUnlink  bool            `json:"isUnlink"`
		},
		struct{},
		struct{},
	]
}
type LinkRoutineItemByIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
type LinkRoutineItemsByIdsRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			LinkedRoutinesAndItems []coretypes.LinkableRoutineAndItem `json:"linkedRoutinesAndItems" validate:"required,min=1,max=1024,dive"`
			IsUnlink               bool                               `json:"isUnlink"`
		},
		struct{},
		struct{},
	]
}
type LinkRoutineItemsByIdsResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
