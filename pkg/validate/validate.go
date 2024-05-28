package validate

import (
	"github.com/kyverno/go-wildcard"

	"github.com/kyverno/policy-reporter/pkg/helper"
)

func Namespace(namespace string, namespaces RuleSets) bool {
	if namespace == "" {
		return true
	}

	return MatchRuleSet(namespace, namespaces)
}

func Kind(kind string, kinds RuleSets) bool {
	if kind == "" {
		return true
	}

	return MatchRuleSet(kind, kinds)
}

func MatchRuleSet(value string, rules RuleSets) bool {
	if len(rules.Include) > 0 {
		for _, ns := range rules.Include {
			if wildcard.Match(ns, value) {
				return true
			}
		}

		return false
	} else if len(rules.Exclude) > 0 {
		for _, ns := range rules.Exclude {
			if wildcard.Match(ns, value) {
				return false
			}
		}
	}

	return true
}

func ContainsRuleSet(value string, rules RuleSets) bool {
	if len(rules.Include) > 0 {
		return helper.Contains(value, rules.Include)
	} else if len(rules.Exclude) > 0 && helper.Contains(value, rules.Exclude) {
		return false
	}

	return true
}
