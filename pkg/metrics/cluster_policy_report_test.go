package metrics_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	ioprometheusclient "github.com/prometheus/client_model/go"
	"k8s.io/apimachinery/pkg/watch"
)

var creport = report.PolicyReport{
	Name:              "cpolr-test",
	Results:           make(map[string]report.Result, 0),
	Summary:           report.Summary{},
	CreationTimestamp: time.Now(),
}

func Test_ClusterPolicyReportMetricGeneration(t *testing.T) {
	report1 := creport
	report1.Summary = report.Summary{Pass: 1, Fail: 1}
	report1.Results = map[string]report.Result{
		result1.GetIdentifier(): result1,
		result2.GetIdentifier(): result2,
	}

	report2 := creport
	report2.Summary = report.Summary{Pass: 0, Fail: 1}
	report2.Results = map[string]report.Result{
		result1.GetIdentifier(): result1,
	}

	handler := metrics.CreateMetricsCallback()

	t.Run("Added Metric", func(t *testing.T) {
		handler(watch.Added, report1, report.PolicyReport{})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "cluster_policy_report_summary")
		if summary == nil {
			t.Fatalf("Metric not found: cluster_policy_report_summary")
		}

		metrics := summary.GetMetric()

		if err = testClusterSummaryMetricLabels(metrics[0], creport, "Error", 0); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[1], creport, "Fail", 1); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[2], creport, "Pass", 1); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[3], creport, "Skip", 0); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[4], creport, "Warn", 0); err != nil {
			t.Error(err)
		}

		results := findMetric(metricFam, "cluster_policy_report_result")
		if summary == nil {
			t.Fatalf("Metric not found: cluster_policy_report_result")
		}

		metrics = results.GetMetric()
		if err = testClusterResultMetricLabels(metrics[0], result2); err != nil {
			t.Error(err)
		}
		if err = testClusterResultMetricLabels(metrics[1], result1); err != nil {
			t.Error(err)
		}
	})

	t.Run("Modified Metric", func(t *testing.T) {
		handler(watch.Added, report1, report.PolicyReport{})
		handler(watch.Modified, report2, report1)

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "cluster_policy_report_summary")
		if summary == nil {
			t.Fatalf("Metric not found: cluster_policy_report_summary")
		}

		metrics := summary.GetMetric()

		if err = testClusterSummaryMetricLabels(metrics[0], creport, "Error", 0); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[1], creport, "Fail", 1); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[2], creport, "Pass", 0); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[3], creport, "Skip", 0); err != nil {
			t.Error(err)
		}
		if err = testClusterSummaryMetricLabels(metrics[4], creport, "Warn", 0); err != nil {
			t.Error(err)
		}

		results := findMetric(metricFam, "cluster_policy_report_result")
		if summary == nil {
			t.Fatalf("Metric not found: cluster_policy_report_result")
		}

		metrics = results.GetMetric()
		if len(metrics) != 1 {
			t.Error("Expected one metric, the second metric should be deleted")
		}
		if err = testClusterResultMetricLabels(metrics[0], result1); err != nil {
			t.Error(err)
		}
	})

	t.Run("Deleted Metric", func(t *testing.T) {
		handler(watch.Added, report1, report.PolicyReport{})
		handler(watch.Modified, report2, report1)
		handler(watch.Deleted, report2, report2)

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

func testClusterResultMetricLabels(metric *ioprometheusclient.Metric, result report.Result) error {
	if name := *metric.Label[0].Name; name != "category" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[0].Value; value != result.Category {
		return fmt.Errorf("unexpected Category Label Value: %s", value)
	}

	if name := *metric.Label[1].Name; name != "kind" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[1].Value; value != result.Resource.Kind {
		return fmt.Errorf("unexpected Kind Label Value: %s", value)
	}

	if name := *metric.Label[2].Name; name != "name" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[2].Value; value != result.Resource.Name {
		return fmt.Errorf("unexpected Name Label Value: %s", value)
	}

	if name := *metric.Label[3].Name; name != "policy" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[3].Value; value != result.Policy {
		return fmt.Errorf("unexpected Policy Label Value: %s", value)
	}

	if name := *metric.Label[4].Name; name != "report" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}

	if name := *metric.Label[5].Name; name != "rule" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[5].Value; value != result.Rule {
		return fmt.Errorf("unexpected Rule Label Value: %s", value)
	}

	if name := *metric.Label[6].Name; name != "severity" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[6].Value; value != result.Severity {
		return fmt.Errorf("unexpected Severity Label Value: %s", value)
	}

	if name := *metric.Label[7].Name; name != "status" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[7].Value; value != result.Status {
		return fmt.Errorf("unexpected Status Label Value: %s", value)
	}

	if value := metric.Gauge.GetValue(); value != 1 {
		return fmt.Errorf("unexpected Metric Value: %v", value)
	}

	return nil
}
