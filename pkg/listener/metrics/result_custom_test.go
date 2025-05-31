package metrics_test

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	ioprometheusclient "github.com/prometheus/client_model/go"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

func Test_CustomResultMetricGeneration(t *testing.T) {
	gauge := metrics.RegisterCustomResultGauge("policy_report_custom_result", []string{"namespace", "policy", "status", "source", "app", "xyz"})

	report1 := &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Labels:            map[string]string{"app": "policy-reporter"},
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Summary: v1alpha1.ReportSummary{Pass: 1, Fail: 1},
		Results: []v1alpha1.ReportResult{fixtures.PassResult, fixtures.PassResult, fixtures.FailPodResult, fixtures.FailDisallowRuleResult},
	}

	report2 := &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Labels:            map[string]string{"app": "policy-reporter"},
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Summary: v1alpha1.ReportSummary{Pass: 1, Fail: 1},
		Results: []v1alpha1.ReportResult{fixtures.FailResult, fixtures.FailPodResult, fixtures.FailDisallowRuleResult},
	}

	filter := metrics.NewResultFilter(validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{Exclude: []string{"disallow-policy"}}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{})

	t.Run("Added Metric", func(t *testing.T) {
		handler := metrics.CreateCustomResultMetricsListener(filter, gauge, metrics.CreateLabelGenerator([]string{"namespace", "policy", "status", "source", "label:app", "property:xyz"}, []string{"namespace", "policy", "status", "source", "app", "xyz"}))
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		assert.NoError(t, err)

		results := findMetric(metricFam, "policy_report_custom_result")
		assert.NotNil(t, results, "Metric not found: policy_report_custom_result")

		metrics := results.GetMetric()
		testCustomResultMetricLabels(t, metrics[0], fixtures.FailPodResult, 1)
		testCustomResultMetricLabels(t, metrics[1], fixtures.PassResult, 2)
	})

	t.Run("Modified Metric", func(t *testing.T) {
		gauge.Reset()

		handler := metrics.CreateCustomResultMetricsListener(filter, gauge, metrics.CreateLabelGenerator([]string{"namespace", "policy", "status", "source", "label:app", "property:xyz"}, []string{"namespace", "policy", "status", "source", "app", "xyz"}))
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Updated, PolicyReport: report2})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		assert.NoError(t, err)

		results := findMetric(metricFam, "policy_report_custom_result")
		assert.NotNil(t, results, "Metric not found: policy_report_custom_result")

		metrics := results.GetMetric()
		assert.Len(t, metrics, 1, "Expected only one metric is left")

		testCustomResultMetricLabels(t, metrics[0], fixtures.FailResult, 2)
	})

	t.Run("Deleted Metric", func(t *testing.T) {
		gauge.Reset()

		handler := metrics.CreateCustomResultMetricsListener(filter, gauge, metrics.CreateLabelGenerator([]string{"namespace", "policy", "status", "source", "label:app", "property:xyz"}, []string{"namespace", "policy", "status", "source", "app", "xyz"}))
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Updated, PolicyReport: report2})
		handler(report.LifecycleEvent{Type: report.Deleted, PolicyReport: report2})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		assert.NoError(t, err)

		results := findMetric(metricFam, "policy_report_custom_result")
		assert.Nil(t, results, "Metric found: policy_report_custom_result")
	})

	t.Run("Decrease Metric", func(t *testing.T) {
		gauge.Reset()

		handler := metrics.CreateCustomResultMetricsListener(filter, gauge, metrics.CreateLabelGenerator([]string{"namespace", "policy", "status", "source", "label:app", "property:xyz"}, []string{"namespace", "policy", "status", "source", "app", "xyz"}))
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Deleted, PolicyReport: report1})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		assert.NoError(t, err)

		results := findMetric(metricFam, "policy_report_custom_result")
		assert.NotNil(t, results, "Metric not found: policy_report_custom_result")

		metrics := results.GetMetric()
		assert.Len(t, metrics, 2, "Expected two metric are left")

		testCustomResultMetricLabels(t, metrics[0], fixtures.FailPodResult, 1)
		testCustomResultMetricLabels(t, metrics[1], fixtures.PassResult, 2)
	})
}

func testCustomResultMetricLabels(t *testing.T, metric *ioprometheusclient.Metric, result v1alpha1.ReportResult, expVal float64) error {
	var index int

	res := &corev1.ObjectReference{}
	if result.HasResource() {
		res = result.GetResource()
	}

	assert.Equal(t, "app", *metric.Label[index].Name, "unexpected name")
	assert.Equal(t, "policy-reporter", *metric.Label[index].Value, "unexpected value")

	index++

	assert.Equal(t, "namespace", *metric.Label[index].Name, "unexpected name")
	assert.Equal(t, res.Namespace, *metric.Label[index].Value, "unexpected value")

	index++

	assert.Equal(t, "policy", *metric.Label[index].Name, "unexpected name")
	assert.Equal(t, result.Policy, *metric.Label[index].Value, "unexpected value")

	index++

	assert.Equal(t, "source", *metric.Label[index].Name, "unexpected name")
	assert.Equal(t, result.Source, *metric.Label[index].Value, "unexpected value")

	index++

	assert.Equal(t, "status", *metric.Label[index].Name, "unexpected name")
	assert.Equal(t, string(result.Result), *metric.Label[index].Value, "unexpected value")

	index++

	assert.Equal(t, "xyz", *metric.Label[index].Name, "unexpected name")
	assert.Equal(t, result.Properties["xyz"], *metric.Label[index].Value, "unexpected value")

	assert.Equal(t, expVal, metric.Gauge.GetValue(), "unexpected metric value")

	return nil
}
