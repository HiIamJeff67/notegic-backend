package schemas

import (
	"time"

	"github.com/google/uuid"

	gqlmodels "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/graphql/models"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"

	enums "github.com/HiIamJeff67/notegic-backend/internal/core/data/postgres/schemas/enums"
)

type BlockPack struct {
	Id                  uuid.UUID            `json:"id" gorm:"column:id; type:uuid; primaryKey; not null; default:gen_random_uuid();"`
	ParentSubShelfId    uuid.UUID            `json:"parentSubShelfId" gorm:"column:parent_sub_shelf_id; type:uuid; not null;"` // Previous unique-name constraint: uniqueIndex:block_pack_idx_parent_sub_shelf_id_name,where:deleted_at IS NULL
	Name                string               `json:"name" gorm:"column:name; size:128; not null; default:'undefined';"`        // Previous unique-name constraint: uniqueIndex:block_pack_idx_parent_sub_shelf_id_name,where:deleted_at IS NULL
	Icon                *enums.SupportedIcon `json:"icon" gorm:"column:icon; type:\"SupportedIcon\"; default:null;"`
	HeaderBackgroundURL *string              `json:"headerBackgroundURL" gorm:"column:header_background_url; default:null;"`
	BlockCount          int64                `json:"blockCount" gorm:"column:block_count; type:bigint; not null; default:0; check:block_pack_check_max_block_count,block_count <= 1000;"`
	DeletedAt           *time.Time           `json:"deletedAt" gorm:"column:deleted_at; type:timestamptz; default:null;"`
	UpdatedAt           time.Time            `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt           time.Time            `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	ParentSubShelf SubShelf              `json:"parentSubShelf" gorm:"foreignKey:ParentSubShelfId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	Blocks         []Block               `json:"blocks" gorm:"foreignKey:BlockPackId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	YjsDocument    *BlockPackYjsDocument `json:"yjsDocument" gorm:"foreignKey:BlockPackId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
	YjsUpdates     []BlockPackYjsUpdate  `json:"yjsUpdates" gorm:"foreignKey:BlockPackId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

// BlockPack Table Name
func (BlockPack) TableName() string {
	return "BlockPackTable"
}

// BlockPack Table Relations
type BlockPackRelation platformpostgres.RelationName

const (
	BlockPackRelation_ParentSubShelf BlockPackRelation = "ParentSubShelf"
	BlockPackRelation_Blocks         BlockPackRelation = "Blocks"
	BlockPackRelation_YjsDocument    BlockPackRelation = "YjsDocument"
	BlockPackRelation_YjsUpdates     BlockPackRelation = "YjsUpdates"
)

/* ============================== Relative Type Conversion ============================== */

func (bp *BlockPack) ToPrivateBlockPack() *gqlmodels.PrivateBlockPack {
	blockIds := make([]uuid.UUID, 0, len(bp.Blocks))
	for _, block := range bp.Blocks {
		blockIds = append(blockIds, block.Id)
	}

	return &gqlmodels.PrivateBlockPack{
		ID:                  bp.Id,
		ParentSubShelfID:    bp.ParentSubShelfId,
		Name:                bp.Name,
		Icon:                bp.Icon.ToContractable(),
		HeaderBackgroundURL: bp.HeaderBackgroundURL,
		BlockCount:          bp.BlockCount,
		DeletedAt:           bp.DeletedAt,
		UpdatedAt:           bp.UpdatedAt,
		CreatedAt:           bp.CreatedAt,
		BlockIds:            blockIds,
	}
}
