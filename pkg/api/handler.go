package api

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/fjogeleit/policy-reporter/pkg/report"
)

// PolicyReportHandler for the PolicyReport REST API
func PolicyReportHandler(s *report.PolicyReportStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		reports := s.List()
		if len(reports) == 0 {
			fmt.Fprint(w, "[]")

			return
		}

		apiReports := make([]PolicyReport, 0, len(reports))
		for _, r := range reports {
			apiReports = append(apiReports, mapPolicyReport(r))
		}

		if err := json.NewEncoder(w).Encode(apiReports); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{ "message": "%s" }`, err.Error())
		}
	}
}

// ClusterPolicyReportHandler for the ClusterPolicyReport REST API
func ClusterPolicyReportHandler(s *report.ClusterPolicyReportStore) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		reports := s.List()
		if len(reports) == 0 {
			fmt.Fprint(w, "[]")

			return
		}

		apiReports := make([]ClusterPolicyReport, 0, len(reports))
		for _, r := range reports {
			apiReports = append(apiReports, mapClusterPolicyReport(r))
		}

		if err := json.NewEncoder(w).Encode(apiReports); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{ "message": "%s" }`, err.Error())
		}
	}
}

// TargetsHandler for the Targets REST API
func TargetsHandler(targets []Target) http.HandlerFunc {
	return func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(http.StatusOK)

		if len(targets) == 0 {
			fmt.Fprint(w, "[]")

			return
		}

		if err := json.NewEncoder(w).Encode(targets); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprintf(w, `{ "message": "%s" }`, err.Error())
		}
	}
}
