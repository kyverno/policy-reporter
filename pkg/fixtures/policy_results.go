package fixtures

import (
	corev1 "k8s.io/api/core/v1"

	"openreports.io/apis/openreports.io/v1alpha1"
)

var PassResult = v1alpha1.ReportResult{
	ID:          "123",
	Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:      "require-requests-and-limits-required",
	Rule:        "autogen-check-for-requests-and-limits",
	Result:      v1alpha1.StatusPass,
	Severity:    v1alpha1.SeverityHigh,
	Category:    "resources",
	Scored:      true,
	Source:      "Kyverno",
	Subjects: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	}},
	Properties: map[string]string{"xyz": "test"},
}

var PassPodResult = v1alpha1.ReportResult{
	ID:          "124",
	Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:      "require-requests-and-limits-required",
	Rule:        "autogen-check-for-requests-and-limits",
	Result:      v1alpha1.StatusPass,
	Category:    "Best Practices",
	Scored:      true,
	Source:      "Kyverno",
	Subjects: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188419",
	}},
	Properties: map[string]string{},
}

var WarnPodResult = v1alpha1.ReportResult{
	ID:          "124",
	Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:      "require-requests-and-limits-required",
	Rule:        "autogen-check-for-requests-and-limits",
	Result:      v1alpha1.StatusWarn,
	Category:    "Best Practices",
	Scored:      true,
	Source:      "Kyverno",
	Subjects: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188419",
	}},
	Properties: map[string]string{},
}

var ErrorPodResult = v1alpha1.ReportResult{
	ID:          "124",
	Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:      "require-requests-and-limits-required",
	Rule:        "autogen-check-for-requests-and-limits",
	Result:      v1alpha1.StatusError,
	Category:    "Best Practices",
	Scored:      true,
	Source:      "Kyverno",
	Subjects: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188419",
	}},
	Properties: map[string]string{},
}

var TrivyResult = v1alpha1.ReportResult{
	ID:          "124",
	Description: "validation error",
	Policy:      "policy",
	Rule:        "rule",
	Result:      v1alpha1.StatusFail,
	Category:    "Best Practices",
	Scored:      true,
	Source:      "Trivy",
}

var FailResult = v1alpha1.ReportResult{
	ID:          "123",
	Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:      "require-requests-and-limits-required",
	Rule:        "autogen-check-for-requests-and-limits",
	Result:      v1alpha1.StatusFail,
	Severity:    v1alpha1.SeverityHigh,
	Category:    "resources",
	Scored:      true,
	Source:      "Kyverno",
	Subjects: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	}},
}

var FailDisallowRuleResult = v1alpha1.ReportResult{
	ID:          "123",
	Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:      "disallow-policy",
	Rule:        "disallow-policy",
	Result:      v1alpha1.StatusFail,
	Severity:    v1alpha1.SeverityHigh,
	Category:    "resources",
	Scored:      true,
	Source:      "Kyverno",
	Subjects: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	}},
}

var FailPodResult = v1alpha1.ReportResult{
	ID:          "124",
	Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:      "require-requests-and-limits-required",
	Rule:        "autogen-check-for-requests-and-limits",
	Result:      v1alpha1.StatusFail,
	Category:    "Best Practices",
	Scored:      true,
	Source:      "Kyverno",
	Subjects: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188419",
	}},
}

var SkipPodResult = v1alpha1.ReportResult{
	ID:       "124",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Result:   v1alpha1.StatusSkip,
	Category: "Best Practices",
	Scored:   true,
	Source:   "Kyverno",
	Subjects: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       "nginx",
		Namespace:  "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188419",
	}},
}

var FailResultWithoutResource = v1alpha1.ReportResult{
	Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:      "require-requests-and-limits-required",
	Rule:        "autogen-check-for-requests-and-limits",
	Result:      v1alpha1.StatusFail,
	Severity:    v1alpha1.SeverityHigh,
	Category:    "resources",
	Scored:      true,
	Source:      "Kyverno",
}

var PassNamespaceResult = v1alpha1.ReportResult{
	ID:          "125",
	Description: "validation error: The label `test` is required. Rule check-for-GetLabels()-on-namespace",
	Policy:      "require-ns-GetLabels()",
	Rule:        "check-for-GetLabels()-on-namespace",
	Result:      v1alpha1.StatusPass,
	Category:    "namespaces",
	Severity:    v1alpha1.SeverityMedium,
	Scored:      true,
	Source:      "Kyverno",
	Subjects: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Namespace",
		Name:       "test",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188411",
	}},
}

var FailNamespaceResult = v1alpha1.ReportResult{
	ID:          "126",
	Description: "validation error: The label `test` is required. Rule check-for-GetLabels()-on-namespace",
	Policy:      "require-ns-GetLabels()",
	Rule:        "check-for-GetLabels()-on-namespace",
	Result:      v1alpha1.StatusFail,
	Category:    "namespaces",
	Severity:    v1alpha1.SeverityHigh,
	Scored:      true,
	Source:      "Kyverno",
	Subjects: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Namespace",
		Name:       "dev",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188412",
	}},
}

var ScopeResult = v1alpha1.ReportResult{
	Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:      "require-requests-and-limits-required",
	Rule:        "autogen-check-for-requests-and-limits",
	Result:      v1alpha1.StatusFail,
	Severity:    v1alpha1.SeverityHigh,
	Category:    "resources",
	Scored:      true,
	Source:      "Kyverno",
}
