package target_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

var result = report.Result{
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.WarningPriority,
	Status:   report.Fail,
	Severity: report.High,
	Category: "resources",
	Scored:   true,
	Source:   "Kyverno",
	Resource: report.Resource{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "default",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	},
}

func Test_BaseClient(t *testing.T) {
	t.Run("Validate Default", func(t *testing.T) {
		filter := &target.Filter{}

		if !filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate MinimumPriority", func(t *testing.T) {
		filter := &target.Filter{MinimumPriority: "error"}

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Source", func(t *testing.T) {
		filter := &target.Filter{Sources: []string{"jsPolicy"}}

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Validate Exclude Namespace match", func(t *testing.T) {
		filter := &target.Filter{Namespace: target.Rules{Exclude: []string{"default"}}}

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Namespace mismatch", func(t *testing.T) {
		filter := &target.Filter{Namespace: target.Rules{Exclude: []string{"team-a"}}}

		if !filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Namespace match", func(t *testing.T) {
		filter := &target.Filter{Namespace: target.Rules{Include: []string{"default"}}}

		if !filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Namespace mismatch", func(t *testing.T) {
		filter := &target.Filter{Namespace: target.Rules{Include: []string{"team-a"}}}

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Validate Exclude Priority match", func(t *testing.T) {
		filter := &target.Filter{Priority: target.Rules{Exclude: []string{report.WarningPriority.String()}}}

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Priority mismatch", func(t *testing.T) {
		filter := &target.Filter{Priority: target.Rules{Exclude: []string{report.ErrorPriority.String()}}}

		if !filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Priority match", func(t *testing.T) {
		filter := &target.Filter{Priority: target.Rules{Include: []string{report.WarningPriority.String()}}}

		if !filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Priority mismatch", func(t *testing.T) {
		filter := &target.Filter{Priority: target.Rules{Include: []string{report.ErrorPriority.String()}}}

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Validate Exclude Policy match", func(t *testing.T) {
		filter := &target.Filter{Policy: target.Rules{Exclude: []string{"require-requests-and-limits-required"}}}

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Policy mismatch", func(t *testing.T) {
		filter := &target.Filter{Policy: target.Rules{Exclude: []string{"policy-test"}}}

		if !filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Policy match", func(t *testing.T) {
		filter := &target.Filter{Policy: target.Rules{Include: []string{"require-requests-and-limits-required"}}}

		if !filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Policy mismatch", func(t *testing.T) {
		filter := &target.Filter{Policy: target.Rules{Include: []string{"policy-test"}}}

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Client Validation", func(t *testing.T) {
		client := target.NewBaseClient("Client", true, &target.Filter{Sources: []string{"jsPolicy"}})

		if client.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("SkipExistingOnStartup", func(t *testing.T) {
		client := target.NewBaseClient("Client", true, &target.Filter{})

		if !client.SkipExistingOnStartup() {
			t.Error("Should return configured SkipExistingOnStartup")
		}
	})
	t.Run("MinimumPriority", func(t *testing.T) {
		client := target.NewBaseClient("Client", true, &target.Filter{MinimumPriority: "error"})

		if client.MinimumPriority() != "error" {
			t.Error("Should return configured MinimumPriority")
		}
	})
	t.Run("Name", func(t *testing.T) {
		client := target.NewBaseClient("Client", true, &target.Filter{MinimumPriority: "error"})

		if client.Name() != "Client" {
			t.Error("Should return configured Name")
		}
	})
	t.Run("Sources", func(t *testing.T) {
		client := target.NewBaseClient("Client", true, &target.Filter{Sources: []string{"Kyverno"}})

		if len(client.Sources()) != 1 {
			t.Fatal("Unexpected length of Sources")
		}
		if client.Sources()[0] != "Kyverno" {
			t.Error("Unexptected Source returned")
		}
	})
}
