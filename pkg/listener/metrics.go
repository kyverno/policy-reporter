package listener

import (
	"strings"

	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var (
	ResultGaugeName        = "policy_report_result"
	ClusterResultGaugeName = "cluster_policy_report_result"
)

const Metrics = "metric_listener"

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
	labels []string,
) []report.PolicyReportListener {
	if mode == metrics.Simple {
		labels = []string{"namespace", "policy", "status", "severity", "category", "source"}
		clusterLabels := []string{"policy", "status", "severity", "category", "source"}

		return []report.PolicyReportListener{
			metrics.CreateCustomResultMetricsListener(
				filter,
				metrics.RegisterCustomResultGauge(ResultGaugeName, labels),
				metrics.CreateLabelGenerator(labels, labels),
			),
			metrics.CreateCustomResultMetricsListener(
				filter,
				metrics.RegisterCustomResultGauge(ClusterResultGaugeName, clusterLabels),
				metrics.CreateLabelGenerator(clusterLabels, clusterLabels),
			),
		}
	}
	if mode == metrics.Custom {
		clusterLabels := make([]string, 0, len(labels))

		clusterLabelNames := make([]string, 0, len(labels))
		labelNames := make([]string, 0, len(labels))

		for _, label := range labels {
			labelName := label
			if strings.HasPrefix(label, metrics.ReportLabelPrefix) {
				replacer := strings.NewReplacer(".", "_", "/", "_", ":", "_", "-", "_", ";", "_")
				labelName = replacer.Replace(strings.TrimPrefix(label, metrics.ReportLabelPrefix))
			}

			labelNames = append(labelNames, labelName)

			if label == "namespace" {
				continue
			}

			clusterLabels = append(clusterLabels, label)
			clusterLabelNames = append(clusterLabelNames, labelName)
		}

		return []report.PolicyReportListener{
			metrics.CreateCustomResultMetricsListener(
				filter,
				metrics.RegisterCustomResultGauge(ResultGaugeName, labelNames),
				metrics.CreateLabelGenerator(labels, labelNames),
			),
			metrics.CreateCustomResultMetricsListener(
				filter,
				metrics.RegisterCustomResultGauge(ClusterResultGaugeName, clusterLabelNames),
				metrics.CreateLabelGenerator(clusterLabels, clusterLabelNames),
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
