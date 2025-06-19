package result_test

import (
	"testing"

	corev1 "k8s.io/api/core/v1"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report/result"
)

func TestResource(t *testing.T) {
	t.Run("resource from scope", func(t *testing.T) {
		resource := &corev1.ObjectReference{Name: "test", Kind: "Pod"}

		res := result.Resource(&openreports.ReportAdapter{Report: &v1alpha1.Report{Scope: resource}}, openreports.ORResultAdapter{})

		if res != resource {
			t.Error("expected function to return scope resource")
		}
	})
	t.Run("resource from result", func(t *testing.T) {
		resource := &corev1.ObjectReference{Name: "test", Kind: "Pod"}

		res := result.Resource(&openreports.ReportAdapter{Report: &v1alpha1.Report{}}, openreports.ORResultAdapter{ReportResult: v1alpha1.ReportResult{Subjects: []corev1.ObjectReference{*resource}}})

		if res.Name != resource.Name {
			t.Error("expected function to return result resource")
		}
	})
	t.Run("empty fallback resource", func(t *testing.T) {
		res := result.Resource(&openreports.ReportAdapter{Report: &v1alpha1.Report{}}, openreports.ORResultAdapter{ReportResult: v1alpha1.ReportResult{Subjects: []corev1.ObjectReference{}}})

		if res == nil {
			t.Error("expected function to return empty fallback resource")
		}
	})
}
