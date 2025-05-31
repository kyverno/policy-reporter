package report

import "openreports.io/apis/openreports.io/v1alpha1"

type ResultValidation = func(v1alpha1.ReportResult) bool

type ResultFilter struct {
	validations     []ResultValidation
	Sources         []string
	MinimumSeverity string
}

func (rf *ResultFilter) AddValidation(v ResultValidation) {
	rf.validations = append(rf.validations, v)
}

func (rf *ResultFilter) Validate(result v1alpha1.ReportResult) bool {
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
