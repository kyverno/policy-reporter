package report

import "k8s.io/apimachinery/pkg/watch"

// PolicyReportCallback is called whenver a new PolicyReport comes in
type PolicyReportCallback = func(watch.EventType, PolicyReport, PolicyReport)

// ClusterPolicyReportCallback is called whenver a new ClusterPolicyReport comes in
type ClusterPolicyReportCallback = func(watch.EventType, ClusterPolicyReport, ClusterPolicyReport)

// PolicyResultCallback is called whenver a new PolicyResult comes in
type PolicyResultCallback = func(Result, bool)

// Client interface for interacting with the Kubernetes API
type Client interface {
	// RegisterPolicyReportCallback register Handlers called on each PolicyReport watch.Event
	RegisterPolicyReportCallback(PolicyReportCallback)
	// RegisterPolicyReportCallback register Handlers called on each PolicyReport- and ClusterPolicyReport watch.Event for each changed PolicyResult
	RegisterPolicyResultCallback(PolicyResultCallback)
	// RegisterPolicyReportCallback register Handlers called on each ClusterPolicyReport watch.Event
	RegisterClusterPolicyReportCallback(ClusterPolicyReportCallback)
	// FetchPolicyReports from the unterlying API
	FetchPolicyReports() ([]PolicyReport, error)
	// FetchPolicyReportResults from the unterlying API
	FetchPolicyReportResults() ([]Result, error)
	// FetchClusterPolicyReport from the unterlying API
	FetchClusterPolicyReports() ([]ClusterPolicyReport, error)
	// RegisterPolicyReportCallback register a handler for ClusterPolicyReports and PolicyReports who call the registered PolicyResultCallbacks
	RegisterPolicyResultWatcher(skipExisting bool)
	// StartWatchClusterPolicyReports calls the WatchAPI, waiting for incoming ClusterPolicyReport watch.Events and call the registered Handlers
	StartWatchClusterPolicyReports() error
	// StartWatchPolicyReports calls the WatchAPI, waiting for incoming PolicyReport watch.Events and call the registered Handlers
	StartWatchPolicyReports() error
}
