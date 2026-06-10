package cache

import "github.com/kyverno/policy-reporter/pkg/openreports"

type Cache interface {
	RemoveReport(id string)
	AddReport(report openreports.ReportInterface)
	GetResults(id string) []string
	Shared() bool
	Clear()
}
