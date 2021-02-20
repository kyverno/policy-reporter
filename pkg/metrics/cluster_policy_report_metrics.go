package metrics

import (
	"sync"

	"github.com/fjogeleit/policy-reporter/pkg/kubernetes"
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"k8s.io/apimachinery/pkg/watch"
)

type ClusterPolicyReportMetrics struct {
	client  kubernetes.Client
	cache   map[string]report.ClusterPolicyReport
	rwmutex *sync.RWMutex
}

func (m ClusterPolicyReportMetrics) getCachedReport(i string) report.ClusterPolicyReport {
	m.rwmutex.RLock()
	defer m.rwmutex.RUnlock()
	return m.cache[i]
}

func (m ClusterPolicyReportMetrics) cachedReport(r report.ClusterPolicyReport) {
	m.rwmutex.Lock()
	m.cache[r.GetIdentifier()] = r
	m.rwmutex.Unlock()
}

func (m ClusterPolicyReportMetrics) removeCachedReport(i string) {
	m.rwmutex.Lock()
	delete(m.cache, i)
	m.rwmutex.Unlock()
}

func (m ClusterPolicyReportMetrics) GenerateMetrics() {
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

	m.client.WatchClusterPolicyReports(func(e watch.EventType, r report.ClusterPolicyReport) {
		go func(event watch.EventType, report report.ClusterPolicyReport) {
			switch event {
			case watch.Added:
				updateClusterPolicyGauge(policyGauge, report)

				for _, rule := range report.Results {
					res := rule.Resources[0]
					ruleGauge.WithLabelValues(rule.Rule, rule.Policy, report.Name, res.Kind, res.Name, rule.Status).Set(1)
				}

				m.cachedReport(report)
			case watch.Modified:
				updateClusterPolicyGauge(policyGauge, report)

				for _, rule := range m.getCachedReport(report.GetIdentifier()).Results {
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
						Set(0)
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
				policyGauge.WithLabelValues(report.Name, "Pass").Set(0)
				policyGauge.WithLabelValues(report.Name, "Fail").Set(0)
				policyGauge.WithLabelValues(report.Name, "Warn").Set(0)
				policyGauge.WithLabelValues(report.Name, "Error").Set(0)
				policyGauge.WithLabelValues(report.Name, "Skip").Set(0)

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
						Set(0)
				}

				m.removeCachedReport(report.GetIdentifier())
			}
		}(e, r)
	})
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

func NewClusterPolicyMetrics(client kubernetes.Client) *ClusterPolicyReportMetrics {
	return &ClusterPolicyReportMetrics{
		client:  client,
		cache:   make(map[string]report.ClusterPolicyReport),
		rwmutex: new(sync.RWMutex),
	}
}
