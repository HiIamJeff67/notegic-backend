package realtimetypes

import (
	"bytes"
	"testing"

	"github.com/google/uuid"
)

func TestYjsPersistenceBatchMarshalBytes(t *testing.T) {
	persistenceBatchId := uuid.MustParse("719ea8f4-4fcb-4cee-b2f2-8652c52c343f")
	originConnectionId := uuid.MustParse("a81c9583-98c7-41b1-8f1c-6c35a1a5f152")
	expected := YjsPersistenceBatch{
		PersistenceBatchId: persistenceBatchId,
		OriginConnectionId: &originConnectionId,
		Payload:            []byte{1, 2, 3},
	}

	payload, err := expected.MarshalBytes()
	if err != nil {
		t.Fatalf("MarshalBytes() error = %v", err)
	}

	var actual YjsPersistenceBatch
	if err := actual.UnmarshalBytes(payload); err != nil {
		t.Fatalf("UnmarshalBytes() error = %v", err)
	}
	if actual.PersistenceBatchId != expected.PersistenceBatchId {
		t.Fatalf("PersistenceBatchId = %s, want %s", actual.PersistenceBatchId, expected.PersistenceBatchId)
	}
	if actual.OriginConnectionId == nil || *actual.OriginConnectionId != *expected.OriginConnectionId {
		t.Fatalf("OriginConnectionId = %v, want %s", actual.OriginConnectionId, expected.OriginConnectionId)
	}
	if !bytes.Equal(actual.Payload, expected.Payload) {
		t.Fatalf("Payload = %v, want %v", actual.Payload, expected.Payload)
	}
}

func TestYjsPersistenceBatchUnmarshalBytesAllowsMixedOrigins(t *testing.T) {
	persistenceBatchId := uuid.MustParse("719ea8f4-4fcb-4cee-b2f2-8652c52c343f")
	payload := make([]byte, 35)
	copy(payload[:16], persistenceBatchId[:])
	copy(payload[32:], []byte{1, 2, 3})

	var batch YjsPersistenceBatch
	if err := batch.UnmarshalBytes(payload); err != nil {
		t.Fatalf("UnmarshalBytes() error = %v", err)
	}
	if batch.OriginConnectionId != nil {
		t.Fatalf("OriginConnectionId = %v, want nil", batch.OriginConnectionId)
	}
}
