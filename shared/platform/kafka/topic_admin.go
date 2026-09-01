package kafka

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/twmb/franz-go/pkg/kerr"
	franzkgo "github.com/twmb/franz-go/pkg/kgo"
	"github.com/twmb/franz-go/pkg/kmsg"

	"github.com/HiIamJeff67/notegic-backend/shared/lib/pointers"
	kafkatopics "github.com/HiIamJeff67/notegic-backend/shared/platform/kafka/topics"
)

type TopicProvisioner struct {
	client *franzkgo.Client
}

func NewTopicProvisioner(kafkaConfig ClientConfig) (*TopicProvisioner, error) {
	options, err := newConnectionOptions(kafkaConfig)
	if err != nil {
		return nil, err
	}

	client, err := franzkgo.NewClient(options...)
	if err != nil {
		return nil, fmt.Errorf("create Kafka topic admin client: %w", err)
	}

	return &TopicProvisioner{client: client}, nil
}

func (p *TopicProvisioner) EnsureTopics(ctx context.Context, specifications []kafkatopics.TopicSpec) error {
	if p == nil || p.client == nil {
		return errors.New("Kafka topic provisioner is unavailable")
	}
	if len(specifications) == 0 {
		return nil
	}

	request := kmsg.CreateTopicsRequest{
		Version:       5,
		TimeoutMillis: 30_000,
		Topics:        make([]kmsg.CreateTopicsRequestTopic, 0, len(specifications)),
	}
	seenTopics := make(map[string]struct{}, len(specifications))
	for _, specification := range specifications {
		if err := validateTopicSpec(specification); err != nil {
			return fmt.Errorf("validate Kafka topic %q: %w", specification.Name, err)
		}
		if _, exists := seenTopics[specification.Name]; exists {
			continue
		}
		seenTopics[specification.Name] = struct{}{}
		request.Topics = append(request.Topics, kmsg.CreateTopicsRequestTopic{
			Topic:             specification.Name,
			NumPartitions:     specification.Partitions,
			ReplicationFactor: specification.ReplicationFactor,
			Configs: []kmsg.CreateTopicsRequestTopicConfig{
				{Name: "cleanup.policy", Value: pointers.ToPtr(specification.CleanupPolicy)},
				{Name: "retention.ms", Value: pointers.ToPtr(fmt.Sprintf("%d", specification.Retention/time.Millisecond))},
				{Name: "min.insync.replicas", Value: pointers.ToPtr(fmt.Sprintf("%d", specification.MinInSyncReplicas))},
			},
		})
		if specification.CreateDeadLetter {
			deadLetter := specification
			deadLetter.Name = GetDeadLetterTopic(specification.Name)
			deadLetter.Retention = specification.DeadLetterRetention
			deadLetter.CreateDeadLetter = false
			deadLetter.DeadLetterRetention = 0
			if _, exists := seenTopics[deadLetter.Name]; !exists {
				seenTopics[deadLetter.Name] = struct{}{}
				request.Topics = append(request.Topics, kmsg.CreateTopicsRequestTopic{
					Topic:             deadLetter.Name,
					NumPartitions:     deadLetter.Partitions,
					ReplicationFactor: deadLetter.ReplicationFactor,
					Configs: []kmsg.CreateTopicsRequestTopicConfig{
						{Name: "cleanup.policy", Value: pointers.ToPtr(deadLetter.CleanupPolicy)},
						{Name: "retention.ms", Value: pointers.ToPtr(fmt.Sprintf("%d", deadLetter.Retention/time.Millisecond))},
						{Name: "min.insync.replicas", Value: pointers.ToPtr(fmt.Sprintf("%d", deadLetter.MinInSyncReplicas))},
					},
				})
			}
		}
	}

	response, err := request.RequestWith(ctx, p.client)
	if err != nil {
		return fmt.Errorf("create Kafka topics: %w", err)
	}
	for _, topic := range response.Topics {
		if topic.ErrorCode == 0 || topic.ErrorCode == kerr.TopicAlreadyExists.Code {
			continue
		}
		brokerError := kerr.ErrorForCode(topic.ErrorCode)
		message := ""
		if topic.ErrorMessage != nil {
			message = ": " + *topic.ErrorMessage
		}
		if brokerError == nil {
			return fmt.Errorf("create Kafka topic %q: broker error code %d%s", topic.Topic, topic.ErrorCode, message)
		}
		return fmt.Errorf("create Kafka topic %q: %w%s", topic.Topic, brokerError, message)
	}

	return nil
}

func (p *TopicProvisioner) Close() {
	if p == nil || p.client == nil {
		return
	}

	p.client.Close()
}

func validateTopicSpec(specification kafkatopics.TopicSpec) error {
	if strings.TrimSpace(specification.Name) == "" {
		return errors.New("Kafka topic name is required")
	}
	if specification.Name != strings.TrimSpace(specification.Name) {
		return errors.New("Kafka topic name must not contain leading or trailing whitespace")
	}
	if specification.Partitions < 1 {
		return fmt.Errorf("Kafka topic %q must have at least one partition", specification.Name)
	}
	if specification.ReplicationFactor < 1 {
		return fmt.Errorf("Kafka topic %q must have a positive replication factor", specification.Name)
	}
	if specification.Retention <= 0 {
		return fmt.Errorf("Kafka topic %q retention must be positive", specification.Name)
	}
	if strings.TrimSpace(specification.CleanupPolicy) == "" {
		return fmt.Errorf("Kafka topic %q cleanup policy is required", specification.Name)
	}
	if specification.CleanupPolicy != strings.TrimSpace(specification.CleanupPolicy) {
		return fmt.Errorf("Kafka topic %q cleanup policy must not contain leading or trailing whitespace", specification.Name)
	}
	if specification.MinInSyncReplicas < 1 {
		return fmt.Errorf("Kafka topic %q must have a positive min.insync.replicas", specification.Name)
	}
	if specification.CreateDeadLetter && specification.DeadLetterRetention <= 0 {
		return fmt.Errorf("Kafka topic %q dead-letter retention must be positive when dead-letter creation is enabled", specification.Name)
	}

	return nil
}
