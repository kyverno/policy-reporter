package metrics_test

import (
	"fmt"
	"testing"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/metrics"
	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/prometheus/client_golang/prometheus"
	io_prometheus_client "github.com/prometheus/client_model/go"
	"k8s.io/apimachinery/pkg/watch"
)

var result1 = report.Result{
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.ErrorPriority,
	Status:   report.Fail,
	Severity: report.High,
	Category: "resources",
	Scored:   true,
	Resources: []report.Resource{
		{
			APIVersion: "v1",
			Kind:       "Deployment",
			Name:       "nginx",
			Namespace:  "test",
			UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
		},
	},
}

var result2 = report.Result{
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "check-requests-and-limits-required",
	Rule:     "check-for-requests-and-limits",
	Priority: report.WarningPriority,
	Status:   report.Pass,
	Category: "resources",
	Scored:   true,
	Resources: []report.Resource{
		{
			APIVersion: "v1",
			Kind:       "Deployment",
			Name:       "nginx",
			Namespace:  "test",
			UID:        "535ab69f-1b3c-4bd9-9ba4-274a56188419",
		},
	},
}

var preport = report.PolicyReport{
	Name:              "polr-test",
	Namespace:         "test",
	Results:           make(map[string]report.Result, 0),
	Summary:           report.Summary{},
	CreationTimestamp: time.Now(),
}

func Test_PolicyReportMetricGeneration(t *testing.T) {
	report1 := preport
	report1.Summary = report.Summary{Pass: 1, Fail: 1}
	report1.Results = map[string]report.Result{
		result1.GetIdentifier(): result1,
		result2.GetIdentifier(): result2,
	}

	report2 := preport
	report2.Summary = report.Summary{Pass: 0, Fail: 1}
	report2.Results = map[string]report.Result{
		result1.GetIdentifier(): result1,
	}

	handler := metrics.CreatePolicyReportMetricsCallback()

	t.Run("Added Metric", func(t *testing.T) {
		handler(watch.Added, report1, report.PolicyReport{})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("Unexpected Error: %s", err)
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
		handler(watch.Added, report1, report.PolicyReport{})
		handler(watch.Modified, report2, report1)

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("Unexpected Error: %s", err)
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
		handler(watch.Added, report1, report.PolicyReport{})
		handler(watch.Modified, report2, report1)
		handler(watch.Deleted, report2, report2)

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("Unexpected Error: %s", err)
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
	metric *io_prometheus_client.Metric,
	preport report.PolicyReport,
	status string,
	gauge float64,
) error {
	if name := *metric.Label[0].Name; name != "name" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[0].Value; value != preport.Name {
		return fmt.Errorf("Unexpected Name Label Value: %s", value)
	}

	if name := *metric.Label[1].Name; name != "namespace" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[1].Value; value != preport.Namespace {
		return fmt.Errorf("Unexpected Namespace Label Value: %s", value)
	}

	if name := *metric.Label[2].Name; name != "status" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[2].Value; value != status {
		return fmt.Errorf("Unexpected Status Label Value: %s", value)
	}

	if value := metric.Gauge.GetValue(); value != gauge {
		return fmt.Errorf("Unexpected Metric Value: %v", value)
	}

	return nil
}

func testResultMetricLabels(metric *io_prometheus_client.Metric, result report.Result) error {
	if name := *metric.Label[0].Name; name != "kind" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[0].Value; value != result.Resources[0].Kind {
		return fmt.Errorf("Unexpected Kind Label Value: %s", value)
	}

	if name := *metric.Label[1].Name; name != "name" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[1].Value; value != result.Resources[0].Name {
		return fmt.Errorf("Unexpected Name Label Value: %s", value)
	}

	if name := *metric.Label[2].Name; name != "namespace" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[2].Value; value != result.Resources[0].Namespace {
		return fmt.Errorf("Unexpected Namespace Label Value: %s", value)
	}

	if name := *metric.Label[3].Name; name != "policy" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}

	if value := *metric.Label[3].Value; value != result.Policy {
		return fmt.Errorf("Unexpected Policy Label Value: %s", value)
	}

	if name := *metric.Label[4].Name; name != "report" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}

	if name := *metric.Label[5].Name; name != "rule" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[5].Value; value != result.Rule {
		return fmt.Errorf("Unexpected Rule Label Value: %s", value)
	}

	if name := *metric.Label[6].Name; name != "severity" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[6].Value; value != result.Severity {
		return fmt.Errorf("Unexpected Severity Label Value: %s", value)
	}

	if name := *metric.Label[7].Name; name != "status" {
		return fmt.Errorf("Unexpected Name Label: %s", name)
	}
	if value := *metric.Label[7].Value; value != result.Status {
		return fmt.Errorf("Unexpected Status Label Value: %s", value)
	}

	if value := metric.Gauge.GetValue(); value != 1 {
		return fmt.Errorf("Unexpected Metric Value: %v", value)
	}

	return nil
}

func findMetric(metrics []*io_prometheus_client.MetricFamily, name string) *io_prometheus_client.MetricFamily {
	for _, metric := range metrics {
		if *metric.Name == name {
			return metric
		}
	}

	return nil
}
