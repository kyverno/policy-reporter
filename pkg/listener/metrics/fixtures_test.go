package metrics_test

import (
	"fmt"

	ioprometheusclient "github.com/prometheus/client_model/go"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

var preport = &openreports.ORReportAdapter{
	Report: &v1alpha1.Report{
		ObjectMeta: v1.ObjectMeta{
			Name:              "polr-test",
			Namespace:         "test",
			CreationTimestamp: v1.Now(),
		},
		Results: make([]v1alpha1.ReportResult, 0),
		Summary: v1alpha1.ReportSummary{},
	},
}

func testSummaryMetricLabels(
	metric *ioprometheusclient.Metric,
	preport v1alpha2.ReportInterface,
	status string,
	gauge float64,
) error {
	if name := *metric.Label[0].Name; name != "name" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[0].Value; value != preport.GetName() {
		return fmt.Errorf("unexpected Name Label Value: %s", value)
	}

	if name := *metric.Label[1].Name; name != "namespace" {
		return fmt.Errorf("unexpected Name Label: %s", name)
	}
	if value := *metric.Label[1].Value; value != preport.GetNamespace() {
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

func findMetric(metrics []*ioprometheusclient.MetricFamily, name string) *ioprometheusclient.MetricFamily {
	for _, metric := range metrics {
		if *metric.Name == name {
			return metric
		}
	}

	return nil
}
