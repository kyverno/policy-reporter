package v1

import (
	"github.com/kyverno/policy-reporter/pkg/report"
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
	Status string `json:"status"`
	Count  int    `json:"count"`
}

type NamespacedStatusCount struct {
	Status string           `json:"status"`
	Items  []NamespaceCount `json:"items"`
}

type NamespaceCount struct {
	Namespace string `json:"namespace"`
	Count     int    `json:"count"`
}

type Resource struct {
	Name string `json:"name"`
	Kind string `json:"kind"`
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
	Timestamp  int               `json:"timestamp,omitempty"`
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
		minPrio = report.Priority(report.DebugPriority).String()
	}

	return Target{
		Name:                  t.Name(),
		MinimumPriority:       minPrio,
		Sources:               t.Sources(),
		SkipExistingOnStartup: t.SkipExistingOnStartup(),
	}
}
