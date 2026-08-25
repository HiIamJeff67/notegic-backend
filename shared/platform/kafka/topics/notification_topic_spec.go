package topics

import (
	"time"

	cnotificationevents "github.com/HiIamJeff67/notegic-backend/contracts/notification/v1/events"
)

func NotificationTopicSpec() TopicSpec {
	return TopicSpec{
		Name:                cnotificationevents.NotificationTopic.String(),
		Partitions:          3,
		ReplicationFactor:   1,
		Retention:           7 * 24 * time.Hour,
		CleanupPolicy:       "delete",
		MinInSyncReplicas:   1,
		CreateDeadLetter:    true,
		DeadLetterRetention: 30 * 24 * time.Hour,
	}
}
