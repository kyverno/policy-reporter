package report

import (
	"bytes"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"
)

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

// ReportType Enum defined for PolicyReport
type ReportType = string

const (
	PolicyReportType        ReportType = "PolicyReport"
	ClusterPolicyReportType ReportType = "ClusterPolicyReport"
)

// Internal Priority definitions and weighting
const (
	DefaultPriority = iota
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
	Message    string
	Policy     string
	Rule       string
	Priority   Priority
	Status     Status
	Severity   Severity `json:",omitempty"`
	Category   string   `json:",omitempty"`
	Scored     bool
	Timestamp  time.Time
	Resource   Resource
	Properties map[string]string
}

// GetIdentifier returns a global unique Result identifier
func (r Result) GetIdentifier() string {
	suffix := ""
	if r.Resource.UID != "" {
		suffix = "__" + r.Resource.UID
	}

	return fmt.Sprintf("%s__%s__%s%s", r.Policy, r.Rule, r.Status, suffix)
}

// HasResource checks if the result has an valid Resource
func (r Result) HasResource() bool {
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
	Name              string
	Namespace         string
	Results           map[string]Result
	Summary           Summary
	CreationTimestamp time.Time
}

// GetIdentifier returns a global unique PolicyReport identifier
func (pr PolicyReport) GetIdentifier() string {
	if pr.Namespace == "" {
		return pr.Name
	}

	return fmt.Sprintf("%s__%s", pr.Namespace, pr.Name)
}

// ResultHash generates a has of the current result set
func (pr PolicyReport) ResultHash() string {
	list := make([]string, 0, len(pr.Results))

	for id := range pr.Results {
		list = append(list, id)
	}

	sort.Strings(list)

	h := sha1.New()
	h.Write([]byte(strings.Join(list, "")))

	return hex.EncodeToString(h.Sum(nil))
}

func (pr PolicyReport) GetResults() map[string]Result {
	return pr.Results
}

func (pr PolicyReport) HasResult(id string) bool {
	_, ok := pr.Results[id]

	return ok
}

func (pr PolicyReport) GetType() ReportType {
	if pr.Namespace == "" {
		return ClusterPolicyReportType
	}

	return PolicyReportType
}

func (pr PolicyReport) GetCreationTimestamp() time.Time {
	return pr.CreationTimestamp
}

// GetNewResults filters already existing Results from the old PolicyReport and returns only the diff with new Results
func (pr PolicyReport) GetNewResults(or PolicyReport) []Result {
	diff := make([]Result, 0)

	for _, r := range pr.Results {
		if or.HasResult(r.GetIdentifier()) {
			continue
		}

		diff = append(diff, r)
	}

	return diff
}
