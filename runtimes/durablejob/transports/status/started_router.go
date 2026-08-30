package status

import (
	"net/http"
)

func ConfigureStartedRouter(mux *http.ServeMux, isStarted func() bool) {
	mux.HandleFunc("/startedz", func(writer http.ResponseWriter, _ *http.Request) {
		if !isStarted() {
			writer.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		writer.WriteHeader(http.StatusOK)
	})
}
