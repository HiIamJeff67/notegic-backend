package kafka

import (
	"testing"
)

func TestNewConsumerRejectsMissingGroupOrTopics(t *testing.T) {
	_, err := NewConsumer(ConsumerConfig{}, "notegic.core.lifecycle.v1")
	if err == nil {
		t.Fatal("expected Kafka consumer group to be required")
	}

	_, err = NewConsumer(ConsumerConfig{
		ConsumerGroup: "test-group",
	})
	if err == nil {
		t.Fatal("expected Kafka consumer topic to be required")
	}
}

func TestGetDeadLetterTopic(t *testing.T) {
	if GetDeadLetterTopic("notegic.core.lifecycle.v1") != "notegic.core.lifecycle.v1.dlq" {
		t.Fatal("unexpected Kafka dead-letter topic name")
	}
}
