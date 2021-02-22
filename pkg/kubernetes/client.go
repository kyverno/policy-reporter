package kubernetes

import (
	"context"
	"errors"
	"log"
	"sync"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
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

type WatchPolicyReportCallback = func(watch.EventType, report.PolicyReport)
type WatchClusterPolicyReportCallback = func(watch.EventType, report.ClusterPolicyReport)
type WatchPolicyResultCallback = func(report.Result)

type Client interface {
	FetchPolicyReports() []report.PolicyReport
	WatchPolicyReports(WatchPolicyReportCallback)
	WatchRuleValidation(WatchPolicyResultCallback, bool)
	WatchClusterPolicyReports(WatchClusterPolicyReportCallback)
}

type DynamicClient struct {
	client             dynamic.Interface
	policyCache        map[string]report.PolicyReport
	clusterPolicyCache map[string]report.ClusterPolicyReport
	priorityMap        map[string]string
	startUp            time.Time
}

func (c *DynamicClient) FetchPolicyReports() []report.PolicyReport {
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

func (c *DynamicClient) WatchClusterPolicyReports(cb WatchClusterPolicyReportCallback) {
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
}

func (c *DynamicClient) WatchPolicyReports(cb WatchPolicyReportCallback) {
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
}

func (c *DynamicClient) WatchRuleValidation(cb WatchPolicyResultCallback, skipExisting bool) {
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

func NewDynamicClient(kubeconfig string, prioties map[string]string, startUp time.Time) (Client, error) {
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

	client, err := dynamic.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &DynamicClient{
		client:             client,
		policyCache:        make(map[string]report.PolicyReport),
		clusterPolicyCache: make(map[string]report.ClusterPolicyReport),
		priorityMap:        prioties,
		startUp:            startUp,
	}, nil
}

func (c *DynamicClient) mapPolicyReport(reportMap map[string]interface{}) report.PolicyReport {
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

func (c *DynamicClient) mapClusterPolicyReport(reportMap map[string]interface{}) report.ClusterPolicyReport {
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

func (c *DynamicClient) mapCreationTime(result map[string]interface{}) (time.Time, error) {
	if metadata, ok := result["metadata"].(map[string]interface{}); ok {
		if created, ok2 := metadata["creationTimestamp"].(string); ok2 {
			return time.Parse("2006-01-02T15:04:05Z", created)
		}

		return time.Time{}, errors.New("No creationTimestamp provided")
	}

	return time.Time{}, errors.New("No metadata provided")
}

func (c *DynamicClient) mapResult(result map[string]interface{}) report.Result {
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
