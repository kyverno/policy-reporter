package metrics

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func RegisterDetailedResultGauge(name string) *prometheus.GaugeVec {
	return promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: "List of all PolicyReport Results",
	}, []string{"namespace", "rule", "policy", "report", "kind", "name", "status", "severity", "category", "source"})
}

func CreateDetailedResultMetricListener(filter *report.ResultFilter, gauge *prometheus.GaugeVec) report.PolicyReportListener {
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

				gauge.With(generateResultLabels(newReport, result)).Set(1)
			}
		case report.Updated:
			for _, result := range oldReport.GetResults() {
				gauge.Delete(generateResultLabels(oldReport, result))
			}

			for _, result := range newReport.GetResults() {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(generateResultLabels(newReport, result)).Set(1)
			}
		case report.Deleted:
			for _, result := range newReport.GetResults() {
				if !filter.Validate(result) {
					continue
				}

				gauge.Delete(generateResultLabels(newReport, result))
			}
		}
	}
}

func generateResultLabels(report v1alpha2.ReportInterface, result v1alpha2.PolicyReportResult) prometheus.Labels {
	labels := prometheus.Labels{
		"namespace": report.GetNamespace(),
		"rule":      result.Rule,
		"policy":    result.Policy,
		"report":    report.GetName(),
		"kind":      "",
		"name":      "",
		"status":    string(result.Result),
		"severity":  string(result.Severity),
		"category":  result.Category,
		"source":    result.Source,
	}

	if result.HasResource() {
		labels["kind"] = result.GetResource().Kind
		labels["name"] = result.GetResource().Name
	}

	return labels
}
