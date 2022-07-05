package metrics_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"
	"github.com/prometheus/client_golang/prometheus"
	ioprometheusclient "github.com/prometheus/client_model/go"
)

func Test_ClusterPolicyReportMetricGeneration(t *testing.T) {
	report1 := report.PolicyReport{
		Name:              "cpolr-test",
		Summary:           report.Summary{Pass: 1, Fail: 1},
		CreationTimestamp: time.Now(),
	}

	report2 := report.PolicyReport{
		Name:              "cpolr-test",
		Summary:           report.Summary{Pass: 0, Fail: 1},
		CreationTimestamp: time.Now(),
	}

	report3 := report.PolicyReport{
		Name:              "cpolr-test",
		Summary:           report.Summary{Pass: 0, Fail: 1},
		CreationTimestamp: time.Now(),
		Results:           []report.Result{{Source: "Kube Bench"}},
	}

	filter := metrics.NewReportFilter(validate.RuleSets{}, validate.RuleSets{Exclude: []string{"Kube Bench"}})
	handler := metrics.CreateClusterPolicyReportMetricsListener(filter)

	t.Run("Added Metric", func(t *testing.T) {
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "cluster_policy_report_summary")
		if summary == nil {
			t.Fatalf("Metric not found: cluster_policy_report_summary")
		}

		metrics := summary.GetMetric()

		if err = testClusterSummaryMetricLabels(metrics[0], report1, "Error", 0); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[1], report1, "Fail", 1); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[2], report1, "Pass", 1); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[3], report1, "Skip", 0); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[4], report1, "Warn", 0); err != nil {
			t.Error(err)
		}
	})

	t.Run("Modified Metric", func(t *testing.T) {
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})
		handler(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: report2, OldPolicyReport: report1})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "cluster_policy_report_summary")
		if summary == nil {
			t.Fatalf("Metric not found: cluster_policy_report_summary")
		}

		metrics := summary.GetMetric()

		if err = testClusterSummaryMetricLabels(metrics[0], report2, "Error", 0); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[1], report2, "Fail", 1); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[2], report2, "Pass", 0); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[3], report2, "Skip", 0); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[4], report2, "Warn", 0); err != nil {
			t.Error(err)
		}
	})

	t.Run("Deleted Metric", func(t *testing.T) {
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})
		handler(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: report2, OldPolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Deleted, NewPolicyReport: report2, OldPolicyReport: report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "cluster_policy_report_summary")
		if summary != nil {
			t.Error("cluster_policy_report_summary should no longer exist", *summary.Name)
		}

		results := metricFam[0]

		if *results.Name == "cluster_policy_report_result" {
			t.Error("cluster_policy_report_result should no longer exist", *results.Name)
		}
	})

	t.Run("Filtered Report", func(t *testing.T) {
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report3, OldPolicyReport: report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "cluster_policy_report_summary")
		if summary != nil {
			t.Error("cluster_policy_report_summary should not be created", *summary.Name)
		}
	})
}

func testClusterSummaryMetricLabels(
	metric *ioprometheusclient.Metric,
	preport report.PolicyReport,
	status string,
	gauge float64,
) error {
	if name := *metric.Label[0].Name; name != "name" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[0].Value; value != preport.Name {
		return fmt.Errorf("unexpected Name Label Value: %s", value)
	}

	if name := *metric.Label[1].Name; name != "status" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[1].Value; value != status {
		return fmt.Errorf("unexpected Status Label Value: %s", value)
	}

	if value := metric.Gauge.GetValue(); value != gauge {
		return fmt.Errorf("unexpected Metric Value: %v", value)
	}

	return nil
}
