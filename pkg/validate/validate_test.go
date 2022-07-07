package validate_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/validate"
)

func Test_Validations(t *testing.T) {
	t.Run("Validate Source", func(t *testing.T) {

		if validate.ContainsRuleSet("test", validate.RuleSets{Include: []string{"jsPolicy"}}) {
			t.Errorf("Unexpected Validation Result")
		}

		if !validate.ContainsRuleSet("jsPolicy", validate.RuleSets{Include: []string{"jsPolicy"}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})

	t.Run("Empty Namespace", func(t *testing.T) {
		if !validate.Namespace("", validate.RuleSets{Exclude: []string{""}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Namespace Exclude Namespace match", func(t *testing.T) {
		if validate.Namespace("default", validate.RuleSets{Exclude: []string{"default"}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Validate Exclude Namespace mismatch", func(t *testing.T) {
		if !validate.Namespace("default", validate.RuleSets{Exclude: []string{"team-a"}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Namespace Include Namespace match", func(t *testing.T) {
		if !validate.Namespace("default", validate.RuleSets{Include: []string{"def*"}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("Namespace Include Namespace mismatch", func(t *testing.T) {
		if validate.Namespace("default", validate.RuleSets{Include: []string{"team-a"}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("MatchRuleSet Exclude Rule match", func(t *testing.T) {
		if !validate.MatchRuleSet("test", validate.RuleSets{Exclude: []string{"team-a"}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("MatchRuleSet Exclude Rule mismatch", func(t *testing.T) {
		if validate.MatchRuleSet("test", validate.RuleSets{Exclude: []string{"test"}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("MatchRuleSet Include Rule match", func(t *testing.T) {
		if !validate.MatchRuleSet("test", validate.RuleSets{Include: []string{"test"}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("MatchRuleSet Include Rule mismatch", func(t *testing.T) {
		if validate.MatchRuleSet("test", validate.RuleSets{Include: []string{"team-a"}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("ContainsRuleSet Include Rule match", func(t *testing.T) {
		if !validate.ContainsRuleSet("test", validate.RuleSets{Include: []string{"test"}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("ContainsRuleSet Include Rule mismatch", func(t *testing.T) {
		if validate.ContainsRuleSet("test", validate.RuleSets{Include: []string{"team-a"}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("ContainsRuleSet Exclude Rule match", func(t *testing.T) {
		if validate.ContainsRuleSet("test", validate.RuleSets{Exclude: []string{"test"}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("ContainsRuleSet Include Rule mismatch", func(t *testing.T) {
		if !validate.ContainsRuleSet("test", validate.RuleSets{Exclude: []string{"team-a"}}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
	t.Run("ContainsRuleSet empty rules", func(t *testing.T) {
		if !validate.ContainsRuleSet("test", validate.RuleSets{}) {
			t.Errorf("Unexpected Validation Result")
		}
	})
}

func Test_RulesCount(t *testing.T) {
	r1 := validate.RuleSets{}
	if r1.Count() != 0 {
		t.Errorf("Unexpected Rules.Count")
	}

	r2 := validate.RuleSets{Include: []string{"dev"}, Exclude: []string{"stage"}}
	if r2.Count() != 2 {
		t.Errorf("Unexpected Rules.Count")
	}
}
