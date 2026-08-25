package realtimegatewaycontract

import (
	"time"

	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

type RequestEnvelope interface {
	GetVersion() string
	GetOperation() string
	GetMetadata() RequestMetadata
}

type Request[D any] struct {
	Version   string          `json:"version"`
	Operation string          `json:"operation"`
	Metadata  RequestMetadata `json:"metadata"`
	Dto       D               `json:"dto"`
}

func (r *Request[D]) GetVersion() string {
	return r.Version
}

func (r *Request[D]) GetOperation() string {
	return r.Operation
}

func (r *Request[D]) GetMetadata() RequestMetadata {
	return r.Metadata
}

type RequestMetadata struct {
	RequestId   string `json:"requestId"`
	TraceParent string `json:"traceParent,omitempty"`
}

type Response[D any] struct {
	Version   string                 `json:"version"`
	Metadata  ResponseMetadata       `json:"metadata"`
	Data      D                      `json:"data"`
	Exception *cexceptions.Exception `json:"exception,omitempty"`
}

type ResponseMetadata struct {
	RequestId   string    `json:"requestId"`
	RespondedAt time.Time `json:"respondedAt"`
}
