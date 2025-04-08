package formatting_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/target/formatting"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestResourceString(t *testing.T) {
	t.Run("namespaced resource", func(t *testing.T) {
		res := formatting.ResourceString(&corev1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Deployment",
			Name:       "nginx",
			Namespace:  "default",
		})

		assert.Equal(t, "v1/Deployment: default/nginx", res)
	})

	t.Run("cluster resource", func(t *testing.T) {
		res := formatting.ResourceString(&corev1.ObjectReference{
			APIVersion: "v1",
			Kind:       "Namespace",
			Name:       "default",
		})

		assert.Equal(t, "v1/Namespace: default", res)
	})
}
