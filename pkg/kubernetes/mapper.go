package kubernetes

import (
	"errors"
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"
)

const ResultIDKey = "resultID"

// Mapper converts maps into report structs
type Mapper interface {
	// MapPolicyReport maps a map into a PolicyReport
	MapPolicyReport(reportMap map[string]interface{}) *report.PolicyReport
}

type mapper struct {
	priorityMap map[string]string
}

func (m *mapper) MapPolicyReport(reportMap map[string]interface{}) *report.PolicyReport {
	summary := &report.Summary{}

	if s, ok := reportMap["summary"].(map[string]interface{}); ok {
		summary.Pass = int(s["pass"].(int64))
		summary.Skip = int(s["skip"].(int64))
		summary.Warn = int(s["warn"].(int64))
		summary.Error = int(s["error"].(int64))
		summary.Fail = int(s["fail"].(int64))
	}

	metadata, ok := reportMap["metadata"].(map[string]interface{})
	if !ok {
		return &report.PolicyReport{}
	}

	r := &report.PolicyReport{
		Name:      toString(metadata["name"]),
		Namespace: toString(metadata["namespace"]),
		Summary:   summary,
		Results:   make(map[string]*report.Result),
	}

	creationTimestamp, err := m.mapCreationTime(reportMap)
	if err == nil {
		r.CreationTimestamp = creationTimestamp
	} else {
		r.CreationTimestamp = time.Now()
	}

	if rs, ok := reportMap["results"].([]interface{}); ok {
		for _, resultItem := range rs {
			results := m.mapResult(resultItem.(map[string]interface{}))
			for _, result := range results {
				r.Results[result.GetIdentifier()] = result
			}
		}
	}

	r.ID = report.GeneratePolicyReportID(r.Name, r.Namespace)

	return r
}

func (m *mapper) SetPriorityMap(priorityMap map[string]string) {
	m.priorityMap = priorityMap
}

func (m *mapper) mapCreationTime(result map[string]interface{}) (time.Time, error) {
	metadata := result["metadata"].(map[string]interface{})
	if created, ok2 := metadata["creationTimestamp"].(string); ok2 {
		return time.Parse("2006-01-02T15:04:05Z", created)
	}

	return time.Time{}, errors.New("no creationTimestamp provided")
}

func (m *mapper) mapResult(result map[string]interface{}) []*report.Result {
	var resources []*report.Resource

	if ress, ok := result["resources"].([]interface{}); ok {
		for _, res := range ress {
			if resMap, ok := res.(map[string]interface{}); ok {
				r := &report.Resource{
					Namespace:  toString(resMap["namespace"]),
					APIVersion: toString(resMap["apiVersion"]),
					Kind:       toString(resMap["kind"]),
					Name:       toString(resMap["name"]),
					UID:        toString(resMap["uid"]),
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

	var results []*report.Result

	factory := func(res *report.Resource) *report.Result {
		r := &report.Result{
			Policy:     toString(result["policy"]),
			Status:     status,
			Priority:   report.PriorityFromStatus(status),
			Resource:   res,
			Properties: make(map[string]string, 0),
		}

		if scored, ok := result["scored"]; ok {
			r.Scored = scored.(bool)
		}

		if severity, ok := result["severity"]; ok {
			r.Severity = severity.(report.Severity)
		}

		if r.Status == report.Fail {
			r.Priority = m.resolvePriority(r.Policy, r.Severity)
		}

		r.Message = toString(result["message"])
		r.Rule = toString(result["rule"])
		r.Category = toString(result["category"])
		r.Source = toString(result["source"])
		r.Timestamp = convertTimestamp(result)

		if props, ok := result["properties"]; ok {
			if properties, ok := props.(map[string]interface{}); ok {
				if id, ok := properties[ResultIDKey]; ok {
					r.ID = toString(id)
					delete(properties, ResultIDKey)
				}

				for property, v := range properties {
					if property == ResultIDKey {
						continue
					}

					value := toString(v)
					if len(value) > 0 {
						r.Properties[property] = value
					}
				}
			}
		}

		if r.ID == "" {
			r.ID = report.GeneratePolicyReportResultID(r.Resource.UID, r.Resource.Name, r.Policy, r.Rule, r.Status, r.Message)
		}

		return r
	}

	for _, resource := range resources {
		results = append(results, factory(resource))
	}

	if len(results) == 0 {
		results = append(results, factory(&report.Resource{}))
	}

	return results
}

func convertTimestamp(result map[string]interface{}) time.Time {
	timestamp, ok := result["timestamp"]
	if !ok {
		return time.Now().UTC()
	}

	seconds, _ := timestamp.(map[string]interface{})["seconds"]

	switch s := seconds.(type) {
	case int64:
		return time.Unix(s, 0).UTC()
	case int:
		return time.Unix(int64(s), 0).UTC()
	default:
		return time.Now().UTC()
	}
}

func toString(value interface{}) string {
	if v, ok := value.(string); ok {
		return v
	}

	return ""
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

// NewMapper creates an new Mapper instance
func NewMapper(priorities map[string]string) Mapper {
	m := &mapper{}
	m.SetPriorityMap(priorities)

	return m
}
