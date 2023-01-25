package metrics_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/fixtures"
	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/validate"
)

func Test_Vaildate(t *testing.T) {
	t.Run("Allow ClusterReport", func(t *testing.T) {
		filter := metrics.NewResultFilter(validate.RuleSets{Include: []string{"test"}}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{})
		if !filter.Validate(fixtures.PassNamespaceResult) {
			t.Error("Expected Validate returns true if Report is a ClusterPolicyReport without namespace")
		}
	})
	t.Run("Disallow if Report include not match", func(t *testing.T) {
		filter := metrics.NewResultFilter(validate.RuleSets{Include: []string{"dev"}}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{})
		if filter.Validate(fixtures.FailPodResult) {
			t.Error("Expected Validate returns false if Report namespace not match include rule")
		}
	})

	t.Run("Allow Report with matching include Namespace", func(t *testing.T) {
		filter := metrics.NewResultFilter(validate.RuleSets{Include: []string{"test"}}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{})
		if !filter.Validate(fixtures.FailPodResult) {
			t.Error("Expected Validate returns true if Report namespace matches include pattern")
		}
	})

	t.Run("Disallow Report with matching exclude Namespace", func(t *testing.T) {
		filter := metrics.NewResultFilter(validate.RuleSets{Exclude: []string{"test"}}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{})
		if filter.Validate(fixtures.FailPodResult) {
			t.Error("Expected Validate returns false if Report namespace matches exclude pattern")
		}
	})

	t.Run("Ignores exclude pattern if include namespaces provided", func(t *testing.T) {
		filter := metrics.NewResultFilter(validate.RuleSets{Exclude: []string{"test"}, Include: []string{"test"}}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{})
		if !filter.Validate(fixtures.FailPodResult) {
			t.Error("Expected Validate returns true because exclude patterns ignored if include patterns provided")
		}
	})

	t.Run("Disallow Report with matching exclude Policy", func(t *testing.T) {
		filter := metrics.NewResultFilter(validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{Exclude: []string{"require-requests-*"}}, validate.RuleSets{}, validate.RuleSets{})
		if filter.Validate(fixtures.FailPodResult) {
			t.Error("Expected Validate returns false if Report policy matches exclude pattern")
		}
	})

	t.Run("Disallow Report with matching exclude Status", func(t *testing.T) {
		filter := metrics.NewResultFilter(validate.RuleSets{}, validate.RuleSets{Exclude: []string{v1alpha2.StatusFail}}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{})
		if filter.Validate(fixtures.FailPodResult) {
			t.Error("Expected Validate returns false if Report status matches exclude pattern")
		}
	})

	t.Run("Disallow Report with matching exclude Severity", func(t *testing.T) {
		filter := metrics.NewResultFilter(validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{Exclude: []string{v1alpha2.SeverityHigh}})
		if filter.Validate(fixtures.FailResult) {
			t.Error("Expected Validate returns false if Report severity matches exclude pattern")
		}
	})

	t.Run("Disallow Report with matching exclude Source", func(t *testing.T) {
		filter := metrics.NewResultFilter(validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{}, validate.RuleSets{Exclude: []string{"Kyverno"}}, validate.RuleSets{})
		if filter.Validate(fixtures.FailPodResult) {
			t.Error("Expected Validate returns false if Report source matches exclude pattern")
		}
	})
}
