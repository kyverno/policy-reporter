package report

import (
	"github.com/kyverno/policy-reporter/pkg/openreports"
)

// PolicyReportListener is called whenever a new PolicyReport comes in
type PolicyReportListener = func(LifecycleEvent)

// PolicyReportResultListener is called whenever a new PolicyResult comes in
type PolicyReportResultListener = func(openreports.ReportInterface, *openreports.ORResultAdapter, bool)

// ScopeResultsListener is called whenever a new PolicyReport with a single resource scope and new results comes in
type ScopeResultsListener = func(openreports.ReportInterface, []*openreports.ORResultAdapter, bool)

// SyncResultsListener is called whenever a PolicyReport event comes in
type SyncResultsListener = func(openreports.ReportInterface)

// PolicyReportClient watches for PolicyReport Events and executes registered callback
type PolicyReportClient interface {
	// Run starts the informer and workerqueue
	Run(worker int, stopper chan struct{}) error
	// Sync Report Informer and start watching for events
	Sync(stopper chan struct{}) error
	// HasSynced the configured PolicyReport
	HasSynced() bool
	// Stop the client
	Stop()
}
