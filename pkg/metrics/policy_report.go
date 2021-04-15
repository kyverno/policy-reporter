package metrics

import (
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/watch"
)

// CreatePolicyMetricsCallback for PolicyReport watch.Events
func CreatePolicyReportMetricsCallback() report.PolicyReportCallback {
	policyGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "policy_report_summary",
		Help: "Summary of all PolicyReports",
	}, []string{"namespace", "name", "status"})

	ruleGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "policy_report_result",
		Help: "List of all PolicyReport Results",
	}, []string{"namespace", "rule", "policy", "report", "kind", "name", "status", "severity", "category"})

	prometheus.Register(policyGauge)
	prometheus.Register(ruleGauge)

	return func(event watch.EventType, report report.PolicyReport, oldReport report.PolicyReport) {
		switch event {
		case watch.Added:
			updatePolicyGauge(policyGauge, report)

			for _, rule := range report.Results {
				ruleGauge.With(generateResultLabels(report, rule)).Set(1)
			}
		case watch.Modified:
			updatePolicyGauge(policyGauge, report)

			for _, rule := range oldReport.Results {
				ruleGauge.Delete(generateResultLabels(oldReport, rule))
			}

			for _, rule := range report.Results {
				ruleGauge.With(generateResultLabels(report, rule)).Set(1)
			}
		case watch.Deleted:
			policyGauge.DeleteLabelValues(report.Namespace, report.Name, "Pass")
			policyGauge.DeleteLabelValues(report.Namespace, report.Name, "Fail")
			policyGauge.DeleteLabelValues(report.Namespace, report.Name, "Warn")
			policyGauge.DeleteLabelValues(report.Namespace, report.Name, "Error")
			policyGauge.DeleteLabelValues(report.Namespace, report.Name, "Skip")

			for _, rule := range report.Results {
				ruleGauge.Delete(generateResultLabels(report, rule))
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
	}

	if len(result.Resources) > 0 {
		res := result.Resources[0]

		labels["kind"] = res.Kind
		labels["name"] = res.Name
	}

	return labels
}

func updatePolicyGauge(policyGauge *prometheus.GaugeVec, report report.PolicyReport) {
	policyGauge.
		WithLabelValues(report.Namespace, report.Name, "Pass").
		Set(float64(report.Summary.Pass))
	policyGauge.
		WithLabelValues(report.Namespace, report.Name, "Fail").
		Set(float64(report.Summary.Fail))
	policyGauge.
		WithLabelValues(report.Namespace, report.Name, "Warn").
		Set(float64(report.Summary.Warn))
	policyGauge.
		WithLabelValues(report.Namespace, report.Name, "Error").
		Set(float64(report.Summary.Error))
	policyGauge.
		WithLabelValues(report.Namespace, report.Name, "Skip").
		Set(float64(report.Summary.Skip))
}
