package filter_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/filter"
)

func Test_BaseClient(t *testing.T) {
	t.Run("Validate Default", func(t *testing.T) {
		filter := filter.New(filter.Rules{}, []string{})

		if !filter.ValidateNamespace("test") {
			t.Errorf("Unexpected Validation Result without configured rules")
		}
		if !filter.ValidateSource("Kyverno") {
			t.Errorf("Unexpected Validation Result without configured rules")
		}
	})
	t.Run("Validate Source", func(t *testing.T) {
		filter := filter.New(filter.Rules{}, []string{"jsPolicy"})

		if filter.ValidateSource("test") {
			t.Errorf("Unexpected Validation Result")
		}

		if !filter.ValidateSource("jsPolicy") {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Validate Exclude Namespace match", func(t *testing.T) {
		filter := filter.New(filter.Rules{Exclude: []string{"default"}}, []string{})

		if filter.ValidateNamespace("default") {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Namespace mismatch", func(t *testing.T) {
		filter := filter.New(filter.Rules{Exclude: []string{"team-a"}}, []string{})

		if !filter.ValidateNamespace("default") {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Namespace match", func(t *testing.T) {
		filter := filter.New(filter.Rules{Include: []string{"default"}}, []string{})

		if !filter.ValidateNamespace("default") {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Namespace mismatch", func(t *testing.T) {
		filter := filter.New(filter.Rules{Include: []string{"team-a"}}, []string{})

		if filter.ValidateNamespace("default") {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Rule match", func(t *testing.T) {
		result := filter.ValidateRule("test", filter.Rules{Exclude: []string{"team-a"}})

		if !result {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Rule mismatch", func(t *testing.T) {
		result := filter.ValidateRule("test", filter.Rules{Exclude: []string{"test"}})

		if result {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Rule match", func(t *testing.T) {
		result := filter.ValidateRule("test", filter.Rules{Include: []string{"test"}})

		if !result {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Include Rule mismatch", func(t *testing.T) {
		result := filter.ValidateRule("test", filter.Rules{Include: []string{"team-a"}})

		if result {
			t.Errorf("Unexpected Validation Result")
		}
	})
}
