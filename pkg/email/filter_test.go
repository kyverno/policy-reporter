package email_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/email"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

func Test_Filters(t *testing.T) {
	t.Run("Validate Default", func(t *testing.T) {
		filter := email.NewFilter(validate.RuleSets{}, validate.RuleSets{})

		if !filter.ValidateNamespace("test") {
			t.Errorf("Unexpected Validation Result without configured rules")
		}
		if !filter.ValidateSource("Kyverno") {
			t.Errorf("Unexpected Validation Result without configured rules")
		}
	})
}
