package v2_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v2 "github.com/kyverno/policy-reporter/pkg/api/v2"
	"github.com/kyverno/policy-reporter/pkg/config"
	"github.com/kyverno/policy-reporter/pkg/database"
)

func TestV2Views(t *testing.T) {
	t.Run("MapValueFilter", func(t *testing.T) {
		empty := v2.MapValueFilter(config.ValueFilter{})

		assert.Nil(t, empty)

		original := config.ValueFilter{
			Include:  []string{"default"},
			Exclude:  []string{"kube-system"},
			Selector: map[string]any{"team": "marketing"},
		}

		filter := v2.MapValueFilter(original)

		assert.Equal(t, original.Include, filter.Include)
		assert.Equal(t, original.Exclude, filter.Exclude)
		assert.Equal(t, original.Selector, filter.Selector)
	})

	t.Run("MapResourceCategoryToSourceDetails", func(t *testing.T) {
		result := v2.MapResourceCategoryToSourceDetails([]database.ResourceCategory{
			{
				Source: "Kyverno",
				Name:   "PSS Baseline",
				Pass:   8,
				Fail:   3,
			},
			{
				Source: "Kyverno",
				Name:   "PSS Restricted",
				Pass:   4,
				Fail:   1,
			},
			{
				Source: "Trivy",
				Name:   "Vulnr",
				Pass:   0,
				Fail:   2,
				Warn:   4,
			},
		})

		assert.Equal(t, 2, len(result))
		assert.Contains(t, result, &v2.SourceDetails{Name: "Kyverno", Categories: []*v2.Category{
			{
				Name: "PSS Baseline",
				Pass: 8,
				Fail: 3,
			},
			{
				Name: "PSS Restricted",
				Pass: 4,
				Fail: 1,
			},
		}})
		assert.Contains(t, result, &v2.SourceDetails{Name: "Trivy", Categories: []*v2.Category{
			{
				Name: "Vulnr",
				Pass: 0,
				Fail: 2,
				Warn: 4,
			},
		}})
	})

	t.Run("MapBaseToTarget", func(t *testing.T) {
		target := v2.MapBaseToTarget(&config.Target[config.WebhookOptions]{
			Name:            "Webhook",
			MinimumPriority: "warning",
			SecretRef:       "ref",
			MountedSecret:   "mounted",
			Sources:         []string{"Kyverno"},
			SkipExisting:    true,
			Valid:           true,
		})

		assert.Equal(t, "Webhook", target.Name)
		assert.Equal(t, "warning", target.MinimumPriority)
		assert.Equal(t, "ref", target.SecretRef)
		assert.Equal(t, "mounted", target.MountedSecret)
		assert.NotNil(t, target.CustomFields)
		assert.NotNil(t, target.Properties)
		assert.Equal(t, []string{"Kyverno"}, target.Filter.Sources.Include)
	})

	t.Run("MapSlackToTarget", func(t *testing.T) {
		target := v2.MapSlackToTarget(&config.Target[config.SlackOptions]{
			Name:            "Slack",
			MinimumPriority: "warning",
			Config: &config.SlackOptions{
				Channel: "general",
				WebhookOptions: config.WebhookOptions{
					Webhook: "http://slack.com/xxxx",
				},
			},
			Valid: true,
		})

		assert.Equal(t, "Slack", target.Name)
		assert.Equal(t, "warning", target.MinimumPriority)
		assert.Equal(t, "Slack", target.Type)
		assert.Equal(t, "general", target.Properties["channel"])
	})

	t.Run("MapLokiToTarget", func(t *testing.T) {
		target := v2.MapLokiToTarget(&config.Target[config.LokiOptions]{
			Name:            "Loki 1",
			MinimumPriority: "warning",
			Config: &config.LokiOptions{
				HostOptions: config.HostOptions{
					Host:        "http://loki.monitoring:3000",
					Certificate: "cert",
					SkipTLS:     true,
				},
				Username: "user",
				Password: "password",
				Path:     "v1/push",
			},
			Valid: true,
		})

		assert.Equal(t, "Loki 1", target.Name)
		assert.Equal(t, "warning", target.MinimumPriority)

		assert.Equal(t, "Loki", target.Type)
		assert.Equal(t, "v1/push", target.Properties["api"])
		assert.Equal(t, "http://loki.monitoring:3000", target.Host)
		assert.True(t, target.SkipTLS)
		assert.True(t, target.UseTLS)
		assert.True(t, target.Auth)
	})

	t.Run("MapElasticsearchToTarget", func(t *testing.T) {
		target := v2.MapElasticsearchToTarget(&config.Target[config.ElasticsearchOptions]{
			Name:            "Target",
			MinimumPriority: "warning",
			Config: &config.ElasticsearchOptions{
				HostOptions: config.HostOptions{
					Host:        "http://elasticsearch.monitoring:3000",
					Certificate: "cert",
					SkipTLS:     true,
					Headers: map[string]string{
						"Authorization": "Bearer 123456",
					},
				},
				Index:    "policy-reporter",
				Rotation: "daily",
			},
			Valid: true,
		})

		assert.Equal(t, "Target", target.Name)
		assert.Equal(t, "warning", target.MinimumPriority)

		assert.Equal(t, "Elasticsearch", target.Type)
		assert.Equal(t, "policy-reporter", target.Properties["index"])
		assert.Equal(t, "daily", target.Properties["rotation"])
		assert.Equal(t, "http://elasticsearch.monitoring:3000", target.Host)
		assert.True(t, target.SkipTLS)
		assert.True(t, target.UseTLS)
		assert.True(t, target.Auth)
	})

	t.Run("MapWebhhokToTarget", func(t *testing.T) {
		target := v2.MapWebhhokToTarget("Discord")(&config.Target[config.WebhookOptions]{
			Name:            "Target",
			MinimumPriority: "warning",
			Config: &config.WebhookOptions{
				Webhook:     "http://discord.com/12345/888XABC",
				Certificate: "cert",
				SkipTLS:     true,
				Headers: map[string]string{
					"Authorization": "Bearer 123456",
				},
			},
			Valid: true,
		})

		assert.Equal(t, "Target", target.Name)
		assert.Equal(t, "warning", target.MinimumPriority)

		assert.Equal(t, "Discord", target.Type)
		assert.Equal(t, "http://discord.com", target.Host)
		assert.True(t, target.SkipTLS)
		assert.True(t, target.UseTLS)
		assert.True(t, target.Auth)
	})

	t.Run("MapTelegramToTarget", func(t *testing.T) {
		target := v2.MapTelegramToTarget(&config.Target[config.TelegramOptions]{
			Name:            "Target",
			MinimumPriority: "warning",
			Config: &config.TelegramOptions{
				Token:  "ABCDE",
				ChatID: "1234567",
				WebhookOptions: config.WebhookOptions{
					Webhook:     "http://telegram.com",
					Certificate: "cert",
					SkipTLS:     true,
				},
			},
			Valid: true,
		})

		assert.Equal(t, "Target", target.Name)
		assert.Equal(t, "warning", target.MinimumPriority)

		assert.Equal(t, "Telegram", target.Type)
		assert.Equal(t, "http://telegram.com", target.Host)
		assert.Equal(t, "1234567", target.Properties["chatID"])
		assert.True(t, target.SkipTLS)
		assert.True(t, target.UseTLS)
		assert.False(t, target.Auth)
	})

	t.Run("MapS3ToTarget", func(t *testing.T) {
		target := v2.MapS3ToTarget(&config.Target[config.S3Options]{
			Name:            "Target",
			MinimumPriority: "warning",
			Config: &config.S3Options{
				Prefix: "policy-reporter",
				Bucket: "kyverno",
				AWSConfig: config.AWSConfig{
					Region:   "eu-central-1",
					Endpoint: "https://s3.aws.com",
				},
			},
			Valid: true,
		})

		assert.Equal(t, "Target", target.Name)
		assert.Equal(t, "warning", target.MinimumPriority)

		assert.Equal(t, "S3", target.Type)
		assert.Equal(t, "https://s3.aws.com", target.Host)
		assert.Equal(t, "kyverno", target.Properties["bucket"])
		assert.Equal(t, "policy-reporter", target.Properties["prefix"])
		assert.Equal(t, "eu-central-1", target.Properties["region"])
		assert.True(t, target.Auth)
	})

	t.Run("MapKinesisToTarget", func(t *testing.T) {
		target := v2.MapKinesisToTarget(&config.Target[config.KinesisOptions]{
			Name:            "Target",
			MinimumPriority: "warning",
			Config: &config.KinesisOptions{
				StreamName: "policy-reporter",
				AWSConfig: config.AWSConfig{
					Region:   "eu-central-1",
					Endpoint: "https://kinesis.aws.com",
				},
			},
			Valid: true,
		})

		assert.Equal(t, "Target", target.Name)
		assert.Equal(t, "warning", target.MinimumPriority)

		assert.Equal(t, "Kinesis", target.Type)
		assert.Equal(t, "https://kinesis.aws.com", target.Host)
		assert.Equal(t, "policy-reporter", target.Properties["stream"])
		assert.Equal(t, "eu-central-1", target.Properties["region"])
		assert.True(t, target.Auth)
	})

	t.Run("MapSecurityHubToTarget", func(t *testing.T) {
		target := v2.MapSecurityHubToTarget(&config.Target[config.SecurityHubOptions]{
			Name:            "Target",
			MinimumPriority: "warning",
			Config: &config.SecurityHubOptions{
				AccountID: "policy-reporter",
				Cleanup:   true,
				AWSConfig: config.AWSConfig{
					Region:   "eu-central-1",
					Endpoint: "https://securityhub.aws.com",
				},
			},
			Valid: true,
		})

		assert.Equal(t, "Target", target.Name)
		assert.Equal(t, "warning", target.MinimumPriority)

		assert.Equal(t, "SecurityHub", target.Type)
		assert.Equal(t, "https://securityhub.aws.com", target.Host)
		assert.Equal(t, "eu-central-1", target.Properties["region"])
		assert.Equal(t, true, target.Properties["cleanup"])
		assert.True(t, target.Auth)
	})

	t.Run("MapGCSToTarget", func(t *testing.T) {
		target := v2.MapGCSToTarget(&config.Target[config.GCSOptions]{
			Name:            "Target",
			MinimumPriority: "warning",
			Config: &config.GCSOptions{
				Prefix: "policy-reporter",
				Bucket: "kyverno",
			},
			Valid: true,
		})

		assert.Equal(t, "Target", target.Name)
		assert.Equal(t, "warning", target.MinimumPriority)

		assert.Equal(t, "GoogleCloudStore", target.Type)
		assert.Equal(t, "kyverno", target.Properties["bucket"])
		assert.Equal(t, "policy-reporter", target.Properties["prefix"])
		assert.True(t, target.Auth)
	})

	t.Run("MapTargets", func(t *testing.T) {
		targets := v2.MapTargets(&config.Target[config.GCSOptions]{
			Name:            "Target",
			MinimumPriority: "warning",
			Config: &config.GCSOptions{
				Prefix: "policy-reporter",
				Bucket: "kyverno",
			},
			Valid: true,
			Channels: []*config.Target[config.GCSOptions]{
				{
					Name:            "Target 2",
					MinimumPriority: "warning",
					Config: &config.GCSOptions{
						Prefix: "policy-reporter",
						Bucket: "trivy",
					},
					Valid: true,
				},
				{
					Name:            "Target 2",
					MinimumPriority: "warning",
					Config: &config.GCSOptions{
						Prefix: "policy-reporter",
						Bucket: "trivy",
					},
					Valid: false,
				},
			},
		}, v2.MapGCSToTarget)

		assert.Equal(t, 2, len(targets))
	})
}
