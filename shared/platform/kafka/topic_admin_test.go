package kafka

import (
	"fmt"
	"testing"
	"time"

	"github.com/twmb/franz-go/pkg/kmsg"

	"github.com/HiIamJeff67/notegic-backend/shared/lib/pointers"
	kafkatopics "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka/topics"
)

func TestValidateTopicSpecRequiresExplicitCreationSettings(t *testing.T) {
	if err := validateTopicSpec(kafkatopics.TopicSpec{Name: "notegic.test.v1"}); err == nil {
		t.Fatal("validate topic spec returned nil for incomplete settings")
	}

	specification := kafkatopics.TopicSpec{
		Name:                "notegic.test.v1",
		Partitions:          3,
		ReplicationFactor:   1,
		Retention:           7 * 24 * time.Hour,
		CleanupPolicy:       "delete",
		MinInSyncReplicas:   1,
		CreateDeadLetter:    true,
		DeadLetterRetention: 30 * 24 * time.Hour,
	}
	if err := validateTopicSpec(specification); err != nil {
		t.Fatalf("validate complete topic spec: %v", err)
	}
}

func TestTopicRequestUsesKafkaMillisecondRetention(t *testing.T) {
	request := kmsg.CreateTopicsRequestTopic{
		Topic:             "notegic.test.v1",
		NumPartitions:     3,
		ReplicationFactor: 1,
		Configs: []kmsg.CreateTopicsRequestTopicConfig{
			{Name: "cleanup.policy", Value: pointers.ToPtr("delete")},
			{Name: "retention.ms", Value: pointers.ToPtr(fmt.Sprintf("%d", 5*time.Minute/time.Millisecond))},
			{Name: "min.insync.replicas", Value: pointers.ToPtr(fmt.Sprintf("%d", 1))},
		},
	}

	if got := *request.Configs[1].Value; got != "300000" {
		t.Fatalf("retention.ms = %q, want 300000", got)
	}
}

func TestValidateTopicSpecRejectsInvalidValues(t *testing.T) {
	tests := []kafkatopics.TopicSpec{
		{Name: ""},
		{Name: " notegic.test.v1 ", Partitions: 3, ReplicationFactor: 1, Retention: time.Hour, CleanupPolicy: "delete", MinInSyncReplicas: 1},
		{Name: "notegic.test.v1", Partitions: -1},
		{Name: "notegic.test.v1", ReplicationFactor: -1},
		{Name: "notegic.test.v1", Retention: -time.Second},
		{Name: "notegic.test.v1", MinInSyncReplicas: -1},
		{Name: "notegic.test.v1", Partitions: 1, ReplicationFactor: 1, Retention: time.Hour, CleanupPolicy: "delete", MinInSyncReplicas: 1, CreateDeadLetter: true},
	}

	for _, specification := range tests {
		if err := validateTopicSpec(specification); err == nil {
			t.Fatalf("validate topic spec %+v returned nil error", specification)
		}
	}
}
