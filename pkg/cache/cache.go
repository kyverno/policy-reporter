package cache

type Cache interface {
	RemoveReport(id string)
	AddReport(report v1alpha2.ReportInterface)
	GetResults(id string) []string
}
