package kubernetes

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
)

var (
	policyReports        = schema.GroupVersionResource{Group: "wgpolicyk8s.io", Version: "v1alpha1", Resource: "policyreports"}
	clusterPolicyReports = schema.GroupVersionResource{Group: "wgpolicyk8s.io", Version: "v1alpha1", Resource: "clusterpolicyreports"}
)

const (
	prioriyConfig = "policy-reporter-priorities"
)

type policyReportClient struct {
	client                 dynamic.Interface
	coreClient             CoreClient
	policyCache            map[string]report.PolicyReport
	clusterPolicyCache     map[string]report.ClusterPolicyReport
	clusterPolicyCallbacks []report.ClusterPolicyReportCallback
	policyCallbacks        []report.PolicyReportCallback
	resultCallbacks        []report.PolicyResultCallback
	mapper                 Mapper
	startUp                time.Time
}

func (c *policyReportClient) RegisterPolicyReportCallback(cb report.PolicyReportCallback) {
	c.policyCallbacks = append(c.policyCallbacks, cb)
}

func (c *policyReportClient) RegisterClusterPolicyReportCallback(cb report.ClusterPolicyReportCallback) {
	c.clusterPolicyCallbacks = append(c.clusterPolicyCallbacks, cb)
}

func (c *policyReportClient) RegisterPolicyResultCallback(cb report.PolicyResultCallback) {
	c.resultCallbacks = append(c.resultCallbacks, cb)
}

func (c *policyReportClient) FetchPolicyReports() ([]report.PolicyReport, error) {
	var reports []report.PolicyReport

	result, err := c.client.Resource(policyReports).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("K8s List Error: %s\n", err.Error())
		return reports, err
	}

	for _, item := range result.Items {
		reports = append(reports, c.mapper.MapPolicyReport(item.Object))
	}

	return reports, nil
}

func (c *policyReportClient) FetchClusterPolicyReports() ([]report.ClusterPolicyReport, error) {
	var reports []report.ClusterPolicyReport

	result, err := c.client.Resource(clusterPolicyReports).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("K8s List Error: %s\n", err.Error())
		return reports, err
	}

	for _, item := range result.Items {
		reports = append(reports, c.mapper.MapClusterPolicyReport(item.Object))
	}

	return reports, nil
}

func (c *policyReportClient) FetchPolicyReportResults() ([]report.Result, error) {
	g := new(errgroup.Group)
	mx := new(sync.Mutex)

	var results []report.Result

	g.Go(func() error {
		reports, err := c.FetchClusterPolicyReports()
		if err != nil {
			return err
		}

		for _, clusterReport := range reports {
			for _, result := range clusterReport.Results {
				mx.Lock()
				results = append(results, result)
				mx.Unlock()
			}
		}

		return nil
	})

	g.Go(func() error {
		reports, err := c.FetchPolicyReports()
		if err != nil {
			return err
		}

		for _, clusterReport := range reports {
			for _, result := range clusterReport.Results {
				mx.Lock()
				results = append(results, result)
				mx.Unlock()
			}
		}

		return nil
	})

	return results, g.Wait()
}

