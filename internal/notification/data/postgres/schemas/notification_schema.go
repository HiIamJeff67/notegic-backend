package schemas

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Notification struct {
	Id                    uuid.UUID      `json:"id" gorm:"column:id;type:uuid;primaryKey;default:gen_random_uuid();index:notification_recipient_created_index,priority:3,sort:desc;"`
	RecipientUserPublicId uuid.UUID      `json:"recipientUserPublicId" gorm:"column:recipient_user_public_id;type:uuid;not null;index:notification_recipient_created_index,priority:1;"`
	Type                  string         `json:"type" gorm:"column:type;type:varchar(64);not null;"`
	Priority              string         `json:"priority" gorm:"column:priority;type:varchar(32);not null;"`
	TemplateKey           string         `json:"templateKey" gorm:"column:template_key;type:varchar(128);not null;"`
	TemplateVersion       int            `json:"templateVersion" gorm:"column:template_version;type:integer;not null;"`
	Payload               datatypes.JSON `json:"payload" gorm:"column:payload;type:jsonb;not null;"`
	DedupeKey             string         `json:"dedupeKey" gorm:"column:dedupe_key;type:varchar(255);not null;uniqueIndex:notification_dedupe_key_index,where:dedupe_key <> ''"`
	CreatedAt             time.Time      `json:"createdAt" gorm:"column:created_at;type:timestamptz;not null;index:notification_recipient_created_index,priority:2,sort:desc;"`
	ReadAt                *time.Time     `json:"readAt" gorm:"column:read_at;type:timestamptz;"`
	DeletedAt             *time.Time     `json:"deletedAt" gorm:"column:deleted_at;type:timestamptz;index;"`
	ExpiresAt             *time.Time     `json:"expiresAt" gorm:"column:expires_at;type:timestamptz;index;"`
}

func (Notification) TableName() string {
	return "NotificationTable"
}
