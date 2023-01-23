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

func Test_ResolvePriority(t *testing.T) {
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
