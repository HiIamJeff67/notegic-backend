package schemas

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"

	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

	postgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

// OutboxEvent is the storage contract shared by runtimes that publish events.
// Runtime-specific event builders belong to the runtime that owns the event.
type OutboxEvent struct {
	Id            uuid.UUID            `json:"id" gorm:"column:id; type:uuid; primaryKey; default:gen_random_uuid();"`
	AggregateType cevent.AggregateType `json:"aggregateType" gorm:"column:aggregate_type; type:varchar(64); not null; index:outbox_event_unpublished_index,priority:2;"`
	AggregateId   uuid.UUID            `json:"aggregateId" gorm:"column:aggregate_id; type:uuid; not null; index:outbox_event_unpublished_index,priority:3;"`
	EventType     cevent.EventType     `json:"eventType" gorm:"column:event_type; type:varchar(128); not null;"`
	Topic         cevent.Topic         `json:"topic" gorm:"column:topic; type:varchar(255); not null;"`
	KafkaKey      string               `json:"kafkaKey" gorm:"column:kafka_key; type:varchar(255); not null;"`
	Payload       datatypes.JSON       `json:"payload" gorm:"column:payload; type:jsonb; not null;"`
	Metadata      datatypes.JSON       `json:"metadata" gorm:"column:metadata; type:jsonb; not null;"`
	AvailableAt   time.Time            `json:"availableAt" gorm:"column:available_at; type:timestamptz; not null; index:outbox_event_unpublished_index,priority:1;"`
	PublishedAt   *time.Time           `json:"publishedAt" gorm:"column:published_at; type:timestamptz; index:outbox_event_unpublished_index,priority:4;"`
	PublishCount  int32                `json:"publishCount" gorm:"column:publish_count; type:integer; not null; default:0;"`
	LastError     *string              `json:"lastError" gorm:"column:last_error; type:text;"`
	ClaimedBy     *string              `json:"claimedBy" gorm:"column:claimed_by; type:varchar(255); index;"`
	ClaimedAt     *time.Time           `json:"claimedAt" gorm:"column:claimed_at; type:timestamptz; index;"`
	CreatedAt     time.Time            `json:"createdAt" gorm:"column:created_at; type:timestamptz; not null; autoCreateTime:true;"`
}

func (OutboxEvent) TableName() string {
	return postgres.TableName_OutboxEventTable.String()
}
