package middlewares

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	cgateway "github.com/HiIamJeff67/notegic-backend/contracts/gateway/v1"
	sharedcontexts "github.com/HiIamJeff67/notegic-backend/shared/lib/contexts"
)

func TestTokenResponseMiddlewareAddsInternalTokenEnvelope(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.Use(TokenResponseMiddleware())
	router.POST("/", func(ctx *gin.Context) {
		ctx.Set(sharedcontexts.ContextFieldName_IsNewTokens.String(), true)
		ctx.Set(sharedcontexts.ContextFieldName_AccessToken.String(), "access-token")
		ctx.Set(sharedcontexts.ContextFieldName_CSRFToken.String(), "csrf-token")
		ctx.JSON(http.StatusOK, cgateway.Response[struct{}]{
			Version: cgateway.Version,
			Data:    struct{}{},
		})
	})

	request := httptest.NewRequest(http.MethodPost, "/", nil)
	responseRecorder := httptest.NewRecorder()
	router.ServeHTTP(responseRecorder, request)

	response := cgateway.Response[struct{}]{}
	if err := json.Unmarshal(responseRecorder.Body.Bytes(), &response); err != nil {
		t.Fatalf("decode internal response: %v", err)
	}
	if response.Tokens == nil || response.Tokens.AccessToken != "access-token" || response.Tokens.CSRFToken != "csrf-token" {
		t.Fatalf("expected internal token envelope, got %#v", response.Tokens)
	}
}
