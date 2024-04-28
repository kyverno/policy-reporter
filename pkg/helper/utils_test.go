package helper_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/helper"
)

func TestContains(t *testing.T) {
	assert.True(t, helper.Contains("kyverno", []string{"test", "kyverno", "trivy"}))
	assert.False(t, helper.Contains("kube-bench", []string{"test", "kyverno", "trivy"}))
}

func TestToList(t *testing.T) {
	result := helper.ToList(map[string]string{
		"first":  "kyverno",
		"second": "trivy",
	})

	assert.Equal(t, 2, len(result))
	assert.Contains(t, result, "kyverno")
	assert.Contains(t, result, "trivy")
}

func TestMap(t *testing.T) {
	assert.Equal(t, []string{"kyverno", "trivy"}, helper.Map([]string{"source_kyverno", "source_trivy"}, func(value string) string {
		return strings.TrimPrefix(value, "source_")
	}))
}

func TestConvertMap(t *testing.T) {
	assert.Equal(t, map[string]string{"first": "kyverno", "second": "trivy"}, helper.ConvertMap(map[string]any{
		"first":  "kyverno",
		"second": "trivy",
		"third":  3,
	}))
}

func TestDetauls(t *testing.T) {
	assert.Equal(t, "fallback", helper.Defaults("", "fallback"))
	assert.Equal(t, "value", helper.Defaults("value", "fallback"))
}

func TestToPointer(t *testing.T) {
	value := "test"
	number := 5

	assert.Equal(t, &value, helper.ToPointer(value))
	assert.Equal(t, &number, helper.ToPointer(number))
}
