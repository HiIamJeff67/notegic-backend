package adapters

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	cexceptions "github.com/HiIamJeff67/notegic-backend/contracts/types/exceptions"
)

func TestCoreAdapterForwardsVersionedEnvelopeAndMetadata(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		if request.Header.Get("Authorization") != "Bearer delegation-token" {
			t.Fatal("expected delegation token header")
		}
		if request.Header.Get("Traceparent") != "00-trace" {
			t.Fatal("expected trace parent header")
		}
		if request.Header.Get("Cookie") != "" {
			t.Fatal("cookies must not cross the Gateway/Core boundary")
		}

		requestEnvelope := cgateway.Request[struct{}]{}
		if err := json.NewDecoder(request.Body).Decode(&requestEnvelope); err != nil {
			t.Fatalf("decode request envelope: %v", err)
		}
		if requestEnvelope.Version != cgateway.Version || requestEnvelope.Operation != "station.get" {
			t.Fatal("expected versioned station request envelope")
		}
		if requestEnvelope.Tokens.AccessToken != "access-token" {
			t.Fatalf("expected typed access token, got %q", requestEnvelope.Tokens.AccessToken)
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(responseWriter).Encode(&cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId: requestEnvelope.Metadata.RequestId,
			},
			Data: struct{}{},
		}); err != nil {
			t.Fatalf("encode response envelope: %v", err)
		}
	}))
	defer server.Close()

	client := NewCoreAdapter(server.URL, time.Second)
	response, err := call[struct{}, struct{}](
		client,
		nil,
		context.Background(),
		http.MethodPost,
		"/v1/operations",
		"delegation-token",
		http.Header{
			"Cookie":     []string{"accessToken=test-token"},
			"User-Agent": []string{"test-agent"},
			"X-Real-IP":  []string{"192.0.2.1"},
		},
		&cgateway.Request[struct{}]{
			Operation: "station.get",
			Metadata: cgateway.RequestMetadata{
				RequestId:   "request-id",
				TraceParent: "00-trace",
			},
			Tokens: cgateway.Tokens{
				AccessToken: "access-token",
			},
		},
	)
	if err != nil {
		t.Fatalf("execute Core service request: %v", err)
	}
	if response.Metadata.RequestId != "request-id" {
		t.Fatalf("expected request ID request-id, got %s", response.Metadata.RequestId)
	}
}

func TestCoreAdapterDecodesErrorEnvelopeBeforeResponseData(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, _ *http.Request) {
		responseWriter.Header().Set("Content-Type", "application/json")
		responseWriter.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(responseWriter).Encode(&cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Metadata: cgateway.ResponseMetadata{
				RequestId: "request-id",
			},
			Data: struct{}{},
			Exception: cexceptions.New(
				"NotFound",
				"Core",
				"Get",
				"The requested resource was not found",
				http.StatusNotFound,
			),
		}); err != nil {
			t.Fatalf("encode error response envelope: %v", err)
		}
	}))
	defer server.Close()

	_, exception := call[struct{}, []struct{}](
		NewCoreAdapter(server.URL, time.Second),
		nil,
		context.Background(),
		http.MethodPost,
		"/v1/operations",
		"delegation-token",
		nil,
		&cgateway.Request[struct{}]{
			Metadata: cgateway.RequestMetadata{RequestId: "request-id"},
		},
	)
	if exception == nil || exception.Reason != "NotFound" {
		t.Fatalf("expected NotFound exception, got %v", exception)
	}
}
