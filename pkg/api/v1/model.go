package v1

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	db "github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/helper"
	"github.com/kyverno/policy-reporter/pkg/target"
)

// Target API Model
type Target struct {
	Name                  string   `json:"name"`
	MinimumPriority       string   `json:"minimumPriority"`
	Sources               []string `json:"sources,omitempty"`
	SkipExistingOnStartup bool     `json:"skipExistingOnStartup"`
}

func mapTarget(t target.Client) Target {
	minPrio := t.MinimumPriority()
	if minPrio == "" {
		minPrio = v1alpha2.DebugPriority.String()
	}

	return Target{
		Name:                  t.Name(),
		MinimumPriority:       minPrio,
		Sources:               t.Sources(),
		SkipExistingOnStartup: t.SkipExistingOnStartup(),
	}
}

type PolicyReport struct {
	ID        string            `json:"id"`
	Name      string            `json:"name"`
	Namespace string            `json:"namespace,omitempty"`
	Source    string            `json:"source"`
	Labels    map[string]string `json:"labels"`
	Pass      int               `json:"pass"`
	Skip      int               `json:"skip"`
	Warn      int               `json:"warn"`
	Error     int               `json:"error"`
	Fail      int               `json:"fail"`
}

func MapPolicyReports(results []db.PolicyReport) []PolicyReport {
	return helper.Map(results, func(res db.PolicyReport) PolicyReport {
		return PolicyReport{
			ID:        res.ID,
			Name:      res.Name,
			Namespace: res.Namespace,
			Source:    res.Source,
			Labels:    res.Labels,
			Pass:      res.Pass,
			Fail:      res.Fail,
			Warn:      res.Warn,
			Error:     res.Error,
			Skip:      res.Skip,
		}
	})
}

type StatusCount struct {
	Source string `json:"source,omitempty"`
	Status string `json:"status"`
	Count  int    `json:"count"`
}

func MapRuleStatusCounts(results []db.StatusCount) []StatusCount {
	mapping := map[string]StatusCount{
		v1alpha2.StatusPass:  {Status: v1alpha2.StatusPass},
		v1alpha2.StatusFail:  {Status: v1alpha2.StatusFail},
		v1alpha2.StatusWarn:  {Status: v1alpha2.StatusWarn},
		v1alpha2.StatusError: {Status: v1alpha2.StatusError},
		v1alpha2.StatusSkip:  {Status: v1alpha2.StatusSkip},
	}

	for _, result := range results {
		mapping[result.Status] = StatusCount{Status: result.Status, Count: result.Count}
	}

	return helper.ToList(mapping)
}

func MapClusterStatusCounts(results []db.StatusCount, status []string) []StatusCount {
	var mapping map[string]StatusCount

	if len(status) == 0 {
		mapping = map[string]StatusCount{
			v1alpha2.StatusPass:  {Status: v1alpha2.StatusPass},
			v1alpha2.StatusFail:  {Status: v1alpha2.StatusFail},
			v1alpha2.StatusWarn:  {Status: v1alpha2.StatusWarn},
			v1alpha2.StatusError: {Status: v1alpha2.StatusError},
			v1alpha2.StatusSkip:  {Status: v1alpha2.StatusSkip},
		}
	} else {
		mapping = map[string]StatusCount{}

		for _, status := range status {
			mapping[status] = StatusCount{Status: status}
		}
	}
	for _, result := range results {
		mapping[result.Status] = StatusCount{Status: result.Status, Count: result.Count}
	}

	return helper.ToList(mapping)
}

type NamespaceCount struct {
	Namespace string `json:"namespace"`
	Count     int    `json:"count"`
	Status    string `json:"-"`
}

type NamespaceStatusCount struct {
	Status string           `json:"status"`
	Items  []NamespaceCount `json:"items"`
}

func MapNamespaceStatusCounts(results []db.StatusCount, status []string) []NamespaceStatusCount {
	var mapping map[string][]NamespaceCount

	if len(status) == 0 {
		mapping = map[string][]NamespaceCount{
			v1alpha2.StatusPass:  make([]NamespaceCount, 0),
			v1alpha2.StatusFail:  make([]NamespaceCount, 0),
			v1alpha2.StatusWarn:  make([]NamespaceCount, 0),
			v1alpha2.StatusError: make([]NamespaceCount, 0),
			v1alpha2.StatusSkip:  make([]NamespaceCount, 0),
		}
	} else {
		mapping = map[string][]NamespaceCount{}

		for _, status := range status {
			mapping[status] = make([]NamespaceCount, 0)
		}
	}
	for _, result := range results {
		mapping[result.Status] = append(mapping[result.Status], NamespaceCount{Status: result.Status, Count: result.Count, Namespace: result.Namespace})
	}

	statusCounts := make([]NamespaceStatusCount, 0, 5)
	for status, items := range mapping {
		statusCounts = append(statusCounts, NamespaceStatusCount{
			Status: status,
			Items:  items,
		})
	}

	return statusCounts
}

type Resource struct {
	ID         string `json:"id,omitempty"`
	UID        string `json:"uid,omitempty"`
	Name       string `json:"name,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

func MapResource(results []db.ResourceResult) []Resource {
	return helper.Map(results, func(res db.ResourceResult) Resource {
		return Resource{
			ID:         res.ID,
			UID:        res.Resource.UID,
			Name:       res.Resource.Name,
			Namespace:  res.Resource.Namespace,
			Kind:       res.Resource.Kind,
			APIVersion: res.Resource.APIVersion,
		}
	})
}

type Result struct {
	ID         string            `json:"id"`
	Namespace  string            `json:"namespace,omitempty"`
	Kind       string            `json:"kind"`
	APIVersion string            `json:"apiVersion"`
	Name       string            `json:"name"`
	ResourceID string            `json:"resourceId"`
	Message    string            `json:"message"`
	Category   string            `json:"category,omitempty"`
	Policy     string            `json:"policy"`
	Rule       string            `json:"rule"`
	Status     string            `json:"status"`
	Severity   string            `json:"severity,omitempty"`
	Timestamp  int64             `json:"timestamp,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

func MapResults(results []db.PolicyReportResult) []Result {
	return helper.Map(results, func(res db.PolicyReportResult) Result {
		return Result{
			ID:         res.ID,
			Namespace:  res.Resource.Namespace,
			Kind:       res.Resource.Kind,
			APIVersion: res.Resource.APIVersion,
			Name:       res.Resource.Name,
			ResourceID: res.Resource.GetID(),
			Message:    res.Message,
			Category:   res.Category,
			Policy:     res.Policy,
			Rule:       res.Rule,
			Status:     res.Result,
			Severity:   res.Severity,
			Timestamp:  res.Created,
			Properties: res.Properties,
		}
	})
}
