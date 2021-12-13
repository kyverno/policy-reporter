package listener

import (
	"sync"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/patrickmn/go-cache"
)

type ResultListener struct {
	skipExisting bool
	listener     []report.PolicyReportResultListener
	cache        *cache.Cache
	startUp      time.Time
}

func (l *ResultListener) RegisterListener(listener report.PolicyReportResultListener) {
	l.listener = append(l.listener, listener)
}

func (l *ResultListener) Listen(event report.LifecycleEvent) {
	if len(event.OldPolicyReport.Results) > 0 {
		for id := range event.OldPolicyReport.Results {
			l.cache.SetDefault(id, true)
		}
	}

	if event.Type != report.Added && event.Type != report.Updated {
		return
	}

	var preExisted bool

	if event.Type == report.Added {
		preExisted = event.NewPolicyReport.CreationTimestamp.Before(l.startUp)

		if l.skipExisting && preExisted {
			return
		}
	}

	if len(event.NewPolicyReport.Results) == 0 {
		return
	}

	diff := event.NewPolicyReport.GetNewResults(event.OldPolicyReport)

	wg := sync.WaitGroup{}

	for _, r := range diff {
		if _, found := l.cache.Get(r.GetIdentifier()); found {
			continue
		}

		wg.Add(len(l.listener))

		for _, cb := range l.listener {
			go func(callback report.PolicyReportResultListener, result *report.Result) {
				callback(result, preExisted)
				wg.Done()
			}(cb, r)
		}
	}

	wg.Wait()
}

func NewResultListener(skipExisting bool, rcache *cache.Cache, startUp time.Time) *ResultListener {
	return &ResultListener{
		skipExisting: skipExisting,
		cache:        rcache,
		startUp:      startUp,
		listener:     make([]report.PolicyReportResultListener, 0),
	}
}
