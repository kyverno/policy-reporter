package fixtures

import (
	corev1 "k8s.io/api/core/v1"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

var PassResult = openreports.ORResultAdapter{
	ID: "123",
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      v1alpha2.StatusPass,
		Severity:    v1alpha2.SeverityHigh,
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

var PassPodResult = openreports.ORResultAdapter{
	ID: "124",
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      v1alpha2.StatusPass,
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

var WarnPodResult = openreports.ORResultAdapter{
	ID: "124",
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      v1alpha2.StatusWarn,
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

var ErrorPodResult = openreports.ORResultAdapter{
	ID: "124",
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      v1alpha2.StatusError,
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

var TrivyResult = openreports.ORResultAdapter{
	ID: "124",
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error",
		Policy:      "policy",
		Rule:        "rule",
		Result:      v1alpha2.StatusFail,
		Category:    "Best Practices",
		Scored:      true,
		Source:      "Trivy",
	},
}

var FailResult = openreports.ORResultAdapter{
	ID: "123",
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      v1alpha2.StatusFail,
		Severity:    v1alpha2.SeverityHigh,
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

var FailDisallowRuleResult = openreports.ORResultAdapter{
	ID: "123",
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "disallow-policy",
		Rule:        "disallow-policy",
		Result:      v1alpha2.StatusFail,
		Severity:    v1alpha2.SeverityHigh,
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

var FailPodResult = openreports.ORResultAdapter{
	ID: "124",
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      v1alpha2.StatusFail,
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

var SkipPodResult = openreports.ORResultAdapter{
	ID: "124",
	ReportResult: v1alpha1.ReportResult{
		Policy:   "require-requests-and-limits-required",
		Rule:     "autogen-check-for-requests-and-limits",
		Result:   v1alpha2.StatusSkip,
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

var FailResultWithoutResource = openreports.ORResultAdapter{
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      v1alpha2.StatusFail,
		Severity:    v1alpha2.SeverityHigh,
		Category:    "resources",
		Scored:      true,
		Source:      "Kyverno",
	},
}

var PassNamespaceResult = openreports.ORResultAdapter{
	ID: "125",
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: The label `test` is required. Rule check-for-GetLabels()-on-namespace",
		Policy:      "require-ns-GetLabels()",
		Rule:        "check-for-GetLabels()-on-namespace",
		Result:      v1alpha2.StatusPass,
		Category:    "namespaces",
		Severity:    v1alpha2.SeverityMedium,
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

var FailNamespaceResult = openreports.ORResultAdapter{
	ID: "126",
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: The label `test` is required. Rule check-for-GetLabels()-on-namespace",
		Policy:      "require-ns-GetLabels()",
		Rule:        "check-for-GetLabels()-on-namespace",
		Result:      v1alpha2.StatusFail,
		Category:    "namespaces",
		Severity:    v1alpha2.SeverityHigh,
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

var ScopeResult = openreports.ORResultAdapter{
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Result:      v1alpha2.StatusFail,
		Severity:    v1alpha2.SeverityHigh,
		Category:    "resources",
		Scored:      true,
		Source:      "Kyverno",
	},
}
