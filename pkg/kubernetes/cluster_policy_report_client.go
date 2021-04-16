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

type clusterPolicyReportClient struct {
	policyAPI       PolicyReportAdapter
	store           *report.ClusterPolicyReportStore
	callbacks       []report.ClusterPolicyReportCallback
	resultCallbacks []report.PolicyResultCallback
	mapper          Mapper
	startUp         time.Time
	skipExisting    bool
	started         bool
	modifyHash      map[string]uint64
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

	for {
		result, err := c.policyAPI.WatchClusterPolicyReports()
		if err != nil {
			c.started = false
			return err
		}

		for result := range result.ResultChan() {
			if item, ok := result.Object.(*unstructured.Unstructured); ok {
				c.executeClusterPolicyReportHandler(result.Type, c.mapper.MapClusterPolicyReport(item.Object))
			}
		}

		// skip existing results when the watcher restarts
		c.skipExisting = true
	}
}

func (c *clusterPolicyReportClient) executeClusterPolicyReportHandler(e watch.EventType, cpr report.ClusterPolicyReport) {
	opr := report.ClusterPolicyReport{}
	if e != watch.Added {
		opr, _ = c.store.Get(cpr.GetIdentifier())
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

			c.modifyHash[cpr.GetIdentifier()] = cpr.ResultHash()

			preExisted := cpr.CreationTimestamp.Before(c.startUp)

			if c.skipExisting && preExisted {
				break
			}

			wg := sync.WaitGroup{}
			wg.Add(len(cpr.Results) * len(c.resultCallbacks))

			for _, r := range cpr.Results {
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

			newHash := cpr.ResultHash()
			if hash, ok := c.modifyHash[cpr.GetIdentifier()]; ok {
				if newHash == hash {
					break
				}
			}

			c.modifyHash[cpr.GetIdentifier()] = newHash

			diff := cpr.GetNewResults(opr)

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
			if _, ok := c.modifyHash[cpr.GetIdentifier()]; ok {
				delete(c.modifyHash, cpr.GetIdentifier())
			}
		}
	})
}

// NewPolicyReportClient creates a new PolicyReportClient based on the kubernetes go-client
func NewClusterPolicyReportClient(client PolicyReportAdapter, store *report.ClusterPolicyReportStore, mapper Mapper, startUp time.Time) report.ClusterPolicyClient {
	return &clusterPolicyReportClient{
		policyAPI:  client,
		store:      store,
		mapper:     mapper,
		startUp:    startUp,
		modifyHash: make(map[string]uint64),
	}
}
