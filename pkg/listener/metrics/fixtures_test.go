package metrics_test

import (
	"github.com/openreports/reports-api/apis/openreports.io/v1alpha1"
	ioprometheusclient "github.com/prometheus/client_model/go"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/openreports"
)

var preport = &openreports.ReportAdapter{
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

func findMetric(metrics []*ioprometheusclient.MetricFamily, name string) *ioprometheusclient.MetricFamily {
	for _, metric := range metrics {
		if *metric.Name == name {
			return metric
		}
	}

	return nil
}
