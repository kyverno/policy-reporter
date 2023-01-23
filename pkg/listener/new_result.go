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
	skipExisting bool
	listener     []report.PolicyReportResultListener
	cache        cache.Cache
	startUp      time.Time
}

func (l *ResultListener) RegisterListener(listener report.PolicyReportResultListener) {
	l.listener = append(l.listener, listener)
}

func (l *ResultListener) Listen(event report.LifecycleEvent) {
	if event.OldPolicyReport != nil && len(event.OldPolicyReport.GetResults()) > 0 {
		for _, result := range event.OldPolicyReport.GetResults() {
			l.cache.Add(result.ID)
		}
	}

	if event.Type != report.Added && event.Type != report.Updated {
		return
	}

	var preExisted bool

	if event.Type == report.Added {
		preExisted = event.NewPolicyReport.GetCreationTimestamp().Local().Before(l.startUp)

		if l.skipExisting && preExisted {
			return
		}
	}

	if len(event.NewPolicyReport.GetResults()) == 0 {
		return
	}

	diff := report.FindNewResults(event.NewPolicyReport, event.OldPolicyReport)

	wg := sync.WaitGroup{}

	for _, r := range diff {
		if found := l.cache.Has(r.GetID()); found {
			continue
		}

		wg.Add(len(l.listener))

		for _, cb := range l.listener {
			go func(callback report.PolicyReportResultListener, result v1alpha2.PolicyReportResult) {
				callback(event.NewPolicyReport, result, preExisted)
				wg.Done()
			}(cb, r)
		}
	}

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
