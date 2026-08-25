package schemas

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	cgqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
)

type Block struct {
	Id            uuid.UUID       `json:"id" gorm:"column:id; type:uuid; primaryKey; not null; default:gen_random_uuid();"`
	BlockPackId   uuid.UUID       `json:"blockPackId" gorm:"column:block_pack_id; type:uuid; not null;"`
	ParentBlockId *uuid.UUID      `json:"parentBlockId" gorm:"column:parent_block_id; type:uuid; check:block_check_parent_block_id_is_not_itself,parent_block_id != id;"`
	PrevBlockId   *uuid.UUID      `json:"prevBlockId" gorm:"column:prev_block_id; type:uuid; check:block_check_prev_block_id_is_not_itself,prev_block_id != id;"`
	NextBlockId   *uuid.UUID      `json:"nextBlockId" gorm:"column:next_block_id; type:uuid; check:block_check_next_block_id_is_not_itself,next_block_id != id;"`
	Type          enums.BlockType `json:"type" gorm:"column:type; type:\"BlockType\"; not null; default:'paragraph';"`
	Props         datatypes.JSON  `json:"props" gorm:"column:props; type:jsonb; not null; default:'{}'; check:block_check_props_size,octet_length(props::text) <= 4096;"`
	Content       datatypes.JSON  `json:"content" gorm:"column:content; type:jsonb; default:'{}'; check:block_check_content_size,octet_length(content::text) <= 16384;"`
	UpdatedAt     time.Time       `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt     time.Time       `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	BlockPack *BlockPack `json:"blockPack" gorm:"foreignKey:BlockPackId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	Parent    *Block     `json:"parent" gorm:"foreignKey:ParentBlockId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	Children  []Block    `json:"children" gorm:"foreignKey:ParentBlockId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	PrevBlock *Block     `json:"prevBlock" gorm:"foreignKey:PrevBlockId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:SET NULL;"`
	NextBlock *Block     `json:"nextBlock" gorm:"foreignKey:NextBlockId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:SET NULL;"`
}

// Root Block Table Name
func (Block) TableName() string {
	return platformpostgres.TableName_BlockTable.String()
}

// Root Block Table Relations
type BlockRelation platformpostgres.RelationName

const (
	BlockRelation_BlockPack BlockRelation = "BlockPack"
	BlockRelation_Parent    BlockRelation = "Parent"
	BlockRelation_Children  BlockRelation = "Children"
	BlockRelation_PrevBlock BlockRelation = "PrevBlock"
	BlockRelation_NextBlock BlockRelation = "NextBlock"
)

/* ============================== Relative Type Conversion ============================== */

func (b *Block) ToPrivateBlock() *cgqlmodels.PrivateBlock {
	childrenIds := make([]uuid.UUID, 0, len(b.Children))
	for _, child := range b.Children {
		childrenIds = append(childrenIds, child.Id)
	}

	return &cgqlmodels.PrivateBlock{
		ID:            b.Id,
		BlockPackID:   b.BlockPackId,
		ParentBlockID: b.ParentBlockId,
		PrevBlockID:   b.PrevBlockId,
		NextBlockID:   b.NextBlockId,
		Type:          b.Type,
		Props:         b.Props,
		Content:       b.Content,
		UpdatedAt:     b.UpdatedAt,
		CreatedAt:     b.CreatedAt,
		ChildrenIds:   childrenIds,
	}
}
