package durablejobcontract

import (
	"github.com/google/uuid"

	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"
)

const ApplyBlockProjectionOperation = "durablejob.apply-block-projection"

type ApplyBlockProjectionRequestDto struct {
	Documents []ApplyBlockProjectionDocumentRequestDto `json:"documents" validate:"required"`
}

type ApplyBlockProjectionDocumentRequestDto struct {
	BlockPackId uuid.UUID                       `json:"blockPackId"`
	Projection  ApplyBlockProjectionDocumentDto `json:"projection"`
}

type ApplyBlockProjectionDocumentDto struct {
	SchemaId          string                              `json:"schemaId"`
	SchemaVersion     int                                 `json:"schemaVersion"`
	ProjectedSequence int64                               `json:"projectedSequence"`
	Blocks            []cblocknote.ArborizedEditableBlock `json:"blocks"`
}

type ApplyBlockProjectionResponseDto struct {
	AppliedBlockPackIds []uuid.UUID `json:"appliedBlockPackIds"`
}
