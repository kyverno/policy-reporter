package report_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_DisableClusterReports(t *testing.T) {
	filter := report.NewFilter(true, make([]string, 0), make([]string, 0))

	if !filter.DisableClusterReports() {
		t.Error("Expected EnableClusterReports to return true as configured")
	}
}
func Test_AllowReport(t *testing.T) {
	t.Run("Allow ClusterReport", func(t *testing.T) {
		filter := report.NewFilter(true, make([]string, 0), []string{"*"})
		if !filter.AllowReport(creport) {
			t.Error("Expected AllowReport returns true if Report is a ClusterPolicyReport without namespace")
		}
	})

	t.Run("Allow Report with matching include Namespace", func(t *testing.T) {
		filter := report.NewFilter(true, []string{"patch", "te*"}, []string{})
		if !filter.AllowReport(preport) {
			t.Error("Expected AllowReport returns true if Report namespace matches include pattern")
		}
	})

	t.Run("Disallow Report with matching exclude Namespace", func(t *testing.T) {
		filter := report.NewFilter(true, []string{}, []string{"patch", "te*"})
		if filter.AllowReport(preport) {
			t.Error("Expected AllowReport returns false if Report namespace matches exclude pattern")
		}
	})

	t.Run("Ignores exclude pattern if include namespaces provided", func(t *testing.T) {
		filter := report.NewFilter(true, []string{"*"}, []string{"te*"})
		if !filter.AllowReport(preport) {
			t.Error("Expected AllowReport returns true because exclude patterns ignored if include patterns provided")
		}
	})

	t.Run("Allow Report when no configuration exists", func(t *testing.T) {
		filter := report.NewFilter(true, []string{}, []string{})
		if !filter.AllowReport(preport) {
			t.Error("Expected AllowReport returns true if no namespace patterns configured")
		}
	})

	t.Run("Disallow Report if no include namespace matches", func(t *testing.T) {
		filter := report.NewFilter(true, []string{"patch", "dev"}, []string{})
		if filter.AllowReport(preport) {
			t.Error("Expected AllowReport returns false if no namespace pattern matches")
		}
	})
}
