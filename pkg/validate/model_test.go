package validate_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/validate"
)

func TestCount(t *testing.T) {
	t.Parallel()
	t.Run("count include rules", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 0, validate.RuleSets{}.Count())
		assert.Equal(t, 2, validate.RuleSets{Include: []string{"kyverno", "falco"}}.Count())
	})
	t.Run("count exclude rules", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, 2, validate.RuleSets{Exclude: []string{"kyverno", "falco"}}.Count())
	})
}

func TestEnabled(t *testing.T) {
	t.Parallel()
	t.Run("enabled when include rule exist", func(t *testing.T) {
		t.Parallel()
		assert.False(t, validate.RuleSets{}.Enabled())
		assert.True(t, validate.RuleSets{Include: []string{"kyverno"}}.Enabled())
	})
	t.Run("enabled when exclude rule exist", func(t *testing.T) {
		t.Parallel()
		assert.True(t, validate.RuleSets{Exclude: []string{"kyverno", "falco"}}.Enabled())
	})
}
