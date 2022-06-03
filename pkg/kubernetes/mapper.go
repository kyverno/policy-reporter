package kubernetes

import (
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"

	"github.com/kyverno/kyverno/api/policyreport/v1alpha2"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ResultIDKey = "resultID"

// Mapper converts maps into report structs
type Mapper interface {
	// MapPolicyReport maps a v1alpha2.PolicyReport into a PolicyReport
	MapPolicyReport(*v1alpha2.PolicyReport) *report.PolicyReport
	// MapClusterPolicyReport maps a v1alpha2.ClusterPolicyReport into a PolicyReport
	MapClusterPolicyReport(*v1alpha2.ClusterPolicyReport) *report.PolicyReport
}

type mapper struct {
	priorityMap map[string]string
}

func (m *mapper) MapPolicyReport(preport *v1alpha2.PolicyReport) *report.PolicyReport {
	r := &report.PolicyReport{
		Name:              preport.Name,
		Namespace:         preport.Namespace,
		Summary:           m.mapSummary(preport.Summary),
		Results:           make(map[string]*report.Result),
		CreationTimestamp: preport.CreationTimestamp.Time,
	}

	for _, resultItem := range preport.Results {
		results := m.mapResult(resultItem.DeepCopy())
		for _, result := range results {
			r.Results[result.GetIdentifier()] = result
		}
	}

	r.ID = report.GeneratePolicyReportID(r.Name, r.Namespace)

	return r
}

func (m *mapper) MapClusterPolicyReport(creport *v1alpha2.ClusterPolicyReport) *report.PolicyReport {
	r := &report.PolicyReport{
		Name:              creport.Name,
		Summary:           m.mapSummary(creport.Summary),
		Results:           make(map[string]*report.Result),
		CreationTimestamp: creport.CreationTimestamp.Time,
	}

	for _, resultItem := range creport.Results {
		results := m.mapResult(resultItem.DeepCopy())
		for _, result := range results {
			r.Results[result.GetIdentifier()] = result
		}
	}

	r.ID = report.GeneratePolicyReportID(r.Name, r.Namespace)

	return r
}

func (m *mapper) SetPriorityMap(priorityMap map[string]string) {
	m.priorityMap = priorityMap
}

func (m *mapper) mapSummary(sum v1alpha2.PolicyReportSummary) *report.Summary {
	summary := &report.Summary{}
	summary.Pass = sum.Pass
	summary.Skip = sum.Skip
	summary.Warn = sum.Warn
	summary.Error = sum.Error
	summary.Fail = sum.Fail

	return summary
}

func (m *mapper) mapResult(result *v1alpha2.PolicyReportResult) []*report.Result {
	var resources []*report.Resource

	for _, res := range result.Resources {
		r := &report.Resource{
			Namespace:  res.Namespace,
			APIVersion: res.APIVersion,
			Kind:       res.Kind,
			Name:       res.Name,
			UID:        string(res.UID),
		}

		resources = append(resources, r)
	}

	var results []*report.Result

	factory := func(res *report.Resource) *report.Result {
		status := string(result.Result)

		r := &report.Result{
			Policy:     result.Policy,
			Status:     string(result.Result),
			Priority:   report.PriorityFromStatus(status),
			Resource:   res,
			Properties: make(map[string]string),
			Scored:     result.Scored,
			Severity:   string(result.Severity),
			Message:    result.Message,
			Rule:       result.Rule,
			Category:   result.Category,
			Source:     result.Source,
			Timestamp:  convertTimestamp(result.Timestamp),
		}

		if r.Status == report.Fail {
			r.Priority = m.resolvePriority(r.Policy, r.Severity)
		}

		if id, ok := result.Properties[ResultIDKey]; ok {
			r.ID = id
			delete(result.Properties, ResultIDKey)
		}

		for property, value := range result.Properties {
			if len(value) > 0 {
				r.Properties[property] = value
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

func convertTimestamp(timestamp v1.Timestamp) time.Time {
	if timestamp.Seconds > 0 {
		return time.Unix(timestamp.Seconds, int64(timestamp.Nanos)).UTC()
	}

	return time.Now().UTC()
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
