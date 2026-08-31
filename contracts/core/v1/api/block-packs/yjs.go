package apicontract

import (
	"github.com/google/uuid"

	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"
)

type InitializeBlockPackYjsDocumentReqDto struct {
	Blocks []cblocknote.ArborizedEditableBlock `json:"blocks"`
}

type InitializeBlockPackYjsDocumentResDto struct {
	Snapshot    []byte `json:"snapshot"`
	StateVector []byte `json:"stateVector"`
}

type UpdateBlockPackYjsDocumentBlockRequestDto struct {
	BlockId uuid.UUID                         `json:"blockId"`
	Block   cblocknote.ArborizedEditableBlock `json:"block"`
}

type UpdateBlockPackYjsDocumentRequestDto struct {
	BlockPackId uuid.UUID                                   `json:"blockPackId"`
	Blocks      []UpdateBlockPackYjsDocumentBlockRequestDto `json:"blocks"`
}

type UpdateBlockPackYjsDocumentBlockResultDto struct {
	BlockId uuid.UUID `json:"blockId"`
	Status  string    `json:"status"`
	Reason  string    `json:"reason,omitempty"`
}

type UpdateBlockPackYjsDocumentResponseDto struct {
	Blocks []UpdateBlockPackYjsDocumentBlockResultDto `json:"blocks"`
}
