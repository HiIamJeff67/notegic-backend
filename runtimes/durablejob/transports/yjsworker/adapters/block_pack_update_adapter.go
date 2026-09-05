package adapters

import (
	"context"
	"errors"

	capi "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/block-packs"

	durablejobconfig "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/configs"
)

type BlockPackUpdateAdapter struct {
	*YjsWorkerAdapter
}

func NewBlockPackUpdateAdapter(
	config durablejobconfig.YjsDocumentInitializationConfig,
) *BlockPackUpdateAdapter {
	return &BlockPackUpdateAdapter{
		YjsWorkerAdapter: NewYjsWorkerAdapter(config.UpdateEndpoint, config.Timeout),
	}
}

func (adapter *BlockPackUpdateAdapter) Call(
	ctx context.Context,
	requestDto capi.UpdateBlockPackYjsDocumentRequestDto,
) (*capi.UpdateBlockPackYjsDocumentResponseDto, error) {
	response := capi.UpdateBlockPackYjsDocumentResponseDto{}
	if err := adapter.call(
		ctx,
		"Yjs block pack update",
		requestDto,
		&response,
	); err != nil {
		return nil, err
	}
	if len(response.Blocks) != len(requestDto.Blocks) {
		return nil, errors.New("Yjs block pack update response is incomplete")
	}

	return &response, nil
}
