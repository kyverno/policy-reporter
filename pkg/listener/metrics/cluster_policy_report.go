package metrics

import (
	"strings"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var clusterPolicyGauge = promauto.NewGaugeVec(prometheus.GaugeOpts{
	Name: "cluster_policy_report_summary",
	Help: "Summary of all ClusterPolicyReports",
}, []string{"name", "status"})

func CreateClusterPolicyReportMetricsListener(filter *report.ReportFilter) report.PolicyReportListener {
	prometheus.Register(clusterPolicyGauge)

	var newReport v1alpha2.ReportInterface

	return func(event report.LifecycleEvent) {
		newReport = event.PolicyReport
		if !filter.Validate(newReport) {
			return
		}

		switch event.Type {
		case report.Added:
			clusterPolicyGauge.WithLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusSkip)).Set(float64(newReport.GetSummary().Skip))
			clusterPolicyGauge.WithLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusPass)).Set(float64(newReport.GetSummary().Pass))
			clusterPolicyGauge.WithLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusWarn)).Set(float64(newReport.GetSummary().Warn))
			clusterPolicyGauge.WithLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusFail)).Set(float64(newReport.GetSummary().Fail))
			clusterPolicyGauge.WithLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusError)).Set(float64(newReport.GetSummary().Error))
		case report.Updated:
			clusterPolicyGauge.WithLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusSkip)).Set(float64(newReport.GetSummary().Skip))
			clusterPolicyGauge.WithLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusPass)).Set(float64(newReport.GetSummary().Pass))
			clusterPolicyGauge.WithLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusWarn)).Set(float64(newReport.GetSummary().Warn))
			clusterPolicyGauge.WithLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusFail)).Set(float64(newReport.GetSummary().Fail))
			clusterPolicyGauge.WithLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusError)).Set(float64(newReport.GetSummary().Error))
		case report.Deleted:
			clusterPolicyGauge.DeleteLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusSkip))
			clusterPolicyGauge.DeleteLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusPass))
			clusterPolicyGauge.DeleteLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusWarn))
			clusterPolicyGauge.DeleteLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusFail))
			clusterPolicyGauge.DeleteLabelValues(newReport.GetName(), strings.Title(v1alpha2.StatusError))
		}
	}
}
