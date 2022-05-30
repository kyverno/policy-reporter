package metrics_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/listener/metrics"
	"github.com/kyverno/policy-reporter/pkg/report"
)

func Test_Vaildate(t *testing.T) {
	t.Run("Allow ClusterReport", func(t *testing.T) {
		filter := metrics.NewFilter(metrics.Rules{Include: []string{"test"}}, metrics.Rules{}, metrics.Rules{}, metrics.Rules{}, metrics.Rules{})
		if !filter.Validate(result1) {
			t.Error("Expected Validate returns true if Report is a ClusterPolicyReport without namespace")
		}
	})
	t.Run("Disallow if Report include not match", func(t *testing.T) {
		filter := metrics.NewFilter(metrics.Rules{Include: []string{"dev"}}, metrics.Rules{}, metrics.Rules{}, metrics.Rules{}, metrics.Rules{})
		if filter.Validate(result1) {
			t.Error("Expected Validate returns false if Report namespace not match include rule")
		}
	})

	t.Run("Allow Report with matching include Namespace", func(t *testing.T) {
		filter := &metrics.Filter{Namespace: metrics.Rules{Include: []string{"test"}}}
		if !filter.Validate(result1) {
			t.Error("Expected Validate returns true if Report namespace matches include pattern")
		}
	})

	t.Run("Disallow Report with matching exclude Namespace", func(t *testing.T) {
		filter := &metrics.Filter{Namespace: metrics.Rules{Exclude: []string{"test"}}}
		if filter.Validate(result1) {
			t.Error("Expected Validate returns false if Report namespace matches exclude pattern")
		}
	})

	t.Run("Ignores exclude pattern if include namespaces provided", func(t *testing.T) {
		filter := &metrics.Filter{Namespace: metrics.Rules{Exclude: []string{"test"}, Include: []string{"test"}}}
		if !filter.Validate(result1) {
			t.Error("Expected Validate returns true because exclude patterns ignored if include patterns provided")
		}
	})

	t.Run("Disallow Report with matching exclude Policy", func(t *testing.T) {
		filter := &metrics.Filter{Policy: metrics.Rules{Exclude: []string{"require-requests-*"}}}
		if filter.Validate(result1) {
			t.Error("Expected Validate returns false if Report policy matches exclude pattern")
		}
	})

	t.Run("Disallow Report with matching exclude Status", func(t *testing.T) {
		filter := &metrics.Filter{Status: metrics.Rules{Exclude: []string{report.Fail}}}
		if filter.Validate(result1) {
			t.Error("Expected Validate returns false if Report status matches exclude pattern")
		}
	})

	t.Run("Disallow Report with matching exclude Severity", func(t *testing.T) {
		filter := &metrics.Filter{Severity: metrics.Rules{Exclude: []string{report.High}}}
		if filter.Validate(result1) {
			t.Error("Expected Validate returns false if Report severity matches exclude pattern")
		}
	})

	t.Run("Disallow Report with matching exclude Source", func(t *testing.T) {
		filter := &metrics.Filter{Source: metrics.Rules{Exclude: []string{"Kyverno"}}}
		if filter.Validate(result1) {
			t.Error("Expected Validate returns false if Report source matches exclude pattern")
		}
	})
}
