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

func TestFind(t *testing.T) {
	list := []string{"test", "find", "item"}

	assert.Equal(t, "find", helper.Find(list, func(t string) bool { return t == "find" }, ""))
	assert.Equal(t, "fallback", helper.Find(list, func(t string) bool { return t == "invalid" }, "fallback"))
}

func TestMapSlice(t *testing.T) {
	assert.Equal(t, []string{"kyverno", "trivy"}, helper.MapSlice(map[int]string{2: "source_kyverno", 3: "source_trivy"}, func(value string) string {
		return strings.TrimPrefix(value, "source_")
	}))
}

func TestFilter(t *testing.T) {
	list := []string{"test", "find", "item", "", ""}

	assert.Equal(t, []string{"test", "find", "item"}, helper.Filter(list, func(t string) bool { return t != "" }))
}
