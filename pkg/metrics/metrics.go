package metrics

import (
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"k8s.io/apimachinery/pkg/watch"
)

// CreateMetricsCallback for PolicyReport watch.Events
func CreateMetricsCallback() report.PolicyReportCallback {
	pCallback := createPolicyReportMetricsCallback()
	cCallback := createClusterPolicyReportMetricsCallback()

	return func(et watch.EventType, pr, opr report.PolicyReport) {
		if pr.Namespace == "" {
			cCallback(et, pr, opr)
		} else {
			pCallback(et, pr, opr)
		}
	}
}
