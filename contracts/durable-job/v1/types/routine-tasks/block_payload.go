package routinetasktypes

import (
	"github.com/google/uuid"

	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"
)

type AppendBlockRoutineTaskPayload struct {
	BlockPackId            uuid.UUID                         `json:"blockPackId" validate:"required"`
	Pattern                RoutineTaskPattern                `json:"pattern" validate:"omitempty,dive"`
	ArborizedEditableBlock cblocknote.ArborizedEditableBlock `json:"arborizedEditableBlock" validate:"required"`
}

type UpdateBlockRoutineTaskPayload struct {
	BlockId                uuid.UUID                          `json:"blockId" validate:"required"`
	Pattern                RoutineTaskPattern                 `json:"pattern" validate:"omitempty,dive"`
	ArborizedEditableBlock *cblocknote.ArborizedEditableBlock `json:"arborizedEditableBlock" validate:"required"`
}

type ResetBlockRoutineTaskPayload struct {
	BlockId uuid.UUID `json:"blockId" validate:"required"`
}
