package inputs

import (
	"github.com/google/uuid"
	"gorm.io/datatypes"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type CreateBlockInput struct {
	Id            uuid.UUID       `json:"id" gorm:"column:id;"`
	BlockPackId   uuid.UUID       `json:"blockPackId" gorm:"column:block_pack_id;"`
	ParentBlockId *uuid.UUID      `json:"parentBlockId" gorm:"column:parent_block_id;"`
	PrevBlockId   *uuid.UUID      `json:"prevBlockId" gorm:"column:prev_block_id;"`
	NextBlockId   *uuid.UUID      `json:"nextBlockId" gorm:"column:next_block_id;"`
	Type          enums.BlockType `json:"type" gorm:"column:type;"`
	Props         datatypes.JSON  `json:"props" gorm:"column:props;"`
	Content       datatypes.JSON  `json:"content" gorm:"column:content;"`
}

type UpdateBlockInput struct {
	BlockPackId   *uuid.UUID       `json:"blockPackId" gorm:"column:block_pack_id;"`
	ParentBlockId *uuid.UUID       `json:"parentBlockId" gorm:"column:parent_block_id;"`
	PrevBlockId   *uuid.UUID       `json:"prevBlockId" gorm:"column:prev_block_id;"`
	NextBlockId   *uuid.UUID       `json:"nextBlockId" gorm:"column:next_block_id;"`
	Type          *enums.BlockType `json:"type" gorm:"column:type;"`
	Props         *datatypes.JSON  `json:"props" gorm:"column:props;"`
	Content       *datatypes.JSON  `json:"content" gorm:"column:content;"`
}

type PartialUpdateBlockInput = PartialUpdateInput[UpdateBlockInput]

/* ============================== System Only Input ============================== */

type BulkCheckBlockPermissionInput struct {
	UserId uuid.UUID `json:"userId" gorm:"column:user_id;"`
	Id     uuid.UUID `json:"id" gorm:"column:id;"`
}

type BulkCreateBlockPackContentInput struct {
	UserId      uuid.UUID `json:"userId" gorm:"column:user_id;"`
	BlockPackId uuid.UUID `json:"blockPackId" gorm:"column:block_pack_id;"`
	Blocks      []CreateBlockInput
}

type BulkUpdateBlockInput struct {
	UserId             uuid.UUID                            `json:"userId" gorm:"column:user_id;"`
	Id                 uuid.UUID                            `json:"id" gorm:"column:id;"`
	PartialUpdateInput PartialUpdateInput[UpdateBlockInput] `json:"partialUpdateInput"`
}
