package summary_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/email/summary"
)

func Test_Source(t *testing.T) {
	source := summary.NewSource("kyverno", true)
	t.Run("Source.ClusterReports", func(t *testing.T) {
		if !source.ClusterReports {
			t.Errorf("Expected Surce.ClusterReports to be true")
		}
	})
	t.Run("Source.AddClusterSummary", func(t *testing.T) {
		source.AddClusterSummary(v1alpha2.PolicyReportSummary{
			Pass:  1,
			Warn:  2,
			Fail:  4,
			Error: 3,
		})

		if source.ClusterScopeSummary.Pass != 1 {
			t.Errorf("Unexpected Pass Summary: %d", source.ClusterScopeSummary.Pass)
		}
		if source.ClusterScopeSummary.Warn != 2 {
			t.Errorf("Unexpected Warn Summary: %d", source.ClusterScopeSummary.Warn)
		}
		if source.ClusterScopeSummary.Fail != 4 {
			t.Errorf("Unexpected Fail Summary: %d", source.ClusterScopeSummary.Fail)
		}
		if source.ClusterScopeSummary.Error != 3 {
			t.Errorf("Unexpected Error Summary: %d", source.ClusterScopeSummary.Error)
		}
	})
	t.Run("Source.AddNamespacedSummary", func(t *testing.T) {
		source.AddNamespacedSummary("test", v1alpha2.PolicyReportSummary{
			Pass:  5,
			Warn:  6,
			Fail:  7,
			Error: 8,
		})

		if source.NamespaceScopeSummary["test"].Pass != 5 {
			t.Errorf("Unexpected Pass Summary: %d", source.ClusterScopeSummary.Pass)
		}
		if source.NamespaceScopeSummary["test"].Warn != 6 {
			t.Errorf("Unexpected Warn Summary: %d", source.ClusterScopeSummary.Warn)
		}
		if source.NamespaceScopeSummary["test"].Fail != 7 {
			t.Errorf("Unexpected Fail Summary: %d", source.ClusterScopeSummary.Fail)
		}
		if source.NamespaceScopeSummary["test"].Error != 8 {
			t.Errorf("Unexpected Error Summary: %d", source.ClusterScopeSummary.Error)
		}

		source.AddNamespacedSummary("test", v1alpha2.PolicyReportSummary{
			Pass:  2,
			Warn:  1,
			Fail:  0,
			Error: 3,
		})

		if source.NamespaceScopeSummary["test"].Pass != 7 {
			t.Errorf("Unexpected Pass Summary: %d", source.ClusterScopeSummary.Pass)
		}
		if source.NamespaceScopeSummary["test"].Warn != 7 {
			t.Errorf("Unexpected Warn Summary: %d", source.ClusterScopeSummary.Warn)
		}
		if source.NamespaceScopeSummary["test"].Fail != 7 {
			t.Errorf("Unexpected Fail Summary: %d", source.ClusterScopeSummary.Fail)
		}
		if source.NamespaceScopeSummary["test"].Error != 11 {
			t.Errorf("Unexpected Error Summary: %d", source.ClusterScopeSummary.Error)
		}
	})
}
