package topics

import (
	"time"

	cdurablejobevents "github.com/HiIamJeff67/notegic-backend/contracts/durable-job/v1/events"
)

func DurableJobRealtimeGatewayRoutineTaskLifecycleTopicSpec() TopicSpec {
	return TopicSpec{
		Name:                cdurablejobevents.DurableJobRealtimeGatewayRoutineTaskLifecycleTopic.String(),
		Partitions:          3,
		ReplicationFactor:   1,
		Retention:           7 * 24 * time.Hour,
		CleanupPolicy:       "delete",
		MinInSyncReplicas:   1,
		CreateDeadLetter:    true,
		DeadLetterRetention: 30 * 24 * time.Hour,
	}
}
