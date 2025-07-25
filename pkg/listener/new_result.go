package listener

import (
	"sync"
	"time"

	"go.uber.org/zap"

	"github.com/kyverno/policy-reporter/pkg/cache"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/openreports"
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

func (l *ResultListener) Validate(r openreports.ResultAdapter) bool {
	if r.Result == openreports.StatusSkip || r.Result == openreports.StatusPass {
		return false
	}

	return true
}

func (l *ResultListener) Listen(event report.LifecycleEvent) {
	logger := zap.L().Sugar()
	logger.Debugf("new event: type %s, report ID %s", event.Type, event.PolicyReport.GetID())
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
		logger.Debugf("report id %s: caching the results and returning because no listeners are configured", event.PolicyReport.GetID())
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
	newResults := make([]openreports.ResultAdapter, 0)

	for _, r := range event.PolicyReport.GetResults() {
		if helper.Contains(r.GetID(), existing) || !l.Validate(r) {
			logger.Debugf("result for %s, policy %s, rule %s: skipping result sending because the result is already in the cached results for this report or is a pass result", r.ResourceString(), r.Policy, r.Rule)
			continue
		}

		if r.Timestamp.Seconds > 0 {
			created := time.Unix(r.Timestamp.Seconds, int64(r.Timestamp.Nanos))
			if l.skipExisting && created.Local().Before(l.startUp) {
				logger.Debugf("result for %s, policy %s, rule %s: skipping result sending because it was created before the reporter started", r.ResourceString(), r.Policy, r.Rule)
				continue
			}
		}

		newResults = append(newResults, r)
	}

	l.cache.AddReport(event.PolicyReport)
	if len(newResults) == 0 {
		logger.Debugf("not calling the listeners because there are no new results to send")
		return
	}

	if scopeListenerCount > 0 {
		wg := sync.WaitGroup{}
		wg.Add(scopeListenerCount)

		for _, cb := range l.scopeListener {
			go func(callback report.ScopeResultsListener, results []openreports.ResultAdapter) {
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
		go func(r openreports.ResultAdapter) {
			defer grp.Done()

			wg := sync.WaitGroup{}
			wg.Add(listenerCount)

			for _, cb := range l.listener {
				go func(callback report.PolicyReportResultListener, result openreports.ResultAdapter) {
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
