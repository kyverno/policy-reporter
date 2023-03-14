package fixtures

import (
	corev1 "k8s.io/api/core/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

var PassResult = v1alpha2.PolicyReportResult{
	ID:       "123",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: v1alpha2.WarningPriority,
	Result:   v1alpha2.StatusPass,
	Severity: v1alpha2.SeverityHigh,
	Category: "resources",
	Scored:   true,
	Source:   "Kyverno",
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	}},
	Properties: map[string]string{"xyz": "test"},
}

var PassPodResult = v1alpha2.PolicyReportResult{
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
	Properties: map[string]string{},
}

var TrivyResult = v1alpha2.PolicyReportResult{
	ID:       "124",
	Message:  "validation error",
	Policy:   "policy",
	Rule:     "rule",
	Priority: v1alpha2.WarningPriority,
	Result:   v1alpha2.StatusFail,
	Category: "Best Practices",
	Scored:   true,
	Source:   "Trivy",
}

var FailResult = v1alpha2.PolicyReportResult{
	ID:       "123",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: v1alpha2.WarningPriority,
	Result:   v1alpha2.StatusFail,
	Severity: v1alpha2.SeverityHigh,
	Category: "resources",
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

var FailDisallowRuleResult = v1alpha2.PolicyReportResult{
	ID:       "123",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "disallow-policy",
	Rule:     "disallow-policy",
	Priority: v1alpha2.WarningPriority,
	Result:   v1alpha2.StatusFail,
	Severity: v1alpha2.SeverityHigh,
	Category: "resources",
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

var FailPodResult = v1alpha2.PolicyReportResult{
	ID:       "124",
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: v1alpha2.WarningPriority,
	Result:   v1alpha2.StatusFail,
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

var FailResultWithoutResource = v1alpha2.PolicyReportResult{
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: v1alpha2.WarningPriority,
	Result:   v1alpha2.StatusFail,
	Severity: v1alpha2.SeverityHigh,
	Category: "resources",
	Scored:   true,
	Source:   "Kyverno",
}

var PassNamespaceResult = v1alpha2.PolicyReportResult{
	ID:       "125",
	Message:  "validation error: The label `test` is required. Rule check-for-GetLabels()-on-namespace",
	Policy:   "require-ns-GetLabels()",
	Rule:     "check-for-GetLabels()-on-namespace",
	Priority: v1alpha2.ErrorPriority,
	Result:   v1alpha2.StatusPass,
	Category: "namespaces",
	Severity: v1alpha2.SeverityMedium,
	Scored:   true,
	Source:   "Kyverno",
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Namespace",
		Name:       "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188411",
	}},
}

var FailNamespaceResult = v1alpha2.PolicyReportResult{
	ID:       "126",
	Message:  "validation error: The label `test` is required. Rule check-for-GetLabels()-on-namespace",
	Policy:   "require-ns-GetLabels()",
	Rule:     "check-for-GetLabels()-on-namespace",
	Priority: v1alpha2.WarningPriority,
	Result:   v1alpha2.StatusFail,
	Category: "namespaces",
	Severity: v1alpha2.SeverityHigh,
	Scored:   true,
	Source:   "Kyverno",
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Namespace",
		Name:       "dev",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188412",
	}},
}

var ScopeResult = v1alpha2.PolicyReportResult{
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: v1alpha2.WarningPriority,
	Result:   v1alpha2.StatusFail,
	Severity: v1alpha2.SeverityHigh,
	Category: "resources",
	Scored:   true,
	Source:   "Kyverno",
}
