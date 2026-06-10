package report

import (
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

type ResultValidation = func(openreports.ResultAdapter) bool

type ResultFilter struct {
	validations     []ResultValidation
	Sources         []string
	MinimumSeverity string
}

func (rf *ResultFilter) AddValidation(v ResultValidation) {
	rf.validations = append(rf.validations, v)
}

func (rf *ResultFilter) Validate(result openreports.ResultAdapter) bool {
	for _, validation := range rf.validations {
		if !validation(result) {
			return false
		}
	}

	return true
}

func NewResultFilter() *ResultFilter {
	return &ResultFilter{}
}
