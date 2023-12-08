package v1

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/target"
)

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
	Type      string            `json:"-"`
	Created   int64             `json:"-"`
}

type PolicyReportList struct {
	Items []*PolicyReport `json:"items"`
	Count int             `json:"count"`
}

type ResultList struct {
	Items []*ListResult `json:"items"`
	Count int           `json:"count"`
}

type StatusCount struct {
	Source string `json:"source,omitempty"`
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type Category struct {
	Name  string `json:"name"`
	Pass  int    `json:"pass"`
	Skip  int    `json:"skip"`
	Warn  int    `json:"warn"`
	Error int    `json:"error"`
	Fail  int    `json:"fail"`
}

type Source struct {
	Name       string     `json:"name"`
	Categories []Category `json:"categories"`
}

type Findings struct {
	Total  int              `json:"total"`
	Counts []*FindingCounts `json:"counts"`
}

type FindingCounts struct {
	Total  int            `json:"total"`
	Source string         `json:"source"`
	Counts map[string]int `json:"counts"`
}

type NamespacedStatusCount struct {
	Status string           `json:"status"`
	Items  []NamespaceCount `json:"items"`
}

type NamespaceCount struct {
	Namespace string `json:"namespace"`
	Count     int    `json:"count"`
	Status    string `json:"-"`
}

type ResourceStatusCount struct {
	Status string          `json:"status"`
	Items  []ResourceCount `json:"items"`
}

type ResourceCount struct {
	Source string `json:"source"`
	Count  int    `json:"count"`
	Status string `json:"-"`
}

type Resource struct {
	ID         string `json:"id,omitempty"`
	UID        string `json:"uid,omitempty"`
	Name       string `json:"name,omitempty"`
	Namespace  string `json:"namespace,omitempty"`
	Kind       string `json:"kind,omitempty"`
	APIVersion string `json:"apiVersion,omitempty"`
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

type ResourceResultList struct {
	Items []*ResourceResult `json:"items"`
	Count int               `json:"count"`
}

type ListResult struct {
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
