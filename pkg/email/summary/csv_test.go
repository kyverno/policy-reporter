package summary_test

import (
	"context"
	"encoding/csv"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	fakeor "github.com/openreports/reports-api/pkg/client/clientset/versioned/fake"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	wg "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	fakewg "github.com/kyverno/policy-reporter/pkg/crd/client/clientset/versioned/fake"
	"github.com/kyverno/policy-reporter/pkg/email/summary"
)

func TestCSVSummaryCompatibility(t *testing.T) {
	dir := t.TempDir()
	require.NoError(t, os.WriteFile(filepath.Join(dir, "summary.html"), []byte("custom {{.ClusterName}}"), 0o600))
	reporter := summary.NewReporter(dir, "cluster", "Report")
	normal, err := reporter.Report(nil, "html")
	require.NoError(t, err)
	disabled, err := reporter.EmailReport(nil, "html", "")
	require.NoError(t, err)
	require.Equal(t, normal, disabled)
	require.Equal(t, "custom cluster", disabled.Message)
	require.Empty(t, disabled.Attachments)
	empty, err := reporter.EmailReport(nil, "html", "csv")
	require.NoError(t, err)
	rows, err := csv.NewReader(strings.NewReader(string(empty.Attachments[0].Data))).ReadAll()
	require.NoError(t, err)
	require.Len(t, rows, 1)
	_, err = reporter.EmailReport(nil, "html", "xlsx")
	require.Error(t, err)
}

func TestCSVSummaryDoesNotMutateSources(t *testing.T) {
	source := summary.NewSource("=source", true)
	source.ClusterScopeSummary = &summary.Summary{Pass: 2, Fail: 1, Warn: 3, Error: 4, Skip: 5}
	source.NamespaceScopeSummary["\t=namespace"] = &summary.Summary{Pass: 7}
	sources := []summary.Source{*source, *summary.NewSource("a", false)}
	reporter := summary.NewReporter("", "=cluster", "Report")
	expected, err := reporter.EmailReport(sources, "html", "csv")
	require.NoError(t, err)
	rows, err := csv.NewReader(strings.NewReader(string(expected.Attachments[0].Data))).ReadAll()
	require.NoError(t, err)
	require.Equal(t, []string{"'=cluster", "'=source", "cluster", "", "2", "1", "3", "4", "5"}, rows[1])
	require.Equal(t, []string{"'=cluster", "'=source", "namespace", "'\t=namespace", "7", "0", "0", "0", "0"}, rows[2])
	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			got, e := reporter.EmailReport(sources, "html", "csv")
			if e != nil {
				t.Error(e)
				return
			}
			if string(got.Attachments[0].Data) != string(expected.Attachments[0].Data) {
				t.Error("unstable CSV")
			}
		}()
	}
	wg.Wait()
	require.Equal(t, "=source", sources[0].Name)
	require.Equal(t, 7, source.NamespaceScopeSummary["\t=namespace"].Pass)
}

func TestCSVSummaryAdapters(t *testing.T) {
	for _, adapter := range []string{"openreports", "wgpolicy"} {
		t.Run(adapter, func(t *testing.T) {
			first := &wg.PolicyReport{ObjectMeta: metav1.ObjectMeta{Name: "first", Namespace: "ns"}, Summary: wg.PolicyReportSummary{Pass: 2, Fail: 3, Warn: 4, Error: 5, Skip: 6}, Results: []wg.PolicyReportResult{{Source: "source", Result: wg.StatusFail}}}
			second := first.DeepCopy()
			second.Name = "second"
			cluster := &wg.ClusterPolicyReport{ObjectMeta: metav1.ObjectMeta{Name: "cluster"}, Summary: wg.PolicyReportSummary{Pass: 1, Fail: 2, Warn: 3, Error: 4, Skip: 5}, Results: []wg.PolicyReportResult{{Source: "source", Result: wg.StatusFail}}}
			var generator *summary.Generator
			if adapter == "wgpolicy" {
				generator = summary.NewGenerator(nil, fakewg.NewSimpleClientset(first, second, cluster).Wgpolicyk8sV1alpha2(), filter, true)
			} else {
				generator = summary.NewGenerator(fakeor.NewSimpleClientset(first.ToOpenReports(), second.ToOpenReports(), cluster.ToOpenReports()).OpenreportsV1alpha1(), nil, filter, true)
			}
			sources, err := generator.GenerateData(context.Background())
			require.NoError(t, err)
			report, err := summary.NewReporter("", "cluster", "Report").EmailReport(sources, "html", "csv")
			require.NoError(t, err)
			rows, err := csv.NewReader(strings.NewReader(string(report.Attachments[0].Data))).ReadAll()
			require.NoError(t, err)
			require.Equal(t, []string{"cluster", "source", "cluster", "", "1", "2", "3", "4", "5"}, rows[1])
			require.Equal(t, []string{"cluster", "source", "namespace", "ns", "4", "6", "8", "10", "12"}, rows[2])
		})
	}
}
