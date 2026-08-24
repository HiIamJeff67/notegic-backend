package schemas

import (
	"time"

	"github.com/google/uuid"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type BlockPackYjsDocument struct {
	Id                     uuid.UUID  `json:"id" gorm:"column:id; type:uuid; primaryKey; not null; default:gen_random_uuid();"`
	BlockPackId            uuid.UUID  `json:"blockPackId" gorm:"column:block_pack_id; type:uuid; not null; uniqueIndex:block_pack_yjs_document_idx_block_pack_id;"`
	Snapshot               []byte     `json:"snapshot" gorm:"column:snapshot; type:bytea; not null; default:'\\x';"`
	StateVector            []byte     `json:"stateVector" gorm:"column:state_vector; type:bytea; not null; default:'\\x';"`
	LastUpdateSequence     int64      `json:"lastUpdateSequence" gorm:"column:last_update_sequence; type:bigint; not null; default:0;"`
	CompactedUntilSequence int64      `json:"compactedUntilSequence" gorm:"column:compacted_until_sequence; type:bigint; not null; default:0;"`
	LastCompactedAt        *time.Time `json:"lastCompactedAt" gorm:"column:last_compacted_at; type:timestamptz; default:null;"`
	ProjectedUntilSequence int64      `json:"projectedUntilSequence" gorm:"column:projected_until_sequence; type:bigint; not null; default:-1;"`
	DeletedAt              *time.Time `json:"deletedAt" gorm:"column:deleted_at; type:timestamptz; default:null;"`
	UpdatedAt              time.Time  `json:"updatedAt" gorm:"column:updated_at; type:timestamptz; not null; autoUpdateTime:true;"`
	CreatedAt              time.Time  `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	BlockPack BlockPack `json:"blockPack" gorm:"foreignKey:BlockPackId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

func (BlockPackYjsDocument) TableName() string {
	return "BlockPackYjsDocumentTable"
}

type BlockPackYjsDocumentRelation platformpostgres.RelationName

const (
	BlockPackYjsDocumentRelation_BlockPack BlockPackYjsDocumentRelation = "BlockPack"
)
