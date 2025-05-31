package result

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	corev1 "k8s.io/api/core/v1"

	"openreports.io/apis/openreports.io/v1alpha1"
)

func Resource(p v1alpha2.ReportInterface, r v1alpha1.ReportResult) *corev1.ObjectReference {
	if r.HasResource() {
		return r.GetResource()
	} else if p.GetScope() != nil {
		return p.GetScope()
	}

	return &corev1.ObjectReference{}
}
