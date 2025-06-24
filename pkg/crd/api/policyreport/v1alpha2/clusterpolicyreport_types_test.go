package v1alpha2_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

func TestClusterPolicyReport(t *testing.T) {
	t.Run("GetSource Fallback", func(t *testing.T) {
		cpolr := &v1alpha2.ClusterPolicyReport{}

		if s := cpolr.GetSource(); s != "" {
			t.Errorf("expected empty source, got: %s", s)
		}
	})

	t.Run("GetSource From Result", func(t *testing.T) {
		cpolr := &v1alpha2.ClusterPolicyReport{Results: []v1alpha2.PolicyReportResult{{Source: "Kyverno"}}}

		if s := cpolr.GetSource(); s != "Kyverno" {
			t.Errorf("expected 'Kyverno' as source, got: %s", s)
		}
	})
}
