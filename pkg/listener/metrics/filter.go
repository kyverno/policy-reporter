package metrics

import (
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/minio/pkg/wildcard"
)

type Rules struct {
	Exclude []string
	Include []string
}

type Filter struct {
	Namespace Rules
	Status    Rules
	Policy    Rules
	Source    Rules
	Severity  Rules
}

func (f *Filter) Validate(result *report.Result) bool {
	if result.Resource != nil &&
		result.Resource.Namespace != "" &&
		!validateRules(result.Resource.Namespace, f.Namespace) {
		return false
	}
	if !validateRules(result.Status, f.Status) {
		return false
	}
	if !validateRules(result.Policy, f.Policy) {
		return false
	}
	if !validateRules(result.Source, f.Source) {
		return false
	}
	if !validateRules(result.Severity, f.Severity) {
		return false
	}

	return true
}

func validateRules(value string, rules Rules) bool {
	if len(rules.Include) > 0 {
		for _, rule := range rules.Include {
			if wildcard.Match(rule, value) {
				return true
			}
		}

		return false
	} else if len(rules.Exclude) > 0 {
		for _, rule := range rules.Exclude {
			if wildcard.Match(rule, value) {
				return false
			}
		}
	}

	return true
}

func NewFilter(namespace, status, policy, source, severity Rules) *Filter {
	return &Filter{namespace, status, policy, source, severity}
}
