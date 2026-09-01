package kafka

import (
	"context"
	"encoding/json"

	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
)

type ConsumerRecord struct {
	Topic     string
	Partition int32
	Offset    int64
	Key       string
}

type ConsumerHandler func(
	ctx context.Context,
	record ConsumerRecord,
	envelope cevent.EventEnvelope[json.RawMessage],
) error
