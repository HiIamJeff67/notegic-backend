package consumers

import (
	"context"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	logs "github.com/HiIamJeff67/notegic-backend/shared/platform/observability/logs"

	realtimelease "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/data/redis/realtimelease"
)

type LifecycleConsumer struct {
	leaseStore  *realtimelease.RealtimeLeaseCacheClient
	kafkaConfig platformkafka.ConsumerConfig
}

func NewLifecycleConsumer(
	leaseStore *realtimelease.RealtimeLeaseCacheClient,
	kafkaConfig platformkafka.ConsumerConfig,
) *LifecycleConsumer {
	return &LifecycleConsumer{
		leaseStore:  leaseStore,
		kafkaConfig: kafkaConfig,
	}
}

func (c *LifecycleConsumer) Start(ctx context.Context) func() {
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

func (c *LifecycleConsumer) run(ctx context.Context) {
	for ctx.Err() == nil {
		consumer, err := platformkafka.NewConsumer(
			c.kafkaConfig,
			coreevents.CoreLifecycleTopic.String(),
		)
		if err == nil {
			err = consumer.Run(ctx, c.handle)
			consumer.Close()
		}
		if ctx.Err() != nil {
			return
		}
		if logs.NotegicLogger != nil {
			logs.NotegicLogger.Error(ctx, err, "RealtimeGateway lifecycle Kafka consumer stopped")
		}

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Second):
		}
	}
}

func (c *LifecycleConsumer) handle(
	ctx context.Context,
	record platformkafka.ConsumerRecord,
	envelope cevent.EventEnvelope[json.RawMessage],
) error {
	claimed, err := c.leaseStore.MarkLifecycleEventProcessed(envelope.EventId)
	if err != nil {
		return err
	}
	if !claimed {
		return nil
	}
	if err := c.process(ctx, record, envelope); err != nil {
		_ = c.leaseStore.ReleaseLifecycleEvent(envelope.EventId)
		return err
	}

	return nil
}

func (c *LifecycleConsumer) process(
	_ context.Context,
	_ platformkafka.ConsumerRecord,
	envelope cevent.EventEnvelope[json.RawMessage],
) error {
	switch envelope.EventType {
	case coreevents.EventType_RoutineTaskCompleted:
		var data coreevents.RoutineTaskCompletedData
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			return &platformkafka.ConsumerError{
				Classification: platformkafka.ErrorClassification_SchemaIncompatible,
				Origin:         err,
			}
		}
		if data.RoutineTaskId == uuid.Nil || data.RoutineTaskRecordId == uuid.Nil ||
			data.RoutineId == uuid.Nil || data.ActorUserPublicId == uuid.Nil ||
			data.Purpose == "" || data.Attempt <= 0 || data.CompletedAt.IsZero() {
			return &platformkafka.ConsumerError{
				Classification: platformkafka.ErrorClassification_SchemaIncompatible,
				Origin:         errors.New("Kafka RoutineTask completed lifecycle event is incomplete"),
			}
		}

		return c.leaseStore.PublishRoutineTaskLifecycleEvent(realtimelease.RoutineTaskLifecycleEvent{
			EventId:             envelope.EventId,
			RoutineTaskId:       data.RoutineTaskId,
			RoutineTaskRecordId: data.RoutineTaskRecordId,
			RoutineId:           data.RoutineId,
			ActorUserPublicId:   data.ActorUserPublicId,
			Purpose:             string(data.Purpose),
			Status:              "completed",
			Attempt:             data.Attempt,
			OccurredAt:          data.CompletedAt,
		})
	case coreevents.EventType_BlockPackAccessRevoked:
		var data coreevents.BlockPackAccessRevokedData
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			return &platformkafka.ConsumerError{
				Classification: platformkafka.ErrorClassification_SchemaIncompatible,
				Origin:         err,
			}
		}
		if data.Reason != coreevents.BlockPackAccessRevocationReason_PermissionRevoked &&
			data.Reason != coreevents.BlockPackAccessRevocationReason_ResourceUnavailable {
			return &platformkafka.ConsumerError{
				Classification: platformkafka.ErrorClassification_SchemaIncompatible,
				Origin:         errors.New("Kafka BlockPack access revocation has an unsupported reason"),
			}
		}

		return c.leaseStore.PublishBlockPackChannelRevocation(realtimelease.BlockPackChannelRevocation{
			EventId:            envelope.EventId,
			BlockPackId:        envelope.AggregateId,
			TargetUserPublicId: data.TargetUserPublicId,
			Reason:             data.Reason,
		})
	case coreevents.EventType_UserSessionsRevoked:
		return c.leaseStore.PublishUserSessionRevocation(realtimelease.UserSessionRevocation{
			EventId:      envelope.EventId,
			UserPublicId: envelope.AggregateId,
		})
	case coreevents.EventType_BlockPackRoomPolicyChanged,
		coreevents.EventType_RootShelfPermissionRevoked:
		if envelope.EventType == coreevents.EventType_RootShelfPermissionRevoked {
			var data coreevents.ResourceChangedData
			if err := json.Unmarshal(envelope.Data, &data); err != nil {
				return &platformkafka.ConsumerError{
					Classification: platformkafka.ErrorClassification_SchemaIncompatible,
					Origin:         err,
				}
			}

			resourceId := data.ResourceId
			if resourceId == uuid.Nil {
				resourceId = envelope.AggregateId
			}
			change := data.Change
			if change == "" {
				change = coreevents.ResourceEventChange_PermissionRevoked
			}

			return c.leaseStore.PublishResourceEvent(realtimelease.ResourceEvent{
				EventId:            envelope.EventId,
				EventType:          string(envelope.EventType),
				ResourceId:         resourceId,
				TargetUserPublicId: data.TargetUserPublicId,
				Change:             string(change),
				Permission:         data.Permission,
			})
		}

		return nil
	case coreevents.EventType_RootShelfPermissionChanged,
		coreevents.EventType_RootShelfDeleted,
		coreevents.EventType_BlockPackChanged,
		coreevents.EventType_BlockPackDeleted:
		var data coreevents.ResourceChangedData
		if err := json.Unmarshal(envelope.Data, &data); err != nil {
			return &platformkafka.ConsumerError{
				Classification: platformkafka.ErrorClassification_SchemaIncompatible,
				Origin:         err,
			}
		}
		resourceId := data.ResourceId
		if resourceId == uuid.Nil {
			resourceId = envelope.AggregateId
		}
		change := data.Change
		if change == "" {
			change = coreevents.ResourceEventChange_Updated
			if envelope.EventType == coreevents.EventType_RootShelfDeleted ||
				envelope.EventType == coreevents.EventType_BlockPackDeleted {
				change = coreevents.ResourceEventChange_Deleted
			}
		}

		return c.leaseStore.PublishResourceEvent(realtimelease.ResourceEvent{
			EventId:            envelope.EventId,
			EventType:          string(envelope.EventType),
			ResourceId:         resourceId,
			TargetUserPublicId: data.TargetUserPublicId,
			Change:             string(change),
			Permission:         data.Permission,
		})
	default:
		return &platformkafka.ConsumerError{
			Classification: platformkafka.ErrorClassification_PoisonMessage,
			Origin:         errors.New("Kafka lifecycle event type is unsupported by RealtimeGateway"),
		}
	}
}
