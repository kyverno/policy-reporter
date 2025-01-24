package target_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyverno/policy-reporter/pkg/crd/api/policyreport/v1alpha2"
	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
)

func TestConfig(t *testing.T) {
	t.Run("return expected secret ref", func(t *testing.T) {
		c := &v1alpha1.Config[v1alpha1.WebhookOptions]{
			SecretRef: "webhook-secret",
		}

		assert.Equal(t, c.Secret(), "webhook-secret")
	})

	t.Run("ignores secret mount", func(t *testing.T) {
		c := &v1alpha1.Config[v1alpha1.WebhookOptions]{
			MountedSecret: "webhook-secret",
		}

		assert.Equal(t, c.Secret(), "")
	})

	t.Run("base mapper set expected fallbacks from parent config", func(t *testing.T) {
		p := &v1alpha1.Config[v1alpha1.WebhookOptions]{
			MinimumSeverity: v1alpha2.SeverityMedium,
			SkipExisting:    true,
		}

		c := &v1alpha1.Config[v1alpha1.WebhookOptions]{}
		c.MapBaseParent(p)

		assert.Equal(t, c.MinimumSeverity, p.MinimumSeverity)
		assert.Equal(t, c.SkipExisting, p.SkipExisting)
	})

	t.Run("base mapper keeps none empty values", func(t *testing.T) {
		p := &v1alpha1.Config[v1alpha1.WebhookOptions]{
			MinimumSeverity: v1alpha2.SeverityMedium,
		}

		c := &v1alpha1.Config[v1alpha1.WebhookOptions]{
			MinimumSeverity: v1alpha2.SeverityInfo,
		}

		c.MapBaseParent(p)

		assert.Equal(t, c.MinimumSeverity, v1alpha2.SeverityInfo)
	})
}

func TestAWSConfig(t *testing.T) {
	t.Run("aws mapper set expected fallbacks from parent config", func(t *testing.T) {
		p := v1alpha1.AWSConfig{
			AccessKeyID:     "access",
			SecretAccessKey: "secret",
			Region:          "eu",
			Endpoint:        "http://localhost:8080",
		}

		c := v1alpha1.AWSConfig{}
		c.MapAWSParent(p)

		assert.Equal(t, c.AccessKeyID, p.AccessKeyID)
		assert.Equal(t, c.SecretAccessKey, p.SecretAccessKey)
		assert.Equal(t, c.Region, p.Region)
		assert.Equal(t, c.Endpoint, p.Endpoint)
	})

	t.Run("base mapper keeps none empty values", func(t *testing.T) {
		p := v1alpha1.AWSConfig{
			AccessKeyID:     "access",
			SecretAccessKey: "secret",
			Region:          "eu",
			Endpoint:        "http://localhost:8080",
		}

		c := v1alpha1.AWSConfig{
			AccessKeyID:     "access_child",
			SecretAccessKey: "secret_child",
			Region:          "de",
			Endpoint:        "http://localhost:9090",
		}
		c.MapAWSParent(p)

		assert.Equal(t, c.AccessKeyID, "access_child")
		assert.Equal(t, c.SecretAccessKey, "secret_child")
		assert.Equal(t, c.Region, "de")
		assert.Equal(t, c.Endpoint, "http://localhost:9090")
	})
}
