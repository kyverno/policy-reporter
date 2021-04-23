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

type policyReportEvent struct {
	report    report.PolicyReport
	eventType watch.EventType
}

type policyReportEventDebouncer struct {
	events  map[string]policyReportEvent
	channel chan<- policyReportEvent
	mutx    *sync.Mutex
}

func (d *policyReportEventDebouncer) Add(e policyReportEvent) {
	_, ok := d.events[e.report.GetIdentifier()]
	if e.eventType != watch.Modified && ok {
		d.mutx.Lock()
		delete(d.events, e.report.GetIdentifier())
		d.mutx.Unlock()
	}

	if e.eventType != watch.Modified {
		d.channel <- e
		return
	}

	if len(e.report.Results) == 0 && !ok {
		d.mutx.Lock()
		d.events[e.report.GetIdentifier()] = e
		d.mutx.Unlock()

		go func() {
			time.Sleep(10 * time.Second)

			d.mutx.Lock()
			if event, ok := d.events[e.report.GetIdentifier()]; ok {
				d.channel <- event
				delete(d.events, e.report.GetIdentifier())
			}
			d.mutx.Unlock()
		}()

		return
	}

	if ok {
		d.mutx.Lock()
		d.events[e.report.GetIdentifier()] = e
		d.mutx.Unlock()

		return
	}

	d.channel <- e
}

type policyReportClient struct {
	policyAPI       PolicyReportAdapter
	store           *report.PolicyReportStore
	callbacks       []report.PolicyReportCallback
	resultCallbacks []report.PolicyResultCallback
	mapper          Mapper
	startUp         time.Time
	skipExisting    bool
	started         bool
	modifyHash      map[string]string
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
	reportChan := make(chan policyReportEvent)
	errorChan := make(chan error)

	go func() {
		for {
			result, err := c.policyAPI.WatchPolicyReports()
			if err != nil {
				c.started = false
				errorChan <- err
			}

			debouncer := policyReportEventDebouncer{
				events:  make(map[string]policyReportEvent, 0),
				mutx:    new(sync.Mutex),
				channel: reportChan,
			}

			for result := range result.ResultChan() {
				if item, ok := result.Object.(*unstructured.Unstructured); ok {
					report := c.mapper.MapPolicyReport(item.Object)
					debouncer.Add(policyReportEvent{report, result.Type})
				}
			}

			// skip existing results when the watcher restarts
			c.skipExisting = true
		}
	}()

	go func() {
		for event := range reportChan {
			c.executePolicyReportHandler(event.eventType, event.report)
		}

		errorChan <- errors.New("Report Channel closed")
	}()

	return <-errorChan
}

func (c *policyReportClient) executePolicyReportHandler(e watch.EventType, pr report.PolicyReport) {
	log.Printf("[INFO] New Event %s for %s", e, pr.Name)
	opr, ok := c.store.Get(pr.GetIdentifier())
	if !ok {
		log.Printf("[INFO] No previous Report for %s found", pr.Name)
		opr = report.PolicyReport{}
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

				newHash := pr.ResultHash()
				if hash, ok := c.modifyHash[pr.GetIdentifier()]; ok {
					if newHash == hash {
						break
					}
				}

				c.modifyHash[pr.GetIdentifier()] = newHash

				preExisted := pr.CreationTimestamp.Before(c.startUp)

				if c.skipExisting && preExisted {
					break
				}

				diff := pr.GetNewResults(or)

				wg := sync.WaitGroup{}
				wg.Add(len(diff) * len(c.resultCallbacks))

				for _, r := range diff {
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
		modifyHash: make(map[string]string),
	}
}
