package target_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/validate"
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

var result2 = report.Result{
	Message:  "validation error: requests and limits required. Rule autogen-check-for-requests-and-limits failed at path /spec/template/spec/containers/0/resources/requests/",
	Policy:   "require-requests-and-limits-required",
	Rule:     "autogen-check-for-requests-and-limits",
	Priority: report.WarningPriority,
	Status:   report.Fail,
	Severity: report.High,
	Category: "resources",
	Scored:   true,
	Source:   "Kyverno",
}

func Test_BaseClient(t *testing.T) {
	t.Run("Validate MinimumPriority", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			"error",
			make([]string, 0),
		)

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Source", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
			[]string{"jsPolicy"},
		)

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Validate ClusterResult", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{Include: []string{"default"}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if !filter.Validate(result2) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Validate Exclude Namespace match", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{Exclude: []string{"default"}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Namespace mismatch", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{Exclude: []string{"team-a"}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if !filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Namespace match", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{Include: []string{"default"}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if !filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Namespace mismatch", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{Include: []string{"team-a"}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Validate Exclude Priority match", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{report.WarningPriority.String()}},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Priority mismatch", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{report.ErrorPriority.String()}},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if !filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Priority match", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{},
			validate.RuleSets{Include: []string{report.WarningPriority.String()}},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if !filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Priority mismatch", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{},
			validate.RuleSets{Include: []string{report.ErrorPriority.String()}},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Validate Exclude Policy match", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{"require-requests-and-limits-required"}},
			"",
			make([]string, 0),
		)

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Policy mismatch", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{"policy-test"}},
			"",
			make([]string, 0),
		)

		if !filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Policy match", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Include: []string{"require-requests-and-limits-required"}},
			"",
			make([]string, 0),
		)

		if !filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Policy mismatch", func(t *testing.T) {
		filter := target.NewClientFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Include: []string{"policy-test"}},
			"",
			make([]string, 0),
		)

		if filter.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Client Validation", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name: "Client",
			Filter: target.NewClientFilter(
				validate.RuleSets{},
				validate.RuleSets{},
				validate.RuleSets{Include: []string{"policy-test"}},
				"",
				[]string{"jsPolicy"},
			),
			SkipExistingOnStartup: true,
		})

		if client.Validate(result) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Client Validation Fallbacks", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			SkipExistingOnStartup: true,
		})

		if !client.Validate(result) {
			t.Errorf("Should fallback to true")
		}
		if client.MinimumPriority() != report.DefaultPriority.String() {
			t.Errorf("Should fallback to default priority")
		}
		if len(client.Sources()) != 0 || client.Sources() == nil {
			t.Errorf("Should fallback to empty list")
		}
	})

	t.Run("SkipExistingOnStartup", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			Filter:                &report.ResultFilter{},
			SkipExistingOnStartup: true,
		})

		if !client.SkipExistingOnStartup() {
			t.Error("Should return configured SkipExistingOnStartup")
		}
	})
	t.Run("MinimumPriority", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			Filter:                &report.ResultFilter{MinimumPriority: "error"},
			SkipExistingOnStartup: true,
		})

		if client.MinimumPriority() != "error" {
			t.Error("Should return configured MinimumPriority")
		}
	})
	t.Run("Name", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			Filter:                &report.ResultFilter{MinimumPriority: "error"},
			SkipExistingOnStartup: true,
		})

		if client.Name() != "Client" {
			t.Error("Should return configured Name")
		}
	})
	t.Run("Sources", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			Filter:                &report.ResultFilter{Sources: []string{"Kyverno"}},
			SkipExistingOnStartup: true,
		})

		if len(client.Sources()) != 1 {
			t.Fatal("Unexpected length of Sources")
		}
		if client.Sources()[0] != "Kyverno" {
			t.Error("Unexptected Source returned")
		}
	})
}
