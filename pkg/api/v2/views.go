package v2

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	db "github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/helper"
)

type Category struct {
	Name  string `json:"name"`
	Pass  int    `json:"pass"`
	Skip  int    `json:"skip"`
	Warn  int    `json:"warn"`
	Error int    `json:"error"`
	Fail  int    `json:"fail"`
}

type SourceDetails struct {
	Name       string     `json:"name"`
	Categories []Category `json:"categories"`
}

func MapToSourceDetails(categories []db.Category) []*SourceDetails {
	list := make(map[string]*SourceDetails, 0)

	for _, r := range categories {
		if s, ok := list[r.Source]; ok {
			s.Categories = append(s.Categories, Category{
				Name:  r.Name,
				Pass:  r.Pass,
				Fail:  r.Fail,
				Warn:  r.Warn,
				Error: r.Error,
				Skip:  r.Skip,
			})
			continue
		}

		list[r.Source] = &SourceDetails{
			Name: r.Source,
			Categories: []Category{{
				Name:  r.Name,
				Pass:  r.Pass,
				Fail:  r.Fail,
				Warn:  r.Warn,
				Error: r.Error,
				Skip:  r.Skip,
			}},
		}
	}

	return helper.ToList(list)
}

type Resource struct {
	ID         string `json:"id,omitempty"`
	UID        string `json:"uid,omitempty"`
	Name       string `json:"name,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
}

func MapResource(result db.ResourceResult) Resource {
	return Resource{
		ID:         result.ID,
		UID:        result.Resource.UID,
		APIVersion: result.Resource.APIVersion,
		Kind:       result.Resource.Kind,
		Name:       result.Resource.Name,
		Namespace:  result.Resource.Namespace,
	}
}

type ResourceStatusCount struct {
	Source string `json:"source,omitempty"`
	Pass   int    `json:"pass"`
	Warn   int    `json:"warn"`
	Fail   int    `json:"fail"`
	Error  int    `json:"error"`
	Skip   int    `json:"skip"`
}

func MapResourceStatusCounts(results []db.ResourceStatusCount) []ResourceStatusCount {
	list := make([]ResourceStatusCount, 0, len(results))
	for _, result := range results {
		list = append(list, ResourceStatusCount{
			Source: result.Source,
			Pass:   result.Pass,
			Fail:   result.Fail,
			Warn:   result.Warn,
			Error:  result.Error,
			Skip:   result.Skip,
		})
	}

	return list
}

type ResourceResult struct {
	ID         string `json:"id"`
	UID        string `json:"uid"`
	Name       string `json:"name"`
	Kind       string `json:"kind"`
	APIVersion string `json:"apiVersion"`
	Namespace  string `json:"namespace,omitempty"`
	Source     string `json:"source,omitempty"`
	Pass       int    `json:"pass"`
	Skip       int    `json:"skip"`
	Warn       int    `json:"warn"`
	Fail       int    `json:"fail"`
	Error      int    `json:"error"`
}

func MapResourceResults(results []db.ResourceResult) []ResourceResult {
	return helper.Map(results, func(res db.ResourceResult) ResourceResult {
		return ResourceResult{
			ID:         res.ID,
			UID:        res.Resource.UID,
			Namespace:  res.Resource.Namespace,
			Kind:       res.Resource.Kind,
			APIVersion: res.Resource.APIVersion,
			Name:       res.Resource.Name,
			Source:     res.Source,
			Pass:       res.Pass,
			Skip:       res.Skip,
			Warn:       res.Warn,
			Fail:       res.Fail,
			Error:      res.Error,
		}
	})
}

type Paginated[T any] struct {
	Items []T `json:"items"`
	Count int `json:"count"`
}

type StatusCount struct {
	Namespace string `json:"namespace,omitempty"`
	Source    string `json:"source,omitempty"`
	Status    string `json:"status"`
	Count     int    `json:"count"`
}

func MapClusterStatusCounts(results []db.StatusCount) map[string]int {
	mapping := map[string]int{
		v1alpha2.StatusPass:  0,
		v1alpha2.StatusFail:  0,
		v1alpha2.StatusWarn:  0,
		v1alpha2.StatusError: 0,
		v1alpha2.StatusSkip:  0,
	}

	for _, result := range results {
		mapping[result.Status] = result.Count
	}

	return mapping
}

func MapNamespaceStatusCounts(results []db.StatusCount) map[string]map[string]int {
	mapping := map[string]map[string]int{}

	for _, result := range results {
		if _, ok := mapping[result.Namespace]; !ok {
			mapping[result.Namespace] = map[string]int{
				v1alpha2.StatusPass:  0,
				v1alpha2.StatusFail:  0,
				v1alpha2.StatusWarn:  0,
				v1alpha2.StatusError: 0,
				v1alpha2.StatusSkip:  0,
			}
		}

		mapping[result.Namespace][result.Status] = result.Count
	}

	return mapping
}

type Policy struct {
	Source   string         `json:"source,omitempty"`
	Category string         `json:"category,omitempty"`
	Name     string         `json:"policy"`
	Severity string         `json:"severity,omitempty"`
	Results  map[string]int `json:"results"`
}

func MapPolicies(results []db.PolicyReportFilter) []*Policy {
	list := make(map[string]*Policy)

	for _, r := range results {
		category := r.Category
		if category == "" {
			category = "Other"
		}

		if _, ok := list[r.Policy]; ok {
			list[r.Policy].Results[r.Result] = r.Count
			continue
		}

		list[r.Policy] = &Policy{
			Source:   r.Source,
			Category: category,
			Name:     r.Policy,
			Severity: r.Severity,
			Results: map[string]int{
				r.Result: r.Count,
			},
		}
	}

	return helper.ToList(list)
}

type PolicyResult struct {
	ID         string            `json:"id"`
	Namespace  string            `json:"namespace,omitempty"`
	Kind       string            `json:"kind"`
	APIVersion string            `json:"apiVersion"`
	Name       string            `json:"name"`
	Message    string            `json:"message"`
	Category   string            `json:"category,omitempty"`
	Policy     string            `json:"policy"`
	Rule       string            `json:"rule"`
	Status     string            `json:"status"`
	Severity   string            `json:"severity,omitempty"`
	Timestamp  int64             `json:"timestamp,omitempty"`
	Properties map[string]string `json:"properties,omitempty"`
}

func MapPolicyResults(results []db.PolicyReportResult) []PolicyResult {
	return helper.Map(results, func(res db.PolicyReportResult) PolicyResult {
		return PolicyResult{
			ID:         res.ID,
			Namespace:  res.Resource.Namespace,
			Kind:       res.Resource.Kind,
			APIVersion: res.Resource.APIVersion,
			Name:       res.Resource.Name,
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

type FindingCounts struct {
	Total  int            `json:"total"`
	Source string         `json:"source"`
	Counts map[string]int `json:"counts"`
}

type Findings struct {
	Total  int              `json:"total"`
	Counts []*FindingCounts `json:"counts"`
}

func MapFindings(results []db.StatusCount) Findings {
	findings := make(map[string]*FindingCounts, 0)
	total := 0

	for _, count := range results {
		if finding, ok := findings[count.Source]; ok {
			finding.Counts[count.Status] = count.Count
			finding.Total = finding.Total + count.Count
		} else {
			findings[count.Source] = &FindingCounts{
				Source: count.Source,
				Total:  count.Count,
				Counts: map[string]int{
					count.Status: count.Count,
				},
			}
		}

		total += count.Count
	}

	return Findings{Counts: helper.ToList(findings), Total: total}
}
