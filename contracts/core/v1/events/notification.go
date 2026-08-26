package eventscontract

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	"github.com/HiIamJeff67/notegic-backend/contracts/types/enums"
	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
)

const CoreNotificationTopic cevent.Topic = "notegic.core.notification.v1"

const (
	AggregateType_Notification      cevent.AggregateType = "Notification"
	EventType_NotificationRequested cevent.EventType     = "NotificationRequested"
)

type NotificationPriority string

const (
	NotificationPriority_Low      NotificationPriority = "low"
	NotificationPriority_Normal   NotificationPriority = "normal"
	NotificationPriority_High     NotificationPriority = "high"
	NotificationPriority_Critical NotificationPriority = "critical"
)

type NotificationType string

const (
	NotificationType_News      NotificationType = "news"
	NotificationType_Warning   NotificationType = "warning"
	NotificationType_Important NotificationType = "important"
)

// UserProjection carries the minimal user snapshot required by Notification
// when Core requests a notification in a separate database.
type UserProjection struct {
	PublicId uuid.UUID        `json:"publicId"`
	Plan     enums.UserPlan   `json:"plan"`
	Status   enums.UserStatus `json:"status"`
}

type NotificationRequestedData struct {
	RecipientUserPublicId uuid.UUID            `json:"recipientUserPublicId"`
	UserProjection        UserProjection       `json:"userProjection"`
	Type                  NotificationType     `json:"type"`
	Priority              NotificationPriority `json:"priority"`
	TemplateKey           string               `json:"templateKey"`
	TemplateVersion       int                  `json:"templateVersion"`
	Payload               json.RawMessage      `json:"payload"`
	DedupeKey             string               `json:"dedupeKey"`
	ExpiresAt             *time.Time           `json:"expiresAt,omitempty"`
}
