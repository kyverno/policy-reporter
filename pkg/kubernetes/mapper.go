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
			res := m.mapResult(resultItem.(map[string]interface{}))
			r.Results[res.GetIdentifier()] = res
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
			res := m.mapResult(resultItem.(map[string]interface{}))
			r.Results[res.GetIdentifier()] = res
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

func (m *mapper) mapResult(result map[string]interface{}) report.Result {
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
		r.Priority = m.resolvePriority(r.Policy)
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

func (m *mapper) resolvePriority(policy string) report.Priority {
	if priority, ok := m.priorityMap[policy]; ok {
		return report.NewPriority(priority)
	}

	if priority, ok := m.priorityMap["default"]; ok {
		return report.NewPriority(priority)
	}

	return report.Priority(report.ErrorPriority)
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
