package target

import (
	"strings"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/minio/pkg/wildcard"
)

// Client for a provided Target
type Client interface {
	// Send the given Result to the configured Target
	Send(result report.Result)
	// SkipExistingOnStartup skips already existing PolicyReportResults on startup
	SkipExistingOnStartup() bool
	// Name is a unique identifier for each Target
	Name() string
	// Validate is a result should send
	Validate(result report.Result) bool
	// MinimumPriority for a triggered Result to send to this target
	MinimumPriority() string
	// Sources of the Results which should send to this target, empty means all sources
	Sources() []string
}

type Rules struct {
	Exclude []string
	Include []string
}

type Filter struct {
	Namespace       Rules
	Priority        Rules
	Policy          Rules
	MinimumPriority string
	Sources         []string
}

func (f *Filter) Validate(result report.Result) bool {
	if len(f.Sources) > 0 && !contains(result.Source, f.Sources) {
		return false
	}

	if result.Priority < report.NewPriority(f.MinimumPriority) {
		return false
	}

	if !f.validateNamespaceRules(result) {
		return false
	}

	if !f.validatePolicyRules(result) {
		return false
	}

	if !f.validatePriorityRules(result) {
		return false
	}

	return true
}

func (f *Filter) validateNamespaceRules(result report.Result) bool {
	if result.HasResource() && len(f.Namespace.Include) > 0 {
		for _, ns := range f.Namespace.Include {
			if wildcard.Match(ns, result.Resource.Namespace) {
				return true
			}
		}

		return false
	} else if result.HasResource() && len(f.Namespace.Exclude) > 0 {
		for _, ns := range f.Namespace.Exclude {
			if wildcard.Match(ns, result.Resource.Namespace) {
				return false
			}
		}
	}

	return true
}

func (f *Filter) validatePolicyRules(result report.Result) bool {
	if len(f.Policy.Include) > 0 {
		for _, ns := range f.Policy.Include {
			if wildcard.Match(ns, result.Policy) {
				return true
			}
		}

		return false
	} else if len(f.Policy.Exclude) > 0 {
		for _, ns := range f.Policy.Exclude {
			if wildcard.Match(ns, result.Policy) {
				return false
			}
		}
	}

	return true
}

func (f *Filter) validatePriorityRules(result report.Result) bool {
	if len(f.Priority.Include) > 0 {
		return contains(result.Priority.String(), f.Priority.Include)
	} else if len(f.Priority.Exclude) > 0 && contains(result.Priority.String(), f.Priority.Exclude) {
		return false
	}

	return true
}

type BaseClient struct {
	name                  string
	skipExistingOnStartup bool
	filter                *Filter
}

func (c *BaseClient) Name() string {
	return c.name
}

func (c *BaseClient) MinimumPriority() string {
	return c.filter.MinimumPriority
}

func (c *BaseClient) Sources() []string {
	return c.filter.Sources
}

func (c *BaseClient) Validate(result report.Result) bool {
	return c.filter.Validate(result)
}

func (c *BaseClient) SkipExistingOnStartup() bool {
	return c.skipExistingOnStartup
}

func NewBaseClient(name string, skipExistingOnStartup bool, filter *Filter) BaseClient {
	return BaseClient{name, skipExistingOnStartup, filter}
}

func contains(source string, sources []string) bool {
	for _, s := range sources {
		if strings.EqualFold(s, source) {
			return true
		}
	}

	return false
}
