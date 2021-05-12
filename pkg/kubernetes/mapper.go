package kubernetes

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/watch"
)

// Mapper converts maps into report structs
type Mapper interface {
	// MapPolicyReport maps a map into a PolicyReport
	MapPolicyReport(reportMap map[string]interface{}) report.PolicyReport
	// MapClusterPolicyReport maps a map into a ClusterPolicyReport
	MapClusterPolicyReport(reportMap map[string]interface{}) report.ClusterPolicyReport
	// SetPriorityMap updates the policy/status to priority mapping
	SetPriorityMap(map[string]string)
	// SyncPriorities when ConfigMap has changed
	SyncPriorities(ctx context.Context) error
	// FetchPriorities from ConfigMap
	FetchPriorities(ctx context.Context) error
}

type mapper struct {
	priorityMap map[string]string
	cmAdapter   ConfigMapAdapter
}

func (m *mapper) MapPolicyReport(reportMap map[string]interface{}) report.PolicyReport {
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

	creationTimestamp, err := m.mapCreationTime(reportMap)
	if err == nil {
		r.CreationTimestamp = creationTimestamp
	} else {
		r.CreationTimestamp = time.Now()
	}

	if rs, ok := reportMap["results"].([]interface{}); ok {
		for _, resultItem := range rs {
			resources := m.mapResult(resultItem.(map[string]interface{}))
			for _, resource := range resources {
				r.Results[resource.GetIdentifier()] = resource
			}
		}
	}

	return r
}

func (m *mapper) MapClusterPolicyReport(reportMap map[string]interface{}) report.ClusterPolicyReport {
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

	creationTimestamp, err := m.mapCreationTime(reportMap)
	if err == nil {
		r.CreationTimestamp = creationTimestamp
	} else {
		r.CreationTimestamp = time.Now()
	}

	if rs, ok := reportMap["results"].([]interface{}); ok {
		for _, resultItem := range rs {
			resources := m.mapResult(resultItem.(map[string]interface{}))
			for _, resource := range resources {
				r.Results[resource.GetIdentifier()] = resource
			}
		}
	}

	return r
}

func (m *mapper) SetPriorityMap(priorityMap map[string]string) {
	m.priorityMap = priorityMap
}

func (m *mapper) mapCreationTime(result map[string]interface{}) (time.Time, error) {
	if metadata, ok := result["metadata"].(map[string]interface{}); ok {
		if created, ok2 := metadata["creationTimestamp"].(string); ok2 {
			return time.Parse("2006-01-02T15:04:05Z", created)
		}

		return time.Time{}, errors.New("No creationTimestamp provided")
	}

	return time.Time{}, errors.New("No metadata provided")
}

func (m *mapper) mapResult(result map[string]interface{}) []report.Result {
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

				if ns, ok := resMap["namespace"]; ok {
					r.Namespace = ns.(string)
				}

				resources = append(resources, r)
			}
		}
	}

	var status report.Status

	if s, ok := result["status"]; ok {
		status = s.(report.Status)
	}
	if r, ok := result["result"]; ok {
		status = r.(report.Status)
	}

	var results = []report.Result{}

	factory := func(res report.Resource) report.Result {
		r := report.Result{
			Message:    result["message"].(string),
			Policy:     result["policy"].(string),
			Status:     status,
			Scored:     result["scored"].(bool),
			Priority:   report.PriorityFromStatus(status),
			Resource:   res,
			Properties: make(map[string]string, 0),
		}

		if severity, ok := result["severity"]; ok {
			r.Severity = severity.(report.Severity)
		}

		if r.Status == report.Error || r.Status == report.Fail {
			r.Priority = m.resolvePriority(r.Policy, r.Severity)
		}

		if rule, ok := result["rule"]; ok {
			r.Rule = rule.(string)
		}

		if category, ok := result["category"]; ok {
			r.Category = category.(string)
		}

		r.Timestamp = convertTimestamp(result)

		if props, ok := result["properties"]; ok {
			if properties, ok := props.(map[string]interface{}); ok {
				for property, value := range properties {
					r.Properties[property] = value.(string)
				}
			}
		}

		return r
	}

	for _, resource := range resources {
		results = append(results, factory(resource))
	}

	if len(results) == 0 {
		results = append(results, factory(report.Resource{}))
	}

	return results
}

func convertTimestamp(result map[string]interface{}) time.Time {
	timestamp, ok := result["timestamp"]
	if !ok {
		return time.Now().UTC()
	}

	seconds, ok := timestamp.(map[string]interface{})["seconds"]

	switch s := seconds.(type) {
	case int64:
		return time.Unix(s, 0).UTC()
	case int:
		return time.Unix(int64(s), 0).UTC()
	default:
		return time.Now().UTC()
	}
}

func (m *mapper) resolvePriority(policy string, severity report.Severity) report.Priority {
	if priority, ok := m.priorityMap[policy]; ok {
		return report.NewPriority(priority)
	}

	if severity != "" {
		return report.PriorityFromSeverity(severity)
	}

	if priority, ok := m.priorityMap["default"]; ok {
		return report.NewPriority(priority)
	}

	return report.Priority(report.WarningPriority)
}

func (m *mapper) FetchPriorities(ctx context.Context) error {
	cm, err := m.cmAdapter.GetConfig(ctx, prioriyConfig)
	if err != nil {
		return err
	}

	if cm != nil {
		m.SetPriorityMap(cm.Data)
		log.Println("[INFO] Priorities loaded")
	}

	return nil
}

func (m *mapper) SyncPriorities(ctx context.Context) error {
	err := m.cmAdapter.WatchConfigs(ctx, func(e watch.EventType, cm *v1.ConfigMap) {
		if cm.Name != prioriyConfig {
			return
		}

		switch e {
		case watch.Added:
			m.SetPriorityMap(cm.Data)
		case watch.Modified:
			m.SetPriorityMap(cm.Data)
		case watch.Deleted:
			m.SetPriorityMap(map[string]string{})
		}

		log.Println("[INFO] Priorities synchronized")
	})

	if err != nil {
		log.Printf("[INFO] Unable to sync Priorities: %s", err.Error())
	}

	return err
}

// NewMapper creates an new Mapper instance
func NewMapper(priorities map[string]string, cmAdapter ConfigMapAdapter) Mapper {
	m := &mapper{cmAdapter: cmAdapter}
	m.SetPriorityMap(priorities)

	return m
}
