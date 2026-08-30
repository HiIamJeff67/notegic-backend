package statuse2etest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	durablejobstatus "github.com/HiIamJeff67/notegic-backend/runtimes/durablejob/transports/status"
)

func TestStatusRoutes(t *testing.T) {
	for _, testCase := range []struct {
		name      string
		configure func(*http.ServeMux, func() bool)
		path      string
		healthy   bool
		expected  int
	}{
		{name: "health ready", configure: durablejobstatus.ConfigureHealthRouter, path: "/healthz", healthy: true, expected: http.StatusOK},
		{name: "health unavailable", configure: durablejobstatus.ConfigureHealthRouter, path: "/healthz", healthy: false, expected: http.StatusServiceUnavailable},
		{name: "started", configure: durablejobstatus.ConfigureStartedRouter, path: "/startedz", healthy: true, expected: http.StatusOK},
		{name: "not started", configure: durablejobstatus.ConfigureStartedRouter, path: "/startedz", healthy: false, expected: http.StatusServiceUnavailable},
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
