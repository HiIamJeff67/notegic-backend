package gatewaycontract

import (
	"github.com/google/uuid"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

/* ============================== Response ============================== */

type Response[D any] struct {
	Version   string                 `json:"version"`
	Metadata  ResponseMetadata       `json:"metadata"`
	Tokens    *Tokens                `json:"tokens,omitempty"`
	Data      D                      `json:"data"`
	Exception *cexceptions.Exception `json:"exception,omitempty"`
}

/* ============================== Client Response ============================== */

// ClientResponse is the public response envelope exposed by the Gateway.
// Access and refresh tokens are intentionally absent from this contract.
type ClientResponse[D any] struct {
	Success           bool                   `json:"success"`
	Data              D                      `json:"data"`
	Exception         *cexceptions.Exception `json:"exception"`
	Embedded          *EmbeddedResponse      `json:"embedded,omitempty"`
	RefreshableTokens *RefreshableTokens     `json:"refreshableTokens,omitempty"`
}

// EmbeddedResponse contains non-sensitive context that the Gateway may attach
// to a public response for client-side identity correlation.
type EmbeddedResponse struct {
	PublicId uuid.UUID `json:"publicId"`
}

// RefreshableTokens contains only non-sensitive values that may be exposed to
// a client after a token refresh. Access and refresh tokens stay in cookies.
type RefreshableTokens struct {
	NewCSRFToken string `json:"newCSRFToken,omitempty"`
}
