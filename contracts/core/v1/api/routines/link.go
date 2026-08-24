package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapicontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
	coretypes "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/types/routines"
	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type LinkRoutineTagByIdRequestDto struct {
	coreapicontract.RequestDto[
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
	coreapicontract.RequestDto[
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
	coreapicontract.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			RoutineId uuid.UUID             `json:"routineId" validate:"required"`
			ItemId    uuid.UUID             `json:"itemId" validate:"required"`
			ItemType  enumcontract.ItemType `json:"itemType" validate:"required,isitemtype"`
			IsUnlink  bool                  `json:"isUnlink"`
		},
		struct{},
		struct{},
	]
}
type LinkRoutineItemByIdResponseDto struct {
	UpdatedAt time.Time `json:"updatedAt"`
}
type LinkRoutineItemsByIdsRequestDto struct {
	coreapicontract.RequestDto[
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
