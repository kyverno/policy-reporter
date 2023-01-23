package listener_test

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var result1 = v1alpha2.PolicyReportResult{
	ID:       "123",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: v1alpha2.ErrorPriority,
	Result:   v1alpha2.StatusFail,
	Category: "Best Practices",
	Severity: v1alpha2.SeverityHigh,
	Scored:   true,
	Source:   "Kyverno",
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	}},
}

var result2 = v1alpha2.PolicyReportResult{
	ID:       "124",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: v1alpha2.WarningPriority,
	Result:   v1alpha2.StatusPass,
	Category: "Best Practices",
	Scored:   true,
	Source:   "Kyverno",
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188419",
	}},
}

var preport1 = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:              "polr-test",
		Namespace:         "test",
		CreationTimestamp: v1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{result1},
	Summary: v1alpha2.PolicyReportSummary{Fail: 1},
}

var preport2 = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:              "polr-test",
		Namespace:         "test",
		CreationTimestamp: v1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{result1, result2},
	Summary: v1alpha2.PolicyReportSummary{Fail: 1, Pass: 1},
}

var preport3 = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:              "polr-test",
		Namespace:         "test",
		CreationTimestamp: v1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{},
}

var creport = &v1alpha2.ClusterPolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Name:              "cpolr-test",
		CreationTimestamp: v1.Now(),
	},
	Results: []v1alpha2.PolicyReportResult{result1, result2},
}
