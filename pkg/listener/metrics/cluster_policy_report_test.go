package metrics_test

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	ioprometheusclient "github.com/prometheus/client_model/go"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

func Test_ClusterPolicyReportMetricGeneration(t *testing.T) {
	report1 := &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "cpolr-test",
			CreationTimestamp: v1.Now(),
		},
		Summary: v1alpha2.PolicyReportSummary{Pass: 1, Fail: 1},
	}

	report2 := &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "cpolr-test",
			CreationTimestamp: v1.Now(),
		},
		Summary: v1alpha2.PolicyReportSummary{Pass: 0, Fail: 1},
	}

	report3 := &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "cpolr-test",
			CreationTimestamp: v1.Now(),
		},
		Summary: v1alpha2.PolicyReportSummary{Pass: 0, Fail: 1},
		Results: []v1alpha2.PolicyReportResult{{Source: "Kube Bench"}},
	}

	filter := metrics.NewReportFilter(validate.RuleSets{}, validate.RuleSets{Exclude: []string{"Kube Bench"}})
	handler := metrics.CreateClusterPolicyReportMetricsListener(filter)

	t.Run("Added Metric", func(t *testing.T) {
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})

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
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Updated, PolicyReport: report2})

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
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Updated, PolicyReport: report2})
		handler(report.LifecycleEvent{Type: report.Deleted, PolicyReport: report2})

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
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report3})

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
	preport v1alpha2.ReportInterface,
	status string,
	gauge float64,
) error {
	if name := *metric.Label[0].Name; name != "name" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[0].Value; value != preport.GetName() {
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
