package target

import (
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"
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

func NewClientFilter(namespace, priority, policy validate.RuleSets, minimumPriority string, sources []string) *report.ResultFilter {
	f := report.NewResultFilter()
	f.Sources = sources
	f.MinimumPriority = minimumPriority

	if len(sources) > 0 {
		f.AddValidation(func(r report.Result) bool {
			return helper.Contains(r.Source, sources)
		})
	}

	if namespace.Count() > 0 {
		f.AddValidation(func(r report.Result) bool {
			return validate.Namespace(r.Resource.Namespace, namespace)
		})
	}

	if minimumPriority != "" {
		f.AddValidation(func(r report.Result) bool {
			return r.Priority >= report.NewPriority(f.MinimumPriority)
		})
	}

	if policy.Count() > 0 {
		f.AddValidation(func(r report.Result) bool {
			return validate.MatchRuleSet(r.Policy, policy)
		})
	}

	if priority.Count() > 0 {
		f.AddValidation(func(r report.Result) bool {
			return validate.ContainsRuleSet(r.Priority.String(), priority)
		})
	}

	return f
}

type BaseClient struct {
	name                  string
	skipExistingOnStartup bool
	filter                *report.ResultFilter
}

type ClientOptions struct {
	Name                  string
	SkipExistingOnStartup bool
	Filter                *report.ResultFilter
}

func (c *BaseClient) Name() string {
	return c.name
}

func (c *BaseClient) MinimumPriority() string {
	if c.filter == nil {
		return report.DefaultPriority.String()
	}

	return c.filter.MinimumPriority
}

func (c *BaseClient) Sources() []string {
	if c.filter == nil {
		return make([]string, 0)
	}

	return c.filter.Sources
}

func (c *BaseClient) Validate(result report.Result) bool {
	if c.filter == nil {
		return true
	}

	return c.filter.Validate(result)
}

func (c *BaseClient) SkipExistingOnStartup() bool {
	return c.skipExistingOnStartup
}

func NewBaseClient(options ClientOptions) BaseClient {
	return BaseClient{options.Name, options.SkipExistingOnStartup, options.Filter}
}
