package report

import "github.com/minio/pkg/wildcard"

type Filter interface {
	DisableClusterReports() bool
	AllowReport(report PolicyReport) bool
}

type filter struct {
	disbaleClusterReports bool
	includeNamespaces     []string
	excludeNamespaces     []string
}

func (f *filter) DisableClusterReports() bool {
	return f.disbaleClusterReports
}

func (f *filter) AllowReport(report PolicyReport) bool {
	if report.Namespace == "" {
		return true
	} else if len(f.includeNamespaces) > 0 {
		for _, ns := range f.includeNamespaces {
			if wildcard.Match(ns, report.Namespace) {
				return true
			}
		}

		return false
	} else if len(f.excludeNamespaces) > 0 {
		for _, ns := range f.excludeNamespaces {
			if wildcard.Match(ns, report.Namespace) {
				return false
			}
		}
	}

	return true
}

func NewFilter(disableClusterReports bool, includeNamespaces []string, excludeNamespaces []string) Filter {
	return &filter{disableClusterReports, includeNamespaces, excludeNamespaces}
}
