package fixtures

import (
	"time"

	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

var seconds = time.Date(2021, time.February, 23, 15, 10, 0, 0, time.UTC).Unix()

var CompleteTargetSendResult = v1alpha2.PolicyReportResult{
	Message:   "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:    "require-requests-and-limits-required",
	Rule:      "autogen-check-for-requests-and-limits",
	Timestamp: v1.Timestamp{Seconds: seconds},
	Priority:  v1alpha2.WarningPriority,
	Result:    v1alpha2.StatusFail,
	Severity:  v1alpha2.SeverityHigh,
	Category:  "resources",
	Scored:    true,
	Source:    "Kyverno",
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "default",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	}},
	Properties: map[string]string{"version": "1.2.0"},
}

var MinimalTargetSendResult = v1alpha2.PolicyReportResult{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: v1alpha2.CriticalPriority,
	Result:   v1alpha2.StatusFail,
	Scored:   true,
}

var EnforceTargetSendResult = v1alpha2.PolicyReportResult{
	Message:   "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:    "require-requests-and-limits-required",
	Rule:      "check-for-requests-and-limits",
	Timestamp: v1.Timestamp{Seconds: seconds},
	Priority:  v1alpha2.WarningPriority,
	Result:    v1alpha2.StatusFail,
	Severity:  v1alpha2.SeverityHigh,
	Category:  "resources",
	Scored:    true,
	Source:    "Kyverno",
	Resources: []corev1.ObjectReference{{
		APIVersion: "",
		Kind:       "Pod",
		Name:       "nginx",
		Namespace:  "default",
		UID:        "",
	}},
	Properties: map[string]string{"version": "1.2.0"},
}

var MissingUIDSendResult = v1alpha2.PolicyReportResult{
	Message:   "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:    "require-requests-and-limits-required",
	Rule:      "check-for-requests-and-limits",
	Timestamp: v1.Timestamp{Seconds: seconds},
	Priority:  v1alpha2.WarningPriority,
	Result:    v1alpha2.StatusFail,
	Severity:  v1alpha2.SeverityHigh,
	Category:  "resources",
	Scored:    true,
	Source:    "Kyverno",
	Resources: []corev1.ObjectReference{{
		APIVersion: "v1",
		Kind:       "Pod",
		Name:       "nginx",
		Namespace:  "default",
		UID:        "",
	}},
	Properties: map[string]string{"version": "1.2.0"},
}

var MissingAPIVersionSendResult = v1alpha2.PolicyReportResult{
	Message:   "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:    "require-requests-and-limits-required",
	Rule:      "check-for-requests-and-limits",
	Timestamp: v1.Timestamp{Seconds: seconds},
	Priority:  v1alpha2.WarningPriority,
	Result:    v1alpha2.StatusFail,
	Severity:  v1alpha2.SeverityHigh,
	Category:  "resources",
	Scored:    true,
	Source:    "Kyverno",
	Resources: []corev1.ObjectReference{{
		APIVersion: "",
		Kind:       "Pod",
		Name:       "nginx",
		Namespace:  "default",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	}},
	Properties: map[string]string{"version": "1.2.0"},
}

var ErrorSendResult = v1alpha2.PolicyReportResult{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: v1alpha2.ErrorPriority,
	Result:   v1alpha2.StatusFail,
	Scored:   true,
}

var CritcalSendResult = v1alpha2.PolicyReportResult{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: v1alpha2.CriticalPriority,
	Result:   v1alpha2.StatusFail,
	Scored:   true,
}

var InfoSendResult = v1alpha2.PolicyReportResult{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: v1alpha2.InfoPriority,
	Result:   v1alpha2.StatusFail,
	Scored:   true,
}

var DebugSendResult = v1alpha2.PolicyReportResult{
	Message:  "validation error: label required. Rule app-label-required failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "app-label-requirement",
	Priority: v1alpha2.DebugPriority,
	Result:   v1alpha2.StatusFail,
	Scored:   true,
}
