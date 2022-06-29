package metrics

import (
	"strings"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var policyGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "policy_report_summary",
	Help: "Summary of all PolicyReports",
}, []string{"namespace", "name", "status"})

var ruleGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "policy_report_result",
	Help: "List of all PolicyReport Results",
}, []string{"namespace", "rule", "policy", "report", "kind", "name", "status", "severity", "category", "source"})

func CreatePolicyReportMetricsListener(filter *Filter) report.PolicyReportListener {
	prometheus.Register(policyGauge)
	prometheus.Register(ruleGauge)

	var newReport report.PolicyReport
	var oldReport report.PolicyReport

	return func(event report.LifecycleEvent) {
		newReport = event.NewPolicyReport
		oldReport = event.OldPolicyReport

		switch event.Type {
		case report.Added:
			resetPolicyGauge(newReport)

			for _, result := range newReport.Results {
				if !filter.Validate(result) {
					continue
				}

				ruleGauge.With(generateResultLabels(newReport, result)).Set(1)
				policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, strings.Title(result.Status)).Add(1)
			}
		case report.Updated:
			resetPolicyGauge(newReport)

			for _, result := range oldReport.Results {
				ruleGauge.Delete(generateResultLabels(oldReport, result))
			}

			for _, result := range newReport.Results {
				if !filter.Validate(result) {
					continue
				}

				ruleGauge.With(generateResultLabels(newReport, result)).Set(1)
				policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, strings.Title(result.Status)).Add(1)
			}
		case report.Deleted:
			policyGauge.DeleteLabelValues(newReport.Namespace, newReport.Name, "Pass")
			policyGauge.DeleteLabelValues(newReport.Namespace, newReport.Name, "Fail")
			policyGauge.DeleteLabelValues(newReport.Namespace, newReport.Name, "Warn")
			policyGauge.DeleteLabelValues(newReport.Namespace, newReport.Name, "Error")
			policyGauge.DeleteLabelValues(newReport.Namespace, newReport.Name, "Skip")

			for _, result := range newReport.Results {
				ruleGauge.Delete(generateResultLabels(newReport, result))
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

func resetPolicyGauge(newReport report.PolicyReport) {
	policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, "Pass").Set(0)
	policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, "Fail").Set(0)
	policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, "Warn").Set(0)
	policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, "Error").Set(0)
	policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, "Skip").Set(0)
}
