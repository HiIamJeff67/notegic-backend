package realtimetypes

import (
	"errors"

	"github.com/google/uuid"
)

type YjsPersistenceBatch struct {
	PersistenceBatchId uuid.UUID
	OriginConnectionId *uuid.UUID
	Payload            []byte
}

func (b *YjsPersistenceBatch) UnmarshalBytes(payload []byte) error {
	*b = YjsPersistenceBatch{}

	// [persistenceBatchId:16][originConnectionId:16, zero UUID when mixed][raw Yjs update:n]
	if len(payload) <= 32 {
		return errors.New("invalid yjs persistence batch")
	}

	copy(b.PersistenceBatchId[:], payload[:16])
	if b.PersistenceBatchId == uuid.Nil {
		*b = YjsPersistenceBatch{}

		return errors.New("invalid yjs persistence batch")
	}

	var originConnectionId uuid.UUID
	copy(originConnectionId[:], payload[16:32])
	if originConnectionId != uuid.Nil {
		b.OriginConnectionId = &originConnectionId
	}

	b.Payload = payload[32:]

	return nil
}

func (b YjsPersistenceBatch) MarshalBytes() ([]byte, error) {
	if b.PersistenceBatchId == uuid.Nil || len(b.Payload) == 0 {
		return nil, errors.New("invalid yjs persistence batch")
	}

	payload := make([]byte, 32+len(b.Payload))
	copy(payload[:16], b.PersistenceBatchId[:])
	if b.OriginConnectionId != nil {
		copy(payload[16:32], b.OriginConnectionId[:])
	}
	copy(payload[32:], b.Payload)

	return payload, nil
}
