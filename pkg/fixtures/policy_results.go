package fixtures

import (
	"github.com/kyverno/policy-reporter/pkg/openreports"
	corev1 "k8s.io/api/core/v1"
	"openreports.io/apis/openreports.io/v1alpha1"
)

var PassResult = &openreports.ORResultAdapter{
	ReportResult: &v1alpha1.ReportResult{
		ID:          "123",
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      openreports.StatusPass,
		Severity:    openreports.SeverityHigh,
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
	},
}

var PassPodResult = &openreports.ORResultAdapter{
	ReportResult: &v1alpha1.ReportResult{
		ID:          "124",
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      openreports.StatusPass,
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
	},
}
var WarnPodResult = &openreports.ORResultAdapter{
	ReportResult: &v1alpha1.ReportResult{
		ID:          "124",
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      openreports.StatusWarn,
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
	},
}

var ErrorPodResult = &openreports.ORResultAdapter{
	ReportResult: &v1alpha1.ReportResult{
		ID:          "124",
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      openreports.StatusError,
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
	},
}

var TrivyResult = &openreports.ORResultAdapter{
	ReportResult: &v1alpha1.ReportResult{
		ID:          "124",
		Description: "validation error",
		Policy:      "policy",
		Rule:        "rule",
		Result:      openreports.StatusFail,
		Category:    "Best Practices",
		Scored:      true,
		Source:      "Trivy",
	},
}

var FailResult = &openreports.ORResultAdapter{
	ReportResult: &v1alpha1.ReportResult{
		ID:          "123",
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      openreports.StatusFail,
		Severity:    openreports.SeverityHigh,
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
	},
}

var FailDisallowRuleResult = &openreports.ORResultAdapter{
	ReportResult: &v1alpha1.ReportResult{
		ID:          "123",
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "disallow-policy",
		Rule:        "disallow-policy",
		Result:      openreports.StatusFail,
		Severity:    openreports.SeverityHigh,
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
	},
}

var FailPodResult = &openreports.ORResultAdapter{
	ReportResult: &v1alpha1.ReportResult{
		ID:          "124",
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      openreports.StatusFail,
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
	},
}

var SkipPodResult = &openreports.ORResultAdapter{
	ReportResult: &v1alpha1.ReportResult{
		ID:       "124",
		Policy:   "require-requests-and-limits-required",
		Rule:     "autogen-check-for-requests-and-limits",
		Result:   openreports.StatusSkip,
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
	},
}

var FailResultWithoutResource = &openreports.ORResultAdapter{
	ReportResult: &v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      openreports.StatusFail,
		Severity:    openreports.SeverityHigh,
		Category:    "resources",
		Scored:      true,
		Source:      "Kyverno",
	},
}

var PassNamespaceResult = &openreports.ORResultAdapter{
	ReportResult: &v1alpha1.ReportResult{
		ID:          "125",
		Description: "validation error: The label `test` is required. Rule check-for-GetLabels()-on-namespace",
		Policy:      "require-ns-GetLabels()",
		Rule:        "check-for-GetLabels()-on-namespace",
		Result:      openreports.StatusPass,
		Category:    "namespaces",
		Severity:    openreports.SeverityMedium,
		Scored:      true,
		Source:      "Kyverno",
		Subjects: []corev1.ObjectReference{{
			APIVersion: "v1",
			Kind:       "Namespace",
			Name:       "test",
			UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188411",
		}},
	},
}

var FailNamespaceResult = &openreports.ORResultAdapter{
	ReportResult: &v1alpha1.ReportResult{
		ID:          "126",
		Description: "validation error: The label `test` is required. Rule check-for-GetLabels()-on-namespace",
		Policy:      "require-ns-GetLabels()",
		Rule:        "check-for-GetLabels()-on-namespace",
		Result:      openreports.StatusFail,
		Category:    "namespaces",
		Severity:    openreports.SeverityHigh,
		Scored:      true,
		Source:      "Kyverno",
		Subjects: []corev1.ObjectReference{{
			APIVersion: "v1",
			Kind:       "Namespace",
			Name:       "dev",
			UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188412",
		}},
	},
}

var ScopeResult = &openreports.ORResultAdapter{
	ReportResult: &v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      openreports.StatusFail,
		Severity:    openreports.SeverityHigh,
		Category:    "resources",
		Scored:      true,
		Source:      "Kyverno",
	},
}
