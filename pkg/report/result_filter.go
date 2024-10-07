package report

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

type ResultValidation = func(v1alpha2.PolicyReportResult) bool

type ResultFilter struct {
	validations     []ResultValidation
	Sources         []string
	MinimumSeverity string
}

func (rf *ResultFilter) AddValidation(v ResultValidation) {
	rf.validations = append(rf.validations, v)
}

func (rf *ResultFilter) Validate(result v1alpha2.PolicyReportResult) bool {
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
