package target_test

import (
	"testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

var preport = &v1alpha2.PolicyReport{
	ObjectMeta: v1.ObjectMeta{
		Labels: map[string]string{"app": "policy-reporter"},
	},
}

func Test_BaseClient(t *testing.T) {
	t.Run("Validate MinimumPriority", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			"error",
			make([]string, 0),
		)

		if filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Source", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
			[]string{"jsPolicy"},
		)

		if filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Validate ClusterResult", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{Include: []string{"default"}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if !filter.Validate(fixtures.FailResultWithoutResource) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Validate Exclude Namespace match", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{Exclude: []string{"test"}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Namespace mismatch", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{Exclude: []string{"team-a"}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if !filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Namespace match", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{Include: []string{"test"}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if !filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Namespace mismatch", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{Include: []string{"team-a"}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Validate Exclude Priority match", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{v1alpha2.WarningPriority.String()}},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Priority mismatch", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{v1alpha2.ErrorPriority.String()}},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if !filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Priority match", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{},
			validate.RuleSets{Include: []string{v1alpha2.WarningPriority.String()}},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if !filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Priority mismatch", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{},
			validate.RuleSets{Include: []string{v1alpha2.ErrorPriority.String()}},
			validate.RuleSets{},
			"",
			make([]string, 0),
		)

		if filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Validate Exclude Policy match", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{"require-requests-and-limits-required"}},
			"",
			make([]string, 0),
		)

		if filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Policy mismatch", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{"policy-test"}},
			"",
			make([]string, 0),
		)

		if !filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Policy match", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Include: []string{"require-requests-and-limits-required"}},
			"",
			make([]string, 0),
		)

		if !filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Policy mismatch", func(t *testing.T) {
		filter := target.NewResultFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Include: []string{"policy-test"}},
			"",
			make([]string, 0),
		)

		if filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Validate Include Label match", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Include: []string{"app:policy-reporter"}},
		)

		if !filter.Validate(preport) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Label match", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Exclude: []string{"app:policy-reporter"}},
		)

		if filter.Validate(preport) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Label mismatch", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Exclude: []string{"app:monitoring"}},
		)

		if !filter.Validate(preport) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Label mismatch", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Include: []string{"app:monitoring"}},
		)

		if filter.Validate(preport) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate label as wildcard filter", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Exclude: []string{"app"}},
		)

		if filter.Validate(preport) {
			t.Errorf("Unexpected Validation Result")
		}

		filter = target.NewReportFilter(
			validate.RuleSets{Include: []string{"app"}},
		)

		if !filter.Validate(preport) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Label wildcard", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Include: []string{"app:*"}},
		)

		if !filter.Validate(preport) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Label wildcard", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Exclude: []string{"app:*"}},
		)

		if filter.Validate(preport) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Client Result Validation", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name: "Client",
			ResultFilter: target.NewResultFilter(
				validate.RuleSets{},
				validate.RuleSets{},
				validate.RuleSets{Include: []string{"policy-test"}},
				"",
				[]string{"jsPolicy"},
			),
			SkipExistingOnStartup: true,
		})

		if client.Validate(&v1alpha2.PolicyReport{}, fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Client Report Validation", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			ReportFilter:          target.NewReportFilter(validate.RuleSets{Include: []string{"app"}}),
			SkipExistingOnStartup: true,
		})

		if client.Validate(&v1alpha2.PolicyReport{}, fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Client nil Validation", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			ReportFilter:          target.NewReportFilter(validate.RuleSets{Include: []string{"app"}}),
			SkipExistingOnStartup: true,
		})

		if client.Validate(nil, fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Client Validation Fallbacks", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			SkipExistingOnStartup: true,
		})

		if !client.Validate(&v1alpha2.PolicyReport{}, fixtures.FailResult) {
			t.Errorf("Should fallback to true")
		}
		if client.MinimumPriority() != v1alpha2.DefaultPriority.String() {
			t.Errorf("Should fallback to default priority")
		}
		if len(client.Sources()) != 0 || client.Sources() == nil {
			t.Errorf("Should fallback to empty list")
		}
	})

	t.Run("SkipExistingOnStartup", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			ResultFilter:          &report.ResultFilter{},
			SkipExistingOnStartup: true,
		})

		if !client.SkipExistingOnStartup() {
			t.Error("Should return configured SkipExistingOnStartup")
		}
	})
	t.Run("MinimumPriority", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			ResultFilter:          &report.ResultFilter{MinimumPriority: "error"},
			SkipExistingOnStartup: true,
		})

		if client.MinimumPriority() != "error" {
			t.Error("Should return configured MinimumPriority")
		}
	})
	t.Run("Name", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			ResultFilter:          &report.ResultFilter{MinimumPriority: "error"},
			SkipExistingOnStartup: true,
		})

		if client.Name() != "Client" {
			t.Error("Should return configured Name")
		}
	})
	t.Run("Sources", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			ResultFilter:          &report.ResultFilter{Sources: []string{"Kyverno"}},
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
