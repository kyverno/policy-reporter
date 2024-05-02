package report

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

type Namespaced interface {
	GetNamespace() string
}

type Filter struct {
	disbaleClusterReports bool
	namespace             validate.RuleSets
}

func (f *Filter) DisableClusterReports() bool {
	return f.disbaleClusterReports
}

func (f *Filter) AllowReport(report Namespaced) bool {
	return validate.Namespace(report.GetNamespace(), f.namespace)
}

func NewFilter(disableClusterReports bool, namespace validate.RuleSets) *Filter {
	return &Filter{disableClusterReports, namespace}
}

type ResultValidation = func(v1alpha2.ReportInterface, v1alpha2.PolicyReportResult) bool

type ResultFilter struct {
	validations     []ResultValidation
	Sources         []string
	MinimumPriority string
}

func (rf *ResultFilter) AddValidation(v ResultValidation) {
	rf.validations = append(rf.validations, v)
}

func (rf *ResultFilter) Validate(report v1alpha2.ReportInterface, result v1alpha2.PolicyReportResult) bool {
	for _, validation := range rf.validations {
		if !validation(report, result) {
			return false
		}
	}

	return true
}

func NewResultFilter() *ResultFilter {
	return &ResultFilter{}
}

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
