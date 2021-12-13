package listener

import (
	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
)

// NewMetricsListener for PolicyReport watch.Events
func NewMetricsListener() report.PolicyReportListener {
	pCallback := metrics.CreatePolicyReportMetricsListener()
	cCallback := metrics.CreateClusterPolicyReportMetricsListener()

	return func(event report.LifecycleEvent) {
		if event.NewPolicyReport.Namespace == "" {
			cCallback(event)
		} else {
			pCallback(event)
		}
	}
}
