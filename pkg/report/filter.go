package report

import (
	"github.com/kyverno/policy-reporter/pkg/validate"
)

type Filter struct {
	disbaleClusterReports bool
	namespace             validate.RuleSets
}

func (f *Filter) DisableClusterReports() bool {
	return f.disbaleClusterReports
}

func (f *Filter) AllowReport(report PolicyReport) bool {
	return validate.Namespace(report.Namespace, f.namespace)
}

func NewFilter(disableClusterReports bool, namespace validate.RuleSets) *Filter {
	return &Filter{disableClusterReports, namespace}
}

type ResultValidation = func(Result) bool

type ResultFilter struct {
	validations     []ResultValidation
	Sources         []string
	MinimumPriority string
}

func (rf *ResultFilter) AddValidation(v ResultValidation) {
	rf.validations = append(rf.validations, v)
}

func (rf *ResultFilter) Validate(result Result) bool {
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
