package v1alpha2_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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

	t.Run("GetID", func(t *testing.T) {
		cpolr := &v1alpha2.ClusterPolicyReport{ObjectMeta: v1.ObjectMeta{Name: "cpolr-namespace-default"}}

		if s := cpolr.GetID(); s != "762077495299640259" {
			t.Errorf("unexpected ID, expected '762077495299640259', got: %s", s)
		}
	})

	t.Run("GetKinds from Scope", func(t *testing.T) {
		cpolr := &v1alpha2.ClusterPolicyReport{Scope: &corev1.ObjectReference{Kind: "Deployment"}}

		if len(cpolr.GetKinds()) != 1 && cpolr.GetKinds()[0] != "Deployment" {
			t.Errorf("expected Deployment, got: %s", cpolr.GetKinds()[0])
		}
	})

	t.Run("GetKinds from Results", func(t *testing.T) {
		cpolr := &v1alpha2.ClusterPolicyReport{Results: []v1alpha2.PolicyReportResult{
			{},
			{Resources: []corev1.ObjectReference{{Kind: "Pod"}}},
			{Resources: []corev1.ObjectReference{{Kind: "Pod"}}},
			{Resources: []corev1.ObjectReference{{Kind: "Deployment"}}},
		}}

		if len(cpolr.GetKinds()) != 2 && cpolr.GetKinds()[1] != "Deployment" {
			t.Errorf("expected Deployment, got: %s", cpolr.GetKinds()[1])
		}
	})

	t.Run("GetSeverities from Results", func(t *testing.T) {
		cpolr := &v1alpha2.ClusterPolicyReport{Results: []v1alpha2.PolicyReportResult{
			{Severity: v1alpha2.SeverityHigh},
			{Severity: v1alpha2.SeverityHigh},
			{Severity: v1alpha2.SeverityCritical},
		}}

		if len(cpolr.GetSeverities()) != 2 && cpolr.GetSeverities()[1] != "critical" {
			t.Errorf("expected critical severity, got: %s", cpolr.GetSeverities()[1])
		}
	})
	t.Run("Results", func(t *testing.T) {
		polr := &v1alpha2.ClusterPolicyReport{}

		if s := len(polr.GetResults()); s != 0 {
			t.Errorf("expected empty results, got: %d", s)
		}

		polr.SetResults([]v1alpha2.PolicyReportResult{
			{Policy: "require-label", Result: v1alpha2.StatusPass},
		})

		if s := len(polr.GetResults()); s != 1 {
			t.Errorf("expected 1 result, got: %d", s)
		}
	})
	t.Run("Summary", func(t *testing.T) {
		polr := &v1alpha2.ClusterPolicyReport{Summary: v1alpha2.PolicyReportSummary{Pass: 1}}

		if s := polr.GetSummary().Pass; s != 1 {
			t.Errorf("expected 1 pass result, got: %d", s)
		}
	})
}
