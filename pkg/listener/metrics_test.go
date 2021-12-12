package listener_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/listener"
	"github.com/kyverno/policy-reporter/pkg/report"

	"github.com/prometheus/client_golang/prometheus"
	ioprometheusclient "github.com/prometheus/client_model/go"
)

func Test_MetricsListener(t *testing.T) {
	slistener := listener.NewMetricsListener()

	t.Run("Add ClusterPolicyReport Metric", func(t *testing.T) {
		slistener(report.LifecycleEvent{Type: report.Added, NewPolicyReport: creport, OldPolicyReport: &report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "cluster_policy_report_summary")
		if summary == nil {
			t.Fatalf("Metric not found: cluster_policy_report_summary")
		}
	})
	t.Run("Add PolicyReport Metric", func(t *testing.T) {
		slistener(report.LifecycleEvent{Type: report.Added, NewPolicyReport: preport1, OldPolicyReport: &report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "policy_report_summary")
		if summary == nil {
			t.Fatalf("Metric not found: policy_report_summary")
		}
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
