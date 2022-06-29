package kubernetes

import (
	"time"

	"github.com/kyverno/policy-reporter/pkg/report"

	"github.com/kyverno/kyverno/api/policyreport/v1alpha2"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const ResultIDKey = "resultID"

// Mapper converts maps into report structs
type Mapper interface {
	// MapPolicyReport maps a v1alpha2.PolicyReport into a PolicyReport
	MapPolicyReport(*v1alpha2.PolicyReport) report.PolicyReport
	// MapClusterPolicyReport maps a v1alpha2.ClusterPolicyReport into a PolicyReport
	MapClusterPolicyReport(*v1alpha2.ClusterPolicyReport) report.PolicyReport
}

type mapper struct {
	priorityMap map[string]string
}

func (m *mapper) MapPolicyReport(preport *v1alpha2.PolicyReport) report.PolicyReport {
	r := report.PolicyReport{
		Name:              preport.Name,
		Namespace:         preport.Namespace,
		Summary:           m.mapSummary(preport.Summary),
		Results:           make([]report.Result, 0),
		CreationTimestamp: preport.CreationTimestamp.Time,
	}

	for _, result := range preport.Results {
		if len(result.Resources) == 0 {
			r.Results = append(r.Results, m.mapResult(result, report.Resource{}))
			continue
		}

		for _, res := range result.Resources {
			r.Results = append(r.Results, m.mapResult(result, mapResource(res)))
		}
	}

	r.ID = report.GeneratePolicyReportID(r.Name, r.Namespace)

	return r
}

func (m *mapper) MapClusterPolicyReport(creport *v1alpha2.ClusterPolicyReport) report.PolicyReport {
	r := report.PolicyReport{
		Name:              creport.Name,
		Summary:           m.mapSummary(creport.Summary),
		Results:           make([]report.Result, 0),
		CreationTimestamp: creport.CreationTimestamp.Time,
	}

	for _, result := range creport.Results {
		if len(result.Resources) == 0 {
			r.Results = append(r.Results, m.mapResult(result, report.Resource{}))
			continue
		}

		for _, res := range result.Resources {
			r.Results = append(r.Results, m.mapResult(result, mapResource(res)))
		}
	}

	r.ID = report.GeneratePolicyReportID(r.Name, r.Namespace)

	return r
}

func (m *mapper) SetPriorityMap(priorityMap map[string]string) {
	m.priorityMap = priorityMap
}

func (m *mapper) mapSummary(sum v1alpha2.PolicyReportSummary) report.Summary {
	summary := report.Summary{}
	summary.Pass = sum.Pass
	summary.Skip = sum.Skip
	summary.Warn = sum.Warn
	summary.Error = sum.Error
	summary.Fail = sum.Fail

	return summary
}

func mapResource(res corev1.ObjectReference) report.Resource {
	return report.Resource{
		Namespace:  res.Namespace,
		APIVersion: res.APIVersion,
		Kind:       res.Kind,
		Name:       res.Name,
		UID:        string(res.UID),
	}
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

func (m *mapper) mapResult(result v1alpha2.PolicyReportResult, res report.Resource) report.Result {
	status := string(result.Result)

	r := report.Result{
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
	}

	for property, value := range result.Properties {
		if property == ResultIDKey {
			continue
		}
		if value != "" {
			r.Properties[property] = value
		}
	}

	if r.ID == "" {
		r.ID = report.GeneratePolicyReportResultID(r.Resource.UID, r.Resource.Name, r.Policy, r.Rule, r.Status, r.Message)
	}

	return r
}

// NewMapper creates an new Mapper instance
func NewMapper(priorities map[string]string) Mapper {
	m := &mapper{}
	m.SetPriorityMap(priorities)

	return m
}
