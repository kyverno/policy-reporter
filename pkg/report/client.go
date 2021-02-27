package report

import "k8s.io/apimachinery/pkg/watch"

// WatchPolicyReportCallback is called whenver a new PolicyReport comes in
type WatchPolicyReportCallback = func(watch.EventType, PolicyReport)

// WatchClusterPolicyReportCallback is called whenver a new ClusterPolicyReport comes in
type WatchClusterPolicyReportCallback = func(watch.EventType, ClusterPolicyReport)

// WatchPolicyResultCallback is called whenver a new PolicyResult comes in
type WatchPolicyResultCallback = func(Result, bool)

// Client interface for interacting with the Kubernetes API
type Client interface {
	// FetchPolicyReports from the unterlying API
	FetchPolicyReports() ([]PolicyReport, error)
	// WatchPolicyReports blocking API to watch for PolicyReport changes
	WatchPolicyReports(WatchPolicyReportCallback) error
	// WatchPolicyReportResults blocking API to watch for PolicyResult changes from PolicyReports and ClusterPolicyReports
	WatchPolicyReportResults(WatchPolicyResultCallback, bool) error
	// FetchPolicyReportResults from the unterlying API
	FetchPolicyReportResults() ([]Result, error)
	// WatchClusterPolicyReports blocking API to watch for ClusterPolicyReport changes
	WatchClusterPolicyReports(WatchClusterPolicyReportCallback) error
	// FetchClusterPolicyReport from the unterlying API
	FetchClusterPolicyReports() ([]ClusterPolicyReport, error)
}
