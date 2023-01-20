package report

// PolicyReportListener is called whenever a new PolicyReport comes in
type PolicyReportListener = func(LifecycleEvent)

// PolicyReportResultListener is called whenever a new PolicyResult comes in
type PolicyReportResultListener = func(PolicyReport, Result, bool)

// PolicyReportClient watches for PolicyReport Events and executes registered callback
type PolicyReportClient interface {
	// Run WatchPolicyReports starts to watch for PolicyReport LifecycleEvent events
	Run(stopper chan struct{}) error
	// HasSynced the configured PolicyReport
	HasSynced() bool
}
