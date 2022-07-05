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

func Test_SimpleClusterResultMetricGeneration(t *testing.T) {
	gauge := metrics.RegisterSimpleClusterResultGauge("cluster_policy_report_simple_result")

	report1 := report.PolicyReport{
		ID:                "1",
		Name:              "polr-test",
		Summary:           report.Summary{Pass: 2, Fail: 1},
		CreationTimestamp: time.Now(),
		Results:           []report.Result{result1, result2, result3},
	}

	report2 := report.PolicyReport{
		ID:                "1",
		Name:              "polr-test",
		Summary:           report.Summary{Pass: 0, Fail: 1},
		CreationTimestamp: time.Now(),
		Results:           []report.Result{result1, result3},
	}

	filter := metrics.NewResultFilter(validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{Exclude: []string{"disallow-policy"}}, validate.RuleSets{}, validate.RuleSets{})

	t.Run("Added Metric", func(t *testing.T) {
		handler := metrics.CreateSimpleClusterResultMetricsListener(filter, gauge)
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "cluster_policy_report_simple_result")
		if results == nil {
			t.Fatalf("Metric not found: cluster_policy_report_simple_result")
		}

		metrics := results.GetMetric()
		if err = testSimpleClusterResultMetricLabels(metrics[0], result2, 1); err != nil {
			t.Error(err)
		}
		if err = testSimpleClusterResultMetricLabels(metrics[1], result1, 1); err != nil {
			t.Error(err)
		}
	})

	t.Run("Modified Metric", func(t *testing.T) {
		gauge.Reset()

		handler := metrics.CreateSimpleClusterResultMetricsListener(filter, gauge)
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})
		handler(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: report2, OldPolicyReport: report1})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "cluster_policy_report_simple_result")
		if results == nil {
			t.Fatalf("Metric not found: cluster_policy_report_simple_result")
		}

		metrics := results.GetMetric()
		if err = testSimpleClusterResultMetricLabels(metrics[0], result2, 0); err != nil {
			t.Error(err)
		}
		if err = testSimpleClusterResultMetricLabels(metrics[1], result1, 1); err != nil {
			t.Error(err)
		}
	})

	t.Run("Deleted Metric", func(t *testing.T) {
		gauge.Reset()

		handler := metrics.CreateSimpleClusterResultMetricsListener(filter, gauge)
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})
		handler(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: report2, OldPolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Deleted, NewPolicyReport: report2, OldPolicyReport: report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "cluster_policy_report_simple_result")
		if results == nil {
			t.Fatalf("Metric not found: cluster_policy_report_simple_result")
		}

		metrics := results.GetMetric()
		if err = testSimpleClusterResultMetricLabels(metrics[0], result2, 0); err != nil {
			t.Error(err)
		}
		if err = testSimpleClusterResultMetricLabels(metrics[1], result1, 0); err != nil {
			t.Error(err)
		}
	})
}

func testSimpleClusterResultMetricLabels(metric *ioprometheusclient.Metric, result report.Result, expV float64) error {
	if name := *metric.Label[0].Name; name != "category" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[0].Value; value != result.Category {
		return fmt.Errorf("unexpected Category Label Value: %s", value)
	}

	if name := *metric.Label[1].Name; name != "policy" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[1].Value; value != result.Policy {
		return fmt.Errorf("unexpected Policy Label Value: %s", value)
	}

	if name := *metric.Label[2].Name; name != "severity" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[2].Value; value != result.Severity {
		return fmt.Errorf("unexpected Severity Label Value: %s", value)
	}

	if name := *metric.Label[3].Name; name != "source" {
		return fmt.Errorf("unexpected Source Label: %s", name)
	}
	if value := *metric.Label[3].Value; value != result.Source {
		return fmt.Errorf("unexpected Source Label Value: %s", value)
	}

	if name := *metric.Label[4].Name; name != "status" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[4].Value; value != result.Status {
		return fmt.Errorf("unexpected Status Label Value: %s", value)
	}
	if value := metric.Gauge.GetValue(); value != expV {
		return fmt.Errorf("unexpected Metric Value: %v", value)
	}

	return nil
}
