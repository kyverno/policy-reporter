package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func RegisterDetailedResultGauge(name string) *prometheus.GaugeVec {
	return promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: name,
		Help: "List of all PolicyReport Results",
	}, []string{"namespace", "rule", "policy", "report", "kind", "name", "status", "severity", "category", "source"})
}

func CreateDetailedResultMetricListener(filter *report.ResultFilter, gauge *prometheus.GaugeVec) report.PolicyReportListener {
	cache := NewCache(filter, generateResultLabels)

	return func(event report.LifecycleEvent) {
		newReport := event.PolicyReport

		switch event.Type {
		case report.Added:
			for _, result := range newReport.GetResults() {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(generateResultLabels(newReport, result)).Set(1)
			}

			cache.AddReport(newReport)
		case report.Updated:
			items := cache.GetReportLabels(newReport.GetID())
			for _, item := range items {
				gauge.Delete(item.Labels)
			}

			for _, result := range newReport.GetResults() {
				if !filter.Validate(result) {
					continue
				}

				gauge.With(generateResultLabels(newReport, result)).Set(1)
			}

			cache.AddReport(newReport)
		case report.Deleted:
			items := cache.GetReportLabels(newReport.GetID())
			for _, item := range items {
				gauge.Delete(item.Labels)
			}

			cache.Remove(newReport.GetID())
		}
	}
}

func generateResultLabels(report openreports.ReportInterface, result *openreports.ORResultAdapter) map[string]string {
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
