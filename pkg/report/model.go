package report

import (
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/openreports"
)

// Event Enum
type Event int

func (e Event) String() string {
	switch e {
	case Added:
		return "add"
	case Updated:
		return "update"
	case Deleted:
		return "delete"
	}

	return "unknown"
}

// Possible PolicyReport Event Enums
const (
	Added Event = iota
	Updated
	Deleted
)

// LifecycleEvent of PolicyReports
type LifecycleEvent struct {
	Type         Event
	PolicyReport openreports.ReportInterface
}

// ResourceType Enum defined for PolicyReport
type ResourceType = string

// ReportType Enum
const (
	PolicyReportType        ResourceType = "PolicyReport"
	ClusterPolicyReportType ResourceType = "ClusterPolicyReport"
)

func GetType(r openreports.ReportInterface) ResourceType {
	if r.GetNamespace() == "" {
		return ClusterPolicyReportType
	}

	return PolicyReportType
}

func FindNewResults(nr, or openreports.ReportInterface) []v1alpha1.ReportResult {
	if or == nil {
		return nr.GetResults()
	}

	diff := make([]v1alpha1.ReportResult, 0)
loop:
	for _, r := range nr.GetResults() {
		for _, o := range or.GetResults() {
			if o.GetID() == r.GetID() {
				continue loop
			}
		}

		diff = append(diff, r)
	}

	return diff
}
