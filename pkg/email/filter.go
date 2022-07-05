package email

import (
	"github.com/kyverno/policy-reporter/pkg/validate"
)

type Filter struct {
	namespace validate.RuleSets
	sources   validate.RuleSets
}

func (f Filter) ValidateSource(source string) bool {
	return validate.ContainsRuleSet(source, f.sources)
}

func (f Filter) ValidateNamespace(namespace string) bool {
	return validate.Namespace(namespace, f.namespace)
}

func NewFilter(namespaces, sources validate.RuleSets) Filter {
	return Filter{namespaces, sources}
}
