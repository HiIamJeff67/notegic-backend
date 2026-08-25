package topics

import (
	"time"

	cemailevents "github.com/HiIamJeff67/notegic-backend/contracts/email/v1/events"
)

func CoreEmailRequestTopicSpec() TopicSpec {
	return TopicSpec{
		Name:                cemailevents.CoreEmailRequestTopic.String(),
		Partitions:          3,
		ReplicationFactor:   1,
		Retention:           7 * 24 * time.Hour,
		CleanupPolicy:       "delete",
		MinInSyncReplicas:   1,
		CreateDeadLetter:    true,
		DeadLetterRetention: 30 * 24 * time.Hour,
	}
}
