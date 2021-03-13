package target

import (
	"github.com/fjogeleit/policy-reporter/pkg/report"
)

// Client for a provided Target
type Client interface {
	// Send the given Result to the configured Target
	Send(result report.Result)
	// SkipExistingOnStartup skips already existing PolicyReportResults on startup
	SkipExistingOnStartup() bool
	// Name is a unique identifier for each Target
	Name() string
	// MinimumPriority for a triggered Result to send to this target
	MinimumPriority() string
}
