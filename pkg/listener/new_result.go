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
	skipExisting bool
	listener     []report.PolicyReportResultListener
	cache        cache.Cache
	startUp      time.Time
}

func (l *ResultListener) RegisterListener(listener report.PolicyReportResultListener) {
	l.listener = append(l.listener, listener)
}

func (l *ResultListener) UnregisterListener() {
	l.listener = make([]report.PolicyReportResultListener, 0)
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
	if listenerCount == 0 {
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

	grp := sync.WaitGroup{}
	grp.Add(resultCount)
	for _, res := range event.PolicyReport.GetResults() {
		go func(r v1alpha2.PolicyReportResult) {
			defer grp.Done()

			if helper.Contains(r.GetID(), existing) {
				return
			}

			if r.Timestamp.Seconds > 0 {
				created := time.Unix(r.Timestamp.Seconds, int64(r.Timestamp.Nanos))
				if l.skipExisting && created.Local().Before(l.startUp) {
					return
				}
			}

			wg := sync.WaitGroup{}
			wg.Add(listenerCount)

			for _, cb := range l.listener {
				go func(callback report.PolicyReportResultListener, result v1alpha2.PolicyReportResult) {
					callback(event.PolicyReport, result, preExisted)
					wg.Done()
				}(cb, r)
			}

			wg.Wait()
		}(res)
	}

	grp.Wait()

	l.cache.AddReport(event.PolicyReport)
}

func NewResultListener(skipExisting bool, rcache cache.Cache, startUp time.Time) *ResultListener {
	return &ResultListener{
		skipExisting: skipExisting,
		cache:        rcache,
		startUp:      startUp,
		listener:     make([]report.PolicyReportResultListener, 0),
	}
}
