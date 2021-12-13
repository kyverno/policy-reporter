package report

import (
	"context"
)

// PolicyReportListener is called whenever a new PolicyReport comes in
type PolicyReportListener = func(LifecycleEvent)

// PolicyReportResultListener is called whenever a new PolicyResult comes in
type PolicyReportResultListener = func(*Result, bool)

// PolicyReportClient watches for PolicyReport Events and executes registered callback
type PolicyReportClient interface {
	// WatchPolicyReports starts to watch for PolicyReport LifecycleEvent events
	WatchPolicyReports(ctx context.Context) <-chan LifecycleEvent
	// GetFoundResources as Map of Names
	GetFoundResources() map[string]string
}
