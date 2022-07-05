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

func Test_DetailedResultMetricGeneration(t *testing.T) {
	gauge := metrics.RegisterDetailedResultGauge("policy_report_result")

	report1 := report.PolicyReport{
		ID:                "1",
		Name:              "polr-test",
		Namespace:         "test",
		Summary:           report.Summary{Pass: 2, Fail: 1},
		CreationTimestamp: time.Now(),
		Results:           []report.Result{result1, result2, result3},
	}

	report2 := report.PolicyReport{
		ID:                "1",
		Name:              "polr-test",
		Namespace:         "test",
		Summary:           report.Summary{Pass: 0, Fail: 1},
		CreationTimestamp: time.Now(),
		Results:           []report.Result{result1, result3},
	}

	filter := metrics.NewResultFilter(validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{Exclude: []string{"disallow-policy"}}, validate.RuleSets{}, validate.RuleSets{})
	handler := metrics.CreateDetailedResultMetricListener(filter, gauge)

	t.Run("Added Metric", func(t *testing.T) {
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		results := findMetric(metricFam, "policy_report_result")
		if results == nil {
			t.Fatalf("Metric not found: policy_report_result")
		}

		metrics := results.GetMetric()
		if err = testResultMetricLabels(metrics[0], result2); err != nil {
			t.Error(err)
		}
		if err = testResultMetricLabels(metrics[1], result1); err != nil {
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

		results := findMetric(metricFam, "policy_report_result")
		if results == nil {
			t.Fatalf("Metric not found: policy_report_result")
		}

		metrics := results.GetMetric()
		if len(metrics) != 1 {
			t.Error("Expected one metric, the second metric should be deleted")
		}
		if err = testResultMetricLabels(metrics[0], result1); err != nil {
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

		results := findMetric(metricFam, "policy_report_result")
		if results != nil {
			t.Error("policy_report_result shoud no longer exist", *results.Name)
		}
	})
}

func testResultMetricLabels(metric *ioprometheusclient.Metric, result report.Result) error {
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

	if name := *metric.Label[3].Name; name != "namespace" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[3].Value; value != result.Resource.Namespace {
		return fmt.Errorf("unexpected Namespace Label Value: %s", value)
	}

	if name := *metric.Label[4].Name; name != "policy" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[4].Value; value != result.Policy {
		return fmt.Errorf("unexpected Policy Label Value: %s", value)
	}

	if name := *metric.Label[5].Name; name != "report" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}

	if name := *metric.Label[6].Name; name != "rule" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[6].Value; value != result.Rule {
		return fmt.Errorf("unexpected Rule Label Value: %s", value)
	}

	if name := *metric.Label[7].Name; name != "severity" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[7].Value; value != result.Severity {
		return fmt.Errorf("unexpected Severity Label Value: %s", value)
	}

	if name := *metric.Label[8].Name; name != "source" {
		return fmt.Errorf("unexpected Source Label: %s", name)
	}
	if value := *metric.Label[8].Value; value != result.Source {
		return fmt.Errorf("unexpected Source Label Value: %s", value)
	}

	if name := *metric.Label[9].Name; name != "status" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[9].Value; value != result.Status {
		return fmt.Errorf("unexpected Status Label Value: %s", value)
	}

	if value := metric.Gauge.GetValue(); value != 1 {
		return fmt.Errorf("unexpected Metric Value: %v", value)
	}

	return nil
}
