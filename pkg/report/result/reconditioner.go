package result

import (
	"strings"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

type ReconditionerConfig struct {
	IDGenerators         IDGenerator
	SelfassignNamespaces bool
}

type Reconditioner struct {
	defaultIDGenerator IDGenerator
	configs            map[string]ReconditionerConfig
}

func (r *Reconditioner) Prepare(polr openreports.ReportInterface) openreports.ReportInterface {
	generator := r.defaultIDGenerator

	config, ok := r.configs[strings.ToLower(polr.GetSource())]
	if ok && config.IDGenerators != nil {
		generator = config.IDGenerators
	}

	scope := polr.GetScope()
	if config.SelfassignNamespaces && scope != nil && scope.GroupVersionKind().GroupVersion().String() == "v1" && scope.Kind == "Namespace" {
		scope.Namespace = scope.Name
		polr.SetNamespace(scope.Name)
	}

	results := polr.GetResults()
	newResults := []openreports.ResultAdapter{}
	for _, r := range results {
		r.ID = generator.Generate(polr, r)
		r.Category = helper.Defaults(r.Category, "Other")

		if len(r.Subjects) == 0 && scope != nil {
			r.Subjects = append(r.Subjects, *scope)
		}

		if r.Source == "" {
			r.Source = polr.GetSource()
		}

		newResults = append(newResults, r)
	}
	polr.SetResults(newResults)
	return polr
}

func NewReconditioner(configs map[string]ReconditionerConfig) *Reconditioner {
	return &Reconditioner{
		defaultIDGenerator: NewIDGenerator(nil),
		configs:            configs,
	}
}
