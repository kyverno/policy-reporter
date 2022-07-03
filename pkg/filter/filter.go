package filter

import (
	"github.com/kyverno/go-wildcard"
	"github.com/kyverno/policy-reporter/pkg/helper"
)

type Filter struct {
	namespace Rules
	sources   []string
}

func (f Filter) ValidateSource(source string) bool {
	return ValidateSource(source, f.sources)
}

func (f Filter) ValidateNamespace(namespace string) bool {
	return ValidateNamespace(namespace, f.namespace)
}

func New(namespaces Rules, sources []string) Filter {
	return Filter{namespaces, sources}
}

func ValidateNamespace(namespace string, namespaces Rules) bool {
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

func ValidateRule(value string, rules Rules) bool {
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

func ValidateSource(source string, sources []string) bool {
	return len(sources) == 0 || helper.Contains(source, sources)
}
