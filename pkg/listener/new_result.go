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

func (l *ResultListener) Listen(event report.LifecycleEvent) {
	if event.Type != report.Added && event.Type != report.Updated {
		l.cache.RemoveReport(event.PolicyReport.GetID())
		return
	}

	if len(event.PolicyReport.GetResults()) == 0 {
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

	wg := sync.WaitGroup{}

	existing := l.cache.GetResults(event.PolicyReport.GetID())

	for _, r := range event.PolicyReport.GetResults() {
		if helper.Contains(r.GetID(), existing) {
			continue
		}

		wg.Add(len(l.listener))

		for _, cb := range l.listener {
			go func(callback report.PolicyReportResultListener, result v1alpha2.PolicyReportResult) {
				callback(event.PolicyReport, result, preExisted)
				wg.Done()
			}(cb, r)
		}
	}

	l.cache.AddReport(event.PolicyReport)

	wg.Wait()
}

func NewResultListener(skipExisting bool, rcache cache.Cache, startUp time.Time) *ResultListener {
	return &ResultListener{
		skipExisting: skipExisting,
		cache:        rcache,
		startUp:      startUp,
		listener:     make([]report.PolicyReportResultListener, 0),
	}
}
