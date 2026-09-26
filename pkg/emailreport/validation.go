package emailreport

import (
	"fmt"
	"strings"

	"github.com/robfig/cron/v3"

	policyreporterv1alpha1 "github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
)

func Validate(spec policyreporterv1alpha1.EmailReportSpec) error {
	switch spec.Type {
	case policyreporterv1alpha1.EmailReportTypeSummary, policyreporterv1alpha1.EmailReportTypeViolations:
	default:
		return fmt.Errorf("unsupported report type %q", spec.Type)
	}

	if _, err := cron.ParseStandard(spec.Schedule); err != nil {
		return fmt.Errorf("invalid schedule: %w", err)
	}

	if len(spec.To) == 0 {
		return fmt.Errorf("at least one recipient is required")
	}
	for _, recipient := range spec.To {
		if strings.TrimSpace(recipient) == "" {
			return fmt.Errorf("recipient must not be empty")
		}
	}

	return nil
}
