package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	durablejobeventscontract "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/events"
	eventcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	realtimelease "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/data/redis/realtimelease"
)

type RoutineTaskLifecycleConsumer struct {
	leaseStore  *realtimelease.RealtimeLeaseCacheClient
	kafkaConfig platformkafka.ConsumerConfig
}

func NewRoutineTaskLifecycleConsumer(
	leaseStore *realtimelease.RealtimeLeaseCacheClient,
	kafkaConfig platformkafka.ConsumerConfig,
) *RoutineTaskLifecycleConsumer {
	return &RoutineTaskLifecycleConsumer{
		leaseStore:  leaseStore,
		kafkaConfig: kafkaConfig,
	}
}

func (c *RoutineTaskLifecycleConsumer) Start(ctx context.Context) func() {
	workerCtx, cancel := context.WithCancel(ctx)
	var waitGroup sync.WaitGroup
	waitGroup.Add(1)
	go func() {
		defer waitGroup.Done()
		c.run(workerCtx)
	}()

	return func() {
		cancel()
		waitGroup.Wait()
	}
}

func (c *RoutineTaskLifecycleConsumer) run(ctx context.Context) {
	for ctx.Err() == nil {
		consumer, err := platformkafka.NewConsumer(
			c.kafkaConfig,
			durablejobeventscontract.DurableJobRealtimeGatewayRoutineTaskLifecycleTopic.String(),
		)
		if err == nil {
			err = consumer.Run(ctx, c.consume)
			consumer.Close()
		}
		if ctx.Err() != nil {
			return
		}
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(
				ctx,
				err,
				"RealtimeGateway DurableJob RoutineTask lifecycle consumer stopped",
			)
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func (c *RoutineTaskLifecycleConsumer) consume(
	_ context.Context,
	_ platformkafka.ConsumerRecord,
	envelope eventcontract.EventEnvelope[json.RawMessage],
) error {
	if envelope.EventType != durablejobeventscontract.EventType_RoutineTaskRunning ||
		envelope.AggregateType != durablejobeventscontract.AggregateType_RoutineTask {
		return &platformkafka.ConsumerError{
			Classification: platformkafka.ErrorClassification_PoisonMessage,
			Origin:         errors.New("Kafka DurableJob RoutineTask lifecycle event is unsupported"),
		}
	}

	var data durablejobeventscontract.RoutineTaskRunningData
	if err := json.Unmarshal(envelope.Data, &data); err != nil {
		return &platformkafka.ConsumerError{
			Classification: platformkafka.ErrorClassification_SchemaIncompatible,
			Origin:         err,
		}
	}
	if data.RoutineTaskId == uuid.Nil || data.RoutineTaskRecordId == uuid.Nil ||
		data.RoutineId == uuid.Nil || data.ActorUserPublicId == uuid.Nil ||
		data.Purpose == "" || data.Attempt <= 0 || data.StartedAt.IsZero() {
		return &platformkafka.ConsumerError{
			Classification: platformkafka.ErrorClassification_SchemaIncompatible,
			Origin:         errors.New("Kafka DurableJob RoutineTask running lifecycle event is incomplete"),
		}
	}

	claimed, err := c.leaseStore.MarkLifecycleEventProcessed(envelope.EventId)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}

	if err := c.leaseStore.PublishRoutineTaskLifecycleEvent(realtimelease.RoutineTaskLifecycleEvent{
		EventId:             envelope.EventId,
		RoutineTaskId:       data.RoutineTaskId,
		RoutineTaskRecordId: data.RoutineTaskRecordId,
		RoutineId:           data.RoutineId,
		ActorUserPublicId:   data.ActorUserPublicId,
		Purpose:             string(data.Purpose),
		Status:              "running",
		Attempt:             data.Attempt,
		OccurredAt:          data.StartedAt,
	}); err != nil {
		_ = c.leaseStore.ReleaseLifecycleEvent(envelope.EventId)
		return err
	}

	return nil
}
