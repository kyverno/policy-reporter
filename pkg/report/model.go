package report

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
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
	PolicyReport v1alpha2.ReportInterface

	NewItems      map[string]v1alpha2.PolicyReportResult
	ModifiedItems map[string]v1alpha2.PolicyReportResult
	DeletedItems  []string
}

// ResourceType Enum defined for PolicyReport
type ResourceType = string

// ReportType Enum
const (
	PolicyReportType        ResourceType = "PolicyReport"
	ClusterPolicyReportType ResourceType = "ClusterPolicyReport"
)

func GetType(r v1alpha2.ReportInterface) ResourceType {
	if r.GetNamespace() == "" {
		return ClusterPolicyReportType
	}

	return PolicyReportType
}

func FindNewResults(nr, or v1alpha2.ReportInterface) []v1alpha2.PolicyReportResult {
	if or == nil {
		return nr.GetResults()
	}

	diff := make([]v1alpha2.PolicyReportResult, 0)
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
