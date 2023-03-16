package api

import (
	"fmt"
	"net/http"

	"go.uber.org/zap"
)

// HealthzHandler for the Halthz REST API
func HealthzHandler(synced func() bool, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if !synced() {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusServiceUnavailable)

			fmt.Fprint(w, `{ "error": "Informers not in sync" }`)

			if logger != nil {
				logger.Warn("informers not synced yet, waiting for k8s client to complete startup")
			}

			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, "{}")
	}
}

// ReadyHandler for the Halthz REST API
func ReadyHandler(synced func() bool, logger *zap.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if !synced() {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusServiceUnavailable)

			fmt.Fprint(w, `{ "error": "Informers not in sync" }`)

			if logger != nil {
				logger.Warn("informers not synced yet, waiting for k8s client to be up")
			}

			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{}")
	}
}
