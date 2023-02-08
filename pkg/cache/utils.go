package cache

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
)

func reportResultsIds(report v1alpha2.ReportInterface) []string {
	list := make([]string, 0, len(report.GetResults()))
	for _, result := range report.GetResults() {
		list = append(list, result.GetID())
	}
	return list
}
