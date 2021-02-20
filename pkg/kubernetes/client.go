package kubernetes

import (
	"context"
	"log"

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
	policyReports = schema.GroupVersionResource{Group: "wgpolicyk8s.io", Version: "v1alpha1", Resource: "policyreports"}
)

type WatchPolicyReportCallback = func(watch.EventType, report.PolicyReport)
type WatchPolicyResultCallback = func(report.Result)

type Client interface {
	FetchPolicyReports() []report.PolicyReport
	WatchPolicyReports(WatchPolicyReportCallback)
	WatchRuleValidation(WatchPolicyResultCallback)
}

type DynamicClient struct {
	client      dynamic.Interface
	reportCache map[string]report.PolicyReport
	priorityMap map[string]report.Priority
}

func (c *DynamicClient) FetchPolicyReports() []report.PolicyReport {
	var reports []report.PolicyReport

	result, err := c.client.Resource(policyReports).List(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("K8s List Error: %s\n", err.Error())
		return reports
	}

	for _, item := range result.Items {
		reports = append(reports, c.mapReport(item.Object))
	}

	return reports
}

func (c *DynamicClient) WatchPolicyReports(cb WatchPolicyReportCallback) {
	result, err := c.client.Resource(policyReports).Watch(context.Background(), metav1.ListOptions{})
	if err != nil {
		log.Printf("K8s Watch Error: %s\n", err.Error())
		return
	}

	for result := range result.ResultChan() {
		if item, ok := result.Object.(*unstructured.Unstructured); ok {
			cb(result.Type, c.mapReport(item.Object))
		}
	}
}

func (c *DynamicClient) WatchRuleValidation(cb WatchPolicyResultCallback) {
	c.WatchPolicyReports(func(s watch.EventType, pr report.PolicyReport) {
		switch s {
		case watch.Added:
			for _, result := range pr.Results {
				cb(result)
			}

			c.reportCache[pr.GetIdentifier()] = pr
		case watch.Modified:
			diff := pr.GetNewValidation(c.reportCache[pr.GetIdentifier()])
			for _, result := range diff {
				cb(result)
			}

			c.reportCache[pr.GetIdentifier()] = pr
		case watch.Deleted:
			delete(c.reportCache, pr.GetIdentifier())
		}
	})
}

func NewDynamicClient(kubeconfig string, prioties map[string]report.Priority) (Client, error) {
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

	return &DynamicClient{client, make(map[string]report.PolicyReport), prioties}, nil
}

func (c *DynamicClient) mapReport(reportMap map[string]interface{}) report.PolicyReport {
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

func (c *DynamicClient) mapResult(result map[string]interface{}) report.Result {
	var resources []report.Resource

	if ress, ok := result["resources"].([]interface{}); ok {
		for _, res := range ress {
			if resMap, ok := res.(map[string]interface{}); ok {
				resources = append(resources, report.Resource{
					APIVersion: resMap["apiVersion"].(string),
					Kind:       resMap["kind"].(string),
					Name:       resMap["name"].(string),
					Namespace:  resMap["namespace"].(string),
					UID:        resMap["uid"].(string),
				})
			}
		}
	}

	r := report.Result{
		Message:   result["message"].(string),
		Policy:    result["policy"].(string),
		Status:    result["status"].(report.Status),
		Scored:    result["scored"].(bool),
		Priority:  report.Alert,
		Resources: resources,
	}

	if priority, ok := c.priorityMap[r.Policy]; ok {
		r.Priority = priority
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
