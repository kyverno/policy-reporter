package report_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

var priorityMap = map[string]string{
	"priority-test": "warning",
}

var mapper = report.NewMapper(priorityMap)

func Test_MapperResolvePriority(t *testing.T) {
	t.Run("priority from map", func(t *testing.T) {
		priority := mapper.ResolvePriority("priority-test", v1alpha2.SeverityHigh)
		if priority != v1alpha2.WarningPriority {
			t.Error("expected priority warning, mapped from priority map")
		}
	})

	t.Run("priority from severity", func(t *testing.T) {
		priority := mapper.ResolvePriority("test", v1alpha2.SeverityCritical)
		if priority != v1alpha2.CriticalPriority {
			t.Error("expected priority critical, mapped from severity")
		}
	})

	t.Run("priority from fallback", func(t *testing.T) {
		priority := mapper.ResolvePriority("test", "")
		if priority != v1alpha2.WarningPriority {
			t.Error("expected priority warning, mapped from fallback")
		}
	})

	t.Run("priority from default", func(t *testing.T) {
		mapper := report.NewMapper(map[string]string{"default": "info"})
		priority := mapper.ResolvePriority("test", "")
		if priority != v1alpha2.InfoPriority {
			t.Error("expected priority info, mapped from default")
		}
	})
}

func Test_ResolvePriority(t *testing.T) {
	t.Run("Status Skip", func(t *testing.T) {
		priority := report.ResolvePriority(v1alpha2.PolicyReportResult{
			Result:   v1alpha2.StatusSkip,
			Severity: v1alpha2.SeverityHigh,
		})

		if priority != v1alpha2.DebugPriority {
			t.Errorf("expected priority debug, got %s", priority.String())
		}
	})

	t.Run("Status Pass", func(t *testing.T) {
		priority := report.ResolvePriority(v1alpha2.PolicyReportResult{
			Result:   v1alpha2.StatusPass,
			Severity: v1alpha2.SeverityHigh,
		})

		if priority != v1alpha2.InfoPriority {
			t.Errorf("expected priority info, got %s", priority.String())
		}
	})

	t.Run("Status Warning", func(t *testing.T) {
		priority := report.ResolvePriority(v1alpha2.PolicyReportResult{
			Result:   v1alpha2.StatusWarn,
			Severity: v1alpha2.SeverityHigh,
		})

		if priority != v1alpha2.WarningPriority {
			t.Errorf("expected priority warning, got %s", priority.String())
		}
	})

	t.Run("Status Error", func(t *testing.T) {
		priority := report.ResolvePriority(v1alpha2.PolicyReportResult{
			Result:   v1alpha2.StatusError,
			Severity: v1alpha2.SeverityHigh,
		})

		if priority != v1alpha2.ErrorPriority {
			t.Errorf("expected priority warning, got %s", priority.String())
		}
	})

	t.Run("Status Fail Fallback", func(t *testing.T) {
		priority := report.ResolvePriority(v1alpha2.PolicyReportResult{
			Result: v1alpha2.StatusFail,
		})

		if priority != v1alpha2.WarningPriority {
			t.Errorf("expected priority warning as fail fallback, got %s", priority.String())
		}
	})

	t.Run("Status Severity", func(t *testing.T) {
		priority := report.ResolvePriority(v1alpha2.PolicyReportResult{
			Result:   v1alpha2.StatusFail,
			Severity: v1alpha2.SeverityCritical,
		})

		if priority != v1alpha2.CriticalPriority {
			t.Errorf("expected priority critical, got %s", priority.String())
		}
	})
}
