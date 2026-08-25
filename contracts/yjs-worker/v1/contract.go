package adapterscontract

import (
	"time"

	"github.com/google/uuid"

	cevent "github.com/HiIamJeff67/notegic-backend/contracts/types/events"
)

const Version = "v1"

type CommandType string

const (
	CommandType_LoadYjsDocument            CommandType = "LoadYjsDocument"
	CommandType_AppendYjsUpdate            CommandType = "AppendYjsUpdate"
	CommandType_LoadCompactableYjsDocument CommandType = "LoadCompactableYjsDocument"
	CommandType_ApplyCompactedYjsDocument  CommandType = "ApplyCompactedYjsDocument"
	CommandType_ApplyBlockProjection       CommandType = "ApplyBlockProjection"
)

type CommandEnvelope[D any] struct {
	SchemaVersion string               `json:"schemaVersion"`
	CommandId     uuid.UUID            `json:"commandId"`
	CommandType   CommandType          `json:"commandType"`
	BlockPackId   uuid.UUID            `json:"blockPackId"`
	CorrelationId string               `json:"correlationId"`
	CausationId   *uuid.UUID           `json:"causationId,omitempty"`
	Trace         cevent.TraceMetadata `json:"trace"`
	Producer      string               `json:"producer"`
	OccurredAt    time.Time            `json:"occurredAt"`
	Data          D                    `json:"data"`
}

type ReplyEnvelope[D any] struct {
	SchemaVersion string               `json:"schemaVersion"`
	CommandId     uuid.UUID            `json:"commandId"`
	CommandType   CommandType          `json:"commandType"`
	BlockPackId   uuid.UUID            `json:"blockPackId"`
	CorrelationId string               `json:"correlationId"`
	CausationId   *uuid.UUID           `json:"causationId,omitempty"`
	Trace         cevent.TraceMetadata `json:"trace"`
	Producer      string               `json:"producer"`
	RespondedAt   time.Time            `json:"respondedAt"`
	Data          D                    `json:"data"`
	Error         *Error               `json:"error,omitempty"`
}

type Error struct {
	Code      string `json:"code"`
	Message   string `json:"message"`
	Retryable bool   `json:"retryable"`
}
