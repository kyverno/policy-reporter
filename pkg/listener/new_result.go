package listener

import (
	"sync"
	"time"

	"github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/report"
)

const NewResults = "new_results_listener"

type ResultListener struct {
	listener      []report.PolicyReportResultListener
	scopeListener []report.ScopeResultsListener
	syncListener  []report.SyncResultsListener
	cache         cache.Cache
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

	newResults := make([]v1alpha2.PolicyReportResult, 0)
	newResults = append(newResults, event.PolicyReport.GetResults()...)

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

				callback(event.PolicyReport, results)
			}(cb, newResults)
		}
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

					callback(event.PolicyReport, result)
				}(cb, r)
			}

			wg.Wait()
		}(res)
	}

	grp.Wait()
}

func NewResultListener(rcache cache.Cache, startUp time.Time) *ResultListener {
	return &ResultListener{
		cache:         rcache,
		listener:      make([]report.PolicyReportResultListener, 0),
		scopeListener: make([]report.ScopeResultsListener, 0),
		syncListener:  make([]report.SyncResultsListener, 0),
	}
}
