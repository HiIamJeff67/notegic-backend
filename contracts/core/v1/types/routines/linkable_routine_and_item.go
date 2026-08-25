package coretypes

import (
	"github.com/google/uuid"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type LinkableRoutineAndItem struct {
	RoutineId uuid.UUID      `json:"routineId" validate:"required"`
	ItemId    uuid.UUID      `json:"itemId" validate:"required"`
	ItemType  enums.ItemType `json:"itemType" validate:"required,isitemtype"`
}
