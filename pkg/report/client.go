package report

import (
	"k8s.io/apimachinery/pkg/watch"
)

// PolicyReportCallback is called whenver a new PolicyReport comes in
type PolicyReportCallback = func(watch.EventType, PolicyReport, PolicyReport)

// ClusterPolicyReportCallback is called whenver a new ClusterPolicyReport comes in
type ClusterPolicyReportCallback = func(watch.EventType, ClusterPolicyReport, ClusterPolicyReport)

// PolicyResultCallback is called whenver a new PolicyResult comes in
type PolicyResultCallback = func(Result, bool)

// Client interface for interacting with the Kubernetes API
type ResultClient interface {
	// FetchPolicyResults from the unterlying API
	FetchPolicyResults() ([]Result, error)
	// RegisterPolicyReportCallback register a handler for ClusterPolicyReports and PolicyReports who call the registered PolicyResultCallbacks
	RegisterPolicyResultWatcher(skipExisting bool)
	// RegisterPolicyResultCallback register Handlers called on each PolicyReport- and ClusterPolicyReport watch.Event for each changed PolicyResult
	RegisterPolicyResultCallback(cb PolicyResultCallback)
}

type PolicyClient interface {
	// RegisterPolicyReportCallback register Handlers called on each PolicyReport watch.Event
	RegisterCallback(PolicyReportCallback)
	// RegisterPolicyReportCallback register Handlers called on each PolicyReport watch.Event for each changed PolicyResult
	RegisterPolicyResultCallback(PolicyResultCallback)
	// FetchPolicyReports from the unterlying API
	FetchPolicyReports() ([]PolicyReport, error)
	// FetchPolicyResults from the unterlying PolicyAPI
	FetchPolicyResults() ([]Result, error)
	// RegisterPolicyReportCallback register a handler for ClusterPolicyReports and PolicyReports who call the registered PolicyResultCallbacks
	RegisterPolicyResultWatcher(skipExisting bool)
	// StartWatching calls the WatchAPI, waiting for incoming PolicyReport watch.Events and call the registered Handlers
	StartWatching() error
}

type ClusterPolicyClient interface {
	// RegisterClusterPolicyReportCallback register Handlers called on each ClusterPolicyReport watch.Event
	RegisterCallback(ClusterPolicyReportCallback)
	// RegisterPolicyReportCallback register Handlers called on each ClusterPolicyReport watch.Event for each changed PolicyResult
	RegisterPolicyResultCallback(PolicyResultCallback)
	// FetchClusterPolicyReports from the unterlying API
	FetchClusterPolicyReports() ([]ClusterPolicyReport, error)
	// FetchPolicyResults from the unterlying ClusterPolicyAPI
	FetchPolicyResults() ([]Result, error)
	// RegisterPolicyReportCallback register a handler for ClusterPolicyReports and PolicyReports who call the registered PolicyResultCallbacks
	RegisterPolicyResultWatcher(skipExisting bool)
	// StartWatchPolicyReports calls the WatchAPI, waiting for incoming PolicyReport watch.Events and call the registered Handlers
	StartWatching() error
}
