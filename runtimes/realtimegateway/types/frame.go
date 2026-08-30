package realtimetypes

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type ControlFrame struct {
	Version   int       `json:"version"`
	Type      FrameType `json:"type"`
	RequestId string    `json:"requestId,omitempty"`
}

type ReadyFrame struct {
	Version             int       `json:"version"`
	Type                FrameType `json:"type"`
	ConnectionId        string    `json:"connectionId"`
	ResubscribeRequired bool      `json:"resubscribeRequired"`
}

type ErrorFrame struct {
	Version            int         `json:"version"`
	Type               FrameType   `json:"type"`
	RequestId          string      `json:"requestId,omitempty"`
	ChannelType        ChannelType `json:"channelType,omitempty"`
	ChannelId          *uuid.UUID  `json:"channelId,omitempty"`
	ConnectorChannelId uint32      `json:"connectorChannelId,omitempty"`
	Code               ErrorCode   `json:"code"`
	Message            string      `json:"message"`
}

type SubscribeFrame struct {
	Version       int         `json:"version"`
	Type          FrameType   `json:"type"`
	RequestId     string      `json:"requestId,omitempty"`
	ChannelType   ChannelType `json:"channelType"`
	ChannelId     uuid.UUID   `json:"channelId"`
	ChannelTicket string      `json:"channelTicket,omitempty"`
}

type SubscribedFrame struct {
	Version                    int                            `json:"version"`
	Type                       FrameType                      `json:"type"`
	RequestId                  string                         `json:"requestId,omitempty"`
	ChannelType                ChannelType                    `json:"channelType"`
	ChannelId                  uuid.UUID                      `json:"channelId"`
	ConnectorChannelId         uint32                         `json:"connectorChannelId"`
	Existing                   bool                           `json:"existing"`
	DocumentQuotaPolicyVersion int                            `json:"documentQuotaPolicyVersion"`
	MaximumBlockCount          int32                          `json:"maximumBlockCount"`
	Participants               []BlockPackPresenceParticipant `json:"participants,omitempty"`
}

type BlockPackPresenceParticipant struct {
	UserPublicId      uuid.UUID `json:"userPublicId"`
	ChannelPermission string    `json:"channelPermission"`
	ConnectionCount   int       `json:"connectionCount"`
}

type BlockPackPresenceDeltaFrame struct {
	Version     int                          `json:"version"`
	Type        FrameType                    `json:"type"`
	ChannelType ChannelType                  `json:"channelType"`
	ChannelId   uuid.UUID                    `json:"channelId"`
	Participant BlockPackPresenceParticipant `json:"participant"`
}

type UnsubscribeFrame struct {
	Version            int       `json:"version"`
	Type               FrameType `json:"type"`
	RequestId          string    `json:"requestId,omitempty"`
	ConnectorChannelId uint32    `json:"connectorChannelId"`
}

type UnsubscribedFrame struct {
	Version            int         `json:"version"`
	Type               FrameType   `json:"type"`
	RequestId          string      `json:"requestId,omitempty"`
	ChannelType        ChannelType `json:"channelType"`
	ChannelId          uuid.UUID   `json:"channelId"`
	ConnectorChannelId uint32      `json:"connectorChannelId"`
}

type AcknowledgeFrame struct {
	Version            int       `json:"version"`
	Type               FrameType `json:"type"`
	RequestId          string    `json:"requestId,omitempty"`
	ConnectorChannelId uint32    `json:"connectorChannelId"`
	Sequence           int64     `json:"sequence"`
}

type AcknowledgedFrame struct {
	Version            int       `json:"version"`
	Type               FrameType `json:"type"`
	RequestId          string    `json:"requestId,omitempty"`
	ConnectorChannelId uint32    `json:"connectorChannelId"`
	Sequence           int64     `json:"sequence"`
}

type HeartbeatFrame struct {
	Version      int       `json:"version"`
	Type         FrameType `json:"type"`
	RequestId    string    `json:"requestId,omitempty"`
	UnixMilliNow int64     `json:"unixMilliNow"`
}

type ResourceEventFrame struct {
	Version            int        `json:"version"`
	Type               FrameType  `json:"type"`
	EventId            uuid.UUID  `json:"eventId"`
	EventType          string     `json:"eventType"`
	ResourceId         uuid.UUID  `json:"resourceId"`
	TargetUserPublicId *uuid.UUID `json:"targetUserPublicId,omitempty"`
	Change             string     `json:"change"`
	Permission         string     `json:"permission,omitempty"`
}

type NotificationFrame struct {
	Version          int             `json:"version"`
	Type             FrameType       `json:"type"`
	EventId          uuid.UUID       `json:"eventId"`
	NotificationId   uuid.UUID       `json:"notificationId"`
	NotificationType string          `json:"notificationType"`
	Priority         string          `json:"priority"`
	TemplateKey      string          `json:"templateKey"`
	TemplateVersion  int             `json:"templateVersion"`
	Payload          json.RawMessage `json:"payload"`
	CreatedAt        time.Time       `json:"createdAt"`
	ExpiresAt        *time.Time      `json:"expiresAt,omitempty"`
}

type RoutineTaskLifecycleFrame struct {
	Version             int       `json:"version"`
	Type                FrameType `json:"type"`
	EventId             uuid.UUID `json:"eventId"`
	RoutineTaskId       uuid.UUID `json:"routineTaskId"`
	RoutineTaskRecordId uuid.UUID `json:"routineTaskRecordId"`
	RoutineId           uuid.UUID `json:"routineId"`
	Purpose             string    `json:"purpose"`
	Status              string    `json:"status"`
	Attempt             int32     `json:"attempt"`
	OccurredAt          time.Time `json:"occurredAt"`
}
