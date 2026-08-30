package statuse2etest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	emailstatus "github.com/HiIamJeff67/notegic-backend/runtimes/email/transports/status"
)

func TestStatusRoutes(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		configure func(*http.ServeMux, func() bool)
		path      string
		healthy   bool
		expected  int
	}{
		{name: "health ready", configure: emailstatus.ConfigureHealthRouter, path: "/healthz", healthy: true, expected: http.StatusOK},
		{name: "health unavailable", configure: emailstatus.ConfigureHealthRouter, path: "/healthz", healthy: false, expected: http.StatusServiceUnavailable},
		{name: "started", configure: emailstatus.ConfigureStartedRouter, path: "/startedz", healthy: true, expected: http.StatusOK},
		{name: "not started", configure: emailstatus.ConfigureStartedRouter, path: "/startedz", healthy: false, expected: http.StatusServiceUnavailable},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			mux := http.NewServeMux()
			testCase.configure(mux, func() bool { return testCase.healthy })

			request := httptest.NewRequest(http.MethodGet, testCase.path, nil)
			response := httptest.NewRecorder()
			mux.ServeHTTP(response, request)

			if response.Code != testCase.expected {
				t.Fatalf("expected status %d, got %d", testCase.expected, response.Code)
			}
		})
	}
}
