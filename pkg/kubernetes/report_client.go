package kubernetes

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
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
	priorityMap        map[string]string
	startUp            time.Time
}

func (c *policyReportClient) FetchPolicyReports() []report.PolicyReport {
	var reports []report.PolicyReport

	result, err := c.client.Resource(policyReports).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("K8s List Error: %s\n", err.Error())
		return reports
	}

	for _, item := range result.Items {
		reports = append(reports, c.mapPolicyReport(item.Object))
	}

	return reports
}

func (c *policyReportClient) WatchClusterPolicyReports(cb report.WatchClusterPolicyReportCallback) {
	for {
		result, err := c.client.Resource(clusterPolicyReports).Watch(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Printf("K8s Watch Error: %s\n", err.Error())
			return
		}

		for result := range result.ResultChan() {
			if item, ok := result.Object.(*unstructured.Unstructured); ok {
				cb(result.Type, c.mapClusterPolicyReport(item.Object))
			}
		}

		log.Println("[WARNING] WatchClusterPolicyReports Stops")
	}
}

func (c *policyReportClient) WatchPolicyReports(cb report.WatchPolicyReportCallback) {
	for {
		result, err := c.client.Resource(policyReports).Watch(context.Background(), metav1.ListOptions{})
		if err != nil {
			log.Printf("K8s Watch Error: %s\n", err.Error())
			return
		}

		for result := range result.ResultChan() {
			if item, ok := result.Object.(*unstructured.Unstructured); ok {
				cb(result.Type, c.mapPolicyReport(item.Object))
			}
		}

		log.Println("[WARNING] WatchPolicyReports Stops")
	}
}

func (c *policyReportClient) WatchRuleValidation(cb report.WatchPolicyResultCallback, skipExisting bool) {
	wg := sync.WaitGroup{}
	wg.Add(2)

	go func(skipExisting bool) {
		c.WatchPolicyReports(func(e watch.EventType, pr report.PolicyReport) {
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
				diff := pr.GetNewValidation(c.policyCache[pr.GetIdentifier()])
				for _, result := range diff {
					cb(result)
				}

				c.policyCache[pr.GetIdentifier()] = pr
			case watch.Deleted:
				delete(c.policyCache, pr.GetIdentifier())
			}
		})

		wg.Done()
	}(skipExisting)

	go func(skipExisting bool) {
		c.WatchClusterPolicyReports(func(s watch.EventType, cpr report.ClusterPolicyReport) {
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
				diff := cpr.GetNewValidation(c.clusterPolicyCache[cpr.GetIdentifier()])
				for _, result := range diff {
					cb(result)
				}

				c.clusterPolicyCache[cpr.GetIdentifier()] = cpr
			case watch.Deleted:
				delete(c.clusterPolicyCache, cpr.GetIdentifier())
			}
		})

		wg.Done()
	}(skipExisting)

	wg.Wait()
}

func (c *policyReportClient) fetchPriorities(ctx context.Context) error {
	cm, err := c.coreClient.GetConfig(ctx, prioriyConfig)
	if err != nil {
		return err
	}

	if cm != nil {
		c.priorityMap = cm.Data
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
			c.priorityMap = cm.Data
		case watch.Modified:
			c.priorityMap = cm.Data
		case watch.Deleted:
			c.priorityMap = map[string]string{}
		}

		log.Println("[INFO] Priorities synchronized")
	})

	if err != nil {
		log.Printf("[INFO] Unable to sync Priorities: %s", err.Error())
	}

	return err
}

