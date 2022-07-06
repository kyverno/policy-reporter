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
func NewMetricsListener(
	filter *report.ResultFilter,
	reportFilter *report.ReportFilter,
	mode metrics.Mode,
	fields []string,
) report.PolicyReportListener {
	resultListeners := ResultListeners(filter, reportFilter, mode, fields)

	return func(event report.LifecycleEvent) {
		if event.NewPolicyReport.Namespace == "" {
			resultListeners[1](event)
		} else {
			resultListeners[0](event)
		}
	}
}

func ResultListeners(
	filter *report.ResultFilter,
	reportFilter *report.ReportFilter,
	mode metrics.Mode,
	fields []string,
) []report.PolicyReportListener {
	if mode == metrics.Simple {
		fields := []string{"namespace", "policy", "status", "severity", "category", "source"}
		clusterFields := []string{"policy", "status", "severity", "category", "source"}

		return []report.PolicyReportListener{
			metrics.CreateCustomResultMetricsListener(
				filter,
				metrics.RegisterCustomResultGauge(ResultGaugeName, fields),
				metrics.CreateLabelGenerator(fields),
			),
			metrics.CreateCustomResultMetricsListener(
				filter,
				metrics.RegisterCustomResultGauge(ClusterResultGaugeName, clusterFields),
				metrics.CreateLabelGenerator(clusterFields),
			),
		}
	}
	if mode == metrics.Custom {
		clusterFields := make([]string, 0, len(fields))
		for _, field := range fields {
			if field == "namespace" {
				continue
			}
			clusterFields = append(clusterFields, field)
		}

		return []report.PolicyReportListener{
			metrics.CreateCustomResultMetricsListener(
				filter,
				metrics.RegisterCustomResultGauge(ResultGaugeName, fields),
				metrics.CreateLabelGenerator(fields),
			),
			metrics.CreateCustomResultMetricsListener(
				filter,
				metrics.RegisterCustomResultGauge(ClusterResultGaugeName, clusterFields),
				metrics.CreateLabelGenerator(clusterFields),
			),
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
