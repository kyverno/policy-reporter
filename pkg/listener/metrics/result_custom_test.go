package metrics_test

import (
	"fmt"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	ioprometheusclient "github.com/prometheus/client_model/go"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

func Test_CustomResultMetricGeneration(t *testing.T) {
	gauge := metrics.RegisterCustomResultGauge("policy_report_custom_result", []string{"namespace", "policy", "status", "source", "app", "xyz"})

	report1 := &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Labels:            map[string]string{"app": "policy-reporter"},
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Summary: v1alpha2.PolicyReportSummary{Pass: 1, Fail: 1},
		Results: []v1alpha2.PolicyReportResult{fixtures.PassResult, fixtures.PassResult, fixtures.FailPodResult, fixtures.FailDisallowRuleResult},
	}

	report2 := &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Labels:            map[string]string{"app": "policy-reporter"},
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Summary: v1alpha2.PolicyReportSummary{Pass: 1, Fail: 1},
		Results: []v1alpha2.PolicyReportResult{fixtures.FailResult, fixtures.FailPodResult, fixtures.FailDisallowRuleResult},
	}

	filter := metrics.NewResultFilter(validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{Exclude: []string{"disallow-policy"}}, validate.RuleSets{}, validate.RuleSets{})

	t.Run("Added Metric", func(t *testing.T) {
		handler := metrics.CreateCustomResultMetricsListener(filter, gauge, metrics.CreateLabelGenerator([]string{"namespace", "policy", "status", "source", "label:app", "property:xyz"}, []string{"namespace", "policy", "status", "source", "app", "xyz"}))
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "policy_report_custom_result")
		if results == nil {
			t.Fatalf("Metric not found: policy_report_custom_result")
		}

		metrics := results.GetMetric()
		if err = testCustomResultMetricLabels(metrics[0], fixtures.FailPodResult, 1); err != nil {
			t.Error(err)
		}
		if err = testCustomResultMetricLabels(metrics[1], fixtures.PassResult, 2); err != nil {
			t.Error(err)
		}
	})

	t.Run("Modified Metric", func(t *testing.T) {
		gauge.Reset()

		handler := metrics.CreateCustomResultMetricsListener(filter, gauge, metrics.CreateLabelGenerator([]string{"namespace", "policy", "status", "source", "label:app", "property:xyz"}, []string{"namespace", "policy", "status", "source", "app", "xyz"}))
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Updated, PolicyReport: report2})

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
		if err = testCustomResultMetricLabels(metrics[0], fixtures.FailResult, 2); err != nil {
			t.Error(err)
		}
	})

	t.Run("Deleted Metric", func(t *testing.T) {
		gauge.Reset()

		handler := metrics.CreateCustomResultMetricsListener(filter, gauge, metrics.CreateLabelGenerator([]string{"namespace", "policy", "status", "source", "label:app", "property:xyz"}, []string{"namespace", "policy", "status", "source", "app", "xyz"}))
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Updated, PolicyReport: report2})
		handler(report.LifecycleEvent{Type: report.Deleted, PolicyReport: report2})

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

		handler := metrics.CreateCustomResultMetricsListener(filter, gauge, metrics.CreateLabelGenerator([]string{"namespace", "policy", "status", "source", "label:app", "property:xyz"}, []string{"namespace", "policy", "status", "source", "app", "xyz"}))
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Deleted, PolicyReport: report1})

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
		if err = testCustomResultMetricLabels(metrics[0], fixtures.FailPodResult, 1); err != nil {
			t.Error(err)
		}
		if err = testCustomResultMetricLabels(metrics[1], fixtures.PassResult, 2); err != nil {
			t.Error(err)
		}
	})
}

func testCustomResultMetricLabels(metric *ioprometheusclient.Metric, result v1alpha2.PolicyReportResult, expVal float64) error {
	var index int

	res := &corev1.ObjectReference{}
	if result.HasResource() {
		res = result.GetResource()
	}

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
	if value := *metric.Label[index].Value; value != res.Namespace {
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
	if value := *metric.Label[index].Value; value != string(result.Result) {
		return fmt.Errorf("unexpected Status Label Value: %s", value)
	}

	index++

	if name := *metric.Label[index].Name; name != "xyz" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[index].Value; value != result.Properties["xyz"] {
		return fmt.Errorf("unexpected Status Label Value: %s", value)
	}

	if value := metric.Gauge.GetValue(); value != expVal {
		return fmt.Errorf("unexpected Metric Value: %v", value)
	}

	return nil
}
