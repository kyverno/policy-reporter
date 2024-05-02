package result_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report/result"
	corev1 "k8s.io/api/core/v1"
)

func TestResource(t *testing.T) {
	t.Run("resource from scope", func(t *testing.T) {
		resource := &corev1.ObjectReference{Name: "test", Kind: "Pod"}

		res := result.Resource(&v1alpha2.PolicyReport{Scope: resource}, v1alpha2.PolicyReportResult{})

		if res != resource {
			t.Error("expected function to return scope resource")
		}
	})
	t.Run("resource from result", func(t *testing.T) {
		resource := &corev1.ObjectReference{Name: "test", Kind: "Pod"}

		res := result.Resource(&v1alpha2.PolicyReport{}, v1alpha2.PolicyReportResult{Resources: []corev1.ObjectReference{*resource}})

		if res.Name != resource.Name {
			t.Error("expected function to return result resource")
		}
	})
	t.Run("empty fallback resource", func(t *testing.T) {
		res := result.Resource(&v1alpha2.PolicyReport{}, v1alpha2.PolicyReportResult{Resources: []corev1.ObjectReference{}})

		if res == nil {
			t.Error("expected function to return empty fallback resource")
		}
	})
}
