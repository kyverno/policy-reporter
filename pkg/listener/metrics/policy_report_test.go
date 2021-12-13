package metrics_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	ioprometheusclient "github.com/prometheus/client_model/go"
)

var result1 = &report.Result{
	ID:       "1",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.ErrorPriority,
	Status:   report.Fail,
	Severity: report.High,
	Category: "resources",
	Scored:   true,
	Resource: &report.Resource{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	},
}

var result2 = &report.Result{
	ID:       "2",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "check-requests-and-limits-required",
	Rule:     "check-for-requests-and-limits",
	Priority: report.WarningPriority,
	Status:   report.Pass,
	Category: "resources",
	Scored:   true,
	Resource: &report.Resource{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "535ab69f-1b3c-4bd9-9ba4-274a56188419",
	},
}

var preport = &report.PolicyReport{
	ID:                "1",
	Name:              "polr-test",
	Namespace:         "test",
	Results:           make(map[string]*report.Result),
	Summary:           &report.Summary{},
	CreationTimestamp: time.Now(),
}

func Test_PolicyReportMetricGeneration(t *testing.T) {
	report1 := &report.PolicyReport{
		ID:                "1",
		Name:              "polr-test",
		Namespace:         "test",
		Summary:           &report.Summary{Pass: 1, Fail: 1},
		CreationTimestamp: time.Now(),
		Results: map[string]*report.Result{
			result1.GetIdentifier(): result1,
			result2.GetIdentifier(): result2,
		},
	}

	report2 := &report.PolicyReport{
		ID:                "1",
		Name:              "polr-test",
		Namespace:         "test",
		Summary:           &report.Summary{Pass: 0, Fail: 1},
		CreationTimestamp: time.Now(),
		Results: map[string]*report.Result{
			result1.GetIdentifier(): result1,
		},
	}

	handler := metrics.CreatePolicyReportMetricsListener()

	t.Run("Added Metric", func(t *testing.T) {
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: &report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "policy_report_summary")
		if summary == nil {
			t.Fatalf("Metric not found: policy_report_summary")
		}

		metrics := summary.GetMetric()

		if err = testSummaryMetricLabels(metrics[0], preport, "Error", 0); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[1], preport, "Fail", 1); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[2], preport, "Pass", 1); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[3], preport, "Skip", 0); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[4], preport, "Warn", 0); err != nil {
			t.Error(err)
		}

		results := findMetric(metricFam, "policy_report_result")
		if summary == nil {
			t.Fatalf("Metric not found: policy_report_result")
		}

		metrics = results.GetMetric()
		if err = testResultMetricLabels(metrics[0], result2); err != nil {
			t.Error(err)
		}
		if err = testResultMetricLabels(metrics[1], result1); err != nil {
			t.Error(err)
		}
	})

	t.Run("Modified Metric", func(t *testing.T) {
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: &report.PolicyReport{}})
		handler(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: report2, OldPolicyReport: report1})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "policy_report_summary")
		if summary == nil {
			t.Fatalf("Metric not found: policy_report_summary")
		}

		metrics := summary.GetMetric()

		if err = testSummaryMetricLabels(metrics[0], preport, "Error", 0); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[1], preport, "Fail", 1); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[2], preport, "Pass", 0); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[3], preport, "Skip", 0); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[4], preport, "Warn", 0); err != nil {
			t.Error(err)
		}

		results := findMetric(metricFam, "policy_report_result")
		if summary == nil {
			t.Fatalf("Metric not found: policy_report_result")
		}

		metrics = results.GetMetric()
		if len(metrics) != 1 {
			t.Error("Expected one metric, the second metric should be deleted")
		}
		if err = testResultMetricLabels(metrics[0], result1); err != nil {
			t.Error(err)
		}
	})

	t.Run("Deleted Metric", func(t *testing.T) {
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: &report.PolicyReport{}})
		handler(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: report2, OldPolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Deleted, NewPolicyReport: report2, OldPolicyReport: &report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "policy_report_summary")
		if summary != nil {
			t.Error("policy_report_summary should no longer exist", *summary.Name)
		}

		results := findMetric(metricFam, "policy_report_result")
		if results != nil {
			t.Error("policy_report_result shoud no longer exist", *results.Name)
		}
	})
}

func testSummaryMetricLabels(
	metric *ioprometheusclient.Metric,
	preport *report.PolicyReport,
	status string,
	gauge float64,
) error {
	if name := *metric.Label[0].Name; name != "name" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[0].Value; value != preport.Name {
		return fmt.Errorf("unexpected Name Label Value: %s", value)
	}

	if name := *metric.Label[1].Name; name != "namespace" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[1].Value; value != preport.Namespace {
		return fmt.Errorf("unexpected Namespace Label Value: %s", value)
	}

	if name := *metric.Label[2].Name; name != "status" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[2].Value; value != status {
		return fmt.Errorf("unexpected Status Label Value: %s", value)
	}

	if value := metric.Gauge.GetValue(); value != gauge {
		return fmt.Errorf("unexpected Metric Value: %v", value)
	}

	return nil
}

func testResultMetricLabels(metric *ioprometheusclient.Metric, result *report.Result) error {
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

	if name := *metric.Label[8].Name; name != "status" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[8].Value; value != result.Status {
		return fmt.Errorf("unexpected Status Label Value: %s", value)
	}

	if value := metric.Gauge.GetValue(); value != 1 {
		return fmt.Errorf("unexpected Metric Value: %v", value)
	}

	return nil
}

func findMetric(metrics []*ioprometheusclient.MetricFamily, name string) *ioprometheusclient.MetricFamily {
	for _, metric := range metrics {
		if *metric.Name == name {
			return metric
		}
	}

	return nil
}
