package cache

import "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"

type Cache interface {
	RemoveReport(id string)
	AddReport(report v1alpha2.ReportInterface)
	GetResults(id string) []string
}

type ItemCache interface {
	Cache
	AddItem(key string, value interface{})
	GetItem(key string) (interface{}, bool)
	RemoveItem(key string)
}
