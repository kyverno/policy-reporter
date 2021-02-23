package metrics

import (
	"sync"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/watch"
)

type PolicyReportMetrics struct {
	client  report.Client
	cache   map[string]report.PolicyReport
	rwmutex *sync.RWMutex
}

func (m PolicyReportMetrics) getCachedReport(i string) report.PolicyReport {
	m.rwmutex.RLock()
	defer m.rwmutex.RUnlock()
	return m.cache[i]
}

func (m PolicyReportMetrics) cachedReport(r report.PolicyReport) {
	m.rwmutex.Lock()
	m.cache[r.GetIdentifier()] = r
	m.rwmutex.Unlock()
}

func (m PolicyReportMetrics) removeCachedReport(i string) {
	m.rwmutex.Lock()
	delete(m.cache, i)
	m.rwmutex.Unlock()
}

func (m PolicyReportMetrics) GenerateMetrics() error {
	policyGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "policy_report_summary",
		Help: "Summary of all PolicyReports",
	}, []string{"namespace", "name", "status"})

	ruleGauge := promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "policy_report_result",
		Help: "List of all PolicyReport Results",
	}, []string{"namespace", "rule", "policy", "report", "kind", "name", "status"})

	prometheus.Register(policyGauge)
	prometheus.Register(ruleGauge)

	return m.client.WatchPolicyReports(func(e watch.EventType, r report.PolicyReport) {
		go func(event watch.EventType, report report.PolicyReport) {
			switch event {
			case watch.Added:
				updatePolicyGauge(policyGauge, report)

				for _, rule := range report.Results {
					res := rule.Resources[0]
					ruleGauge.
						WithLabelValues(
							report.Namespace,
							rule.Rule,
							rule.Policy,
							report.Name,
							res.Kind,
							res.Name,
							rule.Status,
						).
						Set(1)
				}

				m.cachedReport(report)
			case watch.Modified:
				updatePolicyGauge(policyGauge, report)

				for _, rule := range m.getCachedReport(report.GetIdentifier()).Results {
					res := rule.Resources[0]
					ruleGauge.DeleteLabelValues(
						report.Namespace,
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
							report.Namespace,
							rule.Rule,
							rule.Policy,
							report.Name,
							res.Kind,
							res.Name,
							rule.Status,
						).
						Set(1)
				}

				m.cachedReport(report)
			case watch.Deleted:
				policyGauge.DeleteLabelValues(report.Namespace, report.Name, "Pass")
				policyGauge.DeleteLabelValues(report.Namespace, report.Name, "Fail")
				policyGauge.DeleteLabelValues(report.Namespace, report.Name, "Warn")
				policyGauge.DeleteLabelValues(report.Namespace, report.Name, "Error")
				policyGauge.DeleteLabelValues(report.Namespace, report.Name, "Skip")

				for _, rule := range report.Results {
					res := rule.Resources[0]

					ruleGauge.DeleteLabelValues(
						report.Namespace,
						rule.Rule,
						rule.Policy,
						report.Name,
						res.Kind,
						res.Name,
						rule.Status,
					)
				}

				m.removeCachedReport(report.GetIdentifier())
			}
		}(e, r)
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

func NewPolicyReportMetrics(client report.Client) *PolicyReportMetrics {
	return &PolicyReportMetrics{
		client:  client,
		cache:   make(map[string]report.PolicyReport),
		rwmutex: new(sync.RWMutex),
	}
}
