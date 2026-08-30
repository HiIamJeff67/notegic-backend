package middlewares

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestKeyMiddlewareRejectsMissingKeyAndAcceptsFormattedKey(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/", KeyMiddleware(), func(ctx *gin.Context) { ctx.Status(http.StatusNoContent) })

	missing := httptest.NewRecorder()
	router.ServeHTTP(missing, httptest.NewRequest(http.MethodGet, "/", nil))
	if missing.Code != http.StatusUnauthorized {
		t.Fatalf("expected missing key to return 401, got %d", missing.Code)
	}

	valid := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	request.Header.Set("X-API-Key", "nzy_12345678901234567890123456789012")
	router.ServeHTTP(valid, request)
	if valid.Code != http.StatusNoContent {
		t.Fatalf("expected formatted key to pass edge middleware, got %d", valid.Code)
	}
}
