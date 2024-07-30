package target_test

import (
	"testing"

	"github.com/kyverno/policy-reporter/pkg/target"
	"github.com/stretchr/testify/assert"
)

func TestConfig(t *testing.T) {
	t.Run("return expected secret ref", func(t *testing.T) {
		c := &target.Config[target.WebhookOptions]{
			SecretRef: "webhook-secret",
		}

		assert.Equal(t, c.Secret(), "webhook-secret")
	})

	t.Run("ignores secret mount", func(t *testing.T) {
		c := &target.Config[target.WebhookOptions]{
			MountedSecret: "webhook-secret",
		}

		assert.Equal(t, c.Secret(), "")
	})

	t.Run("base mapper set expected fallbacks from parent config", func(t *testing.T) {
		p := &target.Config[target.WebhookOptions]{
			MinimumPriority: "warning",
			SkipExisting:    true,
		}

		c := &target.Config[target.WebhookOptions]{}
		c.MapBaseParent(p)

		assert.Equal(t, c.MinimumPriority, p.MinimumPriority)
		assert.Equal(t, c.SkipExisting, p.SkipExisting)
	})

	t.Run("base mapper keeps none empty values", func(t *testing.T) {
		p := &target.Config[target.WebhookOptions]{
			MinimumPriority: "warning",
		}

		c := &target.Config[target.WebhookOptions]{
			MinimumPriority: "debug",
		}

		c.MapBaseParent(p)

		assert.Equal(t, c.MinimumPriority, "debug")
	})
}

func TestAWSConfig(t *testing.T) {
	t.Run("aws mapper set expected fallbacks from parent config", func(t *testing.T) {
		p := target.AWSConfig{
			AccessKeyID:     "access",
			SecretAccessKey: "secret",
			Region:          "eu",
			Endpoint:        "http://localhost:8080",
		}

		c := target.AWSConfig{}
		c.MapAWSParent(p)

		assert.Equal(t, c.AccessKeyID, p.AccessKeyID)
		assert.Equal(t, c.SecretAccessKey, p.SecretAccessKey)
		assert.Equal(t, c.Region, p.Region)
		assert.Equal(t, c.Endpoint, p.Endpoint)
	})

	t.Run("base mapper keeps none empty values", func(t *testing.T) {
		p := target.AWSConfig{
			AccessKeyID:     "access",
			SecretAccessKey: "secret",
			Region:          "eu",
			Endpoint:        "http://localhost:8080",
		}

		c := target.AWSConfig{
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
