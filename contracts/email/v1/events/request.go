package emaileventscontract

import (
	"time"

	"github.com/google/uuid"
)

type SendEmailRequestDto[P any] struct {
	RequestId  uuid.UUID `json:"requestId"`
	Operation  string    `json:"operation"`
	OccurredAt time.Time `json:"occurredAt"`
	To         string    `json:"to" validate:"required,email"`
	Pattern    P         `json:"pattern"`
}
