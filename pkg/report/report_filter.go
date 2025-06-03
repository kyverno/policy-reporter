package report

import (
	"openreports.io/apis/openreports.io/v1alpha1"
)

type ReportValidation = func(v1alpha1.ReportInterface) bool

type ReportFilter struct {
	validations []ReportValidation
}

func (rf *ReportFilter) AddValidation(v ReportValidation) {
	rf.validations = append(rf.validations, v)
}

func (rf *ReportFilter) Validate(report v1alpha1.ReportInterface) bool {
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
