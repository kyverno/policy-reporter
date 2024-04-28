package v1_test

import (
	"testing"

	v1 "github.com/kyverno/policy-reporter/pkg/api/v1"
	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/stretchr/testify/assert"
)

func TestMapping(t *testing.T) {
	t.Run("MapClusterStatusCounts", func(t *testing.T) {
		result := v1.MapClusterStatusCounts([]database.StatusCount{
			{Source: "kyverno", Status: "pass", Count: 3},
			{Source: "kyverno", Status: "fail", Count: 4},
		}, []string{"pass", "fail"})

		assert.Equal(t, 2, len(result))
		assert.Contains(t, result, v1.StatusCount{Status: "pass", Count: 3})
		assert.Contains(t, result, v1.StatusCount{Status: "fail", Count: 4})
	})

	t.Run("MapNamespaceStatusCounts", func(t *testing.T) {
		result := v1.MapNamespaceStatusCounts([]database.StatusCount{
			{Source: "kyverno", Status: "pass", Count: 3, Namespace: "default"},
			{Source: "kyverno", Status: "fail", Count: 4, Namespace: "default"},
			{Source: "kyverno", Status: "pass", Count: 2, Namespace: "user"},
			{Source: "kyverno", Status: "fail", Count: 2, Namespace: "user"},
		}, []string{"pass", "fail"})

		assert.Equal(t, 2, len(result))

		assert.Contains(t, result, v1.NamespaceStatusCount{Status: "pass", Items: []v1.NamespaceCount{
			{Namespace: "default", Count: 3, Status: "pass"},
			{Namespace: "user", Count: 2, Status: "pass"},
		}})

		assert.Contains(t, result, v1.NamespaceStatusCount{Status: "fail", Items: []v1.NamespaceCount{
			{Namespace: "default", Count: 4, Status: "fail"},
			{Namespace: "user", Count: 2, Status: "fail"},
		}})
	})
}
