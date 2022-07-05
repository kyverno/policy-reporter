package metrics_test

import (
	"testing"
	"time"

	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"
	"github.com/prometheus/client_golang/prometheus"
)

func Test_PolicyReportMetricGeneration(t *testing.T) {
	report1 := report.PolicyReport{
		ID:                "1",
		Name:              "polr-test",
		Namespace:         "test",
		Summary:           report.Summary{Pass: 2, Fail: 1},
		CreationTimestamp: time.Now(),
	}

	report2 := report.PolicyReport{
		ID:                "1",
		Name:              "polr-test",
		Namespace:         "test",
		Summary:           report.Summary{Pass: 3, Fail: 4},
		CreationTimestamp: time.Now(),
	}

	report3 := report.PolicyReport{
		ID:                "1",
		Name:              "polr-dev",
		Namespace:         "dev",
		Summary:           report.Summary{Pass: 0, Fail: 1, Warn: 3},
		CreationTimestamp: time.Now(),
	}

	filter := metrics.NewReportFilter(validate.RuleSets{Exclude: []string{"dev"}}, validate.RuleSets{Exclude: []string{"Test"}})

	t.Run("Added Metric", func(t *testing.T) {
		handler := metrics.CreatePolicyReportMetricsListener(filter)
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "policy_report_summary")
		if summary == nil {
			t.Fatalf("Metric not found: policy_report_summary")
		}

		metrics := summary.GetMetric()

		if err = testSummaryMetricLabels(metrics[0], report1, "Error", 0); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[1], report1, "Fail", 1); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[2], report1, "Pass", 2); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[3], report1, "Skip", 0); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[4], report1, "Warn", 0); err != nil {
			t.Error(err)
		}
	})

	t.Run("Modified Metric", func(t *testing.T) {
		handler := metrics.CreatePolicyReportMetricsListener(filter)
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})
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
		if err = testSummaryMetricLabels(metrics[1], preport, "Fail", 4); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[2], preport, "Pass", 3); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[3], preport, "Skip", 0); err != nil {
			t.Error(err)
		}
		if err = testSummaryMetricLabels(metrics[4], preport, "Warn", 0); err != nil {
			t.Error(err)
		}
	})

	t.Run("Deleted Metric", func(t *testing.T) {
		handler := metrics.CreatePolicyReportMetricsListener(filter)
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: report.PolicyReport{}})
		handler(report.LifecycleEvent{Type: report.Updated, NewPolicyReport: report2, OldPolicyReport: report1})
		handler(report.LifecycleEvent{Type: report.Deleted, NewPolicyReport: report2, OldPolicyReport: report2})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "policy_report_summary")
		if summary != nil {
			t.Error("policy_report_summary should no longer exist", *summary.Name)
		}
	})

	t.Run("Validate Metric Filter", func(t *testing.T) {
		handler := metrics.CreatePolicyReportMetricsListener(filter)
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report3, OldPolicyReport: report.PolicyReport{}})

		metricFam, err := prometheus.DefaultGatherer.Gather()
		if err != nil {
			t.Errorf("unexpected Error: %s", err)
		}

		summary := findMetric(metricFam, "policy_report_summary")
		if summary != nil {
			t.Error("policy_report_summary should not created", *summary.Name)
		}
	})
}
