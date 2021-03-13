package api

import (
	"time"

	"github.com/fjogeleit/policy-reporter/pkg/report"
	"github.com/fjogeleit/policy-reporter/pkg/target"
)

// Resource API Model
type Resource struct {
	APIVersion string `json:"apiVersion"`
	Kind       string `json:"kind"`
	Name       string `json:"name"`
	Namespace  string `json:"namespace,omitempty"`
	UID        string `json:"uid"`
}

// Result API Model
type Result struct {
	Message  string   `json:"message"`
	Policy   string   `json:"policy"`
	Rule     string   `json:"rule"`
	Priority string   `json:"priority"`
	Status   string   `json:"status"`
	Severity string   `json:"severity,omitempty"`
	Category string   `json:"category,omitempty"`
	Scored   bool     `json:"scored"`
	Resource Resource `json:"resource"`
}

// Summary API Model
type Summary struct {
	Pass  int `json:"pass"`
	Skip  int `json:"skip"`
	Warn  int `json:"warn"`
	Error int `json:"error"`
	Fail  int `json:"fail"`
}

// PolicyReport API Model
type PolicyReport struct {
	Name              string    `json:"name"`
	Namespace         string    `json:"namespace"`
	Results           []Result  `json:"results"`
	Summary           Summary   `json:"summary"`
	CreationTimestamp time.Time `json:"creationTimestamp"`
}

// ClusterPolicyReport API Model
type ClusterPolicyReport struct {
	Name              string    `json:"name"`
	Results           []Result  `json:"results"`
	Summary           Summary   `json:"summary"`
	CreationTimestamp time.Time `json:"creationTimestamp"`
}

func mapPolicyReport(p report.PolicyReport) PolicyReport {
	results := make([]Result, 0, len(p.Results))

	for _, r := range p.Results {

		results = append(results, Result{
			Message:  r.Message,
			Policy:   r.Policy,
			Rule:     r.Rule,
			Priority: r.Priority.String(),
			Status:   r.Status,
			Severity: r.Severity,
			Category: r.Category,
			Scored:   r.Scored,
			Resource: Resource{
				Namespace:  r.Resources[0].Namespace,
				APIVersion: r.Resources[0].APIVersion,
				Kind:       r.Resources[0].Kind,
				Name:       r.Resources[0].Name,
				UID:        r.Resources[0].UID,
			},
		})
	}

	return PolicyReport{
		Name:              p.Name,
		Namespace:         p.Namespace,
		CreationTimestamp: p.CreationTimestamp,
		Summary: Summary{
			Skip:  p.Summary.Skip,
			Pass:  p.Summary.Pass,
			Warn:  p.Summary.Warn,
			Fail:  p.Summary.Fail,
			Error: p.Summary.Error,
		},
		Results: results,
	}
}

func mapClusterPolicyReport(c report.ClusterPolicyReport) ClusterPolicyReport {
	results := make([]Result, 0, len(c.Results))

	for _, r := range c.Results {
		results = append(results, Result{
			Message:  r.Message,
			Policy:   r.Policy,
			Rule:     r.Rule,
			Priority: r.Priority.String(),
			Status:   r.Status,
			Severity: r.Severity,
			Category: r.Category,
			Scored:   r.Scored,
			Resource: Resource{
				Namespace:  r.Resources[0].Namespace,
				APIVersion: r.Resources[0].APIVersion,
				Kind:       r.Resources[0].Kind,
				Name:       r.Resources[0].Name,
				UID:        r.Resources[0].UID,
			},
		})
	}

	return ClusterPolicyReport{
		Name:              c.Name,
		CreationTimestamp: c.CreationTimestamp,
		Summary: Summary{
			Skip:  c.Summary.Skip,
			Pass:  c.Summary.Pass,
			Warn:  c.Summary.Warn,
			Fail:  c.Summary.Fail,
			Error: c.Summary.Error,
		},
		Results: results,
	}
}

// Target API Model
type Target struct {
	Name                  string `json:"name"`
	MinimumPriority       string `json:"minimumPriority"`
	SkipExistingOnStartup bool   `json:"skipExistingOnStartup"`
}

func mapTarget(t target.Client) Target {
	minPrio := t.MinimumPriority()
	if minPrio == "" {
		minPrio = report.Priority(report.DebugPriority).String()
	}

	return Target{
		Name:                  t.Name(),
		MinimumPriority:       minPrio,
		SkipExistingOnStartup: t.SkipExistingOnStartup(),
	}
}
