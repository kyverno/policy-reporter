package listener

import (
	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var (
	ResultGaugeName        = "policy_report_result"
	ClusterResultGaugeName = "cluster_policy_report_result"
)

// NewMetricsListener for PolicyReport watch.Events
func NewMetricsListener(filter *report.ResultFilter, reportFilter *report.ReportFilter, mode metrics.Mode) report.PolicyReportListener {
	resultListeners := ResultListeners(filter, reportFilter, mode)

	return func(event report.LifecycleEvent) {
		if event.NewPolicyReport.Namespace == "" {
			resultListeners[1](event)
		} else {
			resultListeners[0](event)
		}
	}
}

func ResultListeners(filter *report.ResultFilter, reportFilter *report.ReportFilter, mode metrics.Mode) []report.PolicyReportListener {
	if mode == metrics.Simple {
		return []report.PolicyReportListener{
			metrics.CreateSimpleResultMetricsListener(filter, metrics.RegisterSimpleResultGauge(ResultGaugeName)),
			metrics.CreateSimpleClusterResultMetricsListener(filter, metrics.RegisterSimpleClusterResultGauge(ClusterResultGaugeName)),
		}
	}

	prCallback := metrics.CreateDetailedResultMetricListener(filter, metrics.RegisterDetailedResultGauge(ResultGaugeName))
	pCallback := metrics.CreatePolicyReportMetricsListener(reportFilter)

	crCallback := metrics.CreateDetailedClusterResultMetricListener(filter, metrics.RegisterDetailedClusterResultGauge(ClusterResultGaugeName))
	cCallback := metrics.CreateClusterPolicyReportMetricsListener(reportFilter)

	return []report.PolicyReportListener{
		func(event report.LifecycleEvent) {
			pCallback(event)
			prCallback(event)
		},
		func(event report.LifecycleEvent) {
			cCallback(event)
			crCallback(event)
		},
	}
}
