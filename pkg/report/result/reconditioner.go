package result

import (
	"strings"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
)

type Reconditioner struct {
	defaultIDGenerator IDGenerator
	customIDGenerators map[string]IDGenerator
}

func (r *Reconditioner) Prepare(polr v1alpha2.ReportInterface) v1alpha2.ReportInterface {
	generator := r.defaultIDGenerator
	if g, ok := r.customIDGenerators[strings.ToLower(polr.GetSource())]; ok {
		generator = g
	}

	results := polr.GetResults()
	for i, r := range results {
		r.ID = generator.Generate(polr, r)
		r.Priority = ResolvePriority(r)
		r.Category = helper.Defaults(r.Category, "Other")

		scope := polr.GetScope()
		if len(r.Resources) == 0 && scope != nil {
			r.Resources = append(r.Resources, *scope)
		}

		results[i] = r
	}

	return polr
}

func NewReconditioner(generators map[string]IDGenerator) *Reconditioner {
	return &Reconditioner{
		defaultIDGenerator: NewIDGenerator(nil),
		customIDGenerators: generators,
	}
}
