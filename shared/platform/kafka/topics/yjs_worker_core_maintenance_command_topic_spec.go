package topics

import (
	"time"

	cyjsworkerevents "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1/events"
)

func YjsWorkerCoreMaintenanceCommandTopicSpec() TopicSpec {
	return TopicSpec{
		Name:                cyjsworkerevents.YjsWorkerCoreMaintenanceCommandTopic.String(),
		Partitions:          3,
		ReplicationFactor:   1,
		Retention:           7 * 24 * time.Hour,
		CleanupPolicy:       "delete",
		MinInSyncReplicas:   1,
		CreateDeadLetter:    true,
		DeadLetterRetention: 30 * 24 * time.Hour,
	}
}
