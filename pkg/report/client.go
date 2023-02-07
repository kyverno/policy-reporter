package report

import "github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"

// PolicyReportListener is called whenever a new PolicyReport comes in
type PolicyReportListener = func(LifecycleEvent)

// PolicyReportResultListener is called whenever a new PolicyResult comes in
type PolicyReportResultListener = func(v1alpha2.ReportInterface, v1alpha2.PolicyReportResult, bool)

// PolicyReportClient watches for PolicyReport Events and executes registered callback
type PolicyReportClient interface {
	// Run WatchPolicyReports starts to watch for PolicyReport LifecycleEvent events
	Run(stopper chan struct{}) error
	// HasSynced the configured PolicyReport
	HasSynced() bool
}
