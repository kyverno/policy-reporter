package email

import "openreports.io/apis/openreports.io/v1alpha1"

const (
	PassColor    = "#198754"
	WarnColor    = "#fd7e14"
	FailColor    = "#dc3545"
	ErrorColor   = "#b02a37"
	DefaultColor = "#cccccc"
)

func ColorFromStatus(status string) string {
	switch status {
	case v1alpha1.StatusPass:
		return PassColor
	case v1alpha1.StatusWarn:
		return WarnColor
	case v1alpha1.StatusFail:
		return FailColor
	case v1alpha1.StatusError:
		return ErrorColor
	default:
		return DefaultColor
	}
}