func (c *policyReportClient) mapPolicyReport(reportMap map[string]interface{}) report.PolicyReport {
	summary := report.Summary{}

	if s, ok := reportMap["summary"].(map[string]interface{}); ok {
		summary.Pass = int(s["pass"].(int64))
		summary.Skip = int(s["skip"].(int64))
		summary.Warn = int(s["warn"].(int64))
		summary.Error = int(s["error"].(int64))
		summary.Fail = int(s["fail"].(int64))
	}

	r := report.PolicyReport{
		Name:      reportMap["metadata"].(map[string]interface{})["name"].(string),
		Namespace: reportMap["metadata"].(map[string]interface{})["namespace"].(string),
		Summary:   summary,
		Results:   make(map[string]report.Result),
	}

	if rs, ok := reportMap["results"].([]interface{}); ok {
		for _, resultItem := range rs {
			res := c.mapResult(resultItem.(map[string]interface{}))
			r.Results[res.GetIdentifier()] = res
		}
	}

	return r
}

func (c *policyReportClient) mapClusterPolicyReport(reportMap map[string]interface{}) report.ClusterPolicyReport {
	summary := report.Summary{}

	if s, ok := reportMap["summary"].(map[string]interface{}); ok {
		summary.Pass = int(s["pass"].(int64))
		summary.Skip = int(s["skip"].(int64))
		summary.Warn = int(s["warn"].(int64))
		summary.Error = int(s["error"].(int64))
		summary.Fail = int(s["fail"].(int64))
	}

	r := report.ClusterPolicyReport{
		Name:    reportMap["metadata"].(map[string]interface{})["name"].(string),
		Summary: summary,
		Results: make(map[string]report.Result),
	}

	creationTimestamp, err := c.mapCreationTime(reportMap)
	if err == nil {
		r.CreationTimestamp = creationTimestamp
	} else {
		r.CreationTimestamp = time.Now()
	}

	if rs, ok := reportMap["results"].([]interface{}); ok {
		for _, resultItem := range rs {
			res := c.mapResult(resultItem.(map[string]interface{}))
			r.Results[res.GetIdentifier()] = res
		}
	}

	return r
}

func (c *policyReportClient) mapCreationTime(result map[string]interface{}) (time.Time, error) {
	if metadata, ok := result["metadata"].(map[string]interface{}); ok {
		if created, ok2 := metadata["creationTimestamp"].(string); ok2 {
			return time.Parse("2006-01-02T15:04:05Z", created)
		}

		return time.Time{}, errors.New("No creationTimestamp provided")
	}

	return time.Time{}, errors.New("No metadata provided")
}

func (c *policyReportClient) mapResult(result map[string]interface{}) report.Result {
	var resources []report.Resource

	if ress, ok := result["resources"].([]interface{}); ok {
		for _, res := range ress {
			if resMap, ok := res.(map[string]interface{}); ok {
				r := report.Resource{
					APIVersion: resMap["apiVersion"].(string),
					Kind:       resMap["kind"].(string),
					Name:       resMap["name"].(string),
					UID:        resMap["uid"].(string),
				}

				if ns, ok := result["namespace"]; ok {
					r.Namespace = ns.(string)
				}

				resources = append(resources, r)
			}
		}
	}

	status := result["status"].(report.Status)

	r := report.Result{
		Message:   result["message"].(string),
		Policy:    result["policy"].(string),
		Status:    status,
		Scored:    result["scored"].(bool),
		Priority:  report.PriorityFromStatus(status),
		Resources: resources,
	}

	if r.Status == report.Error || r.Status == report.Fail {
		if priority, ok := c.priorityMap[r.Policy]; ok {
			r.Priority = report.NewPriority(priority)
		}
	}

	if rule, ok := result["rule"]; ok {
		r.Rule = rule.(string)
	}

	if category, ok := result["category"]; ok {
		r.Category = category.(string)
	}

	if severity, ok := result["severity"]; ok {
		r.Severity = severity.(report.Severity)
	}

	return r
}

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
		priorityMap:        make(map[string]string),
		startUp:            startUp,
	}

	err = reportClient.fetchPriorities(ctx)
	if err != nil {
		log.Printf("[INFO] No PriorityConfig found: %s", err.Error())
	}

	go reportClient.syncPriorities(ctx)

	return reportClient, nil
}
