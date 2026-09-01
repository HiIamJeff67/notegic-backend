package routinetasktypes

import (
	"github.com/google/uuid"

	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

const MaxRoutineTaskBlockPackUpdates = 1000

type GetBlockPackRoutineTaskPayload struct {
	BlockPackId uuid.UUID `json:"blockPackId" validate:"required"`
}

type DeleteBlockPackRoutineTaskPayload struct {
	BlockPackId uuid.UUID `json:"blockPackId" validate:"required"`
}

type CreateBlockPackRoutineTaskTemplateBlock struct {
	ClientId               string                            `json:"clientId" validate:"required"`
	PrevClientId           *string                           `json:"prevClientId" validate:"omitnil"`
	ArborizedEditableBlock cblocknote.ArborizedEditableBlock `json:"arborizedEditableBlock" validate:"required"`
}

type CreateBlockPackRoutineTaskTemplate struct {
	Name                string                                    `json:"name" validate:"required,min=1,max=128"`
	Icon                *cenums.SupportedIcon                     `json:"icon" validate:"omitnil,issupportedicon"`
	HeaderBackgroundURL *string                                   `json:"headerBackgroundURL" validate:"omitnil"`
	Blocks              []CreateBlockPackRoutineTaskTemplateBlock `json:"blocks" validate:"required,min=1"`
}

type CreateBlockPackRoutineTaskPayload struct {
	Id               *uuid.UUID                         `json:"id,omitempty" validate:"omitnil"`
	TargetSubShelfId RoutineTaskObjectReference         `json:"targetSubShelfId" validate:"required"`
	Template         CreateBlockPackRoutineTaskTemplate `json:"template" validate:"required"`
	Pattern          RoutineTaskPattern                 `json:"pattern" validate:"omitempty,dive"`
}

type UpdateBlockPackRoutineTaskPayloadBlock struct {
	BlockId                uuid.UUID                          `json:"blockId" validate:"required"`
	ArborizedEditableBlock *cblocknote.ArborizedEditableBlock `json:"arborizedEditableBlock" validate:"required"`
}

type UpdateBlockPackRoutineTaskPayload struct {
	BlockPackId uuid.UUID                                `json:"blockPackId" validate:"required"`
	Pattern     RoutineTaskPattern                       `json:"pattern" validate:"omitempty,dive"`
	Blocks      []UpdateBlockPackRoutineTaskPayloadBlock `json:"blocks" validate:"required,min=1"`
}
