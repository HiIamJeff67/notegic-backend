package inputs

import (
	"time"

	"github.com/google/uuid"
)

type AppendBlockPackYjsUpdateInput struct {
	PersistenceBatchId uuid.UUID  `json:"persistenceBatchId" gorm:"column:persistence_batch_id;"`
	OriginConnectionId *uuid.UUID `json:"originConnectionId" gorm:"column:origin_connection_id;"`
	Payload            []byte     `json:"payload" gorm:"column:payload;"`
}

type ApplyCompactedBlockPackYjsDocumentInput struct {
	BaseCompactedUntilSequence int64  `json:"baseCompactedUntilSequence" gorm:"column:base_compacted_until_sequence;"`
	CutoffSequence             int64  `json:"cutoffSequence" gorm:"column:cutoff_sequence;"`
	Snapshot                   []byte `json:"snapshot" gorm:"column:snapshot;"`
	StateVector                []byte `json:"stateVector" gorm:"column:state_vector;"`
}

type BulkApplyCompactedBlockPackYjsDocumentInput struct {
	BlockPackId uuid.UUID `json:"blockPackId" gorm:"column:block_pack_id;"`
	ApplyCompactedBlockPackYjsDocumentInput
}

type DeleteCompactedBlockPackYjsUpdatesInput struct {
	Before time.Time `json:"before" gorm:"column:compacted_at;"`
	Limit  int       `json:"limit"`
}
