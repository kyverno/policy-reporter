package report

import (
	"context"
	"k8s.io/apimachinery/pkg/watch"
)

// PolicyReportCallback is called whenever a new PolicyReport comes in
type PolicyReportCallback = func(watch.EventType, PolicyReport, PolicyReport)

// PolicyResultCallback is called whenever a new PolicyResult comes in
type PolicyResultCallback = func(Result, bool)

// PolicyResultClient watches for PolicyReport Events and executes registered callback
type PolicyResultClient interface {
	// RegisterCallback register Handlers called on each PolicyReport watch.Event
	RegisterCallback(PolicyReportCallback)
	// RegisterPolicyResultCallback register Handlers called on each PolicyReport watch.Event for each changed PolicyResult
	RegisterPolicyResultCallback(PolicyResultCallback)
	// RegisterPolicyResultWatcher register a handler for ClusterPolicyReports and PolicyReports who call the registered PolicyResultCallbacks
	RegisterPolicyResultWatcher(skipExisting bool)
	// StartWatching calls the WatchAPI, waiting for incoming PolicyReport watch.Events and call the registered Handlers
	StartWatching(ctx context.Context) error
	// GetFoundResources as Map of Names
	GetFoundResources() map[string]string
}
