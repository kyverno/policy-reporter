package target_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
)

var result = &report.Result{
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.WarningPriority,
	Status:   report.Fail,
	Severity: report.High,
	Category: "resources",
	Scored:   true,
	Source:   "Kyverno",
	Resource: &report.Resource{
		APIVersion: "v1",
		Kind:       "Deployment",
		Name:       "nginx",
		Namespace:  "default",
		UID:        "536ab69f-1b3c-4bd9-9ba4-274a56188409",
	},
}

func Test_BaseClient(t *testing.T) {
	t.Run("Validate Default", func(t *testing.T) {
		client := target.NewBaseClient("", []string{}, false)

		if !client.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate MinimumPriority", func(t *testing.T) {
		client := target.NewBaseClient("error", []string{}, false)

		if client.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Source", func(t *testing.T) {
		client := target.NewBaseClient("", []string{"jsPolicy"}, false)

		if client.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("SkipExistingOnStartup", func(t *testing.T) {
		client := target.NewBaseClient("", []string{}, true)

		if !client.SkipExistingOnStartup() {
			t.Error("Should return configured SkipExistingOnStartup")
		}
	})
	t.Run("MinimumPriority", func(t *testing.T) {
		client := target.NewBaseClient("error", []string{}, true)

		if client.MinimumPriority() != "error" {
			t.Error("Should return configured MinimumPriority")
		}
	})
	t.Run("Sources", func(t *testing.T) {
		client := target.NewBaseClient("", []string{"Kyverno"}, true)

		if len(client.Sources()) != 1 {
			t.Fatal("Unexpected length of Sources")
		}
		if client.Sources()[0] != "Kyverno" {
			t.Error("Unexptected Source returned")
		}
	})
}
