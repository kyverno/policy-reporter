package metrics_test

import (
	"fmt"
	"testing"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"
	"github.com/prometheus/client_golang/prometheus"
	ioprometheusclient "github.com/prometheus/client_model/go"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_DetailedClusterResultMetricGeneration(t *testing.T) {
	gauge := metrics.RegisterDetailedClusterResultGauge("cluster_policy_report_result")

	report1 := &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			CreationTimestamp: v1.Now(),
		},
		Summary: v1alpha2.PolicyReportSummary{Pass: 1, Fail: 2},
		Results: []v1alpha2.PolicyReportResult{fixtures.PassResult, fixtures.FailResultWithoutResource, fixtures.FailDisallowRuleResult},
	}

	report2 := &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			CreationTimestamp: v1.Now(),
		},
		Summary: v1alpha2.PolicyReportSummary{Pass: 0, Fail: 2},
		Results: []v1alpha2.PolicyReportResult{fixtures.FailResult, fixtures.FailDisallowRuleResult},
	}

	filter := metrics.NewResultFilter(validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{Exclude: []string{"disallow-policy"}}, validate.RuleSets{}, validate.RuleSets{})
	handler := metrics.CreateDetailedClusterResultMetricListener(filter, gauge)

	t.Run("Added Metric", func(t *testing.T) {
		handler(report.LifecycleEvent{Type: report.Added, PolicyReport: report1})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "cluster_policy_report_result")
		if results == nil {
			t.Fatalf("Metric not found: cluster_policy_report_result")
		}

		metrics := results.GetMetric()
		if err = testClusterResultMetricLabels(metrics[0], fixtures.FailResultWithoutResource); err != nil {
			t.Error(err)
		}
		if err = testClusterResultMetricLabels(metrics[1], fixtures.PassResult); err != nil {
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

		results := findMetric(metricFam, "cluster_policy_report_result")
		if results == nil {
			t.Fatalf("Metric not found: cluster_policy_report_result")
		}

		metrics := results.GetMetric()
		if len(metrics) != 1 {
			t.Error("Expected one metric, the second metric should be deleted")
		}
		if err = testClusterResultMetricLabels(metrics[0], fixtures.FailResult); err != nil {
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

		results := findMetric(metricFam, "cluster_policy_report_result")
		if results != nil {
			t.Error("cluster_policy_report_result shoud no longer exist", *results.Name)
		}
	})
}

func testClusterResultMetricLabels(metric *ioprometheusclient.Metric, result v1alpha2.PolicyReportResult) error {
	res := &corev1.ObjectReference{}
	if result.HasResource() {
		res = result.GetResource()
	}
	if name := *metric.Label[0].Name; name != "category" {
		return fmt.Errorf("unexpected Category Label: %s", name)
	}
	if value := *metric.Label[0].Value; value != result.Category {
		return fmt.Errorf("unexpected Category Label Value: %s", value)
	}

	if name := *metric.Label[1].Name; name != "kind" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[1].Value; value != res.Kind {
		return fmt.Errorf("unexpected Kind Label Value: %s", value)
	}

	if name := *metric.Label[2].Name; name != "name" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[2].Value; value != res.Name {
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
	if value := *metric.Label[6].Value; value != string(result.Severity) {
		return fmt.Errorf("unexpected Severity Label Value: %s", value)
	}

	if name := *metric.Label[7].Name; name != "source" {
		return fmt.Errorf("unexpected Source Label: %s", name)
	}
	if value := *metric.Label[7].Value; value != result.Source {
		return fmt.Errorf("unexpected Source Label Value: %s", value)
	}

	if name := *metric.Label[8].Name; name != "status" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[8].Value; value != string(result.Result) {
		return fmt.Errorf("unexpected Status Label Value: %s", value)
	}

	if value := metric.Gauge.GetValue(); value != 1 {
		return fmt.Errorf("unexpected Metric Value: %v", value)
	}

	return nil
}
