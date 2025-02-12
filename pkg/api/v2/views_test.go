package v2_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	v2 "github.com/kyverno/policy-reporter/pkg/api/v2"
	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
	"github.com/kyverno/policy-reporter/pkg/database"
	"github.com/kyverno/policy-reporter/pkg/filters"
)

func TestV2Views(t *testing.T) {
	t.Run("MapValueFilter", func(t *testing.T) {
		empty := v2.MapValueFilter(filters.ValueFilter{})

		assert.Nil(t, empty)

		original := filters.ValueFilter{
			Include:  []string{"default"},
			Exclude:  []string{"kube-system"},
			Selector: map[string]string{"team": "marketing"},
		}

		filter := v2.MapValueFilter(original)

		assert.Equal(t, original.Include, filter.Include)
		assert.Equal(t, original.Exclude, filter.Exclude)
		assert.Equal(t, map[string]interface{}{"team": "marketing"}, filter.Selector)
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
				Status: &v2.StatusList{
					Pass: 8,
					Fail: 3,
				},
			},
			{
				Name: "PSS Restricted",
				Status: &v2.StatusList{
					Pass: 4,
					Fail: 1,
				},
			},
		}})
		assert.Contains(t, result, &v2.SourceDetails{Name: "Trivy", Categories: []*v2.Category{
			{
				Name: "Vulnr",
				Status: &v2.StatusList{
					Pass: 0,
					Fail: 2,
					Warn: 4,
				},
			},
		}})
	})

	t.Run("MapBaseToTarget", func(t *testing.T) {
		target := v2.MapBaseToTarget(&v1alpha1.Config[v1alpha1.WebhookOptions]{
			Name:            "Webhook",
			MinimumSeverity: "medium",
			SecretRef:       "ref",
			MountedSecret:   "mounted",
			Sources:         []string{"Kyverno"},
			SkipExisting:    true,
			Valid:           true,
		})

		assert.Equal(t, "Webhook", target.Name)
		assert.Equal(t, "medium", target.MinimumSeverity)
		assert.Equal(t, "ref", target.SecretRef)
		assert.Equal(t, "mounted", target.MountedSecret)
		assert.NotNil(t, target.CustomFields)
		assert.NotNil(t, target.Properties)
		assert.Equal(t, []string{"Kyverno"}, target.Filter.Sources.Include)
	})

	t.Run("MapSlackToTarget", func(t *testing.T) {
		target := v2.MapSlackToTarget(&v1alpha1.Config[v1alpha1.SlackOptions]{
			Name:            "Slack",
			MinimumSeverity: "medium",
			Config: &v1alpha1.SlackOptions{
				Channel: "general",
				WebhookOptions: v1alpha1.WebhookOptions{
					Webhook: "http://slack.com/xxxx",
				},
			},
			Valid: true,
		})

		assert.Equal(t, "Slack", target.Name)
		assert.Equal(t, "medium", target.MinimumSeverity)
		assert.Equal(t, "Slack", target.Type)
		assert.Equal(t, "general", target.Properties["channel"])
	})

	t.Run("MapLokiToTarget", func(t *testing.T) {
		target := v2.MapLokiToTarget(&v1alpha1.Config[v1alpha1.LokiOptions]{
			Name:            "Loki 1",
			MinimumSeverity: "medium",
			Config: &v1alpha1.LokiOptions{
				HostOptions: v1alpha1.HostOptions{
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
		assert.Equal(t, "medium", target.MinimumSeverity)

		assert.Equal(t, "Loki", target.Type)
		assert.Equal(t, "v1/push", target.Properties["api"])
		assert.Equal(t, "http://loki.monitoring:3000", target.Host)
		assert.True(t, target.SkipTLS)
		assert.True(t, target.UseTLS)
		assert.True(t, target.Auth)
	})

	t.Run("MapElasticsearchToTarget", func(t *testing.T) {
		target := v2.MapElasticsearchToTarget(&v1alpha1.Config[v1alpha1.ElasticsearchOptions]{
			Name:            "Target",
			MinimumSeverity: "medium",
			Config: &v1alpha1.ElasticsearchOptions{
				HostOptions: v1alpha1.HostOptions{
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
		assert.Equal(t, "medium", target.MinimumSeverity)

		assert.Equal(t, "Elasticsearch", target.Type)
		assert.Equal(t, "policy-reporter", target.Properties["index"])
		assert.Equal(t, "daily", target.Properties["rotation"])
		assert.Equal(t, "http://elasticsearch.monitoring:3000", target.Host)
		assert.True(t, target.SkipTLS)
		assert.True(t, target.UseTLS)
		assert.True(t, target.Auth)
	})

	t.Run("MapWebhhokToTarget", func(t *testing.T) {
		target := v2.MapWebhhokToTarget("Discord")(&v1alpha1.Config[v1alpha1.WebhookOptions]{
			Name:            "Target",
			MinimumSeverity: "medium",
			Config: &v1alpha1.WebhookOptions{
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
		assert.Equal(t, "medium", target.MinimumSeverity)

		assert.Equal(t, "Discord", target.Type)
		assert.Equal(t, "http://discord.com", target.Host)
		assert.True(t, target.SkipTLS)
		assert.True(t, target.UseTLS)
		assert.True(t, target.Auth)
	})

	t.Run("MapTelegramToTarget", func(t *testing.T) {
		target := v2.MapTelegramToTarget(&v1alpha1.Config[v1alpha1.TelegramOptions]{
			Name:            "Target",
			MinimumSeverity: "medium",
			Config: &v1alpha1.TelegramOptions{
				Token:  "ABCDE",
				ChatID: "1234567",
				WebhookOptions: v1alpha1.WebhookOptions{
					Webhook:     "http://telegram.com",
					Certificate: "cert",
					SkipTLS:     true,
				},
			},
			Valid: true,
		})

		assert.Equal(t, "Target", target.Name)
		assert.Equal(t, "medium", target.MinimumSeverity)

		assert.Equal(t, "Telegram", target.Type)
		assert.Equal(t, "http://telegram.com", target.Host)
		assert.Equal(t, "1234567", target.Properties["chatId"])
		assert.True(t, target.SkipTLS)
		assert.True(t, target.UseTLS)
		assert.False(t, target.Auth)
	})

	t.Run("MapS3ToTarget", func(t *testing.T) {
		target := v2.MapS3ToTarget(&v1alpha1.Config[v1alpha1.S3Options]{
			Name:            "Target",
			MinimumSeverity: "medium",
			Config: &v1alpha1.S3Options{
				Prefix: "policy-reporter",
				Bucket: "kyverno",
				AWSConfig: v1alpha1.AWSConfig{
					Region:   "eu-central-1",
					Endpoint: "https://s3.aws.com",
				},
			},
			Valid: true,
		})

		assert.Equal(t, "Target", target.Name)
		assert.Equal(t, "medium", target.MinimumSeverity)

		assert.Equal(t, "S3", target.Type)
		assert.Equal(t, "https://s3.aws.com", target.Host)
		assert.Equal(t, "kyverno", target.Properties["bucket"])
		assert.Equal(t, "policy-reporter", target.Properties["prefix"])
		assert.Equal(t, "eu-central-1", target.Properties["region"])
		assert.True(t, target.Auth)
	})

	t.Run("MapKinesisToTarget", func(t *testing.T) {
		target := v2.MapKinesisToTarget(&v1alpha1.Config[v1alpha1.KinesisOptions]{
			Name:            "Target",
			MinimumSeverity: "medium",
			Config: &v1alpha1.KinesisOptions{
				StreamName: "policy-reporter",
				AWSConfig: v1alpha1.AWSConfig{
					Region:   "eu-central-1",
					Endpoint: "https://kinesis.aws.com",
				},
			},
			Valid: true,
		})

		assert.Equal(t, "Target", target.Name)
		assert.Equal(t, "medium", target.MinimumSeverity)

		assert.Equal(t, "Kinesis", target.Type)
		assert.Equal(t, "https://kinesis.aws.com", target.Host)
		assert.Equal(t, "policy-reporter", target.Properties["stream"])
		assert.Equal(t, "eu-central-1", target.Properties["region"])
		assert.True(t, target.Auth)
	})

	t.Run("MapSecurityHubToTarget", func(t *testing.T) {
		target := v2.MapSecurityHubToTarget(&v1alpha1.Config[v1alpha1.SecurityHubOptions]{
			Name:            "Target",
			MinimumSeverity: "medium",
			Config: &v1alpha1.SecurityHubOptions{
				AccountID:   "policy-reporter",
				Synchronize: true,
				AWSConfig: v1alpha1.AWSConfig{
					Region:   "eu-central-1",
					Endpoint: "https://securityhub.aws.com",
				},
			},
			Valid: true,
		})

		assert.Equal(t, "Target", target.Name)
		assert.Equal(t, "medium", target.MinimumSeverity)

		assert.Equal(t, "SecurityHub", target.Type)
		assert.Equal(t, "https://securityhub.aws.com", target.Host)
		assert.Equal(t, "eu-central-1", target.Properties["region"])
		assert.Equal(t, true, target.Properties["synchronize"])
		assert.True(t, target.Auth)
	})

	t.Run("MapGCSToTarget", func(t *testing.T) {
		target := v2.MapGCSToTarget(&v1alpha1.Config[v1alpha1.GCSOptions]{
			Name:            "Target",
			MinimumSeverity: "medium",
			Config: &v1alpha1.GCSOptions{
				Prefix: "policy-reporter",
				Bucket: "kyverno",
			},
			Valid: true,
		})

		assert.Equal(t, "Target", target.Name)
		assert.Equal(t, "medium", target.MinimumSeverity)

		assert.Equal(t, "GoogleCloudStore", target.Type)
		assert.Equal(t, "kyverno", target.Properties["bucket"])
		assert.Equal(t, "policy-reporter", target.Properties["prefix"])
		assert.True(t, target.Auth)
	})

	t.Run("MapTargets", func(t *testing.T) {
		targets := v2.MapTargets(&v1alpha1.Config[v1alpha1.GCSOptions]{
			Name:            "Target",
			MinimumSeverity: "medium",
			Config: &v1alpha1.GCSOptions{
				Prefix: "policy-reporter",
				Bucket: "kyverno",
			},
			Valid: true,
			Channels: []*v1alpha1.Config[v1alpha1.GCSOptions]{
				{
					Name:            "Target 2",
					MinimumSeverity: "medium",
					Config: &v1alpha1.GCSOptions{
						Prefix: "policy-reporter",
						Bucket: "trivy",
					},
					Valid: true,
				},
				{
					Name:            "Target 2",
					MinimumSeverity: "medium",
					Config: &v1alpha1.GCSOptions{
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
