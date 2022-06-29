package api

import (
	"fmt"
	"net/http"
)

// HealthzHandler for the Halthz REST API
func HealthzHandler(synced func() bool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if !synced() {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusServiceUnavailable)

			fmt.Fprint(w, `{ "error": "Informers not in sync" }`)

			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, "{}")
	}
}

// ReadyHandler for the Halthz REST API
func ReadyHandler(synced func() bool) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if !synced() {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusServiceUnavailable)

			fmt.Fprint(w, `{ "error": "Informers not in sync" }`)

			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{}")
	}
}
