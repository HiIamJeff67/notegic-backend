package adapterstransport

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/google/uuid"

	cblockpacks "github.com/HiIamJeff67/notegic-backend/contracts/core/v1/api/block-packs"
	cblocknote "github.com/HiIamJeff67/notegic-backend/contracts/types/blocknote"
	cenums "github.com/HiIamJeff67/notegic-backend/contracts/types/enums"

	coreconfig "github.com/HiIamJeff67/notegic-backend/runtimes/core/configs"
)

func TestDocumentInitializationClientInitializeDocuments(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(responseWriter http.ResponseWriter, request *http.Request) {
		var requestBody struct {
			Documents []cblockpacks.InitializeBlockPackYjsDocumentReqDto `json:"documents"`
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
		[]cblockpacks.InitializeBlockPackYjsDocumentReqDto{
			{
				Blocks: []cblocknote.ArborizedEditableBlock{
					{
						Id:       uuid.New(),
						Type:     cenums.BlockType_Paragraph,
						Props:    &cblocknote.BaseProps{},
						Content:  cblocknote.InlineContentList{},
						Children: []cblocknote.ArborizedEditableBlock{},
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
