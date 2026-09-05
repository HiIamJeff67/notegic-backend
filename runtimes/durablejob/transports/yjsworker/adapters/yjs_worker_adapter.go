package adapters

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	cyjsworker "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1"
)

type YjsWorkerAdapter struct {
	endpoint   string
	httpClient *http.Client
}

func NewYjsWorkerAdapter(endpoint string, timeout time.Duration) *YjsWorkerAdapter {
	return &YjsWorkerAdapter{
		endpoint: endpoint,
		httpClient: &http.Client{
			Timeout: timeout,
		},
	}
}

func (adapter *YjsWorkerAdapter) call(
	ctx context.Context,
	operation string,
	request any,
	response any,
) error {
	payload, err := json.Marshal(request)
	if err != nil {
		return fmt.Errorf("encode %s request: %w", operation, err)
	}
	if len(payload) > cyjsworker.YjsMaintenanceMaximumPayloadBytes {
		return fmt.Errorf("%s request exceeds the worker payload limit", operation)
	}

	httpRequest, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		adapter.endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return fmt.Errorf("create %s request: %w", operation, err)
	}
	httpRequest.Header.Set("Content-Type", "application/json")
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(httpRequest.Header))

	httpResponse, err := adapter.httpClient.Do(httpRequest)
	if err != nil {
		return fmt.Errorf("send %s request: %w", operation, err)
	}
	defer httpResponse.Body.Close()

	if httpResponse.StatusCode != http.StatusOK {
		return fmt.Errorf("Yjs document worker returned %s for %s", httpResponse.Status, operation)
	}

	responsePayload, err := io.ReadAll(io.LimitReader(
		httpResponse.Body,
		int64(cyjsworker.YjsMaintenanceMaximumPayloadBytes)+1,
	))
	if err != nil {
		return fmt.Errorf("read %s response: %w", operation, err)
	}
	if len(responsePayload) > cyjsworker.YjsMaintenanceMaximumPayloadBytes {
		return fmt.Errorf("%s response exceeds the worker payload limit", operation)
	}
	if err := json.Unmarshal(responsePayload, response); err != nil {
		return fmt.Errorf("decode %s response: %w", operation, err)
	}

	return nil
}
