package report

import "github.com/kyverno/policy-reporter/pkg/openreports"

type ReportValidation = func(openreports.ReportInterface) bool

type ReportFilter struct {
	validations []ReportValidation
}

func (rf *ReportFilter) AddValidation(v ReportValidation) {
	rf.validations = append(rf.validations, v)
}

func (rf *ReportFilter) Validate(report openreports.ReportInterface) bool {
	for _, validation := range rf.validations {
		if !validation(report) {
			return false
		}
	}

	return true
}

func NewReportFilter() *ReportFilter {
	return &ReportFilter{}
}
