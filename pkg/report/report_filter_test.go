package report_test

import (
	"testing"

	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_ReportFilter(t *testing.T) {
	t.Run("don't filter any result without validations", func(t *testing.T) {
		filter := report.NewReportFilter()
		if !filter.Validate(preport) {
			t.Error("Expected result validates to true")
		}
	})
	t.Run("filter result with a false validation", func(t *testing.T) {
		filter := report.NewReportFilter()
		filter.AddValidation(func(r v1alpha1.ReportInterface) bool { return false })
		if filter.Validate(preport) {
			t.Error("Expected result validates to false")
		}
	})
}
