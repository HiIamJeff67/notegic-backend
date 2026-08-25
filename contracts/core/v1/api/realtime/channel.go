package apicontract

import (
	"time"

	"github.com/google/uuid"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type CreateMyBlockPackChannelTicketRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct {
			BlockPackId uuid.UUID `json:"blockPackId" validate:"required"`
			Permission  string    `json:"permission" validate:"required,oneof=read write"`
		},
		struct{},
		struct{},
	]
}

type CreateMyBlockPackChannelTicketResponseDto struct {
	ChannelTicket              string    `json:"channelTicket"`
	ExpiresAt                  time.Time `json:"expiresAt"`
	ChannelType                string    `json:"channelType"`
	ChannelId                  uuid.UUID `json:"channelId"`
	Permission                 string    `json:"permission"`
	RoomName                   string    `json:"roomName"`
	FragmentName               string    `json:"fragmentName"`
	SchemaId                   string    `json:"schemaId"`
	SchemaVersion              int       `json:"schemaVersion"`
	RealtimeProtocolVersion    int       `json:"realtimeProtocolVersion"`
	DocumentQuotaPolicyVersion int       `json:"documentQuotaPolicyVersion"`
	MaximumBlockCount          int32     `json:"maximumBlockCount"`
	LastUpdateSequence         int64     `json:"lastUpdateSequence"`
	CompactedUntilSequence     int64     `json:"compactedUntilSequence"`
}
