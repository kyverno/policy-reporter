package listener_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	ioprometheusclient "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_SimpleMetricsListener(t *testing.T) {
	listener.ResultGaugeName = "policy_report_simple_result"
	listener.ClusterResultGaugeName = "cluster_policy_report_simple_result"

	slistener := listener.NewMetricsListener(&report.ResultFilter{}, &report.ReportFilter{}, metrics.Simple, make([]string, 0))

	t.Run("Add ClusterPolicyReport Metric", func(t *testing.T) {
		slistener(report.LifecycleEvent{Type: report.Added, PolicyReport: creport})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "cluster_policy_report_summary")
		assert.Nil(t, summary, "Metric should not be created: cluster_policy_report_summary")

		result := findMetric(metricFam, "cluster_policy_report_simple_result")
		assert.NotNil(t, result, "Metric not found: cluster_policy_report_simple_result")
	})
	t.Run("Add PolicyReport Metric", func(t *testing.T) {
		slistener(report.LifecycleEvent{Type: report.Added, PolicyReport: preport1})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "policy_report_summary")
		assert.Nil(t, summary, "Metric should not be created: policy_report_summary")

		result := findMetric(metricFam, "policy_report_simple_result")
		assert.NotNil(t, result, "Metric not found: policy_report_simple_result")
	})
}

func Test_CustomMetricsListener(t *testing.T) {
	listener.ResultGaugeName = "policy_report_custom_result"
	listener.ClusterResultGaugeName = "cluster_policy_report_custom_result"
	customFields := []string{"namespace", "policy", "status", "source", "label:app", "property:xyz"}

	slistener := listener.NewMetricsListener(&report.ResultFilter{}, &report.ReportFilter{}, metrics.Custom, customFields)

	t.Run("Add ClusterPolicyReport Metric", func(t *testing.T) {
		slistener(report.LifecycleEvent{Type: report.Added, PolicyReport: creport})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "cluster_policy_report_summary")
		assert.Nil(t, summary, "Metric should not be created: cluster_policy_report_summary")

		result := findMetric(metricFam, "cluster_policy_report_custom_result")
		assert.NotNil(t, result, "Metric not found: cluster_policy_report_custom_result")
	})
	t.Run("Add PolicyReport Metric", func(t *testing.T) {
		slistener(report.LifecycleEvent{Type: report.Added, PolicyReport: preport1})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "policy_report_summary")
		assert.Nil(t, summary, "Metric should not be created: policy_report_summary")

		result := findMetric(metricFam, "policy_report_custom_result")
		assert.NotNil(t, result, "Metric not found: policy_report_custom_result")
	})
}

func Test_MetricsListener(t *testing.T) {
	listener.ResultGaugeName = "policy_report_result"
	listener.ClusterResultGaugeName = "cluster_policy_report_result"

	slistener := listener.NewMetricsListener(&report.ResultFilter{}, &report.ReportFilter{}, metrics.Detailed, make([]string, 0))

	t.Run("Add ClusterPolicyReport Metric", func(t *testing.T) {
		slistener(report.LifecycleEvent{Type: report.Added, PolicyReport: creport})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		result := findMetric(metricFam, "cluster_policy_report_result")
		assert.NotNil(t, result, "Metric not found: cluster_policy_report_result")
	})
	t.Run("Add PolicyReport Metric", func(t *testing.T) {
		slistener(report.LifecycleEvent{Type: report.Added, PolicyReport: preport1})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		result := findMetric(metricFam, "policy_report_result")
		assert.NotNil(t, result, "Metric not found: policy_report_result")
	})
}

func findMetric(metrics []*ioprometheusclient.MetricFamily, name string) *ioprometheusclient.MetricFamily {
	for _, metric := range metrics {
		if *metric.Name == name {
			return metric
		}
	}

	return nil
}
