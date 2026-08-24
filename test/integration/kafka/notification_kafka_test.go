package kafka_test

import (
	"context"
	"encoding/json"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"

	coreeventscontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
	notificationtypescontract "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/types"
	eventcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
	platformkafka "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka"
)

func TestCoreNotificationKafkaContract(t *testing.T) {
	if os.Getenv("NOTEGIC_RUN_INTEGRATION") != "1" {
		t.Skip("set NOTEGIC_RUN_INTEGRATION=1 to run Kafka broker integration tests")
	}

	brokers := configuredKafkaBrokers(t)
	producer, err := platformkafka.NewProducer(platformkafka.ClientConfig{
		ConnectionConfig: platformkafka.ConnectionConfig{
			Brokers:     brokers,
			DialTimeout: 10 * time.Second,
		},
		ClientId: "notegic-test-notification-producer",
	})
	if err != nil {
		t.Fatalf("create Kafka producer: %v", err)
	}
	t.Cleanup(producer.Close)

	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	t.Cleanup(cancel)
	if err := producer.Ping(ctx); err != nil {
		t.Skipf("Kafka broker is unavailable: %v", err)
	}

	consumer, err := platformkafka.NewConsumer(platformkafka.ConsumerConfig{
		ClientConfig: platformkafka.ClientConfig{
			ConnectionConfig: platformkafka.ConnectionConfig{
				Brokers:     brokers,
				DialTimeout: 10 * time.Second,
			},
			ClientId: "notegic-test-notification-consumer",
		},
		ConsumerGroup:       "notegic-test-notification-" + uuid.NewString(),
		MaximumAttempts:     2,
		InitialRetryBackoff: 10 * time.Millisecond,
		MaximumRetryBackoff: 25 * time.Millisecond,
		MaximumPollRecords:  20,
	}, coreeventscontract.CoreNotificationTopic.String())
	if err != nil {
		t.Fatalf("create Kafka consumer: %v", err)
	}
	t.Cleanup(consumer.Close)

	userPublicId := uuid.New()
	correlationId := uuid.NewString()
	var (
		mu                 sync.Mutex
		eventsReceivedOnce sync.Once
		notificationSeen   bool
		invalidContractErr error
		receivedEventCount int
		eventsReceived     = make(chan struct{})
	)
	consumerContext, stopConsumer := context.WithCancel(ctx)
	t.Cleanup(stopConsumer)
	go func() {
		_ = consumer.Run(consumerContext, func(
			_ context.Context,
			_ platformkafka.ConsumerRecord,
			event eventcontract.EventEnvelope[json.RawMessage],
		) error {
			if event.CorrelationId != correlationId {
				return nil
			}

			mu.Lock()
			defer mu.Unlock()
			receivedEventCount++
			switch event.EventType {
			case coreeventscontract.EventType_NotificationRequested:
				var data coreeventscontract.NotificationRequestedData
				if err := json.Unmarshal(event.Data, &data); err != nil {
					invalidContractErr = err
					return nil
				}
				if data.RecipientUserPublicId != userPublicId || data.TemplateKey != notificationtypescontract.TemplateKey_News {
					invalidContractErr = &notificationContractError{message: "notification request contract fields are invalid"}
					return nil
				}
				notificationSeen = true
			}
			if notificationSeen {
				eventsReceivedOnce.Do(func() { close(eventsReceived) })
			}
			return nil
		})
	}()

	notificationPayload, err := json.Marshal(coreeventscontract.NotificationRequestedData{
		RecipientUserPublicId: userPublicId,
		Type:                  coreeventscontract.NotificationType_News,
		Priority:              coreeventscontract.NotificationPriority_Normal,
		TemplateKey:           notificationtypescontract.TemplateKey_News,
		TemplateVersion:       1,
		Payload:               json.RawMessage(`{"title":"Release update","summary":"A new release is available.","body":"Read the release notes."}`),
		DedupeKey:             "integration:" + userPublicId.String(),
	})
	if err != nil {
		t.Fatalf("marshal notification contract: %v", err)
	}
	publishNotificationContract(t, ctx, producer, coreeventscontract.CoreNotificationTopic.String(), correlationId, userPublicId, coreeventscontract.EventType_NotificationRequested, coreeventscontract.AggregateType_Notification, notificationPayload)

	select {
	case <-eventsReceived:
	case <-ctx.Done():
		t.Fatalf("timed out waiting for notification contracts: %v", ctx.Err())
	}

	mu.Lock()
	defer mu.Unlock()
	if invalidContractErr != nil {
		t.Fatalf("received invalid notification contract: %v", invalidContractErr)
	}
	if receivedEventCount != 1 {
		t.Fatalf("received event count = %d, want 1", receivedEventCount)
	}
}

type notificationContractError struct {
	message string
}

func (e *notificationContractError) Error() string {
	return e.message
}

func publishNotificationContract(
	t *testing.T,
	ctx context.Context,
	producer *platformkafka.Producer,
	topic string,
	correlationId string,
	aggregateId uuid.UUID,
	eventType eventcontract.EventType,
	aggregateType eventcontract.AggregateType,
	data []byte,
) {
	t.Helper()

	payload, err := json.Marshal(eventcontract.EventEnvelope[json.RawMessage]{
		SchemaVersion: eventcontract.Version,
		EventId:       uuid.New(),
		EventType:     eventType,
		AggregateType: aggregateType,
		AggregateId:   aggregateId,
		KafkaKey:      aggregateId.String(),
		OccurredAt:    time.Now().UTC(),
		CorrelationId: correlationId,
		Data:          data,
	})
	if err != nil {
		t.Fatalf("marshal notification event envelope: %v", err)
	}
	if err := producer.Produce(ctx, topic, aggregateId.String(), payload); err != nil {
		t.Fatalf("publish notification event: %v", err)
	}
}
