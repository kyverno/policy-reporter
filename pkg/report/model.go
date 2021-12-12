package report

import (
	"bytes"
	"crypto/sha1"
	"fmt"
	"time"
)

// Event Enum
type Event = int

// Possible PolicyReport Event Enums
const (
	Added Event = iota
	Updated
	Deleted
)

// LifecycleEvent of PolicyReports
type LifecycleEvent struct {
	Type            Event
	NewPolicyReport *PolicyReport
	OldPolicyReport *PolicyReport
}

// Status Enum defined for PolicyReport
type Status = string

// Severity Enum defined for PolicyReport
type Severity = string

// Enums for predefined values from the PolicyReport spec
const (
	Fail  Status = "fail"
	Warn  Status = "warn"
	Error Status = "error"
	Pass  Status = "pass"
	Skip  Status = "skip"

	Low    Severity = "low"
	Medium Severity = "medium"
	High   Severity = "high"

	defaultString  = ""
	debugString    = "debug"
	infoString     = "info"
	warningString  = "warning"
	errorString    = "error"
	criticalString = "critical"
)

// ResourceType Enum defined for PolicyReport
type ResourceType = string

// ReportType Enum
const (
	PolicyReportType        ResourceType = "PolicyReport"
	ClusterPolicyReportType ResourceType = "ClusterPolicyReport"
)

// Internal Priority definitions and weighting
const (
	DefaultPriority Priority = iota
	DebugPriority
	InfoPriority
	WarningPriority
	CriticalPriority
	ErrorPriority
)

// Priority Enum for internal Result weighting
type Priority int

// String maps the internal weighting of Priorities to a String representation
func (p Priority) String() string {
	switch p {
	case DebugPriority:
		return debugString
	case InfoPriority:
		return infoString
	case WarningPriority:
		return warningString
	case ErrorPriority:
		return errorString
	case CriticalPriority:
		return criticalString
	default:
		return defaultString
	}
}

// MarshalJSON marshals the enum as a quoted json string
func (p Priority) MarshalJSON() ([]byte, error) {
	buffer := bytes.NewBufferString(`"`)
	buffer.WriteString(p.String())
	buffer.WriteString(`"`)

	return buffer.Bytes(), nil
}

// PriorityFromStatus creates a Priority based on a Status
func PriorityFromStatus(s Status) Priority {
	switch s {
	case Fail:
		return CriticalPriority
	case Error:
		return ErrorPriority
	case Warn:
		return WarningPriority
	case Pass:
		return InfoPriority
	default:
		return DefaultPriority
	}
}

// PriorityFromSeverity creates a Priority based on a Severity
func PriorityFromSeverity(s Severity) Priority {
	switch s {
	case High:
		return CriticalPriority
	case Medium:
		return WarningPriority
	default:
		return InfoPriority
	}
}

// NewPriority creates a new Priority based an its string representation
func NewPriority(p string) Priority {
	switch p {
	case debugString:
		return DebugPriority
	case infoString:
		return InfoPriority
	case warningString:
		return WarningPriority
	case errorString:
		return ErrorPriority
	case criticalString:
		return CriticalPriority
	default:
		return DefaultPriority
	}
}

// Resource from the Kubernetes spec k8s.io/api/core/v1.ObjectReference
type Resource struct {
	APIVersion string
	Kind       string
	Name       string
	Namespace  string
	UID        string
}

// Result from the PolicyReport spec wgpolicyk8s.io/v1alpha1.PolicyReportResult
type Result struct {
	ID         string `json:"-"`
	Message    string
	Policy     string
	Rule       string
	Priority   Priority
	Status     Status
	Severity   Severity `json:",omitempty"`
	Category   string   `json:",omitempty"`
	Source     string   `json:",omitempty"`
	Scored     bool
	Timestamp  time.Time
	Resource   *Resource
	Properties map[string]string
}

// GetIdentifier returns a global unique Result identifier
func (r Result) GetIdentifier() string {
	return r.ID
}

// HasResource checks if the result has an valid Resource
func (r Result) HasResource() bool {
	if r.Resource == nil {
		return false
	}

	return r.Resource.UID != ""
}

// Summary from the PolicyReport spec wgpolicyk8s.io/v1alpha1.PolicyReportSummary
type Summary struct {
	Pass  int
	Skip  int
	Warn  int
	Error int
	Fail  int
}

// PolicyReport from the PolicyReport spec wgpolicyk8s.io/v1alpha1.PolicyReport
type PolicyReport struct {
	ID                string
	Name              string
	Namespace         string
	Results           map[string]*Result
	Summary           *Summary
	CreationTimestamp time.Time
}

// GetIdentifier returns a global unique PolicyReport identifier
func (pr PolicyReport) GetIdentifier() string {
	return pr.ID
}

// HasResult returns if the Report has an Rusult with the given ID
func (pr PolicyReport) HasResult(id string) bool {
	_, ok := pr.Results[id]

	return ok
}

// GetType returns the Type of the Report
func (pr PolicyReport) GetType() ResourceType {
	if pr.Namespace == "" {
		return ClusterPolicyReportType
	}

	return PolicyReportType
}

// GetNewResults filters already existing Results from the old PolicyReport and returns only the diff with new Results
func (pr PolicyReport) GetNewResults(or *PolicyReport) []*Result {
	diff := make([]*Result, 0)

	for _, r := range pr.Results {
		if or.HasResult(r.GetIdentifier()) {
			continue
		}

		diff = append(diff, r)
	}

	return diff
}

func GeneratePolicyReportID(name, namespace string) string {
	id := name

	if namespace != "" {
		id = fmt.Sprintf("%s__%s", namespace, name)
	}

	h := sha1.New()

	h.Write([]byte(id))

	return fmt.Sprintf("%x", h.Sum(nil))
}

func GeneratePolicyReportResultID(uid, policy, rule, status, suffix string) string {
	if uid != "" {
		suffix = "__" + uid
	}

	id := fmt.Sprintf("%s__%s__%s%s", policy, rule, status, suffix)

	h := sha1.New()
	h.Write([]byte(id))

	return fmt.Sprintf("%x", h.Sum(nil))
}
