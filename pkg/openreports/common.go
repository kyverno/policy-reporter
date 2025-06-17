package openreports

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"openreports.io/apis/openreports.io/v1alpha1"
)

// Status specifies state of a policy result
const (
	StatusPass  = "pass"
	StatusFail  = "fail"
	StatusWarn  = "warn"
	StatusError = "error"
	StatusSkip  = "skip"
)

// Severity specifies priority of a policy result
const (
	SeverityCritical = "critical"
	SeverityHigh     = "high"
	SeverityMedium   = "medium"
	SeverityLow      = "low"
	SeverityInfo     = "info"
)

// +kubebuilder:validation:Enum=pass;fail;warn;error;skip

// PolicyResult has one of the following values:
//   - pass: indicates that the policy requirements are met
//   - fail: indicates that the policy requirements are not met
//   - warn: indicates that the policy requirements and not met, and the policy is not scored
//   - error: indicates that the policy could not be evaluated
//   - skip: indicates that the policy was not selected based on user inputs or applicability
type PolicyResult string

// +kubebuilder:validation:Enum=critical;high;low;medium;info

// PolicySeverity has one of the following values:
// - critical
// - high
// - low
// - medium
// - info
type Severity string

var SeverityLevel = map[v1alpha1.ResultSeverity]int{
	"":               -1,
	SeverityInfo:     0,
	SeverityLow:      1,
	SeverityMedium:   2,
	SeverityHigh:     3,
	SeverityCritical: 4,
}

type ReportInterface interface {
	metav1.Object
	GetID() string
	GetKey() string
	GetScope() *corev1.ObjectReference
	GetResults() []ORResultAdapter
	SetResults([]ORResultAdapter)
	HasResult(id string) bool
	GetSummary() v1alpha1.ReportSummary
	GetSource() string
	GetKinds() []string
	GetSeverities() []string
}
