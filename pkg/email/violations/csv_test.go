package violations_test

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	fakeor "github.com/openreports/reports-api/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	wg "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	fakewg "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/fake"
	"github.com/kyverno/policy-reporter/pkg/email/violations"
)

func TestCSVViolationAdaptersAndExpansion(t *testing.T) {
	// WGPolicy's Message becomes OpenReports Description. Both must survive the
	// reduced model, including no-resource and scope-fallback branches.
	for _, adapter := range []string{"openreports", "wgpolicy"} {
		t.Run(adapter, func(t *testing.T) {
			result := wg.PolicyReportResult{Source: "kyverno", Policy: "p", Rule: "r", Result: wg.StatusFail, Severity: wg.SeverityHigh, Message: "=formula, \"雪\"\nnext"}
			multi := result
			multi.Resources = []corev1.ObjectReference{{Kind: "Pod", Name: "z"}, {Kind: "Pod", Name: "a"}}
			warn := result
			warn.Result = wg.StatusWarn
			warn.Rule = ""
			warn.Message = "fallback rule"
			pass := result
			pass.Result = wg.StatusPass
			skip := result
			skip.Result = wg.StatusSkip
			polr := &wg.PolicyReport{ObjectMeta: metav1.ObjectMeta{Name: "p", Namespace: "ns"}, Summary: wg.PolicyReportSummary{Pass: 1, Skip: 1, Fail: 2, Warn: 1}, Results: []wg.PolicyReportResult{multi, result, warn, pass, skip}}
			scoped := polr.DeepCopy()
			scoped.Name = "scoped"
			scoped.Scope = &corev1.ObjectReference{Kind: "Deployment", Name: "fallback"}
			scoped.Results = []wg.PolicyReportResult{result}
			scoped.Summary = wg.PolicyReportSummary{Fail: 1}
			var generator *violations.Generator
			if adapter == "wgpolicy" {
				generator = violations.NewGenerator(nil, fakewg.NewSimpleClientset(polr, scoped).Wgpolicyk8sV1alpha2(), filter, false)
			} else {
				generator = violations.NewGenerator(fakeor.NewSimpleClientset(polr.ToOpenReports(), scoped.ToOpenReports()).OpenreportsV1alpha1(), nil, filter, false)
			}
			sources, err := generator.GenerateData(context.Background())
			require.NoError(t, err)
			before := fmt.Sprint(sources)
			reporter := violations.NewReporter("", "cluster", "Report")
			report, err := reporter.EmailReport(sources, "html", "csv")
			require.NoError(t, err)
			rows, err := csv.NewReader(strings.NewReader(string(report.Attachments[0].Data))).ReadAll()
			require.NoError(t, err)
			require.Len(t, rows, 6)
			require.Equal(t, []string{"cluster", "kyverno", "namespace", "ns", "p", "fallback rule", "", "", "warn", "high", "fallback rule"}, rows[1])
			require.Equal(t, []string{"cluster", "kyverno", "namespace", "ns", "p", "r", "", "", "fail", "high", "'=formula, \"雪\"\nnext"}, rows[2])
			require.Equal(t, "fallback", rows[3][7])
			require.Equal(t, "a", rows[4][7])
			require.Equal(t, "z", rows[5][7])
			for range 5 {
				got, e := reporter.EmailReport(sources, "html", "csv")
				require.NoError(t, e)
				require.Equal(t, report.Attachments, got.Attachments)
			}
			require.Equal(t, before, fmt.Sprint(sources))
			require.Equal(t, result.Message, polr.Results[0].Message)
		})
	}
}

func TestCSVViolationCompatibility(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "violations.html"), []byte("custom {{.ClusterName}}"), 0o600))
	reporter := violations.NewReporter(dir, "cluster", "Report")
	normal, err := reporter.Report(nil, "html")
	require.NoError(t, err)
	disabled, err := reporter.EmailReport(nil, "html", "")
	require.NoError(t, err)
	require.Equal(t, normal, disabled)
	require.Equal(t, "custom cluster", disabled.Message)
	empty, err := reporter.EmailReport(nil, "html", "csv")
	require.NoError(t, err)
	rows, err := csv.NewReader(strings.NewReader(string(empty.Attachments[0].Data))).ReadAll()
	require.NoError(t, err)
	require.Len(t, rows, 1)
	// Report, used by the HTML API, stays a template render after CSV calls.
	html, err := reporter.Report(nil, "HTML")
	require.NoError(t, err)
	require.Equal(t, "custom cluster", html.Message)
	require.Empty(t, html.Attachments)
	_, err = reporter.EmailReport(nil, "html", "xlsx")
	require.Error(t, err)
}
