package v1_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"

	"github.com/kyverno/policy-reporter/pkg/api"
	v1 "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/target"
)

func TestHTMLReportWithCSVEmailConfiguration(t *testing.T) {
	db, err := database.NewSQLiteDB(filepath.Join(t.TempDir(), "reports.db"))
	require.NoError(t, err)
	t.Cleanup(func() { require.NoError(t, db.Close()) })
	store, err := database.NewStore(db, "test")
	require.NoError(t, err)
	require.NoError(t, store.PrepareDatabase(context.Background()))
	store.Add(context.Background(), reconditioner.Prepare(&openreports.ClusterReportAdapter{ClusterReport: fixtures.KyvernoClusterPolicyReport.DeepCopy()}))
	var defaultBody string
	for _, format := range []string{"", "csv"} {
		cfg := &config.Config{EmailReports: config.EmailReports{ClusterName: "test cluster", TitlePrefix: "Report", Violations: config.EmailReport{AttachmentFormat: format}}}
		reporter := config.NewResolver(cfg, nil).ViolationsReporter()
		// Exercise the same reporter's email path before serving HTML.
		_, err := reporter.EmailReport(nil, "html", format)
		require.NoError(t, err)
		server := api.NewServer(gin.New(), v1.WithAPI(store, target.NewCollection(), reporter))
		response := httptest.NewRecorder()
		server.Serve(response, httptest.NewRequest(http.MethodGet, "/v1/html-report/violations", nil))
		require.Equal(t, http.StatusOK, response.Code)
		require.Equal(t, "text/html; charset=utf-8", response.Header().Get("Content-Type"))
		require.Contains(t, response.Body.String(), "test cluster")
		require.Contains(t, response.Body.String(), "<table")
		require.NotContains(t, response.Body.String(), "attached as")
		if format == "" {
			defaultBody = response.Body.String()
		} else {
			require.Equal(t, defaultBody, response.Body.String())
		}
	}
}
