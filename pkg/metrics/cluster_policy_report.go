package metrics

import (
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/watch"
)

// CreateClusterPolicyReportMetricsCallback for ClusterPolicy watch.Events
func CreateClusterPolicyReportMetricsCallback() report.ClusterPolicyReportCallback {
	policyGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_policy_report_summary",
		Help: "Summary of all ClusterPolicyReports",
	}, []string{"name", "status"})

	ruleGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cluster_policy_report_result",
		Help: "List of all ClusterPolicyReport Results",
	}, []string{"rule", "policy", "report", "kind", "name", "status", "severity", "category"})

	prometheus.Register(policyGauge)
	prometheus.Register(ruleGauge)

	return func(event watch.EventType, report report.ClusterPolicyReport, oldReport report.ClusterPolicyReport) {
		switch event {
		case watch.Added:
			updateClusterPolicyGauge(policyGauge, report)

			for _, rule := range report.Results {
				ruleGauge.With(generateClusterResultLabels(report, rule)).Set(1)
			}
		case watch.Modified:
			updateClusterPolicyGauge(policyGauge, report)

			for _, rule := range oldReport.Results {
				ruleGauge.Delete(generateClusterResultLabels(oldReport, rule))
			}

			for _, rule := range report.Results {
				ruleGauge.With(generateClusterResultLabels(report, rule)).Set(1)
			}
		case watch.Deleted:
			policyGauge.DeleteLabelValues(report.Name, "Pass")
			policyGauge.DeleteLabelValues(report.Name, "Fail")
			policyGauge.DeleteLabelValues(report.Name, "Warn")
			policyGauge.DeleteLabelValues(report.Name, "Error")
			policyGauge.DeleteLabelValues(report.Name, "Skip")

			for _, rule := range report.Results {
				ruleGauge.Delete(generateClusterResultLabels(report, rule))
			}
		}
	}
}

func generateClusterResultLabels(report report.ClusterPolicyReport, result report.Result) prometheus.Labels {
	labels := prometheus.Labels{
		"rule":     result.Rule,
		"policy":   result.Policy,
		"report":   report.Name,
		"kind":     "",
		"name":     "",
		"status":   result.Status,
		"severity": result.Severity,
		"category": result.Category,
	}

	if result.HasResource() {
		labels["kind"] = result.Resource.Kind
		labels["name"] = result.Resource.Name
	}

	return labels
}

func updateClusterPolicyGauge(policyGauge *prometheus.GaugeVec, report report.ClusterPolicyReport) {
	policyGauge.
		WithLabelValues(report.Name, "Pass").
		Set(float64(report.Summary.Pass))
	policyGauge.
		WithLabelValues(report.Name, "Fail").
		Set(float64(report.Summary.Fail))
	policyGauge.
		WithLabelValues(report.Name, "Warn").
		Set(float64(report.Summary.Warn))
	policyGauge.
		WithLabelValues(report.Name, "Error").
		Set(float64(report.Summary.Error))
	policyGauge.
		WithLabelValues(report.Name, "Skip").
		Set(float64(report.Summary.Skip))
}
