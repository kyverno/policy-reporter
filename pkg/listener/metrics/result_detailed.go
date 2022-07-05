package metrics

import (
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

				gauge.With(generateResultLabels(newReport, result)).Set(1)
			}
		case report.Updated:
			for _, result := range oldReport.Results {
				gauge.Delete(generateResultLabels(oldReport, result))
			}

			for _, result := range newReport.Results {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(generateResultLabels(newReport, result)).Set(1)
			}
		case report.Deleted:
			for _, result := range newReport.Results {
				if !filter.Validate(result) {
					continue
				}

				gauge.Delete(generateResultLabels(newReport, result))
			}
		}
	}
}

func generateResultLabels(report report.PolicyReport, result report.Result) prometheus.Labels {
	labels := prometheus.Labels{
		"namespace": report.Namespace,
		"rule":      result.Rule,
		"policy":    result.Policy,
		"report":    report.Name,
		"kind":      "",
		"name":      "",
		"status":    result.Status,
		"severity":  result.Severity,
		"category":  result.Category,
		"source":    result.Source,
	}

	if result.HasResource() {
		labels["kind"] = result.Resource.Kind
		labels["name"] = result.Resource.Name
	}

	return labels
}
