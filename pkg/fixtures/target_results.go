package fixtures

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

var seconds = time.Date(2021, time.February, 23, 15, 10, 0, 0, time.UTC).Unix()

var CompleteTargetSendResult = openreports.ORResultAdapter{
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "autogen-check-for-requests-and-limits",
		Timestamp:   v1.Timestamp{Seconds: seconds},
		Result:      v1alpha2.StatusFail,
		Severity:    v1alpha2.SeverityHigh,
		Category:    "resources",
		Scored:      true,
		Source:      "Kyverno",
		Subjects: []corev1.ObjectReference{{
			APIVersion: "v1",
			Kind:       "Deployment",
			Name:       "nginx",
			Namespace:  "default",
			UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
		}},
		Properties: map[string]string{"version": "1.2.0"},
	},
}

var MinimalTargetSendResult = openreports.ORResultAdapter{
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "app-label-requirement",
		Result:      v1alpha2.StatusFail,
		Scored:      true,
	},
}

var EnforceTargetSendResult = openreports.ORResultAdapter{
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "check-for-requests-and-limits",
		Timestamp:   v1.Timestamp{Seconds: seconds},
		Result:      v1alpha2.StatusFail,
		Severity:    v1alpha2.SeverityHigh,
		Category:    "resources",
		Scored:      true,
		Source:      "Kyverno",
		Subjects: []corev1.ObjectReference{{
			APIVersion: "",
			Kind:       "Pod",
			Name:       "nginx",
			Namespace:  "default",
			UID:        "",
		}},
		Properties: map[string]string{"version": "1.2.0"},
	},
}

var MissingUIDSendResult = openreports.ORResultAdapter{
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "check-for-requests-and-limits",
		Timestamp:   v1.Timestamp{Seconds: seconds},
		Result:      v1alpha2.StatusFail,
		Severity:    v1alpha2.SeverityHigh,
		Category:    "resources",
		Scored:      true,
		Source:      "Kyverno",
		Subjects: []corev1.ObjectReference{{
			APIVersion: "v1",
			Kind:       "Pod",
			Name:       "nginx",
			Namespace:  "default",
			UID:        "",
		}},
		Properties: map[string]string{"version": "1.2.0"},
	},
}

var MissingAPIVersionSendResult = openreports.ORResultAdapter{
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "require-requests-and-limits-required",
		Rule:        "check-for-requests-and-limits",
		Timestamp:   v1.Timestamp{Seconds: seconds},
		Result:      v1alpha2.StatusFail,
		Severity:    v1alpha2.SeverityHigh,
		Category:    "resources",
		Scored:      true,
		Source:      "Kyverno",
		Subjects: []corev1.ObjectReference{{
			APIVersion: "",
			Kind:       "Pod",
			Name:       "nginx",
			Namespace:  "default",
			UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
		}},
		Properties: map[string]string{"version": "1.2.0"},
	},
}

var ErrorSendResult = openreports.ORResultAdapter{
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "app-label-requirement",
		Result:      v1alpha2.StatusFail,
		Scored:      true,
	},
}

var CritcalSendResult = openreports.ORResultAdapter{
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "app-label-requirement",
		Result:      v1alpha2.StatusFail,
		Scored:      true,
	},
}

var InfoSendResult = openreports.ORResultAdapter{
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "app-label-requirement",
		Result:      v1alpha2.StatusFail,
		Scored:      true,
	},
}

var DebugSendResult = openreports.ORResultAdapter{
	ReportResult: v1alpha1.ReportResult{
		Description: "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
		Policy:      "app-label-requirement",
		Result:      v1alpha2.StatusFail,
		Scored:      true,
	},
}
