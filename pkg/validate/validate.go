package validate

import (
	"github.com/kyverno/go-wildcard"
	"github.com/kyverno/policy-reporter/pkg/helper"
)

func Namespace(namespace string, namespaces RuleSets) bool {
	if namespace != "" && len(namespaces.Include) > 0 {
		for _, ns := range namespaces.Include {
			if wildcard.Match(ns, namespace) {
				return true
			}
		}

		return false
	} else if namespace != "" && len(namespaces.Exclude) > 0 {
		for _, ns := range namespaces.Exclude {
			if wildcard.Match(ns, namespace) {
				return false
			}
		}
	}

	return true
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

func ContainsString(value string, list []string) bool {
	return len(list) == 0 || helper.Contains(value, list)
}
