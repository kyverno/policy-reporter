package report

import "k8s.io/apimachinery/pkg/watch"

type WatchPolicyReportCallback = func(watch.EventType, PolicyReport)
type WatchClusterPolicyReportCallback = func(watch.EventType, ClusterPolicyReport)
type WatchPolicyResultCallback = func(Result)

type Client interface {
	FetchPolicyReports() []PolicyReport
	WatchPolicyReports(WatchPolicyReportCallback)
	WatchRuleValidation(WatchPolicyResultCallback, bool)
	WatchClusterPolicyReports(WatchClusterPolicyReportCallback)
}
