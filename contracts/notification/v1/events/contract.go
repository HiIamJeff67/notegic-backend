package notificationeventscontract

import cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"

const NotificationTopic cevent.Topic = "notegic.notification.v1"

const (
	AggregateType_Notification    cevent.AggregateType = "Notification"
	EventType_NotificationCreated cevent.EventType     = "NotificationCreated"
	EventType_NotificationUpdated cevent.EventType     = "NotificationUpdated"
)
