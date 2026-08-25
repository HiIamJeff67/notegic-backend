package consumers

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis"
	"github.com/google/uuid"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	skafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
	sredis "github.com/HiIamJeff67/notegic-backend/shared/platform/redis"

	realtimelease "github.com/HiIamJeff67/notegic-backend/internal/realtimegateway/data/redis/realtimelease"
)

func TestLifecycleConsumerPublishesCompletedRoutineTaskToRealtimeGateway(t *testing.T) {
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
			sredis.NewClientSetFromClients(redisClient),
		),
	)
	consumer := NewLifecycleConsumer(leaseStore, skafka.ConsumerConfig{})
	received := make(chan realtimelease.RoutineTaskLifecycleEvent, 1)
	shutdown, err := leaseStore.SubscribeRoutineTaskLifecycleEvents(func(event realtimelease.RoutineTaskLifecycleEvent) {
		received <- event
	})
	if err != nil {
		t.Fatalf("subscribe to lifecycle events: %v", err)
	}
	defer shutdown()

	data := coreevents.RoutineTaskCompletedData{
		RoutineTaskId:       uuid.New(),
		RoutineTaskRecordId: uuid.New(),
		RoutineId:           uuid.New(),
		ActorUserPublicId:   uuid.New(),
		Purpose:             cenums.RoutineTaskPurpose_CreateBlockPack,
		WorkerId:            uuid.New(),
		Attempt:             1,
		CompletedAt:         time.Now().UTC(),
	}
	payload, err := json.Marshal(data)
	if err != nil {
		t.Fatalf("marshal completed lifecycle event: %v", err)
	}

	if err := consumer.process(
		context.Background(),
		skafka.ConsumerRecord{},
		cevent.EventEnvelope[json.RawMessage]{
			EventId:     uuid.New(),
			EventType:   coreevents.EventType_RoutineTaskCompleted,
			AggregateId: data.RoutineTaskId,
			Data:        payload,
		},
	); err != nil {
		t.Fatalf("process completed lifecycle event: %v", err)
	}

	select {
	case event := <-received:
		if event.Status != "completed" || event.RoutineTaskId != data.RoutineTaskId ||
			event.ActorUserPublicId != data.ActorUserPublicId {
			t.Fatalf("unexpected realtime lifecycle event: %#v", event)
		}
	case <-time.After(time.Second):
		t.Fatal("expected completed RoutineTask lifecycle event")
	}
}