func (c *policyReportClient) StartWatchClusterPolicyReports() error {
	for {
		result, err := c.client.Resource(clusterPolicyReports).Watch(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		for result := range result.ResultChan() {
			if item, ok := result.Object.(*unstructured.Unstructured); ok {
				c.executeClusterPolicyReportHandler(result.Type, c.mapper.MapClusterPolicyReport(item.Object))
			}
		}
	}
}

func (c *policyReportClient) executeClusterPolicyReportHandler(e watch.EventType, cpr report.ClusterPolicyReport) {
	wg := sync.WaitGroup{}
	wg.Add(len(c.clusterPolicyCallbacks))

	opr := report.ClusterPolicyReport{}
	if e != watch.Added {
		opr = c.clusterPolicyCache[cpr.GetIdentifier()]
	}

	for _, cb := range c.clusterPolicyCallbacks {
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
		delete(c.clusterPolicyCache, cpr.GetIdentifier())
	} else {
		c.clusterPolicyCache[cpr.GetIdentifier()] = cpr
	}
}

func (c *policyReportClient) StartWatchPolicyReports() error {
	for {
		result, err := c.client.Resource(policyReports).Watch(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		for result := range result.ResultChan() {
			if item, ok := result.Object.(*unstructured.Unstructured); ok {
				c.executePolicyReportHandler(result.Type, c.mapper.MapPolicyReport(item.Object))
			}
		}
	}
}

func (c *policyReportClient) executePolicyReportHandler(e watch.EventType, pr report.PolicyReport) {
	wg := sync.WaitGroup{}
	wg.Add(len(c.policyCallbacks))

	opr := report.PolicyReport{}
	if e != watch.Added {
		opr = c.policyCache[pr.GetIdentifier()]
	}

	for _, cb := range c.policyCallbacks {
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
		delete(c.policyCache, pr.GetIdentifier())
	} else {
		c.policyCache[pr.GetIdentifier()] = pr
	}
}

func (c *policyReportClient) RegisterPolicyResultWatcher(skipExisting bool) {
	c.RegisterPolicyReportCallback(
		func(e watch.EventType, pr report.PolicyReport, or report.PolicyReport) {
			switch e {
			case watch.Added:
				preExisted := pr.CreationTimestamp.Before(c.startUp)

				if skipExisting && preExisted {
					break
				}

				for _, result := range pr.Results {
					for _, cb := range c.resultCallbacks {
						cb(result, preExisted)
					}
				}
			case watch.Modified:
				diff := pr.GetNewResults(or)
				for _, result := range diff {
					for _, cb := range c.resultCallbacks {
						cb(result, false)
					}
				}
			}
		})

	c.RegisterClusterPolicyReportCallback(func(s watch.EventType, cpr report.ClusterPolicyReport, opr report.ClusterPolicyReport) {
		switch s {
		case watch.Added:
			preExisted := cpr.CreationTimestamp.Before(c.startUp)

			if skipExisting && preExisted {
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
			diff := cpr.GetNewResults(c.clusterPolicyCache[cpr.GetIdentifier()])

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
		}
	})
}

func (c *policyReportClient) fetchPriorities(ctx context.Context) error {
	cm, err := c.coreClient.GetConfig(ctx, prioriyConfig)
	if err != nil {
		return err
	}

	if cm != nil {
		c.mapper.SetPriorityMap(cm.Data)
		log.Println("[INFO] Priorities loaded")
	}

	return nil
}

func (c *policyReportClient) syncPriorities(ctx context.Context) error {
	err := c.coreClient.WatchConfigs(ctx, func(e watch.EventType, cm *v1.ConfigMap) {
		if cm.Name != prioriyConfig {
			return
		}

		switch e {
		case watch.Added:
			c.mapper.SetPriorityMap(cm.Data)
		case watch.Modified:
			c.mapper.SetPriorityMap(cm.Data)
		case watch.Deleted:
			c.mapper.SetPriorityMap(map[string]string{})
		}

		log.Println("[INFO] Priorities synchronized")
	})

	if err != nil {
		log.Printf("[INFO] Unable to sync Priorities: %s", err.Error())
	}

	return err
}

// NewPolicyReportClient creates a new ReportClient based on the kubernetes go-client
func NewPolicyReportClient(ctx context.Context, dynamicClient dynamic.Interface, coreClient CoreClient, startUp time.Time) (report.Client, error) {
	reportClient := &policyReportClient{
		client:             dynamicClient,
		coreClient:         coreClient,
		policyCache:        make(map[string]report.PolicyReport),
		clusterPolicyCache: make(map[string]report.ClusterPolicyReport),
		mapper:             NewMapper(make(map[string]string)),
		startUp:            startUp,
	}

	err := reportClient.fetchPriorities(ctx)
	if err != nil {
		log.Printf("[INFO] No PriorityConfig found: %s", err.Error())
	}

	go reportClient.syncPriorities(ctx)

	return reportClient, nil
}
