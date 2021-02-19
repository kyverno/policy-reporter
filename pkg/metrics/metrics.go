package metrics

import (
	"github.com/fjogeleit/policy-reporter/pkg/kubernetes"
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/watch"
)

func GenerateMetrics(client kubernetes.Client) {
	policyGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "policy_report",
	}, []string{"namespace", "name", "status"})

	ruleGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "rule_validation",
	}, []string{"namespace", "rule", "policy", "kind", "name", "status"})

	prometheus.Register(policyGauge)
	prometheus.Register(ruleGauge)

	cache := make(map[string]report.PolicyReport)

	client.WatchPolicyReports(func(s watch.EventType, report report.PolicyReport) {
		switch s {
		case watch.Added:
			updatePolicyGauge(policyGauge, report)

			for _, rule := range report.Results {
				res := rule.Resources[0]
				ruleGauge.WithLabelValues(report.Namespace, rule.Rule, rule.Policy, res.Kind, res.Name, rule.Status).Set(1)
			}

			cache[report.GetIdentifier()] = report
		case watch.Modified:
			updatePolicyGauge(policyGauge, report)

			for _, rule := range cache[report.GetIdentifier()].Results {
				res := rule.Resources[0]
				ruleGauge.WithLabelValues(report.Namespace, rule.Rule, rule.Policy, res.Kind, res.Name, rule.Status).Set(0)
			}

			for _, rule := range report.Results {
				res := rule.Resources[0]
				ruleGauge.WithLabelValues(report.Namespace, rule.Rule, rule.Policy, res.Kind, res.Name, rule.Status).Set(1)
			}
		case watch.Deleted:
			policyGauge.WithLabelValues(report.Namespace, report.Name, "Pass").Set(0)
			policyGauge.WithLabelValues(report.Namespace, report.Name, "Fail").Set(0)
			policyGauge.WithLabelValues(report.Namespace, report.Name, "Warn").Set(0)
			policyGauge.WithLabelValues(report.Namespace, report.Name, "Error").Set(0)
			policyGauge.WithLabelValues(report.Namespace, report.Name, "Skip").Set(0)

			for _, rule := range report.Results {
				res := rule.Resources[0]
				ruleGauge.WithLabelValues(report.Namespace, rule.Rule, rule.Policy, res.Kind, res.Name, rule.Status).Set(0)
			}

			delete(cache, report.GetIdentifier())
		}
	})
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
