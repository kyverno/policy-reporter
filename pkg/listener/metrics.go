package listener

import (
	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
)

// NewMetricsListener for PolicyReport watch.Events
func NewMetricsListener(filter *metrics.Filter) report.PolicyReportListener {
	pCallback := metrics.CreatePolicyReportMetricsListener(filter)
	cCallback := metrics.CreateClusterPolicyReportMetricsListener(filter)

	return func(event report.LifecycleEvent) {
		if event.NewPolicyReport.Namespace == "" {
			cCallback(event)
		} else {
			pCallback(event)
		}
	}
}
