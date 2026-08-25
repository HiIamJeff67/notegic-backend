package statuse2etest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	gin "github.com/gin-gonic/gin"

	gatewaystatus "github.com/HiIamJeff67/notegic-backend/internal/apigateway/transports/status"
)

func TestStatusRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)

	for _, testCase := range []struct {
		name      string
		configure func(gin.IRouter, func() bool)
		path      string
		healthy   bool
		expected  int
	}{
		{name: "health ready", configure: gatewaystatus.ConfigureHealthRouter, path: "/healthz", healthy: true, expected: http.StatusOK},
		{name: "health unavailable", configure: gatewaystatus.ConfigureHealthRouter, path: "/healthz", healthy: false, expected: http.StatusServiceUnavailable},
		{name: "started", configure: gatewaystatus.ConfigureStartedRouter, path: "/startedz", healthy: true, expected: http.StatusOK},
		{name: "not started", configure: gatewaystatus.ConfigureStartedRouter, path: "/startedz", healthy: false, expected: http.StatusServiceUnavailable},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			router := gin.New()
			testCase.configure(router, func() bool { return testCase.healthy })

			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			response := httptest.NewRecorder()
			router.ServeHTTP(response, request)

			if response.Code != testCase.expected {
				t.Fatalf("expected status %d, got %d", testCase.expected, response.Code)
			}
		})
	}
}
