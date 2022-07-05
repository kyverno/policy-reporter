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

func CreatePolicyReportMetricsListener(filter *report.ReportFilter) report.PolicyReportListener {
	prometheus.Register(policyGauge)

	var newReport report.PolicyReport

	return func(event report.LifecycleEvent) {
		newReport = event.NewPolicyReport
		if !filter.Validate(newReport) {
			return
		}

		switch event.Type {
		case report.Added:
			policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Skip)).Set(float64(newReport.Summary.Skip))
			policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Pass)).Set(float64(newReport.Summary.Pass))
			policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Warn)).Set(float64(newReport.Summary.Warn))
			policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Fail)).Set(float64(newReport.Summary.Fail))
			policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Error)).Set(float64(newReport.Summary.Error))
		case report.Updated:
			policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Skip)).Set(float64(newReport.Summary.Skip))
			policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Pass)).Set(float64(newReport.Summary.Pass))
			policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Warn)).Set(float64(newReport.Summary.Warn))
			policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Fail)).Set(float64(newReport.Summary.Fail))
			policyGauge.WithLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Error)).Set(float64(newReport.Summary.Error))
		case report.Deleted:
			policyGauge.DeleteLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Skip))
			policyGauge.DeleteLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Pass))
			policyGauge.DeleteLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Warn))
			policyGauge.DeleteLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Fail))
			policyGauge.DeleteLabelValues(newReport.Namespace, newReport.Name, strings.Title(report.Error))
		}
	}
}
