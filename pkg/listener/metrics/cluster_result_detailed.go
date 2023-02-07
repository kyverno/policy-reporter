package metrics

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func RegisterDetailedClusterResultGauge(name string) *prometheus.GaugeVec {
	return promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: "List of all PolicyReport Results",
	}, []string{"rule", "policy", "report", "kind", "name", "status", "severity", "category", "source"})
}

func CreateDetailedClusterResultMetricListener(filter *report.ResultFilter, gauge *prometheus.GaugeVec) report.PolicyReportListener {
	var newReport v1alpha2.ReportInterface
	var oldReport v1alpha2.ReportInterface

	return func(event report.LifecycleEvent) {
		newReport = event.NewPolicyReport
		oldReport = event.OldPolicyReport

		switch event.Type {
		case report.Added:
			for _, result := range newReport.GetResults() {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(generateClusterResultLabels(newReport, result)).Set(1)
			}
		case report.Updated:
			for _, result := range oldReport.GetResults() {
				gauge.Delete(generateClusterResultLabels(oldReport, result))
			}

			for _, result := range newReport.GetResults() {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(generateClusterResultLabels(newReport, result)).Set(1)
			}
		case report.Deleted:
			for _, result := range newReport.GetResults() {
				if !filter.Validate(result) {
					continue
				}

				gauge.Delete(generateClusterResultLabels(newReport, result))
			}
		}
	}
}

func generateClusterResultLabels(report v1alpha2.ReportInterface, result v1alpha2.PolicyReportResult) prometheus.Labels {
	labels := prometheus.Labels{
		"rule":     result.Rule,
		"policy":   result.Policy,
		"report":   report.GetName(),
		"kind":     "",
		"name":     "",
		"status":   string(result.Result),
		"severity": string(result.Severity),
		"category": result.Category,
		"source":   result.Source,
	}

	if result.HasResource() {
		labels["kind"] = result.GetResource().Kind
		labels["name"] = result.GetResource().Name
	}

	return labels
}
