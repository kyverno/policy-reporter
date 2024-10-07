package email

import (
	"context"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/kubernetes/namespaces"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

type Filter struct {
	client    namespaces.Client
	namespace validate.RuleSets
	sources   validate.RuleSets
}

func (f Filter) ValidateSource(source string) bool {
	return validate.MatchRuleSet(source, f.sources)
}

func (f Filter) ValidateNamespace(namespace string) bool {
	ruleset := f.namespace

	if len(f.namespace.Selector) > 0 {
		list, err := f.client.List(context.Background(), f.namespace.Selector)
		if err != nil {
			zap.L().Error("failed to resolve namespace selector", zap.Error(err))
		}

		ruleset = validate.RuleSets{
			Include: list,
		}
	}

	return validate.Namespace(namespace, ruleset)
}

func NewFilter(client namespaces.Client, namespaces, sources validate.RuleSets) Filter {
	return Filter{client, namespaces, sources}
}
