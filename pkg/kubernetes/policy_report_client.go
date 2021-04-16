package kubernetes

import (
	"errors"
	"log"
	"sync"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
)

type policyReportClient struct {
	policyAPI       PolicyReportAdapter
	store           *report.PolicyReportStore
	callbacks       []report.PolicyReportCallback
	resultCallbacks []report.PolicyResultCallback
	mapper          Mapper
	startUp         time.Time
	skipExisting    bool
	started         bool
	modifyHash      map[string]uint64
}

func (c *policyReportClient) RegisterCallback(cb report.PolicyReportCallback) {
	c.callbacks = append(c.callbacks, cb)
}

func (c *policyReportClient) RegisterPolicyResultCallback(cb report.PolicyResultCallback) {
	c.resultCallbacks = append(c.resultCallbacks, cb)
}

func (c *policyReportClient) FetchPolicyReports() ([]report.PolicyReport, error) {
	var reports []report.PolicyReport

	result, err := c.policyAPI.ListPolicyReports()
	if err != nil {
		log.Printf("K8s List Error: %s\n", err.Error())
		return reports, err
	}

	for _, item := range result.Items {
		reports = append(reports, c.mapper.MapPolicyReport(item.Object))
	}

	return reports, nil
}

func (c *policyReportClient) FetchPolicyResults() ([]report.Result, error) {
	var results []report.Result

	reports, err := c.FetchPolicyReports()
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

func (c *policyReportClient) StartWatching() error {
	if c.started {
		return errors.New("PolicyClient.StartWatching was already started")
	}

	c.started = true

	for {
		result, err := c.policyAPI.WatchPolicyReports()
		if err != nil {
			c.started = false
			return err
		}

		for result := range result.ResultChan() {
			if item, ok := result.Object.(*unstructured.Unstructured); ok {
				c.executePolicyReportHandler(result.Type, c.mapper.MapPolicyReport(item.Object))
			}
		}

		// skip existing results when the watcher restarts
		c.skipExisting = true
	}
}

func (c *policyReportClient) executePolicyReportHandler(e watch.EventType, pr report.PolicyReport) {
	opr := report.PolicyReport{}
	if e != watch.Added {
		opr, _ = c.store.Get(pr.GetIdentifier())
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
		c.store.Remove(pr.GetIdentifier())
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

				c.modifyHash[pr.GetIdentifier()] = pr.ResultHash()

				preExisted := pr.CreationTimestamp.Before(c.startUp)

				if c.skipExisting && preExisted {
					break
				}

				wg := sync.WaitGroup{}
				wg.Add(len(pr.Results) * len(c.resultCallbacks))

				for _, r := range pr.Results {
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

				newHash := pr.ResultHash()
				if hash, ok := c.modifyHash[pr.GetIdentifier()]; ok {
					if newHash == hash {
						break
					}
				}

				c.modifyHash[pr.GetIdentifier()] = newHash

				diff := pr.GetNewResults(or)

				wg := sync.WaitGroup{}
				wg.Add(len(diff) * len(c.resultCallbacks))

				for _, r := range diff {
					for _, cb := range c.resultCallbacks {
						go func(callback report.PolicyResultCallback, result report.Result) {
							callback(result, false)
							wg.Done()
						}(cb, r)
					}
				}

				wg.Wait()
			case watch.Deleted:
				if _, ok := c.modifyHash[pr.GetIdentifier()]; ok {
					delete(c.modifyHash, pr.GetIdentifier())
				}
			}
		})
}

// NewPolicyReportClient creates a new PolicyReportClient based on the kubernetes go-client
func NewPolicyReportClient(client PolicyReportAdapter, store *report.PolicyReportStore, mapper Mapper, startUp time.Time) report.PolicyClient {
	return &policyReportClient{
		policyAPI:  client,
		store:      store,
		mapper:     mapper,
		startUp:    startUp,
		modifyHash: make(map[string]uint64),
	}
}
