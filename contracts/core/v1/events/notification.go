package eventscontract

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

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

type NotificationRequestedData struct {
	RecipientUserPublicId uuid.UUID            `json:"recipientUserPublicId"`
	Type                  NotificationType     `json:"type"`
	Priority              NotificationPriority `json:"priority"`
	TemplateKey           string               `json:"templateKey"`
	TemplateVersion       int                  `json:"templateVersion"`
	Payload               json.RawMessage      `json:"payload"`
	DedupeKey             string               `json:"dedupeKey"`
	ExpiresAt             *time.Time           `json:"expiresAt,omitempty"`
}
