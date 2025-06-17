package report_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/openreports"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_ResultFilter(t *testing.T) {
	t.Run("don't filter any result without validations", func(t *testing.T) {
		filter := report.NewResultFilter()
		if !filter.Validate(fixtures.FailResult) {
			t.Error("Expected result validates to true")
		}
	})
	t.Run("filter result with a false validation", func(t *testing.T) {
		filter := report.NewResultFilter()
		filter.AddValidation(func(r openreports.ORResultAdapter) bool { return false })
		if filter.Validate(fixtures.FailResult) {
			t.Error("Expected result validates to false")
		}
	})
}
