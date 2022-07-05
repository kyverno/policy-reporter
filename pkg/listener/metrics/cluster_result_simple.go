package metrics

import (
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func RegisterSimpleClusterResultGauge(name string) *prometheus.GaugeVec {
	return promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: "Gauge of ClusterPolicyReportResults by Policy",
	}, []string{"policy", "status", "severity", "category", "source"})
}

func CreateSimpleClusterResultMetricsListener(filter *report.ResultFilter, gauge *prometheus.GaugeVec) report.PolicyReportListener {
	var newReport report.PolicyReport
	var oldReport report.PolicyReport

	return func(event report.LifecycleEvent) {
		newReport = event.NewPolicyReport
		oldReport = event.OldPolicyReport

		switch event.Type {
		case report.Added:
			for _, result := range newReport.Results {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(generateSimpleClusterResultLabels(result)).Inc()
			}
		case report.Updated:
			for _, result := range oldReport.Results {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(generateSimpleClusterResultLabels(result)).Dec()
			}

			for _, result := range newReport.Results {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(generateSimpleClusterResultLabels(result)).Inc()
			}
		case report.Deleted:
			for _, result := range newReport.Results {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(generateSimpleClusterResultLabels(result)).Dec()
			}
		}
	}
}

func generateSimpleClusterResultLabels(result report.Result) prometheus.Labels {
	labels := prometheus.Labels{
		"policy":   result.Policy,
		"status":   result.Status,
		"severity": result.Severity,
		"category": result.Category,
		"source":   result.Source,
	}

	return labels
}
