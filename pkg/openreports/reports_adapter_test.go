package openreports_test

import (
	"testing"

	"github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

func TestReport(t *testing.T) {
	t.Run("GetSource Fallback", func(t *testing.T) {
		creport := openreports.ReportAdapter{Report: &v1alpha1.Report{}}

		if s := creport.GetSource(); s != "" {
			t.Errorf("expected empty source, got: %s", s)
		}
	})

	t.Run("GetSource from Result", func(t *testing.T) {
		creport := openreports.ReportAdapter{Report: &v1alpha1.Report{Results: []v1alpha1.ReportResult{{Source: "Kyverno"}}}}

		if s := creport.GetSource(); s != "Kyverno" {
			t.Errorf("expected 'Kyverno' as source, got: %s", s)
		}
	})

	t.Run("GetSource from root property", func(t *testing.T) {
		creport := openreports.ReportAdapter{Report: &v1alpha1.Report{Source: "Kyverno"}}

		if s := creport.GetSource(); s != "Kyverno" {
			t.Errorf("expected 'Kyverno' as source, got: %s", s)
		}
	})

	t.Run("GetID", func(t *testing.T) {
		creport := openreports.ReportAdapter{Report: &v1alpha1.Report{ObjectMeta: v1.ObjectMeta{Name: "report-pod-nginx", Namespace: "default"}}}

		if s := creport.GetID(); s != "17831693618079313969" {
			t.Errorf("unexpected ID, expected '17831693618079313969', got: %s", s)
		}
	})

	t.Run("GetKinds from Scope", func(t *testing.T) {
		creport := openreports.ReportAdapter{Report: &v1alpha1.Report{Scope: &corev1.ObjectReference{Kind: "Deployment"}}}

		if len(creport.GetKinds()) != 1 && creport.GetKinds()[0] != "Deployment" {
			t.Errorf("expected Deployment, got: %s", creport.GetKinds()[0])
		}
	})

	t.Run("GetKinds from Results", func(t *testing.T) {
		creport := openreports.ReportAdapter{Report: &v1alpha1.Report{
			Results: []v1alpha1.ReportResult{
				{},
				{Subjects: []corev1.ObjectReference{{Kind: "Pod"}}},
				{Subjects: []corev1.ObjectReference{{Kind: "Pod"}}},
				{Subjects: []corev1.ObjectReference{{Kind: "Deployment"}}},
			},
		}}

		if len(creport.GetKinds()) != 2 && creport.GetKinds()[1] != "Deployment" {
			t.Errorf("expected Deployment, got: %s", creport.GetKinds()[1])
		}
	})

	t.Run("GetSeverities from Results", func(t *testing.T) {
		creport := openreports.ReportAdapter{Report: &v1alpha1.Report{
			Results: []v1alpha1.ReportResult{
				{Severity: v1alpha2.SeverityHigh},
				{Severity: v1alpha2.SeverityHigh},
				{Severity: v1alpha2.SeverityCritical},
			},
		}}

		if len(creport.GetSeverities()) != 2 && creport.GetSeverities()[1] != "critical" {
			t.Errorf("expected critical severity, got: %s", creport.GetSeverities()[1])
		}
	})
	t.Run("Results", func(t *testing.T) {
		report := openreports.ReportAdapter{Report: &v1alpha1.Report{}}

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
		report := openreports.ReportAdapter{Report: &v1alpha1.Report{Summary: v1alpha1.ReportSummary{Pass: 1}}}

		if s := report.GetSummary().Pass; s != 1 {
			t.Errorf("expected 1 pass result, got: %d", s)
		}
	})
}
