package api

import (
	"fmt"
	"log"
	"net/http"
)

// HealthzHandler for the Halthz REST API
func HealthzHandler(found map[string]string) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		if len(found) == 0 {
			w.Header().Set("Content-Type", "application/json; charset=UTF-8")
			w.WriteHeader(http.StatusServiceUnavailable)

			log.Println("[WARNING] - Healthz Check: No policyreport.wgpolicyk8s.io and clusterpolicyreport.wgpolicyk8s.io crds are found")

			fmt.Fprint(w, `{ "error": "No PolicyReport CRDs found" }`)

			return
		}

		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, "{}")
	}
}

// ReadyHandler for the Halthz REST API
func ReadyHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "{}")
	}
}
