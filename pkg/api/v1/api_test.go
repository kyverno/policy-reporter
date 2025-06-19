package v1_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/api"
	v1 "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/email/violations"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/report/result"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/target/webhook"
)

var reconditioner = result.NewReconditioner(nil)

func TestV1(t *testing.T) {
	db, err := database.NewSQLiteDB("db_v2.db")
	if err != nil {
		assert.Fail(t, "failed to init SQLite DB")
	}

	store, err := database.NewStore(db, "1.0")
	if err != nil {
		assert.Fail(t, "failed to init Store")
	}

	if err := store.PrepareDatabase(context.Background()); err != nil {
		assert.Fail(t, "failed to prepare Store")
	}

	store.Add(context.Background(), reconditioner.Prepare(fixtures.DefaultPolicyReport))
	store.Add(context.Background(), reconditioner.Prepare(&openreports.ReportAdapter{Report: fixtures.KyvernoPolicyReport}))
	store.Add(context.Background(), reconditioner.Prepare(&openreports.ClusterReportAdapter{ClusterReport: fixtures.KyvernoClusterPolicyReport}))

	gin.SetMode(gin.ReleaseMode)

	server := api.NewServer(gin.New(), v1.WithAPI(store, target.NewCollection(&target.Target{
		Client: webhook.NewClient(webhook.Options{
			ClientOptions: target.ClientOptions{
				Name:                  "Webhook",
				SkipExistingOnStartup: true,
				ResultFilter: &report.ResultFilter{
					MinimumSeverity: "",
					Sources:         []string{"Kyverno"},
				},
			},
			Host: "http://localhost:8080",
		}),
	}), violations.NewReporter("../../../templates", "Cluster", "Report")))

	t.Run("TargetResponse", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/targets", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]v1.Target, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 1, len(resp))
			assert.Contains(t, resp, v1.Target{
				Name:                  "Webhook",
				MinimumSeverity:       v1alpha2.SeverityInfo,
				Sources:               []string{"Kyverno"},
				SkipExistingOnStartup: true,
			})
		}
	})

	t.Run("ListPolicyReports", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/policy-reports", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := api.Paginated[v1.PolicyReport]{}

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 2, resp.Count)
		}
	})

	t.Run("ListClusterPolicyReports", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/cluster-policy-reports", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := api.Paginated[v1.PolicyReport]{}

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 1, resp.Count)
		}
	})

	t.Run("ListNamespaces", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/namespaces", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]string, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 2, len(resp))
		}
	})

	t.Run("RuleStatusCounts", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/rule-status-count?policy=required-limit&rule=resource-limit-required", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]v1.StatusCount, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 5, len(resp))
			assert.Contains(t, resp, v1.StatusCount{Status: "pass", Count: 1})
			assert.Contains(t, resp, v1.StatusCount{Status: "warn", Count: 1})
		}
	})

	t.Run("ListClusterFilter(Source)", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/cluster-resources/sources", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]string, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 1, len(resp))
			assert.Contains(t, resp, "Kyverno")
		}
	})

	t.Run("ListNamespacedFilter(Source)", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/namespaced-resources/sources", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]string, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 2, len(resp))
			assert.Contains(t, resp, "Kyverno")
		}
	})

	t.Run("ListClusterResources", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/cluster-resources/resources", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]v1.Resource, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 1, len(resp))
		}
	})

	t.Run("ListNamespacedResources", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/namespaced-resources/resources", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]v1.Resource, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 4, len(resp))
		}
	})

	t.Run("ListClusterStatusCounts", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/cluster-resources/status-counts", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]v1.StatusCount, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 5, len(resp))
		}
	})

	t.Run("ListNamespacedStatusCounts", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/namespaced-resources/status-counts", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := make([]v1.NamespaceCount, 0)

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 5, len(resp))
		}
	})

	t.Run("ListClusterResults", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/cluster-resources/results", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := api.Paginated[v1.Result]{}

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 1, resp.Count)
		}
	})

	t.Run("ListNamespacedResults", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/namespaced-resources/results", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)

		if ok := assert.Equal(t, http.StatusOK, w.Code); ok {
			resp := api.Paginated[v1.Result]{}

			json.NewDecoder(w.Body).Decode(&resp)

			assert.Equal(t, 5, resp.Count)
		}
	})

	t.Run("HTMLViolationsReport", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/v1/html-report/violations", nil)
		w := httptest.NewRecorder()

		server.Serve(w, req)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}
