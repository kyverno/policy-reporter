package target

import (
	"strings"

	"github.com/kyverno/policy-reporter/pkg/report"
)

// Client for a provided Target
type Client interface {
	// Send the given Result to the configured Target
	Send(result *report.Result)
	// SkipExistingOnStartup skips already existing PolicyReportResults on startup
	SkipExistingOnStartup() bool
	// Name is a unique identifier for each Target
	Name() string
	// Validate is a result should send
	Validate(result *report.Result) bool
	// MinimumPriority for a triggered Result to send to this target
	MinimumPriority() string
	// Sources of the Results which should send to this target, empty means all sources
	Sources() []string
}

type BaseClient struct {
	minimumPriority       string
	sources               []string
	skipExistingOnStartup bool
}

func (c *BaseClient) MinimumPriority() string {
	return c.minimumPriority
}

func (c *BaseClient) Sources() []string {
	return c.sources
}

func (c *BaseClient) SkipExistingOnStartup() bool {
	return c.skipExistingOnStartup
}

func (c *BaseClient) Validate(result *report.Result) bool {
	if result.Priority < report.NewPriority(c.minimumPriority) {
		return false
	}

	if len(c.sources) > 0 && !contains(result.Source, c.sources) {
		return false
	}

	return true
}

func contains(source string, sources []string) bool {
	for _, s := range sources {
		if strings.EqualFold(s, source) {
			return true
		}
	}

	return false
}

func NewBaseClient(minimumPriority string, sources []string, skipExistingOnStartup bool) BaseClient {
	return BaseClient{minimumPriority, sources, skipExistingOnStartup}
}
