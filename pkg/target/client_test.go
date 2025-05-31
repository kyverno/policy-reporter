package target_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"openreports.io/apis/openreports.io/v1alpha1"

	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/report"
	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

var preport = &v1alpha1.Report{
	ObjectMeta: v1.ObjectMeta{
		Labels: map[string]string{"app": "policy-reporter"},
	},
}

var factory = target.NewResultFilterFactory(nil)

func Test_BaseClient(t *testing.T) {
	t.Run("Validate MinimumSeverity", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			v1alpha1.SeverityCritical,
		)

		assert.False(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})
	t.Run("Validate Source", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Include: []string{"jsPolicy"}},
			"",
		)

		assert.False(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})

	t.Run("Validate ClusterResult", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{Include: []string{"default"}},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
		)

		assert.True(t, filter.Validate(fixtures.FailResultWithoutResource), "Unexpected Validation Result")
	})

	t.Run("Validate Exclude Namespace match", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{Exclude: []string{"test"}},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
		)

		assert.False(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})
	t.Run("Validate Exclude Namespace mismatch", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{Exclude: []string{"team-a"}},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
		)

		assert.True(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})
	t.Run("Validate Include Namespace match", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{Include: []string{"test"}},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
		)

		assert.True(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})
	t.Run("Validate Exclude Namespace mismatch", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{Include: []string{"team-a"}},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
		)

		assert.False(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})

	t.Run("Validate Exclude Status match", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{v1alpha1.StatusFail}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
		)

		assert.False(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})
	t.Run("Validate Exclude Status mismatch", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{v1alpha1.StatusSkip}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
		)

		assert.True(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})
	t.Run("Validate Include Status match", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Include: []string{v1alpha1.StatusFail}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
		)

		assert.True(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})
	t.Run("Validate Exclude Status mismatch", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{v1alpha1.StatusFail}},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
		)

		assert.False(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})

	t.Run("Validate Exclude Severity match", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{v1alpha1.SeverityHigh}},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
		)

		assert.False(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})
	t.Run("Validate Exclude Severity mismatch", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{v1alpha1.SeverityCritical}},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
		)

		assert.True(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})
	t.Run("Validate Include Severity match", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{Include: []string{v1alpha1.SeverityHigh}},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
		)

		assert.True(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})
	t.Run("Validate Exclude Severity mismatch", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{Include: []string{v1alpha1.SeverityCritical}},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			"",
		)

		assert.False(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})

	t.Run("Validate Exclude Policy match", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{"require-requests-and-limits-required"}},
			validate.RuleSets{},
			"",
		)

		assert.False(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})
	t.Run("Validate Exclude Policy mismatch", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Exclude: []string{"policy-test"}},
			validate.RuleSets{},
			"",
		)

		assert.True(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})
	t.Run("Validate Include Policy match", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Include: []string{"require-requests-and-limits-required"}},
			validate.RuleSets{},
			"",
		)

		if !filter.Validate(fixtures.FailResult) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Policy mismatch", func(t *testing.T) {
		filter := factory.CreateFilter(
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{},
			validate.RuleSets{Include: []string{"policy-test"}},
			validate.RuleSets{},
			"",
		)

		assert.False(t, filter.Validate(fixtures.FailResult), "Unexpected Validation Result")
	})

	t.Run("Validate Include Label match", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Include: []string{"app:policy-reporter"}},
			validate.RuleSets{},
		)

		assert.True(t, filter.Validate(preport), "Unexpected Validation Result")
	})
	t.Run("Validate Exclude Label match", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Exclude: []string{"app:policy-reporter"}},
			validate.RuleSets{},
		)

		if filter.Validate(preport) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Label mismatch", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Exclude: []string{"app:monitoring"}},
			validate.RuleSets{},
		)

		assert.True(t, filter.Validate(preport), "Unexpected Validation Result")
	})
	t.Run("Validate Include Label mismatch", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Include: []string{"app:monitoring"}},
			validate.RuleSets{},
		)

		assert.False(t, filter.Validate(preport), "Unexpected Validation Result")
	})
	t.Run("Validate label as wildcard filter", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Exclude: []string{"app"}},
			validate.RuleSets{},
		)

		assert.False(t, filter.Validate(preport), "Unexpected Validation Result")

		filter = target.NewReportFilter(
			validate.RuleSets{Include: []string{"app"}},
			validate.RuleSets{},
		)

		assert.True(t, filter.Validate(preport), "Unexpected Validation Result")
	})
	t.Run("Validate Include Label wildcard", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Include: []string{"app:*"}},
			validate.RuleSets{},
		)

		assert.True(t, filter.Validate(preport), "Unexpected Validation Result")
	})
	t.Run("Validate Exclude Label wildcard", func(t *testing.T) {
		filter := target.NewReportFilter(
			validate.RuleSets{Exclude: []string{"app:*"}},
			validate.RuleSets{},
		)

		assert.False(t, filter.Validate(preport), "Unexpected Validation Result")
	})

	t.Run("Client Result Validation", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name: "Client",
			ResultFilter: factory.CreateFilter(
				validate.RuleSets{},
				validate.RuleSets{},
				validate.RuleSets{Include: []string{"policy-test"}},
				validate.RuleSets{},
				validate.RuleSets{Include: []string{"jsPolicy"}},
				"",
			),
			SkipExistingOnStartup: true,
		})

		assert.False(t, client.Validate(&v1alpha1.Report{}, fixtures.FailResult), "Unexpected Validation Result")
	})

	t.Run("Client Report Validation", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name: "Client",
			ReportFilter: target.NewReportFilter(
				validate.RuleSets{Include: []string{"app"}},
				validate.RuleSets{},
			),
			SkipExistingOnStartup: true,
		})

		assert.False(t, client.Validate(&v1alpha1.Report{}, fixtures.FailResult), "Unexpected Validation Result")
	})

	t.Run("Client nil Validation", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name: "Client",
			ReportFilter: target.NewReportFilter(
				validate.RuleSets{Include: []string{"app"}},
				validate.RuleSets{},
			),
			SkipExistingOnStartup: true,
		})

		assert.False(t, client.Validate(nil, fixtures.FailResult), "Unexpected Validation Result")
	})

	t.Run("Client Validation Fallbacks", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			SkipExistingOnStartup: true,
		})

		assert.True(t, client.Validate(&v1alpha1.Report{}, fixtures.FailResult), "Should fallback to true")
		assert.Equal(t, client.MinimumSeverity(), v1alpha1.SeverityInfo, "Should fallback to severity info")
		assert.NotNil(t, client.Sources(), "Should fallback to empty list")
	})

	t.Run("SkipExistingOnStartup", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			ResultFilter:          &report.ResultFilter{},
			SkipExistingOnStartup: true,
		})

		assert.True(t, client.SkipExistingOnStartup(), "Should return configured SkipExistingOnStartup")
	})
	t.Run("MinimumSeverity", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			ResultFilter:          &report.ResultFilter{MinimumSeverity: v1alpha1.SeverityHigh},
			SkipExistingOnStartup: true,
		})

		assert.Equal(t, client.MinimumSeverity(), v1alpha1.SeverityHigh, "Should return configured MinimumSeverity")
	})
	t.Run("Name", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			ResultFilter:          &report.ResultFilter{MinimumSeverity: "error"},
			SkipExistingOnStartup: true,
		})

		assert.Equal(t, client.Name(), "Client", "Should return configured Name")
	})
	t.Run("Sources", func(t *testing.T) {
		client := target.NewBaseClient(target.ClientOptions{
			Name:                  "Client",
			ResultFilter:          &report.ResultFilter{Sources: []string{"Kyverno"}},
			SkipExistingOnStartup: true,
		})

		assert.Len(t, client.Sources(), 1)
		assert.Equal(t, client.Sources()[0], "Kyverno")
	})
}
