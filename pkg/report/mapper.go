package report

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

func ResolvePriority(result v1alpha2.PolicyReportResult) v1alpha2.Priority {
	if result.Result == v1alpha2.StatusSkip {
		return v1alpha2.DebugPriority
	}

	if result.Result == v1alpha2.StatusPass {
		return v1alpha2.InfoPriority
	}

	if result.Result == v1alpha2.StatusError {
		return v1alpha2.ErrorPriority
	}

	if result.Result == v1alpha2.StatusWarn {
		return v1alpha2.WarningPriority
	}

	if result.Severity != "" {
		return v1alpha2.PriorityFromSeverity(result.Severity)
	}

	return v1alpha2.WarningPriority
}
