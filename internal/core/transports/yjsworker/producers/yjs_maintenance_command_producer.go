package adaptersproducers

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"

	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	cyjsworkerevents "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1/events"

	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
)

type YjsMaintenanceCommandProducer struct {
	producer *platformkafka.Producer
}

func NewYjsMaintenanceCommandProducer(producer *platformkafka.Producer) *YjsMaintenanceCommandProducer {
	return &YjsMaintenanceCommandProducer{producer: producer}
}

func (p *YjsMaintenanceCommandProducer) Produce(
	ctx context.Context,
	source cevent.EventEnvelope[json.RawMessage],
	request cyjsworkerevents.YjsMaintenanceCommandData,
) error {
	command := cevent.EventEnvelope[cyjsworkerevents.YjsMaintenanceCommandData]{
		SchemaVersion: cevent.Version,
		EventId:       uuid.New(),
		EventType:     cyjsworkerevents.EventType_YjsMaintenanceCommand,
		AggregateType: cyjsworkerevents.AggregateType_BlockPack,
		AggregateId:   request.BlockPackId,
		KafkaKey:      request.BlockPackId.String(),
		OccurredAt:    time.Now().UTC(),
		CorrelationId: request.CorrelationId,
		CausationId:   &source.EventId,
		Trace:         source.Trace,
		Data: cyjsworkerevents.YjsMaintenanceCommandData{
			RequestId:      request.RequestId,
			BlockPackId:    request.BlockPackId,
			DocumentId:     request.DocumentId,
			Operation:      request.Operation,
			TargetSequence: request.TargetSequence,
			CorrelationId:  request.CorrelationId,
		},
	}
	payload, err := json.Marshal(command)
	if err != nil {
		return err
	}

	return p.producer.Produce(
		ctx,
		cyjsworkerevents.YjsWorkerCoreMaintenanceCommandTopic.String(),
		request.BlockPackId.String(),
		payload,
	)
}
