package kubernetes

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/patrickmn/go-cache"
	"k8s.io/apimachinery/pkg/watch"
)

type policyReportClient struct {
	policyAPI       PolicyReportAdapter
	store           *report.PolicyReportStore
	callbacks       []report.PolicyReportCallback
	resultCallbacks []report.PolicyResultCallback
	debouncer       *debouncer
	startUp         time.Time
	skipExisting    bool
	started         bool
	resultCache     *cache.Cache
}

func (c *policyReportClient) RegisterCallback(cb report.PolicyReportCallback) {
	c.callbacks = append(c.callbacks, cb)
}

func (c *policyReportClient) RegisterPolicyResultCallback(cb report.PolicyResultCallback) {
	c.resultCallbacks = append(c.resultCallbacks, cb)
}

func (c *policyReportClient) GetFoundResources() map[string]string {
	return c.policyAPI.GetFoundResources()
}

func (c *policyReportClient) StartWatching(ctx context.Context) error {
	if c.started {
		return errors.New("StartWatching was already started")
	}

	c.started = true

	events, err := c.policyAPI.WatchPolicyReports(ctx)
	if err != nil {
		c.started = false
		return err
	}

	go func() {
		for event := range events {
			c.debouncer.Add(event)
		}

		close(c.debouncer.channel)
	}()

	for event := range c.debouncer.ReportChan() {
		c.executeReportHandler(event.Type, event.Report)
	}

	c.started = false

	return errors.New("watching stopped")
}

func (c *policyReportClient) cacheResults(opr report.PolicyReport) {
	for id := range opr.Results {
		c.resultCache.SetDefault(id, true)
	}
}

func (c *policyReportClient) executeReportHandler(e watch.EventType, pr report.PolicyReport) {
	opr, ok := c.store.Get(pr.GetType(), pr.GetIdentifier())
	if !ok {
		opr = report.PolicyReport{}
	}

	if len(opr.Results) > 0 {
		c.cacheResults(opr)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(c.callbacks))

	for _, cb := range c.callbacks {
		go func(
			callback report.PolicyReportCallback,
			event watch.EventType,
			creport report.PolicyReport,
			oreport report.PolicyReport,
		) {
			callback(event, creport, oreport)
			wg.Done()
		}(cb, e, pr, opr)
	}

	wg.Wait()

	if e == watch.Deleted {
		c.store.Remove(pr.GetType(), pr.GetIdentifier())
		return
	}

	c.store.Add(pr)
}

func (c *policyReportClient) RegisterPolicyResultWatcher(skipExisting bool) {
	c.skipExisting = skipExisting

	c.RegisterCallback(
		func(e watch.EventType, pr report.PolicyReport, or report.PolicyReport) {
			switch e {
			case watch.Added:
				if len(pr.Results) == 0 {
					break
				}

				preExisted := pr.CreationTimestamp.Before(c.startUp)

				if c.skipExisting && preExisted {
					break
				}

				diff := pr.GetNewResults(or)

				wg := sync.WaitGroup{}

				for _, r := range diff {
					if _, found := c.resultCache.Get(r.GetIdentifier()); found {
						continue
					}

					wg.Add(len(c.resultCallbacks))

					for _, cb := range c.resultCallbacks {
						go func(callback report.PolicyResultCallback, result report.Result) {
							callback(result, preExisted)
							wg.Done()
						}(cb, r)
					}
				}

				wg.Wait()
			case watch.Modified:
				if len(pr.Results) == 0 {
					break
				}

				diff := pr.GetNewResults(or)

				wg := sync.WaitGroup{}

				for _, r := range diff {
					if _, found := c.resultCache.Get(r.GetIdentifier()); found {
						continue
					}

					wg.Add(len(c.resultCallbacks))

					for _, cb := range c.resultCallbacks {
						go func(callback report.PolicyResultCallback, result report.Result) {
							callback(result, false)
							wg.Done()
						}(cb, r)
					}
				}

				wg.Wait()
			}
		})
}

// NewPolicyReportClient creates a new PolicyReportClient based on the kubernetes go-client
func NewPolicyReportClient(
	client PolicyReportAdapter,
	store *report.PolicyReportStore,
	startUp time.Time,
	resultCache *cache.Cache,
) report.PolicyResultClient {
	return &policyReportClient{
		policyAPI:   client,
		store:       store,
		startUp:     startUp,
		resultCache: resultCache,
		debouncer:   newDebouncer(),
	}
}
