package result

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/openreports"
)

func Resource(p openreports.ReportInterface, r openreports.ResultAdapter) *corev1.ObjectReference {
	if r.HasResource() {
		return r.GetResource()
	} else if p.GetScope() != nil {
		return p.GetScope()
	}

	return &corev1.ObjectReference{}
}
