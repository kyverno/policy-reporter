package report

import (
	"github.com/kyverno/policy-reporter/pkg/validate"
)

type Namespaced interface {
	GetNamespace() string
}

type MetaFilter struct {
	disbaleClusterReports bool
	namespace             validate.RuleSets
}

func (f *MetaFilter) DisableClusterReports() bool {
	return f.disbaleClusterReports
}

func (f *MetaFilter) AllowReport(report Namespaced) bool {
	return validate.Namespace(report.GetNamespace(), f.namespace)
}

func NewMetaFilter(disableClusterReports bool, namespace validate.RuleSets) *MetaFilter {
	return &MetaFilter{disableClusterReports, namespace}
}
