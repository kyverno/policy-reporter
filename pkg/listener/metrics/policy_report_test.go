package metrics_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"
	"github.com/prometheus/client_golang/prometheus"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_PolicyReportMetricGeneration(t *testing.T) {
	report1 := &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Summary: v1alpha2.PolicyReportSummary{Pass: 2, Fail: 1},
	}

	report2 := &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Summary: v1alpha2.PolicyReportSummary{Pass: 3, Fail: 4},
	}

	report3 := &v1alpha2.PolicyReport{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-dev",
			Namespace:         "dev",
			CreationTimestamp: v1.Now(),
		},
		Summary: v1alpha2.PolicyReportSummary{Pass: 0, Fail: 1, Warn: 3},
	}

	filter := metrics.NewReportFilter(validate.RuleSets{Exclude: []string{"dev"}}, validate.RuleSets{Exclude: []string{"Test"}})

	t.Run("Added Metric", func(t *testing.T) {
		handler := metrics.CreatePolicyReportMetricsListener(filter)
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: nil})

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
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: nil})
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
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report1, OldPolicyReport: nil})
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
		handler(report.LifecycleEvent{Type: report.Added, NewPolicyReport: report3, OldPolicyReport: nil})

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
