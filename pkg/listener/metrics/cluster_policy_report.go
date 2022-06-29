package metrics

import (
	"strings"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var clusterPolicyGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "cluster_policy_report_summary",
	Help: "Summary of all ClusterPolicyReports",
}, []string{"name", "status"})

var clusterRuleGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "cluster_policy_report_result",
	Help: "List of all ClusterPolicyReport Results",
}, []string{"rule", "policy", "report", "kind", "name", "status", "severity", "category", "source"})

func CreateClusterPolicyReportMetricsListener(filter *Filter) report.PolicyReportListener {
	prometheus.Register(clusterPolicyGauge)
	prometheus.Register(clusterRuleGauge)

	var newReport report.PolicyReport
	var oldReport report.PolicyReport

	return func(event report.LifecycleEvent) {
		newReport = event.NewPolicyReport
		oldReport = event.OldPolicyReport

		switch event.Type {
		case report.Added:
			resetClusterPolicyGauge(newReport)

			for _, result := range newReport.Results {
				if !filter.Validate(result) {
					continue
				}

				clusterRuleGauge.With(generateClusterResultLabels(newReport, result)).Set(1)
				clusterPolicyGauge.WithLabelValues(newReport.Name, strings.Title(result.Status)).Add(1)
			}
		case report.Updated:
			resetClusterPolicyGauge(newReport)

			for _, result := range oldReport.Results {
				clusterRuleGauge.Delete(generateClusterResultLabels(oldReport, result))
			}

			for _, result := range newReport.Results {
				if !filter.Validate(result) {
					continue
				}

				clusterRuleGauge.With(generateClusterResultLabels(newReport, result)).Set(1)
				clusterPolicyGauge.WithLabelValues(newReport.Name, strings.Title(result.Status)).Add(1)
			}
		case report.Deleted:
			clusterPolicyGauge.DeleteLabelValues(newReport.Name, "Pass")
			clusterPolicyGauge.DeleteLabelValues(newReport.Name, "Fail")
			clusterPolicyGauge.DeleteLabelValues(newReport.Name, "Warn")
			clusterPolicyGauge.DeleteLabelValues(newReport.Name, "Error")
			clusterPolicyGauge.DeleteLabelValues(newReport.Name, "Skip")

			for _, result := range newReport.Results {
				clusterRuleGauge.Delete(generateClusterResultLabels(newReport, result))
			}
		}
	}
}

func generateClusterResultLabels(newReport report.PolicyReport, result report.Result) prometheus.Labels {
	labels := prometheus.Labels{
		"rule":     result.Rule,
		"policy":   result.Policy,
		"report":   newReport.Name,
		"kind":     "",
		"name":     "",
		"status":   result.Status,
		"severity": result.Severity,
		"category": result.Category,
		"source":   result.Source,
	}

	if result.HasResource() {
		labels["kind"] = result.Resource.Kind
		labels["name"] = result.Resource.Name
	}

	return labels
}

func resetClusterPolicyGauge(newReport report.PolicyReport) {
	clusterPolicyGauge.
		WithLabelValues(newReport.Name, "Pass").
		Set(0)
	clusterPolicyGauge.
		WithLabelValues(newReport.Name, "Fail").
		Set(0)
	clusterPolicyGauge.
		WithLabelValues(newReport.Name, "Warn").
		Set(0)
	clusterPolicyGauge.
		WithLabelValues(newReport.Name, "Error").
		Set(0)
	clusterPolicyGauge.
		WithLabelValues(newReport.Name, "Skip").
		Set(0)
}
