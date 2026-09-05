package developmentroutes

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestConfigureStaticRoutesServesLogo(t *testing.T) {
	workingDirectory, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(filepath.Join(workingDirectory, "../../../../../..")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.Chdir(workingDirectory); err != nil {
			t.Fatal(err)
		}
	})

	router := gin.New()
	configureStaticRoutes(router.Group("/api/development/v1"))

	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/development/v1/static/logo", nil))

	if response.Code != http.StatusOK {
		t.Fatalf("GET /static/logo status = %d, want %d", response.Code, http.StatusOK)
	}
	if !strings.HasPrefix(response.Header().Get("Content-Type"), "image/svg+xml") {
		t.Fatalf("GET /static/logo Content-Type = %q, want image/svg+xml", response.Header().Get("Content-Type"))
	}
	if response.Header().Get("Cross-Origin-Resource-Policy") != "cross-origin" {
		t.Fatalf("GET /static/logo Cross-Origin-Resource-Policy = %q, want cross-origin", response.Header().Get("Cross-Origin-Resource-Policy"))
	}
	if !strings.Contains(response.Body.String(), "<svg") {
		t.Fatal("GET /static/logo body does not contain SVG markup")
	}

	response = httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/development/v1/static/global-images/avatars/1", nil))
	if response.Code != http.StatusNotFound {
		t.Fatalf("GET removed avatar route status = %d, want %d", response.Code, http.StatusNotFound)
	}
}
