package schemas

import (
	"time"

	"github.com/google/uuid"

	platformpostgres "github.com/HiIamJeff67/notegic-backend/shared/platform/postgres"
)

// InboxEvent is the shared idempotency record used by runtimes consuming events.
type InboxEvent struct {
	EventId    uuid.UUID `json:"eventId" gorm:"column:event_id;type:uuid;primaryKey;"`
	ConsumedAt time.Time `json:"consumedAt" gorm:"column:consumed_at;type:timestamptz;not null;autoCreateTime:true;"`
}

func (InboxEvent) TableName() string {
	return platformpostgres.TableName_InboxEventTable.String()
}
