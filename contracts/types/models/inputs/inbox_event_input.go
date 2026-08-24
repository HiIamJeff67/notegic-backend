package inputs

import "github.com/google/uuid"

type CreateInboxEventInput struct {
	EventId uuid.UUID
}
