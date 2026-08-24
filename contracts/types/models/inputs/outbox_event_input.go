package inputs

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"

	eventcontract "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
)

type CreateOutboxEventInput struct {
	Id            uuid.UUID                   `json:"id"`
	AggregateType eventcontract.AggregateType `json:"aggregateType"`
	AggregateId   uuid.UUID                   `json:"aggregateId"`
	EventType     eventcontract.EventType     `json:"eventType"`
	Topic         eventcontract.Topic         `json:"topic"`
	KafkaKey      string                      `json:"kafkaKey"`
	Payload       json.RawMessage             `json:"payload"`
	Metadata      json.RawMessage             `json:"metadata"`
	AvailableAt   time.Time                   `json:"availableAt"`
}

type FailedOutboxEventInput struct {
	Id          uuid.UUID `json:"id"`
	LastError   string    `json:"lastError"`
	AvailableAt time.Time `json:"availableAt"`
}
