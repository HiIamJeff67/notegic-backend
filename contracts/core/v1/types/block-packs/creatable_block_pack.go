package coretypes

import (
	"github.com/google/uuid"

	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type CreatableBlockPack struct {
	Id                  *uuid.UUID            `json:"id" validate:"omitnil"`
	ParentSubShelfId    uuid.UUID             `json:"parentSubShelfId" validate:"required"`
	Name                string                `json:"name" validate:"required,min=1,max=128"`
	Icon                *cenums.SupportedIcon `json:"icon" validate:"omitnil,issupportedicon"`
	HeaderBackgroundURL *string               `json:"headerBackgroundURL" validate:"omitnil"`
}
