package coretypes

import (
	"github.com/google/uuid"

	enumcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"
)

type UpdatableBlockPack struct {
	BlockPackId uuid.UUID `json:"blockPackId" validate:"required"`
	Values      struct {
		Name                *string                     `json:"name" validate:"omitnil,min=1,max=128"`
		Icon                *enumcontract.SupportedIcon `json:"icon" validate:"omitnil,issupportedicon"`
		HeaderBackgroundURL *string                     `json:"headerBackgroundURL" validate:"omitnil"`
	} `json:"values"`
	SetNull *map[string]bool `json:"setNull,omitempty"`
}
