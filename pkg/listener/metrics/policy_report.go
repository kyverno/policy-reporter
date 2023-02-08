package metrics

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var policyGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "policy_report_summary",
	Help: "Summary of all PolicyReports",
}, []string{"namespace", "name", "status"})

func CreatePolicyReportMetricsListener(filter *report.ReportFilter) report.PolicyReportListener {
	prometheus.Register(policyGauge)

	var newReport v1alpha2.ReportInterface

	return func(event report.LifecycleEvent) {
		newReport = event.PolicyReport
		if !filter.Validate(newReport) {
			return
		}

		switch event.Type {
		case report.Added:
			policyGauge.WithLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusSkip)).Set(float64(newReport.GetSummary().Skip))
			policyGauge.WithLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusPass)).Set(float64(newReport.GetSummary().Pass))
			policyGauge.WithLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusWarn)).Set(float64(newReport.GetSummary().Warn))
			policyGauge.WithLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusFail)).Set(float64(newReport.GetSummary().Fail))
			policyGauge.WithLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusError)).Set(float64(newReport.GetSummary().Error))
		case report.Updated:
			policyGauge.WithLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusSkip)).Set(float64(newReport.GetSummary().Skip))
			policyGauge.WithLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusPass)).Set(float64(newReport.GetSummary().Pass))
			policyGauge.WithLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusWarn)).Set(float64(newReport.GetSummary().Warn))
			policyGauge.WithLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusFail)).Set(float64(newReport.GetSummary().Fail))
			policyGauge.WithLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusError)).Set(float64(newReport.GetSummary().Error))
		case report.Deleted:
			policyGauge.DeleteLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusSkip))
			policyGauge.DeleteLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusPass))
			policyGauge.DeleteLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusWarn))
			policyGauge.DeleteLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusFail))
			policyGauge.DeleteLabelValues(newReport.GetNamespace(), newReport.GetName(), strings.Title(v1alpha2.StatusError))
		}
	}
}
