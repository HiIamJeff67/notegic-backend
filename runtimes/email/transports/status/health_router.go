package status

import (
	"net/http"
)

func ConfigureHealthRouter(mux *http.ServeMux, isHealthy func() bool) {
	mux.HandleFunc("/healthz", func(writer http.ResponseWriter, _ *http.Request) {
		if !isHealthy() {
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		writer.WriteHeader(http.StatusOK)
	})
}
