package routinetasktypes

import (
	"encoding/json"
	"time"
)

type ExecutionItemStatus string

const (
	ExecutionItemStatus_Updated   ExecutionItemStatus = "updated"
	ExecutionItemStatus_Skipped   ExecutionItemStatus = "skipped"
	ExecutionItemStatus_Failed    ExecutionItemStatus = "failed"
	ExecutionItemStatus_Retrieved ExecutionItemStatus = "retrieved"
)

type ExecutionItemResult struct {
	ItemId string              `json:"itemId" validate:"required"`
	Status ExecutionItemStatus `json:"status" validate:"required"`
	Reason string              `json:"reason,omitempty"`
	Data   json.RawMessage     `json:"data,omitempty"`
}

type ExecutionResult struct {
	Retrieved int                   `json:"retrieved"`
	Updated   int                   `json:"updated"`
	Skipped   int                   `json:"skipped"`
	Failed    int                   `json:"failed"`
	Items     []ExecutionItemResult `json:"items" validate:"dive"`
	At        time.Time             `json:"at"`
}
