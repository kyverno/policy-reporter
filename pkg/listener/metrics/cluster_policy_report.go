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

func CreateClusterPolicyReportMetricsListener(filter *report.ReportFilter) report.PolicyReportListener {
	prometheus.Register(clusterPolicyGauge)

	var newReport report.PolicyReport

	return func(event report.LifecycleEvent) {
		newReport = event.NewPolicyReport
		if !filter.Validate(newReport) {
			return
		}

		switch event.Type {
		case report.Added:
			clusterPolicyGauge.WithLabelValues(newReport.Name, strings.Title(report.Skip)).Set(float64(newReport.Summary.Skip))
			clusterPolicyGauge.WithLabelValues(newReport.Name, strings.Title(report.Pass)).Set(float64(newReport.Summary.Pass))
			clusterPolicyGauge.WithLabelValues(newReport.Name, strings.Title(report.Warn)).Set(float64(newReport.Summary.Warn))
			clusterPolicyGauge.WithLabelValues(newReport.Name, strings.Title(report.Fail)).Set(float64(newReport.Summary.Fail))
			clusterPolicyGauge.WithLabelValues(newReport.Name, strings.Title(report.Error)).Set(float64(newReport.Summary.Error))
		case report.Updated:
			clusterPolicyGauge.WithLabelValues(newReport.Name, strings.Title(report.Skip)).Set(float64(newReport.Summary.Skip))
			clusterPolicyGauge.WithLabelValues(newReport.Name, strings.Title(report.Pass)).Set(float64(newReport.Summary.Pass))
			clusterPolicyGauge.WithLabelValues(newReport.Name, strings.Title(report.Warn)).Set(float64(newReport.Summary.Warn))
			clusterPolicyGauge.WithLabelValues(newReport.Name, strings.Title(report.Fail)).Set(float64(newReport.Summary.Fail))
			clusterPolicyGauge.WithLabelValues(newReport.Name, strings.Title(report.Error)).Set(float64(newReport.Summary.Error))
		case report.Deleted:
			clusterPolicyGauge.DeleteLabelValues(newReport.Name, strings.Title(report.Skip))
			clusterPolicyGauge.DeleteLabelValues(newReport.Name, strings.Title(report.Pass))
			clusterPolicyGauge.DeleteLabelValues(newReport.Name, strings.Title(report.Warn))
			clusterPolicyGauge.DeleteLabelValues(newReport.Name, strings.Title(report.Fail))
			clusterPolicyGauge.DeleteLabelValues(newReport.Name, strings.Title(report.Error))
		}
	}
}
