package metrics_test

import (
	"fmt"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	ioprometheusclient "github.com/prometheus/client_model/go"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var result1 = v1alpha2.PolicyReportResult{
	ID:       "1",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: v1alpha2.ErrorPriority,
	Result:   v1alpha2.StatusFail,
	Severity: v1alpha2.SeverityHigh,
	Category: "resources",
	Scored:   true,
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	}},
	Source: "Kyverno",
}

var result2 = v1alpha2.PolicyReportResult{
	ID:       "2",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "check-requests-and-limits-required",
	Rule:     "check-for-requests-and-limits",
	Priority: v1alpha2.WarningPriority,
	Result:   v1alpha2.StatusPass,
	Category: "resources",
	Scored:   true,
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "535ab69f-1b3c-4bd9-9ba4-274a56188419",
	}},
	Source: "Kyverno",
}

var result3 = v1alpha2.PolicyReportResult{
	ID:       "3",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "disallow-policy",
	Rule:     "check-for-requests-and-limits",
	Priority: v1alpha2.WarningPriority,
	Result:   v1alpha2.StatusPass,
	Category: "resources",
	Scored:   true,
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "535ab69f-1b3c-4bd9-9ba4-274a56188419",
	}},
	Source: "Kyverno",
}

var preport = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:              "polr-test",
		Namespace:         "test",
		CreationTimestamp: v1.Now(),
	},
	Results: make([]v1alpha2.PolicyReportResult, 0),
	Summary: v1alpha2.PolicyReportSummary{},
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
