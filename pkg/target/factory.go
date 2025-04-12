package target

import (
	"github.com/kyverno/policy-reporter/pkg/crd/api/targetconfig/v1alpha1"
)

type Factory interface {
	CreateClients(config *Targets) *Collection
	CreateLokiTarget(config, parent *v1alpha1.Config[v1alpha1.LokiOptions]) *Target
	CreateSingleClient(*v1alpha1.TargetConfig) (*Target, error)
	CreateElasticsearchTarget(config, parent *v1alpha1.Config[v1alpha1.ElasticsearchOptions]) *Target
	CreateSlackTarget(config, parent *v1alpha1.Config[v1alpha1.SlackOptions]) *Target
	CreateDiscordTarget(config, parent *v1alpha1.Config[v1alpha1.WebhookOptions]) *Target
	CreateTeamsTarget(config, parent *v1alpha1.Config[v1alpha1.WebhookOptions]) *Target
	CreateWebhookTarget(config, parent *v1alpha1.Config[v1alpha1.WebhookOptions]) *Target
	CreateTelegramTarget(config, parent *v1alpha1.Config[v1alpha1.TelegramOptions]) *Target
	CreateGoogleChatTarget(config, parent *v1alpha1.Config[v1alpha1.WebhookOptions]) *Target
	CreateJiraTarget(config, parent *v1alpha1.Config[v1alpha1.JiraOptions]) *Target
	CreateS3Target(config, parent *v1alpha1.Config[v1alpha1.S3Options]) *Target
	CreateKinesisTarget(config, parent *v1alpha1.Config[v1alpha1.KinesisOptions]) *Target
	CreateSecurityHubTarget(config, parent *v1alpha1.Config[v1alpha1.SecurityHubOptions]) *Target
	CreateGCSTarget(config, parent *v1alpha1.Config[v1alpha1.GCSOptions]) *Target
}
