package openreports_test

import (
	"testing"

	"github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

func TestClusterReport(t *testing.T) {
	t.Run("GetSource Fallback", func(t *testing.T) {
		report := openreports.ClusterReportAdapter{ClusterReport: &v1alpha1.ClusterReport{}}

		if s := report.GetSource(); s != "" {
			t.Errorf("expected empty source, got: %s", s)
		}
	})

	t.Run("GetSource From Result", func(t *testing.T) {
		cpolr := openreports.ClusterReportAdapter{ClusterReport: &v1alpha1.ClusterReport{Results: []v1alpha1.ReportResult{{Source: "Kyverno"}}}}

		if s := cpolr.GetSource(); s != "Kyverno" {
			t.Errorf("expected 'Kyverno' as source, got: %s", s)
		}
	})

	t.Run("GetSource from root property", func(t *testing.T) {
		report := openreports.ClusterReportAdapter{ClusterReport: &v1alpha1.ClusterReport{Source: "Kyverno"}}

		if s := report.GetSource(); s != "Kyverno" {
			t.Errorf("expected 'Kyverno' as source, got: %s", s)
		}
	})

	t.Run("GetID", func(t *testing.T) {
		report := openreports.ClusterReportAdapter{ClusterReport: &v1alpha1.ClusterReport{ObjectMeta: v1.ObjectMeta{Name: "cpolr-pod-nginx"}}}

		if s := report.GetID(); s != "10821080135567234638" {
			t.Errorf("unexpected ID, expected '10821080135567234638', got: %s", s)
		}
	})

	t.Run("GetKinds from Scope", func(t *testing.T) {
		report := openreports.ClusterReportAdapter{ClusterReport: &v1alpha1.ClusterReport{Scope: &corev1.ObjectReference{Kind: "ClusterRole"}}}

		if len(report.GetKinds()) != 1 && report.GetKinds()[0] != "ClusterRole" {
			t.Errorf("expected ClusterRole, got: %s", report.GetKinds()[0])
		}
	})

	t.Run("GetKinds from Results", func(t *testing.T) {
		report := openreports.ClusterReportAdapter{ClusterReport: &v1alpha1.ClusterReport{
			Results: []v1alpha1.ReportResult{
				{},
				{Subjects: []corev1.ObjectReference{{Kind: "ClusterRole"}}},
				{Subjects: []corev1.ObjectReference{{Kind: "ClusterRole"}}},
				{Subjects: []corev1.ObjectReference{{Kind: "Namespace"}}},
			},
		}}

		if len(report.GetKinds()) != 2 && report.GetKinds()[1] != "Namespace" {
			t.Errorf("expected Namespace, got: %s", report.GetKinds()[1])
		}
	})

	t.Run("GetSeverities from Results", func(t *testing.T) {
		report := openreports.ClusterReportAdapter{ClusterReport: &v1alpha1.ClusterReport{
			Results: []v1alpha1.ReportResult{
				{Severity: v1alpha2.SeverityHigh},
				{Severity: v1alpha2.SeverityHigh},
				{Severity: v1alpha2.SeverityCritical},
			},
		}}

		if len(report.GetSeverities()) != 2 && report.GetSeverities()[1] != "critical" {
			t.Errorf("expected critical severity, got: %s", report.GetSeverities()[1])
		}
	})
	t.Run("Results", func(t *testing.T) {
		report := openreports.ClusterReportAdapter{ClusterReport: &v1alpha1.ClusterReport{}}

		if s := len(report.GetResults()); s != 0 {
			t.Errorf("expected empty results, got: %d", s)
		}

		report.SetResults([]openreports.ResultAdapter{
			{ReportResult: v1alpha1.ReportResult{Policy: "require-label", Result: v1alpha2.StatusPass}},
		})

		if s := len(report.GetResults()); s != 1 {
			t.Errorf("expected 1 result, got: %d", s)
		}
	})
	t.Run("Summary", func(t *testing.T) {
		report := openreports.ClusterReportAdapter{ClusterReport: &v1alpha1.ClusterReport{Summary: v1alpha1.ReportSummary{Pass: 1}}}

		if s := report.GetSummary().Pass; s != 1 {
			t.Errorf("expected 1 pass result, got: %d", s)
		}
	})
}
