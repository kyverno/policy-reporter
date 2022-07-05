package metrics

import (
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

func NewFilter(namespace, status, policy, source, severity validate.RuleSets) *report.ResultFilter {
	f := &report.ResultFilter{}
	if namespace.Count() > 0 {
		f.AddValidation(func(r report.Result) bool {
			return validate.Namespace(r.Resource.Namespace, namespace)
		})
	}

	if status.Count() > 0 {
		f.AddValidation(func(r report.Result) bool {
			return validate.MatchRuleSet(r.Status, status)
		})
	}

	if policy.Count() > 0 {
		f.AddValidation(func(r report.Result) bool {
			return validate.MatchRuleSet(r.Policy, policy)
		})
	}

	if source.Count() > 0 {
		f.AddValidation(func(r report.Result) bool {
			return validate.MatchRuleSet(r.Source, source)
		})
	}

	if severity.Count() > 0 {
		f.AddValidation(func(r report.Result) bool {
			return validate.MatchRuleSet(r.Severity, severity)
		})
	}

	return f
}
