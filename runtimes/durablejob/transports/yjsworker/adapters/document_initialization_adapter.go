package adapters

import (
	"context"
	"errors"

	cblockpacks "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/block-packs"
	durablejobconfig "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/configs"
)

type DocumentInitializationAdapter struct {
	*YjsWorkerAdapter
}

func NewDocumentInitializationAdapter(
	config durablejobconfig.YjsDocumentInitializationConfig,
) *DocumentInitializationAdapter {
	return &DocumentInitializationAdapter{
		YjsWorkerAdapter: NewYjsWorkerAdapter(config.Endpoint, config.Timeout),
	}
}

func (adapter *DocumentInitializationAdapter) Call(
	ctx context.Context,
	requestDtos []cblockpacks.InitializeBlockPackYjsDocumentReqDto,
) ([]cblockpacks.InitializeBlockPackYjsDocumentResDto, error) {
	if len(requestDtos) == 0 {
		return []cblockpacks.InitializeBlockPackYjsDocumentResDto{}, nil
	}

	request := struct {
		Documents []cblockpacks.InitializeBlockPackYjsDocumentReqDto `json:"documents"`
	}{
		Documents: requestDtos,
	}
	response := struct {
		Documents []cblockpacks.InitializeBlockPackYjsDocumentResDto `json:"documents"`
	}{}
	if err := adapter.call(
		ctx,
		"Yjs document initialization",
		request,
		&response,
	); err != nil {
		return nil, err
	}
	if len(response.Documents) != len(requestDtos) {
		return nil, errors.New("Yjs document initialization response is incomplete")
	}
	for _, document := range response.Documents {
		if len(document.Snapshot) == 0 || len(document.StateVector) == 0 {
			return nil, errors.New("Yjs document initialization response is invalid")
		}
	}

	return response.Documents, nil
}
