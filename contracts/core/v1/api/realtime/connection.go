package apicontract

import (
	"time"

	coreapi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api"
)

type CreateMyRealtimeConnectionTicketRequestDto struct {
	coreapi.RequestDto[
		struct {
			UserAgent string `json:"userAgent" validate:"required,isuseragent"`
		},
		struct{},
		struct{},
		struct{},
	]
}

type CreateMyRealtimeConnectionTicketResponseDto struct {
	RealtimeEndpoint        string    `json:"realtimeEndpoint"`
	RealtimeProtocolVersion int       `json:"realtimeProtocolVersion"`
	ConnectionTicket        string    `json:"connectionTicket"`
	ExpiresAt               time.Time `json:"expiresAt"`
}
