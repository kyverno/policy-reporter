package cache

import "openreports.io/apis/openreports.io/v1alpha1"

type Cache interface {
	RemoveReport(id string)
	AddReport(report v1alpha1.ReportInterface)
	GetResults(id string) []string
	Shared() bool
	Clear()
}
