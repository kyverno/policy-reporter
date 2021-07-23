package metrics

import (
	"github.com/kyverno/policy-reporter/pkg/report"
	"k8s.io/apimachinery/pkg/watch"
)

var (
	pCallback report.PolicyReportCallback
	cCallback report.PolicyReportCallback
)

// CreateMetricsCallback for PolicyReport watch.Events
func CreateMetricsCallback() report.PolicyReportCallback {
	if pCallback == nil {
		pCallback = createPolicyReportMetricsCallback()
	}
	if cCallback == nil {
		cCallback = createClusterPolicyReportMetricsCallback()
	}

	return func(et watch.EventType, pr, opr report.PolicyReport) {
		if pr.Namespace == "" {
			cCallback(et, pr, opr)
		} else {
			pCallback(et, pr, opr)
		}
	}
}
