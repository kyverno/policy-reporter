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
	Message    string            `json:"message"`
	Policy     string            `json:"policy"`
	Rule       string            `json:"rule"`
	Priority   string            `json:"priority"`
	Status     string            `json:"status"`
	Severity   string            `json:"severity,omitempty"`
	Category   string            `json:"category,omitempty"`
	Scored     bool              `json:"scored"`
	Properties map[string]string `json:"properties,omitempty"`
	Source     string            `json:"source,omitempty"`
	Resource   *Resource         `json:"resource,omitempty"`
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
	Namespace         string    `json:"namespace,omitempty"`
	Results           []Result  `json:"results"`
	Summary           Summary   `json:"summary"`
	CreationTimestamp time.Time `json:"creationTimestamp"`
}

func mapPolicyReport(p report.PolicyReport) PolicyReport {
	results := make([]Result, 0, len(p.Results))

	for _, r := range p.Results {
		result := Result{
			Message:    r.Message,
			Policy:     r.Policy,
			Rule:       r.Rule,
			Priority:   r.Priority.String(),
			Status:     r.Status,
			Severity:   r.Severity,
			Category:   r.Category,
			Scored:     r.Scored,
			Properties: r.Properties,
			Source:     r.Source,
		}

		if r.HasResource() {
			result.Resource = &Resource{
				Namespace:  r.Resource.Namespace,
				APIVersion: r.Resource.APIVersion,
				Kind:       r.Resource.Kind,
				Name:       r.Resource.Name,
				UID:        r.Resource.UID,
			}
		}

		results = append(results, result)
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
