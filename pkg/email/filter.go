package email

import (
	"github.com/kyverno/policy-reporter/pkg/validate"
)

type Filter struct {
	namespace validate.RuleSets
	sources   []string
}

func (f Filter) ValidateSource(source string) bool {
	return validate.ContainsString(source, f.sources)
}

func (f Filter) ValidateNamespace(namespace string) bool {
	return validate.Namespace(namespace, f.namespace)
}

func NewFilter(namespaces validate.RuleSets, sources []string) Filter {
	return Filter{namespaces, sources}
}
