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
	}, []string{"rule", "policy", "report", "kind", "name", "status"})

	prometheus.Register(policyGauge)
	prometheus.Register(ruleGauge)

	return func(event watch.EventType, report report.ClusterPolicyReport, oldReport report.ClusterPolicyReport) {
		switch event {
		case watch.Added:
			updateClusterPolicyGauge(policyGauge, report)

			for _, rule := range report.Results {
				res := rule.Resources[0]
				ruleGauge.WithLabelValues(rule.Rule, rule.Policy, report.Name, res.Kind, res.Name, rule.Status).Set(1)
			}
		case watch.Modified:
			updateClusterPolicyGauge(policyGauge, report)

			for _, rule := range oldReport.Results {
				res := rule.Resources[0]
				ruleGauge.DeleteLabelValues(
					rule.Rule,
					rule.Policy,
					report.Name,
					res.Kind,
					res.Name,
					rule.Status,
				)
			}

			for _, rule := range report.Results {
				res := rule.Resources[0]
				ruleGauge.
					WithLabelValues(
						rule.Rule,
						rule.Policy,
						report.Name,
						res.Kind,
						res.Name,
						rule.Status,
					).
					Set(1)
			}
		case watch.Deleted:
			policyGauge.DeleteLabelValues(report.Name, "Pass")
			policyGauge.DeleteLabelValues(report.Name, "Fail")
			policyGauge.DeleteLabelValues(report.Name, "Warn")
			policyGauge.DeleteLabelValues(report.Name, "Error")
			policyGauge.DeleteLabelValues(report.Name, "Skip")

			for _, rule := range report.Results {
				res := rule.Resources[0]
				ruleGauge.DeleteLabelValues(
					rule.Rule,
					rule.Policy,
					report.Name,
					res.Kind,
					res.Name,
					rule.Status,
				)
			}
		}
	}
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
