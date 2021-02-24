package kubernetes

import (
	"context"
	"log"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"golang.org/x/sync/errgroup"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

var (
	policyReports        = schema.GroupVersionResource{Group: "wgpolicyk8s.io", Version: "v1alpha1", Resource: "policyreports"}
	clusterPolicyReports = schema.GroupVersionResource{Group: "wgpolicyk8s.io", Version: "v1alpha1", Resource: "clusterpolicyreports"}
)

const (
	prioriyConfig = "policy-reporter-priorities"
)

type policyReportClient struct {
	client             dynamic.Interface
	coreClient         CoreClient
	policyCache        map[string]report.PolicyReport
	clusterPolicyCache map[string]report.ClusterPolicyReport
	mapper             Mapper
	startUp            time.Time
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

func (c *policyReportClient) WatchClusterPolicyReports(cb report.WatchClusterPolicyReportCallback) error {
	for {
		result, err := c.client.Resource(clusterPolicyReports).Watch(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		for result := range result.ResultChan() {
			if item, ok := result.Object.(*unstructured.Unstructured); ok {
				cb(result.Type, c.mapper.MapClusterPolicyReport(item.Object))
			}
		}
	}
}

func (c *policyReportClient) WatchPolicyReports(cb report.WatchPolicyReportCallback) error {
	for {
		result, err := c.client.Resource(policyReports).Watch(context.Background(), metav1.ListOptions{})
		if err != nil {
			return err
		}

		for result := range result.ResultChan() {
			if item, ok := result.Object.(*unstructured.Unstructured); ok {
				cb(result.Type, c.mapper.MapPolicyReport(item.Object))
			}
		}
	}
}

func (c *policyReportClient) WatchRuleValidation(cb report.WatchPolicyResultCallback, skipExisting bool) error {
	wg := new(errgroup.Group)

	wg.Go(func() error {
		return c.WatchPolicyReports(func(e watch.EventType, pr report.PolicyReport) {
			switch e {
			case watch.Added:
				if skipExisting && pr.CreationTimestamp.Before(c.startUp) {
					c.policyCache[pr.GetIdentifier()] = pr
					break
				}

				for _, result := range pr.Results {
					cb(result)
				}

				c.policyCache[pr.GetIdentifier()] = pr
			case watch.Modified:
				diff := pr.GetNewResults(c.policyCache[pr.GetIdentifier()])
				for _, result := range diff {
					cb(result)
				}

				c.policyCache[pr.GetIdentifier()] = pr
			case watch.Deleted:
				delete(c.policyCache, pr.GetIdentifier())
			}
		})
	})

	wg.Go(func() error {
		return c.WatchClusterPolicyReports(func(s watch.EventType, cpr report.ClusterPolicyReport) {
			switch s {
			case watch.Added:
				if skipExisting && cpr.CreationTimestamp.Before(c.startUp) {
					c.clusterPolicyCache[cpr.GetIdentifier()] = cpr
					break
				}

				for _, result := range cpr.Results {
					cb(result)
				}

				c.clusterPolicyCache[cpr.GetIdentifier()] = cpr
			case watch.Modified:
				diff := cpr.GetNewResults(c.clusterPolicyCache[cpr.GetIdentifier()])
				for _, result := range diff {
					cb(result)
				}

				c.clusterPolicyCache[cpr.GetIdentifier()] = cpr
			case watch.Deleted:
				delete(c.clusterPolicyCache, cpr.GetIdentifier())
			}
		})
	})

	return wg.Wait()
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
func NewPolicyReportClient(ctx context.Context, kubeconfig, namespace string, startUp time.Time) (report.Client, error) {
	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
	}
	if err != nil {
		return nil, err
	}

	dynamicClient, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	coreClient, err := NewCoreClient(kubeconfig, namespace)
	if err != nil {
		return nil, err
	}

	reportClient := &policyReportClient{
		client:             dynamicClient,
		coreClient:         coreClient,
		policyCache:        make(map[string]report.PolicyReport),
		clusterPolicyCache: make(map[string]report.ClusterPolicyReport),
		mapper:             NewMapper(make(map[string]string)),
		startUp:            startUp,
	}

	err = reportClient.fetchPriorities(ctx)
	if err != nil {
		log.Printf("[INFO] No PriorityConfig found: %s", err.Error())
	}

	go reportClient.syncPriorities(ctx)

	return reportClient, nil
}
