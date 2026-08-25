package schemas

import (
	"time"

	"github.com/google/uuid"

	postgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

type BlockPackYjsUpdate struct {
	Id                 uuid.UUID  `json:"id" gorm:"column:id; type:uuid; primaryKey; not null; default:gen_random_uuid();"`
	BlockPackId        uuid.UUID  `json:"blockPackId" gorm:"column:block_pack_id; type:uuid; not null; uniqueIndex:block_pack_yjs_update_idx_block_pack_id_update_sequence,priority:1;"`
	UpdateSequence     int64      `json:"updateSequence" gorm:"column:update_sequence; type:bigint; not null; uniqueIndex:block_pack_yjs_update_idx_block_pack_id_update_sequence,priority:2;"`
	PersistenceBatchId uuid.UUID  `json:"-" gorm:"column:persistence_batch_id; type:uuid; not null; default:gen_random_uuid(); uniqueIndex:block_pack_yjs_update_idx_block_pack_id_persistence_batch_id,priority:1;"`
	Payload            []byte     `json:"payload" gorm:"column:payload; type:bytea; not null;"`
	OriginConnectionId *uuid.UUID `json:"originConnectionId" gorm:"column:origin_connection_id; type:uuid; default:null;"`
	OriginClientId     *string    `json:"originClientId" gorm:"column:origin_client_id; default:null;"`
	CompactedAt        *time.Time `json:"compactedAt" gorm:"column:compacted_at; type:timestamptz; default:null;"`
	CreatedAt          time.Time  `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`

	// relations
	BlockPack BlockPack `json:"blockPack" gorm:"foreignKey:BlockPackId; references:Id; constraint:OnUpdate:CASCADE, OnDelete:CASCADE;"`
}

func (BlockPackYjsUpdate) TableName() string {
	return postgres.TableName_BlockPackYjsUpdateTable.String()
}

type BlockPackYjsUpdateRelation postgres.RelationName

const (
	BlockPackYjsUpdateRelation_BlockPack BlockPackYjsUpdateRelation = "BlockPack"
)
