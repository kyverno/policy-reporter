package email

import "github.com/kyverno/kyverno/api/policyreport/v1alpha2"

const (
	PassColor    = "#198754"
	WarnColor    = "#fd7e14"
	FailColor    = "#dc3545"
	ErrorColor   = "#b02a37"
	DefaultColor = "#cccccc"
)

func ColorFromStatus(status string) string {
	switch status {
	case v1alpha2.StatusPass:
		return PassColor
	case v1alpha2.StatusWarn:
		return WarnColor
	case v1alpha2.StatusFail:
		return FailColor
	case v1alpha2.StatusError:
		return ErrorColor
	default:
		return DefaultColor
	}
}
