package result

import (
	"strings"

	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

type Reconditioner struct {
	defaultIDGenerator IDGenerator
	customIDGenerators map[string]IDGenerator
}

func (r *Reconditioner) Prepare(polr openreports.ReportInterface) openreports.ReportInterface {
	generator := r.defaultIDGenerator
	if g, ok := r.customIDGenerators[strings.ToLower(polr.GetSource())]; ok {
		generator = g
	}

	results := polr.GetResults()
	newResults := []openreports.ResultAdapter{}
	for _, r := range results {
		r.ID = generator.Generate(polr, r)
		r.Category = helper.Defaults(r.Category, "Other")

		scope := polr.GetScope()
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

func NewReconditioner(generators map[string]IDGenerator) *Reconditioner {
	return &Reconditioner{
		defaultIDGenerator: NewIDGenerator(nil),
		customIDGenerators: generators,
	}
}
