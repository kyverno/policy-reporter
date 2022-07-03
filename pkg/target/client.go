package target

import (
	"github.com/kyverno/policy-reporter/pkg/filter"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
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
	Namespace       filter.Rules
	Priority        filter.Rules
	Policy          filter.Rules
	MinimumPriority string
	Sources         []string
}

func (f *Filter) Validate(result report.Result) bool {
	if len(f.Sources) > 0 && !helper.Contains(result.Source, f.Sources) {
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
	if !result.HasResource() {
		return true
	}

	return filter.ValidateNamespace(result.Resource.Namespace, f.Namespace)
}

func (f *Filter) validatePolicyRules(result report.Result) bool {
	return filter.ValidateRule(result.Policy, f.Policy)
}

func (f *Filter) validatePriorityRules(result report.Result) bool {
	if len(f.Priority.Include) > 0 {
		return helper.Contains(result.Priority.String(), f.Priority.Include)
	} else if len(f.Priority.Exclude) > 0 && helper.Contains(result.Priority.String(), f.Priority.Exclude) {
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
