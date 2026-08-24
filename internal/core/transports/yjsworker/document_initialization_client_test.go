package adapterstransport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	blockpackscontract "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/block-packs"
	blocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"
	blockenums "github.com/HiIamJeff67/notegic-backend/contracts/types/models/enums"

	coreconfig "github.com/HiIamJeff67/notegic-backend/internal/core/configs"
)

func TestDocumentInitializationClientInitializeDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		var requestBody struct {
			Documents []blockpackscontract.InitializeBlockPackYjsDocumentReqDto `json:"documents"`
		}
		if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
			t.Fatalf("Decode() error = %v", err)
		}
		if len(requestBody.Documents) != 1 {
			t.Fatalf("documents = %d, want 1", len(requestBody.Documents))
		}

		responseWriter.Header().Set("Content-Type", "application/json")
		_, _ = responseWriter.Write([]byte(`{"documents":[{"snapshot":"AQ==","stateVector":"Ag=="}]}`))
	}))
	defer server.Close()

	client := NewDocumentInitializationClient(coreconfig.YjsDocumentInitializationConfig{
		Endpoint: server.URL,
		Timeout:  time.Second,
	})
	responseDtos, err := client.InitializeDocuments(
		context.Background(),
		[]blockpackscontract.InitializeBlockPackYjsDocumentReqDto{
			{
				Blocks: []blocknote.ArborizedEditableBlock{
					{
						Id:       uuid.New(),
						Type:     blockenums.BlockType_Paragraph,
						Props:    &blocknote.BaseProps{},
						Content:  blocknote.InlineContentList{},
						Children: []blocknote.ArborizedEditableBlock{},
					},
				},
			},
		},
	)
	if err != nil {
		t.Fatalf("InitializeDocuments() error = %v", err)
	}
	if len(responseDtos) != 1 || len(responseDtos[0].Snapshot) != 1 || len(responseDtos[0].StateVector) != 1 {
		t.Fatalf("InitializeDocuments() = %#v", responseDtos)
	}
}
