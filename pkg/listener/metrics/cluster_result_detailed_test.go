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
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

func Test_DetailedClusterResultMetricGeneration(t *testing.T) {
	gauge := metrics.RegisterDetailedClusterResultGauge("cluster_policy_report_result")

	report1 := &openreports.ORReportAdapter{
		Report: &v1alpha1.Report{
			ObjectMeta: v1.ObjectMeta{
				Name:              "polr-test",
				CreationTimestamp: v1.Now(),
			},
			Summary: v1alpha1.ReportSummary{Pass: 1, Fail: 2},
			Results: []v1alpha1.ReportResult{*fixtures.PassResult.ReportResult, *fixtures.FailResultWithoutResource.ReportResult, *fixtures.FailDisallowRuleResult.ReportResult},
		},
	}

	report2 := &openreports.ORReportAdapter{
		Report: &v1alpha1.Report{
			ObjectMeta: v1.ObjectMeta{
				Name:              "polr-test",
				CreationTimestamp: v1.Now(),
			},
			Summary: v1alpha1.ReportSummary{Pass: 0, Fail: 2},
			Results: []v1alpha1.ReportResult{*fixtures.FailResult.ReportResult, *fixtures.FailDisallowRuleResult.ReportResult},
		},
	}

	filter := metrics.NewResultFilter(validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{Exclude: []string{"disallow-policy"}}, validate.RuleSets{}, validate.RuleSets{}, (validate.RuleSets{}))
	handler := metrics.CreateDetailedClusterResultMetricListener(filter, gauge)

	t.Run("Added Metric", func(t *testing.T) {
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		assert.NoError(t, err)

		results := findMetric(metricFam, "cluster_policy_report_result")
		assert.NotNil(t, results, "Metric not found: cluster_policy_report_result")

		metrics := results.GetMetric()
		testClusterResultMetricLabels(t, metrics[0], fixtures.FailResultWithoutResource)
		testClusterResultMetricLabels(t, metrics[1], fixtures.PassResult)
	})

	t.Run("Modified Metric", func(t *testing.T) {
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Updated, PolicyReport: report2})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		assert.NoError(t, err)

		results := findMetric(metricFam, "cluster_policy_report_result")
		assert.NotNil(t, results, "Metric not found: cluster_policy_report_result")

		metrics := results.GetMetric()
		assert.Len(t, metrics, 1, "Expected one metric, the second metric should be deleted")
		testClusterResultMetricLabels(t, metrics[0], fixtures.FailResult)
	})

	t.Run("Deleted Metric", func(t *testing.T) {
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Updated, PolicyReport: report2})
		handler(report.LifecycleEvent{Type: report.Deleted, PolicyReport: report2})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		assert.NoError(t, err)

		results := findMetric(metricFam, "cluster_policy_report_result")
		assert.Nil(t, results, "Metric found: cluster_policy_report_result")
	})
}

func testClusterResultMetricLabels(t *testing.T, metric *ioprometheusclient.Metric, result *openreports.ORResultAdapter) error {
	res := &corev1.ObjectReference{}
	if result.HasResource() {
		res = result.GetResource()
	}

	assert.Equal(t, "category", *metric.Label[0].Name, "unexpected name")
	assert.Equal(t, result.Category, *metric.Label[0].Value, "unexpected value")

	assert.Equal(t, "kind", *metric.Label[1].Name, "unexpected name")
	assert.Equal(t, res.Kind, *metric.Label[1].Value, "unexpected value")

	assert.Equal(t, "name", *metric.Label[2].Name, "unexpected name")
	assert.Equal(t, res.Name, *metric.Label[2].Value, "unexpected value")

	assert.Equal(t, "policy", *metric.Label[3].Name, "unexpected name")
	assert.Equal(t, result.Policy, *metric.Label[3].Value, "unexpected value")

	assert.Equal(t, "report", *metric.Label[4].Name, "unexpected name")

	assert.Equal(t, "rule", *metric.Label[5].Name, "unexpected name")
	assert.Equal(t, result.Rule, *metric.Label[5].Value, "unexpected value")

	assert.Equal(t, "severity", *metric.Label[6].Name, "unexpected name")
	assert.Equal(t, string(result.Severity), *metric.Label[6].Value, "unexpected value")

	assert.Equal(t, "source", *metric.Label[7].Name, "unexpected name")
	assert.Equal(t, result.Source, *metric.Label[7].Value, "unexpected value")

	assert.Equal(t, "status", *metric.Label[8].Name, "unexpected name")
	assert.Equal(t, string(result.Result), *metric.Label[8].Value, "unexpected value")

	assert.Equal(t, float64(1), metric.Gauge.GetValue(), "unexpected metric value")

	return nil
}
