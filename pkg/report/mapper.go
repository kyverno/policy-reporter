package report

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

type Mapper interface {
	ResolvePriority(policy string, severity v1alpha2.PolicySeverity) v1alpha2.Priority
}

type mapper struct {
	priorityMap map[string]string
}

func (m *mapper) ResolvePriority(policy string, severity v1alpha2.PolicySeverity) v1alpha2.Priority {
	if priority, ok := m.priorityMap[policy]; ok {
		return v1alpha2.NewPriority(priority)
	}

	if severity != "" {
		return v1alpha2.PriorityFromSeverity(severity)
	}

	if priority, ok := m.priorityMap["default"]; ok {
		return v1alpha2.NewPriority(priority)
	}

	return v1alpha2.WarningPriority
}

// NewMapper creates an new Mapper instance
func NewMapper(priorities map[string]string) Mapper {
	return &mapper{priorityMap: priorities}
}
