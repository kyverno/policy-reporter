package email

import (
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

const (
	PassColor    = "#198754"
	WarnColor    = "#fd7e14"
	FailColor    = "#dc3545"
	ErrorColor   = "#b02a37"
	DefaultColor = "#cccccc"
)

func ColorFromStatus(status string) string {
	switch status {
	case openreports.StatusPass:
		return PassColor
	case openreports.StatusWarn:
		return WarnColor
	case openreports.StatusFail:
		return FailColor
	case openreports.StatusError:
		return ErrorColor
	default:
		return DefaultColor
	}
}
