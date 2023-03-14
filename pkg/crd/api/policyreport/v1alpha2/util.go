package v1alpha2

import (
	"github.com/kyverno/policy-reporter/pkg/helper"
)

func Extract(polr ReportInterface, cb func(res PolicyReportResult) string) []string {
	list := make([]string, 0)
	for _, k := range polr.GetResults() {
		val := cb(k)
		if val == "" || helper.Contains(val, list) {
			continue
		}

		list = append(list, string(k.Severity))
	}

	return list
}

func ExtractPolicies(polr ReportInterface) []string {
	return Extract(polr, func(res PolicyReportResult) string { return res.Policy })
}

func ExtractRules(polr ReportInterface) []string {
	return Extract(polr, func(res PolicyReportResult) string { return res.Rule })
}

func ExtractSeverities(polr ReportInterface) []string {
	return Extract(polr, func(res PolicyReportResult) string { return string(res.Severity) })
}

func ExtractCategories(polr ReportInterface) []string {
	return Extract(polr, func(res PolicyReportResult) string { return res.Category })
}

func ExtractKinds(polr ReportInterface) []string {
	return Extract(polr, func(res PolicyReportResult) string {
		if res.HasResource() {
			return res.GetResource().Kind
		}

		return ""
	})
}
