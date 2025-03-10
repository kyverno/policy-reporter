package listener

import (
	"sync"
	"time"

	"github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/report"
)

const NewResults = "new_results_listener"

type ResultListener struct {
	skipExisting  bool
	listener      []report.PolicyReportResultListener
	scopeListener []report.ScopeResultsListener
	syncListener  []report.SyncResultsListener
	cache         cache.Cache
	startUp       time.Time
}

func (l *ResultListener) RegisterListener(listener report.PolicyReportResultListener) {
	l.listener = append(l.listener, listener)
}

func (l *ResultListener) UnregisterListener() {
	l.listener = make([]report.PolicyReportResultListener, 0)
}

func (l *ResultListener) RegisterScopeListener(listener report.ScopeResultsListener) {
	l.scopeListener = append(l.scopeListener, listener)
}

func (l *ResultListener) UnregisterScopeListener() {
	l.scopeListener = make([]report.ScopeResultsListener, 0)
}

func (l *ResultListener) RegisterSyncListener(listener report.SyncResultsListener) {
	l.syncListener = append(l.syncListener, listener)
}

func (l *ResultListener) UnregisterSyncListener() {
	l.syncListener = make([]report.SyncResultsListener, 0)
}

func (l *ResultListener) Validate(r v1alpha2.PolicyReportResult) bool {
	if r.Result == v1alpha2.StatusSkip || r.Result == v1alpha2.StatusPass {
		return false
	}

	return true
}

func (l *ResultListener) Listen(event report.LifecycleEvent) {
	if event.Type != report.Added && event.Type != report.Updated {
		l.cache.RemoveReport(event.PolicyReport.GetID())
		return
	}

	resultCount := len(event.PolicyReport.GetResults())
	if resultCount == 0 {
		return
	}

	listenerCount := len(l.listener)
	scopeListenerCount := len(l.scopeListener)
	syncListenerCount := len(l.syncListener)

	if syncListenerCount > 0 {
		wg := sync.WaitGroup{}
		wg.Add(syncListenerCount)

		for _, cb := range l.syncListener {
			go func(callback report.SyncResultsListener) {
				defer wg.Done()

				callback(event.PolicyReport)
			}(cb)
		}

		wg.Wait()
	}

	if listenerCount == 0 && scopeListenerCount == 0 {
		l.cache.AddReport(event.PolicyReport)
		return
	}

	var preExisted bool

	if event.Type == report.Added {
		preExisted = event.PolicyReport.GetCreationTimestamp().Local().Before(l.startUp)

		if l.skipExisting && preExisted {
			l.cache.AddReport(event.PolicyReport)
			return
		}
	}

	existing := l.cache.GetResults(event.PolicyReport.GetID())
	newResults := make([]v1alpha2.PolicyReportResult, 0)

	for _, r := range event.PolicyReport.GetResults() {
		if helper.Contains(r.GetID(), existing) || !l.Validate(r) {
			continue
		}

		if r.Timestamp.Seconds > 0 {
			created := time.Unix(r.Timestamp.Seconds, int64(r.Timestamp.Nanos))
			if l.skipExisting && created.Local().Before(l.startUp) {
				continue
			}
		}

		newResults = append(newResults, r)
	}

	l.cache.AddReport(event.PolicyReport)
	if len(newResults) == 0 {
		return
	}

	if scopeListenerCount > 0 {
		wg := sync.WaitGroup{}
		wg.Add(scopeListenerCount)

		for _, cb := range l.scopeListener {
			go func(callback report.ScopeResultsListener, results []v1alpha2.PolicyReportResult) {
				defer wg.Done()

				callback(event.PolicyReport, results, preExisted)
			}(cb, newResults)
		}

		wg.Wait()
	}

	if len(l.listener) == 0 {
		return
	}

	grp := sync.WaitGroup{}
	grp.Add(len(newResults))
	for _, res := range newResults {
		go func(r v1alpha2.PolicyReportResult) {
			defer grp.Done()

			wg := sync.WaitGroup{}
			wg.Add(listenerCount)

			for _, cb := range l.listener {
				go func(callback report.PolicyReportResultListener, result v1alpha2.PolicyReportResult) {
					defer wg.Done()

					callback(event.PolicyReport, result, preExisted)
				}(cb, r)
			}

			wg.Wait()
		}(res)
	}

	grp.Wait()
}

func NewResultListener(skipExisting bool, rcache cache.Cache, startUp time.Time) *ResultListener {
	return &ResultListener{
		skipExisting:  skipExisting,
		cache:         rcache,
		startUp:       startUp,
		listener:      make([]report.PolicyReportResultListener, 0),
		scopeListener: make([]report.ScopeResultsListener, 0),
		syncListener:  make([]report.SyncResultsListener, 0),
	}
}
