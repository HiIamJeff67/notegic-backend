package coretypes

import (
	"github.com/google/uuid"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type CreatableBlockPack struct {
	Id                  *uuid.UUID                  `json:"id" validate:"omitnil"`
	ParentSubShelfId    uuid.UUID                   `json:"parentSubShelfId" validate:"required"`
	Name                string                      `json:"name" validate:"required,min=1,max=128"`
	Icon                *enumcontract.SupportedIcon `json:"icon" validate:"omitnil,issupportedicon"`
	HeaderBackgroundURL *string                     `json:"headerBackgroundURL" validate:"omitnil"`
}
