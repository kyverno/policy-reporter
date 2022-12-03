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

func Test_CustomResultMetricGeneration(t *testing.T) {
	gauge := metrics.RegisterCustomResultGauge("policy_report_custom_result", []string{"namespace", "policy", "status", "source", "app"})

	report1 := report.PolicyReport{
		ID:                "1",
		Labels:            map[string]string{"app": "policy-reporter"},
		Name:              "polr-test",
		Namespace:         "test",
		Summary:           report.Summary{Pass: 2, Fail: 1},
		CreationTimestamp: time.Now(),
		Results:           []report.Result{result1, result2, result3},
	}

	report2 := report.PolicyReport{
		ID:                "1",
		Labels:            map[string]string{"app": "policy-reporter"},
		Name:              "polr-test",
		Namespace:         "test",
		Summary:           report.Summary{Pass: 0, Fail: 1},
		CreationTimestamp: time.Now(),
		Results:           []report.Result{result1, result3},
	}

	filter := metrics.NewResultFilter(validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{Exclude: []string{"disallow-policy"}}, validate.RuleSets{}, validate.RuleSets{})

	t.Run("Added Metric", func(t *testing.T) {
		handler := metrics.CreateCustomResultMetricsListener(filter, gauge, metrics.CreateLabelGenerator([]string{"namespace", "policy", "status", "source", "label:app"}, []string{"namespace", "policy", "status", "source", "app"}))
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "policy_report_custom_result")
		if results == nil {
			t.Fatalf("Metric not found: policy_report_custom_result")
		}

		metrics := results.GetMetric()
		if err = testCustomResultMetricLabels(metrics[0], result2, 1); err != nil {
			t.Error(err)
		}
		if err = testCustomResultMetricLabels(metrics[1], result1, 1); err != nil {
			t.Error(err)
		}
	})

	t.Run("Modified Metric", func(t *testing.T) {
		gauge.Reset()

		handler := metrics.CreateCustomResultMetricsListener(filter, gauge, metrics.CreateLabelGenerator([]string{"namespace", "policy", "status", "source", "label:app"}, []string{"namespace", "policy", "status", "source", "app"}))
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})
		handler(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: report2, OldPolicyReport: report1})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "policy_report_custom_result")
		if results == nil {
			t.Fatalf("Metric not found: policy_report_custom_result")
		}

		metrics := results.GetMetric()
		if len(metrics) != 1 {
			t.Fatalf("Expected only one metric is left, got %d\n", len(metrics))
		}
		if err = testCustomResultMetricLabels(metrics[0], result1, 1); err != nil {
			t.Error(err)
		}
	})

	t.Run("Deleted Metric", func(t *testing.T) {
		gauge.Reset()

		handler := metrics.CreateCustomResultMetricsListener(filter, gauge, metrics.CreateLabelGenerator([]string{"namespace", "policy", "status", "source", "label:app"}, []string{"namespace", "policy", "status", "source", "app"}))
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})
		handler(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: report2, OldPolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Deleted, NewPolicyReport: report2, OldPolicyReport: report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "policy_report_custom_result")
		if results != nil {
			t.Fatalf(" expected metric policy_report_custom_result no longer exists")
		}
	})

	t.Run("Decrease Metric", func(t *testing.T) {
		gauge.Reset()

		handler := metrics.CreateCustomResultMetricsListener(filter, gauge, metrics.CreateLabelGenerator([]string{"namespace", "policy", "status", "source", "label:app"}, []string{"namespace", "policy", "status", "source", "app"}))
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})
		handler(report.LifecycleEvent{Type: report.Deleted, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "policy_report_custom_result")
		if results == nil {
			t.Fatalf("Metric not found: policy_report_custom_result")
		}

		metrics := results.GetMetric()
		if len(metrics) != 2 {
			t.Fatalf("Expected only one metric is left, got %d\n", len(metrics))
		}
		if err = testCustomResultMetricLabels(metrics[0], result2, 1); err != nil {
			t.Error(err)
		}
		if err = testCustomResultMetricLabels(metrics[1], result1, 1); err != nil {
			t.Error(err)
		}
	})
}

func testCustomResultMetricLabels(metric *ioprometheusclient.Metric, result report.Result, expVal float64) error {
	var index int

	if name := *metric.Label[index].Name; name != "app" {
		return fmt.Errorf("unexpected App Label: %s", name)
	}
	if value := *metric.Label[index].Value; value != "policy-reporter" {
		return fmt.Errorf("unexpected Namespace Label Value: %s", value)
	}

	index++

	if name := *metric.Label[index].Name; name != "namespace" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[index].Value; value != result.Resource.Namespace {
		return fmt.Errorf("unexpected Namespace Label Value: %s", value)
	}

	index++

	if name := *metric.Label[index].Name; name != "policy" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[index].Value; value != result.Policy {
		return fmt.Errorf("unexpected Policy Label Value: %s", value)
	}

	index++

	if name := *metric.Label[index].Name; name != "source" {
		return fmt.Errorf("unexpected Source Label: %s", name)
	}
	if value := *metric.Label[index].Value; value != result.Source {
		return fmt.Errorf("unexpected Source Label Value: %s", value)
	}

	index++

	if name := *metric.Label[index].Name; name != "status" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[index].Value; value != result.Status {
		return fmt.Errorf("unexpected Status Label Value: %s", value)
	}

	if value := metric.Gauge.GetValue(); value != expVal {
		return fmt.Errorf("unexpected Metric Value: %v", value)
	}

	return nil
}
