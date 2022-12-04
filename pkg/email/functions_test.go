package email_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/email"
)

func Test_ColorFromStatus(t *testing.T) {
	t.Run("ColorFromStatus.Pass", func(t *testing.T) {
		color := email.ColorFromStatus(v1alpha2.StatusPass)
		if color != email.PassColor {
			t.Errorf("Unexpected pass color: %s", color)
		}
	})
	t.Run("ColorFromStatus.Warn", func(t *testing.T) {
		color := email.ColorFromStatus(v1alpha2.StatusWarn)
		if color != email.WarnColor {
			t.Errorf("Unexpected warn color: %s", color)
		}
	})
	t.Run("ColorFromStatus.Fail", func(t *testing.T) {
		color := email.ColorFromStatus(v1alpha2.StatusFail)
		if color != email.FailColor {
			t.Errorf("Unexpected fail color: %s", color)
		}
	})
	t.Run("ColorFromStatus.Error", func(t *testing.T) {
		color := email.ColorFromStatus(v1alpha2.StatusError)
		if color != email.ErrorColor {
			t.Errorf("Unexpected error color: %s", color)
		}
	})
	t.Run("ColorFromStatus.Default", func(t *testing.T) {
		color := email.ColorFromStatus("")
		if color != email.DefaultColor {
			t.Errorf("Unexpected error color: %s", color)
		}
	})
}
