package topics

import (
	"time"

	coreevents "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/events"
)

func CoreLifecycleTopicSpec() TopicSpec {
	return TopicSpec{
		Name:                coreevents.CoreLifecycleTopic.String(),
		Partitions:          3,
		ReplicationFactor:   1,
		Retention:           7 * 24 * time.Hour,
		CleanupPolicy:       "delete",
		MinInSyncReplicas:   1,
		CreateDeadLetter:    true,
		DeadLetterRetention: 30 * 24 * time.Hour,
	}
}
