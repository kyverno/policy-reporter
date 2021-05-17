package kubernetes

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/patrickmn/go-cache"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

type clusterPolicyReportClient struct {
	policyAPI       PolicyReportAdapter
	store           *report.ClusterPolicyReportStore
	callbacks       []report.ClusterPolicyReportCallback
	resultCallbacks []report.PolicyResultCallback
	mapper          Mapper
	startUp         time.Time
	skipExisting    bool
	started         bool
	debouncer       clusterPolicyReportEventDebouncer
	resultCache     *cache.Cache
}

func (c *clusterPolicyReportClient) RegisterCallback(cb report.ClusterPolicyReportCallback) {
	c.callbacks = append(c.callbacks, cb)
}

func (c *clusterPolicyReportClient) RegisterPolicyResultCallback(cb report.PolicyResultCallback) {
	c.resultCallbacks = append(c.resultCallbacks, cb)
}

func (c *clusterPolicyReportClient) FetchClusterPolicyReports() ([]report.ClusterPolicyReport, error) {
	var reports []report.ClusterPolicyReport

	result, err := c.policyAPI.ListClusterPolicyReports()
	if err != nil {
		log.Printf("K8s List Error: %s\n", err.Error())
		return reports, err
	}

	for _, item := range result.Items {
		reports = append(reports, c.mapper.MapClusterPolicyReport(item.Object))
	}

	return reports, nil
}

func (c *clusterPolicyReportClient) FetchPolicyResults() ([]report.Result, error) {
	var results []report.Result

	reports, err := c.FetchClusterPolicyReports()
	if err != nil {
		return results, err
	}

	for _, clusterReport := range reports {
		for _, result := range clusterReport.Results {
			results = append(results, result)
		}
	}

	return results, nil
}

func (c *clusterPolicyReportClient) StartWatching() error {
	if c.started {
		return errors.New("ClusterPolicyClient.StartWatching was already started")
	}

	c.started = true
	errorChan := make(chan error)
	go func() {
		for {
			result, err := c.policyAPI.WatchClusterPolicyReports()
			if err != nil {
				c.started = false
				errorChan <- err
			}

			c.debouncer.Reset()

			for result := range result.ResultChan() {
				if item, ok := result.Object.(*unstructured.Unstructured); ok {
					report := c.mapper.MapClusterPolicyReport(item.Object)
					c.debouncer.Add(clusterPolicyReportEvent{report, result.Type})
				}
			}

			// skip existing results when the watcher restarts
			c.skipExisting = true
		}
	}()

	go func() {
		for event := range c.debouncer.ReportChan() {
			c.executeClusterPolicyReportHandler(event.eventType, event.report)
		}

		errorChan <- errors.New("Report Channel closed")
	}()

	return <-errorChan
}

func (c *clusterPolicyReportClient) cacheResults(opr report.ClusterPolicyReport) {
	for id := range opr.Results {
		c.resultCache.SetDefault(id, true)
	}
}

func (c *clusterPolicyReportClient) executeClusterPolicyReportHandler(e watch.EventType, cpr report.ClusterPolicyReport) {
	opr, ok := c.store.Get(cpr.GetIdentifier())
	if !ok {
		opr = report.ClusterPolicyReport{}
	}

	if e == watch.Modified {
		c.cacheResults(opr)
	}

	wg := sync.WaitGroup{}
	wg.Add(len(c.callbacks))

	for _, cb := range c.callbacks {
		go func(
			callback report.ClusterPolicyReportCallback,
			event watch.EventType,
			creport report.ClusterPolicyReport,
			oreport report.ClusterPolicyReport,
		) {
			callback(event, creport, oreport)
			wg.Done()
		}(cb, e, cpr, opr)
	}

	wg.Wait()

	if e == watch.Deleted {
		c.store.Remove(cpr.GetIdentifier())
		return
	}

	c.store.Add(cpr)
}

func (c *clusterPolicyReportClient) RegisterPolicyResultWatcher(skipExisting bool) {
	c.skipExisting = skipExisting

	c.RegisterCallback(func(s watch.EventType, cpr report.ClusterPolicyReport, opr report.ClusterPolicyReport) {
		switch s {
		case watch.Added:
			if len(cpr.Results) == 0 {
				break
			}

			preExisted := cpr.CreationTimestamp.Before(c.startUp)

			if c.skipExisting && preExisted {
				break
			}

			diff := cpr.GetNewResults(opr)

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
			if len(cpr.Results) == 0 {
				break
			}

			diff := cpr.GetNewResults(opr)

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
func NewClusterPolicyReportClient(
	client PolicyReportAdapter,
	store *report.ClusterPolicyReportStore,
	mapper Mapper,
	startUp time.Time,
	debounceTime time.Duration,
	resultCache *cache.Cache,
) report.ClusterPolicyClient {
	return &clusterPolicyReportClient{
		policyAPI: client,
		store:     store,
		mapper:    mapper,
		startUp:   startUp,
		debouncer: clusterPolicyReportEventDebouncer{
			events:       make(map[string]clusterPolicyReportEvent, 0),
			mutx:         new(sync.Mutex),
			channel:      make(chan clusterPolicyReportEvent),
			debounceTime: debounceTime,
		},
		resultCache: resultCache,
	}
}
