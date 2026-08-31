package yjsworkertransport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/block-packs"
	cyjsworker "github.com/HiIamJeff67/notegic-backend/contracts/yjs-worker/v1"

	durablejobconfig "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/configs"
)

type BlockPackUpdateClient struct {
	endpoint   string
	httpClient *http.Client
}

func NewBlockPackUpdateClient(config durablejobconfig.YjsDocumentInitializationConfig) *BlockPackUpdateClient {
	return &BlockPackUpdateClient{
		endpoint: config.UpdateEndpoint,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

func (c *BlockPackUpdateClient) UpdateBlockPack(
	ctx context.Context,
	requestDto capi.UpdateBlockPackYjsDocumentRequestDto,
) (*capi.UpdateBlockPackYjsDocumentResponseDto, error) {
	payload, err := json.Marshal(requestDto)
	if err != nil {
		return nil, fmt.Errorf("encode Yjs block pack update request: %w", err)
	}
	if len(payload) > cyjsworker.YjsMaintenanceMaximumPayloadBytes {
		return nil, errors.New("Yjs block pack update request exceeds the worker payload limit")
	}
	request, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.endpoint,
		bytes.NewReader(payload),
	)
	if err != nil {
		return nil, fmt.Errorf("create Yjs block pack update request: %w", err)
	}
	request.Header.Set("Content-Type", "application/json")
	otel.GetTextMapPropagator().Inject(ctx, propagation.HeaderCarrier(request.Header))
	response, err := c.httpClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("send Yjs block pack update request: %w", err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Yjs block pack update worker returned %s", response.Status)
	}
	responsePayload, err := io.ReadAll(io.LimitReader(
		response.Body,
		int64(cyjsworker.YjsMaintenanceMaximumPayloadBytes)+1,
	))
	if err != nil {
		return nil, fmt.Errorf("read Yjs block pack update response: %w", err)
	}
	if len(responsePayload) > cyjsworker.YjsMaintenanceMaximumPayloadBytes {
		return nil, errors.New("Yjs block pack update response exceeds the worker payload limit")
	}
	var responseDto capi.UpdateBlockPackYjsDocumentResponseDto
	if err := json.Unmarshal(responsePayload, &responseDto); err != nil {
		return nil, fmt.Errorf("decode Yjs block pack update response: %w", err)
	}
	if len(responseDto.Blocks) != len(requestDto.Blocks) {
		return nil, errors.New("Yjs block pack update response is incomplete")
	}

	return &responseDto, nil
}
