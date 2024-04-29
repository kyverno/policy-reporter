package report

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

type ReportValidation = func(v1alpha2.ReportInterface) bool

type ReportFilter struct {
	validations []ReportValidation
}

func (rf *ReportFilter) AddValidation(v ReportValidation) {
	rf.validations = append(rf.validations, v)
}

func (rf *ReportFilter) Validate(report v1alpha2.ReportInterface) bool {
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
