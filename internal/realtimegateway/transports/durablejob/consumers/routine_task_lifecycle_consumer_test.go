package consumers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
	"github.com/google/uuid"

	cdurablejobevents "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/events"
	enums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	platformredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	realtimelease "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/data/redis/realtimelease"
)

func TestRoutineTaskLifecycleConsumerPublishesRunningTaskToRealtimeGateway(t *testing.T) {
	server, err := miniredis.Run()
	if err != nil {
		t.Fatalf("start Redis: %v", err)
	}
	defer server.Close()

	redisClient := redis.NewClient(&redis.Options{
		Addr: server.Addr(),
	})
	defer redisClient.Close()

	leaseStore := realtimelease.NewRealtimeLeaseCacheClient(
		realtimelease.NewRealtimeLeaseCacheStore(
			platformredis.NewClientSetFromClients(redisClient),
		),
	)
	consumer := NewRoutineTaskLifecycleConsumer(leaseStore, platformkafka.ConsumerConfig{})
	received := make(chan realtimelease.RoutineTaskLifecycleEvent, 1)
	shutdown, err := leaseStore.SubscribeRoutineTaskLifecycleEvents(func(event realtimelease.RoutineTaskLifecycleEvent) {
		received <- event
	})
	if err != nil {
		t.Fatalf("subscribe to lifecycle events: %v", err)
	}
	defer shutdown()

	data := cdurablejobevents.RoutineTaskRunningData{
		RoutineTaskId:       uuid.New(),
		RoutineTaskRecordId: uuid.New(),
		RoutineId:           uuid.New(),
		ActorUserPublicId:   uuid.New(),
		Purpose:             enums.RoutineTaskPurpose_CreateBlockPack,
		Attempt:             1,
		StartedAt:           time.Now().UTC(),
	}
	payload, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal running lifecycle event: %v", err)
	}

	if err := consumer.consume(
		context.Background(),
		platformkafka.ConsumerRecord{},
		cevent.EventEnvelope[json.RawMessage]{
			EventId:       uuid.New(),
			EventType:     cdurablejobevents.EventType_RoutineTaskRunning,
			AggregateType: cdurablejobevents.AggregateType_RoutineTask,
			AggregateId:   data.RoutineTaskId,
			Data:          payload,
		},
	); err != nil {
		t.Fatalf("consume running lifecycle event: %v", err)
	}

	select {
	case event := <-received:
		if event.Status != "running" || event.RoutineTaskId != data.RoutineTaskId ||
			event.ActorUserPublicId != data.ActorUserPublicId {
			t.Fatalf("unexpected realtime lifecycle event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("expected running RoutineTask lifecycle event")
	}
}
